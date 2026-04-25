package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/Edu963/ocultar/apps/sombra/pkg/connector"
	"github.com/Edu963/ocultar/apps/sombra/pkg/router"
	"github.com/Edu963/ocultar/apps/sombra/pkg/scrubber"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/proxy"
	"github.com/Edu963/ocultar/pkg/audit"
	"github.com/Edu963/ocultar/vault"
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
	actor := extractActor(r)
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
		body, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read uploaded file", http.StatusInternalServerError)
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
			redactedBytes, _ := json.MarshalIndent(processed, "", "  ")
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
		http.Error(w, fmt.Sprintf("ai model request failed: %v", err), http.StatusBadGateway)
		return
	}

	// 5. Re-hydrate the tokens in the AI's response using the vault.
	rehydratedResponse, err := proxy.RehydrateString(g.vault, g.masterKey, aiResponse)
	if err != nil {
		if g.auditor != nil {
			g.auditor.Log(fetchReq.Actor, "AI_ROUTING", modelName, "FAILED", "Re-hydration error")
		}
		http.Error(w, fmt.Sprintf("re-hydration failed: %v", err), http.StatusInternalServerError)
		return
	}

	if g.auditor != nil {
		g.auditor.Log(fetchReq.Actor, "AI_ROUTING", modelName, "SUCCESS", fmt.Sprintf("Connector=%s", connName))
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

// extractActor pulls the user's identity from the Authorization header.
func extractActor(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		// [CISO-AUDIT] Zero-Trust: Refuse to process requests without explicit actor identity.
		return ""
	}

	// Simple extraction: return the part after the type (Bearer/Basic)
	parts := regexp.MustCompile(`\s+`).Split(auth, 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return auth
}
