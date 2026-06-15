package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/connector"
	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/router"
	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/scrubber"
	"github.com/ocultar-dev/ocultar/pkg/audit"
	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/pkg/proxy"
	"github.com/ocultar-dev/ocultar/pkg/refinery"
	"github.com/ocultar-dev/ocultar/vault"
	"github.com/golang-jwt/jwt/v5"
)

// Gateway ties together the Data Connectors, the OCULTAR Refinery,
// and the Multi-Model Router.
type Gateway struct {
	connectors map[string]connector.Connector
	router     *router.Router
	eng        *refinery.Refinery
	vault      vault.Provider
	masterKey  []byte
	auditor    *audit.ImmutableLogger
}

// NewGateway creates a new Sombra gateway.
func NewGateway(eng *refinery.Refinery, v vault.Provider, masterKey []byte, r *router.Router, auditor *audit.ImmutableLogger) *Gateway {
	return &Gateway{
		connectors: make(map[string]connector.Connector),
		router:     r,
		eng:        eng,
		vault:      v,
		masterKey:  masterKey,
		auditor:    auditor,
	}
}

// RegisterConnector adds a new data source adapter.
func (g *Gateway) RegisterConnector(c connector.Connector) {
	g.connectors[c.Name()] = c
}

// HandleQuery is the main HTTP endpoint for agentic interactions.
// Expected multi-part form or JSON:
//   - connector: name of the registered connector (e.g. "file", "banking")
//   - model: optional name of the AI model to route to
//   - prompt: the user's question or instruction
//   - source_id: connector-specific ID (or file upload)
func (g *Gateway) HandleQuery(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request.
	if err := r.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	connName := r.FormValue("connector")
	modelName := r.FormValue("model")
	prompt := r.FormValue("prompt")
	sourceID := r.FormValue("source_id")

	if connName == "" {
		http.Error(w, "missing 'connector' parameter", http.StatusBadRequest)
		return
	}

	conn, ok := g.connectors[connName]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown connector: %q", connName), http.StatusBadRequest)
		return
	}

	policy := conn.Policy()
	if !policy.IsModelAllowed(modelName) {
		http.Error(w, fmt.Sprintf("connector policy forbids sending data to model %q", modelName), http.StatusForbidden)
		return
	}

	// 2. Fetch data from the connector.
	actor := g.extractActor(r)
	if actor == "" {
		http.Error(w, "unauthorized: invalid or missing token", http.StatusUnauthorized)
		return
	}
	fetchReq := connector.FetchRequest{
		SourceID:   sourceID,
		Parameters: make(map[string]string),
		Actor:      actor,
	}

	// Forward all other form values as parameters
	for k, v := range r.Form {
		if k != "connector" && k != "model" && k != "prompt" && k != "source_id" && len(v) > 0 {
			fetchReq.Parameters[k] = v[0]
		}
	}

	// Handle file uploads. We intentionally do NOT pass the HTTP Content-Type header
	// because multipart uploads always report "application/octet-stream" regardless
	// of the actual file type. The connector sniffs the real type from the bytes.
	file, _, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		// Cap the in-memory read to the connector's policy limit before allocating.
		// Without this, a large upload OOMs the process before the connector's size
		// check fires.
		limit := int64(10 << 20) // 10 MB safety cap when policy is unset
		if policy.MaxBodyBytes > 0 {
			limit = policy.MaxBodyBytes + 1
		}
		body, err := io.ReadAll(io.LimitReader(file, limit))
		if err != nil {
			http.Error(w, "failed to read uploaded file", http.StatusInternalServerError)
			return
		}
		if policy.MaxBodyBytes > 0 && int64(len(body)) > policy.MaxBodyBytes {
			http.Error(w, "uploaded file exceeds size limit", http.StatusRequestEntityTooLarge)
			return
		}
		fetchReq.RawBody = body
		// ContentType intentionally left empty — connector auto-detects from content.
	}

	ctx := r.Context()
	fetchResp, err := conn.Fetch(ctx, fetchReq)
	if err != nil {
		// If the error is specifically "no file body or source_id provided", treat
		// the prompt itself as the data context (direct/prompt-only mode).
		if strings.Contains(err.Error(), "no file body or source_id provided") {
			fetchResp = &connector.FetchResponse{
				ContentType: "text/plain",
				Body:        []byte(""), // Empty context; the prompt carries all the data.
			}
		} else {
			http.Error(w, fmt.Sprintf("connector fetch failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// 3. Pre-scrub: tokenise emails and account numbers before the OCULTAR
	// refinery runs. This prevents Tier 1.1 (libphonenumber) from tagging
	// account numbers as [PHONE_...] and Tier 1.5 (greeting scanner) from
	// splitting emails at the @ sign.
	sc := scrubber.New(g.vault, g.masterKey)
	prescrubbedData, err := sc.Prescrub(string(fetchResp.Body))
	if err != nil {
		http.Error(w, fmt.Sprintf("pre-scrub failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Redact remaining PII using the OCULTAR refinery.
	// Redact both the data context AND the user's prompt.
	var redactedData string
	if fetchResp.ContentType == "application/json" {
		// Structured Redaction for JSON data
		var jsonData interface{}
		if err := json.Unmarshal([]byte(prescrubbedData), &jsonData); err == nil {
			processed, err := g.eng.ProcessInterface(jsonData, fetchReq.Actor)
			if err != nil {
				http.Error(w, fmt.Sprintf("structured redaction failed: %v", err), http.StatusInternalServerError)
				return
			}
			redactedBytes, err := json.MarshalIndent(processed, "", "  ")
			if err != nil {
				http.Error(w, "structured redaction marshal failed", http.StatusInternalServerError)
				return
			}
			redactedData = string(redactedBytes)
		} else {
			// Fallback to string-based redaction if JSON is malformed
			redactedData, err = g.eng.RefineString(prescrubbedData, fetchReq.Actor, nil)
		}
	} else {
		redactedData, err = g.eng.RefineString(prescrubbedData, fetchReq.Actor, nil)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("redaction refinery failed (data): %v", err), http.StatusInternalServerError)
		return
	}

	redactedPrompt, err := g.eng.RefineString(prompt, fetchReq.Actor, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("redaction refinery failed (prompt): %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Policy Enforcement: Strip categories marked for removal.
	if len(policy.StripCategories) > 0 {
		redactedData = g.stripCategories(redactedData, policy.StripCategories)
		redactedPrompt = g.stripCategories(redactedPrompt, policy.StripCategories)
	}

	// 4. Send to Multi-Model Router.
	// We combine the redacted user's prompt with the redacted data.
	fullPrompt := fmt.Sprintf("%s\n\nData Context:\n%s", redactedPrompt, redactedData)
	messages := []router.Message{
		{Role: "system", Content: "You are a helpful AI assistant analyzing user data. Answer questions accurately based on the provided Data Context."},
		{Role: "user", Content: fullPrompt},
	}
	opts := router.ModelOpts{}

	aiResponse, err := g.router.Send(ctx, modelName, messages, opts)
	if err != nil {
		log.Printf("[ERROR] HandleQuery router.Send failed: %v", err)
		http.Error(w, "ai model request failed", http.StatusBadGateway)
		return
	}

	// 5. Re-hydrate the tokens in the AI's response using the vault.
	rehydratedResponse, err := proxy.RehydrateString(g.vault, g.masterKey, aiResponse)
	if err != nil {
		if g.auditor != nil {
			g.auditor.Log(actor, "AI_ROUTING", modelName, "FAILED", "Re-hydration error")
		}
		http.Error(w, fmt.Sprintf("re-hydration failed: %v", err), http.StatusInternalServerError)
		return
	}

	if g.auditor != nil {
		g.auditor.Log(actor, "AI_ROUTING", modelName, "SUCCESS", fmt.Sprintf("Connector=%s", connName))
	}

	// 6. Return the safe response to the user.
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"response": rehydratedResponse,
		"metadata": map[string]interface{}{
			"model":            modelName,
			"connector":        connName,
			"pii_was_redacted": proxy.ContainsTokens(redactedData) || proxy.ContainsTokens(redactedPrompt),
		},
	}

	// [CISO-AUDIT] Only include the trace in debug/demo mode to prevent metadata leakage.
	// Requires the OCULTAR_DEBUG=true environment variable.
	if config.Global.ShowDebugMetadata {
		resp["metadata"].(map[string]interface{})["ai_saw"] = fullPrompt
		resp["metadata"].(map[string]interface{})["prompt_redacted"] = redactedPrompt
		resp["metadata"].(map[string]interface{})["original_prompt"] = prompt
	}

	json.NewEncoder(w).Encode(resp)
}

// stripCategories removes tokens associated with the given PII categories from the text.
// Tokens follow the pattern: [CATEGORY_HASH] (e.g. [SSN_1234abcd])
func (g *Gateway) stripCategories(text string, categories []string) string {
	for _, cat := range categories {
		// Create a regex to match tokens for this specific category.
		// Pattern: [CAT_...]
		pattern := fmt.Sprintf(`\[%s_[0-9a-f]+\]`, regexp.QuoteMeta(cat))
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "[STRIPPED_"+cat+"]")
	}
	return text
}

// ─── Entity Registry API ──────────────────────────────────────────────────────

// entityRegisterRequest is the JSON body for POST /v1/entities.
type entityRegisterRequest struct {
	EntityType    string   `json:"entity_type"`
	CanonicalName string   `json:"canonical_name"`
	Variants      []string `json:"variants"`
}

// entityRegisterResponse is returned by POST /v1/entities.
type entityRegisterResponse struct {
	CanonicalToken string `json:"canonical_token"`
}

// entitySeedResponse is returned by POST /v1/entities/seed.
type entitySeedResponse struct {
	Seeded  int      `json:"seeded"`
	Skipped int      `json:"skipped,omitempty"`
	Tokens  []string `json:"tokens"`
}

// HandleEntityRegister handles POST /v1/entities.
// Registers a single canonical entity and returns its token (e.g. "[PERSON_1]").
func (g *Gateway) HandleEntityRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actor := g.extractActor(r)
	if actor == "" {
		http.Error(w, "unauthorized: invalid or missing token", http.StatusUnauthorized)
		return
	}

	var req entityRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.EntityType == "" || req.CanonicalName == "" {
		http.Error(w, "entity_type and canonical_name are required", http.StatusBadRequest)
		return
	}

	token, err := g.vault.RegisterEntity(req.EntityType, req.CanonicalName, req.Variants)
	if err != nil {
		http.Error(w, fmt.Sprintf("registration failed: %v", err), http.StatusInternalServerError)
		return
	}

	if g.auditor != nil {
		g.auditor.Log(actor, "ENTITY_REGISTER", req.EntityType, "SUCCESS", req.CanonicalName)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entityRegisterResponse{CanonicalToken: token})
}

// HandleEntitySeed handles POST /v1/entities/seed.
// Bulk-registers a list of entities from a CRM roster or patient list.
func (g *Gateway) HandleEntitySeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actor := g.extractActor(r)
	if actor == "" {
		http.Error(w, "unauthorized: invalid or missing token", http.StatusUnauthorized)
		return
	}

	// Accept both a flat JSON array and a {"entities": [...]} wrapper.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var seeds []entityRegisterRequest
	if json.Unmarshal(body, &seeds) != nil {
		// Try wrapper format
		var wrapper struct {
			Entities []entityRegisterRequest `json:"entities"`
		}
		if err := json.Unmarshal(body, &wrapper); err != nil || len(wrapper.Entities) == 0 {
			http.Error(w, "body must be a JSON array of entity objects or {\"entities\":[...]}", http.StatusBadRequest)
			return
		}
		seeds = wrapper.Entities
	}

	var tokens []string
	var skipped int
	for _, s := range seeds {
		if s.EntityType == "" || s.CanonicalName == "" {
			skipped++
			continue
		}
		tok, err := g.vault.RegisterEntity(s.EntityType, s.CanonicalName, s.Variants)
		if err != nil {
			http.Error(w, fmt.Sprintf("seed failed for %q: %v", s.CanonicalName, err), http.StatusInternalServerError)
			return
		}
		tokens = append(tokens, tok)
	}

	if g.auditor != nil {
		g.auditor.Log(actor, "ENTITY_SEED", "BATCH", "SUCCESS", fmt.Sprintf("seeded=%d skipped=%d", len(tokens), skipped))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entitySeedResponse{Seeded: len(tokens), Skipped: skipped, Tokens: tokens})
}

// HandleEntityList handles GET /v1/entities.
// Returns all registered canonical entities with their variant lists.
func (g *Gateway) HandleEntityList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actor := g.extractActor(r)
	if actor == "" {
		http.Error(w, "unauthorized: invalid or missing token", http.StatusUnauthorized)
		return
	}

	records, err := g.vault.ListEntities()
	if err != nil {
		http.Error(w, fmt.Sprintf("list failed: %v", err), http.StatusInternalServerError)
		return
	}
	if records == nil {
		records = []vault.EntityRecord{}
	}

	if g.auditor != nil {
		g.auditor.Log(actor, "ENTITY_LIST", "BATCH", "SUCCESS", fmt.Sprintf("count=%d", len(records)))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// extractActor pulls the user's identity from the Authorization header.
// It validates the HS256 signature against the config.Global.JWTSecret.
func (g *Gateway) extractActor(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	// Simple extraction: return the part after the type (Bearer/Basic)
	parts := regexp.MustCompile(`\s+`).Split(auth, 2)
	if len(parts) < 2 {
		return ""
	}
	tokenString := parts[1]

	// If no secret is configured, return a fixed actor name rather than echoing the
	// caller-supplied token string — prevents actor spoofing in audit logs in dev mode.
	if config.Global.JWTSecret == "" {
		return "dev-anonymous"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Global.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Prefer 'sub', fallback to 'email'
		if sub, ok := claims["sub"].(string); ok {
			return sub
		}
		if email, ok := claims["email"].(string); ok {
			return email
		}
	}

	return ""
}
