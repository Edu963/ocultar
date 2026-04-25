package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Edu963/ocultar/apps/sombra/pkg/router"
	"github.com/Edu963/ocultar/apps/sombra/pkg/scrubber"
	"github.com/Edu963/ocultar/pkg/proxy"
)

// OpenAIChatCompletionRequest represents a standard OpenAI incoming request.
type OpenAIChatCompletionRequest struct {
	Model    string           `json:"model"`
	Messages []router.Message `json:"messages"`
}

// HandleV1ChatCompletions acts as a drop-in multi-model proxy replacement for standard OpenAI SDKs.
// It intercepts standard OpenAI chat complete payloads, iteratively scrubs every message,
// forwards them to the requested model (OpenAI, Claude, Gemini, Local, etc.),
// and then re-hydrates the tokens in the response before streaming back to the client.
func (g *Gateway) HandleV1ChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actor := extractActor(r)
	if actor == "" {
		actor = "sdk-proxy" // basic default for unauthenticated proxy use in dev
	}

	var req OpenAIChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid open_ai formatted json request: %v", err), http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	sc := scrubber.New(g.vault, g.masterKey)

	// 1. Iterative PII Redaction
	// Clean every message's content independently
	for i, msg := range req.Messages {
		if msg.Content == "" {
			continue
		}
		
		// Tier 1.5 - Prescrub to prevent token fragmentation
		prescrubbed, err := sc.Prescrub(msg.Content)
		if err != nil {
			http.Error(w, fmt.Sprintf("pre-scrub failed on message %d: %v", i, err), http.StatusInternalServerError)
			return
		}

		// Full Refinery Core
		redacted, err := g.eng.RefineString(prescrubbed, actor, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("redaction failed on message %d: %v", i, err), http.StatusInternalServerError)
			return
		}

		req.Messages[i].Content = redacted
	}

	// 2. Mult-Model Orchestration
	opts := router.ModelOpts{}
	aiRespString, err := g.router.Send(ctx, req.Model, req.Messages, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("AI Routing failed: %v", err), http.StatusBadGateway)
		return
	}

	// 3. Security Re-Hydration
	rehydratedResponse, err := proxy.RehydrateString(g.vault, g.masterKey, aiRespString)
	if err != nil {
		if g.auditor != nil {
			g.auditor.Log(actor, "PROXY_CHAT_COMPLETION", req.Model, "FAILED", "Re-hydration error")
		}
		http.Error(w, fmt.Sprintf("re-hydration failed: %v", err), http.StatusInternalServerError)
		return
	}

	if g.auditor != nil {
		g.auditor.Log(actor, "PROXY_CHAT_COMPLETION", req.Model, "SUCCESS", fmt.Sprintf("Tokens Redacted/Rehydrated. Messages: %d", len(req.Messages)))
	}

	// 4. Return Compatible Payload to Client SDK
	resp := map[string]interface{}{
		"id":      "chatcmpl-ocultar-sombra",
		"object":  "chat.completion",
		"created": 1677652288,
		"model":   req.Model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": rehydratedResponse,
				},
				"finish_reason": "stop",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
