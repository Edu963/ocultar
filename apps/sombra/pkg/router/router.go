// Package router implements the multi-model AI routing layer for Sombra.
// It dispatches sanitised prompts to the correct AI backend and returns
// the raw response for re-hydration.
package router

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
)

// ModelAdapter is the contract every AI backend must satisfy.
type ModelAdapter interface {
	// Name returns the identifier used in config (e.g. "gpt-4o", "gemini-pro").
	Name() string
	// Send dispatches a prompt to the model and returns the text response.
	Send(ctx context.Context, messages []Message, opts ModelOpts) (string, error)
	// Endpoint returns the base URL for the model.
	Endpoint() string
	// HealthCheck verifies the backend is reachable.
	HealthCheck(ctx context.Context) error
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelOpts contains per-request options.
type ModelOpts struct {
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// Router manages the registered model adapters and selects the right one
// based on the request.
type Router struct {
	mu             sync.RWMutex
	adapters       map[string]ModelAdapter
	fallback       string   // name of the default model
	allowedDomains []string // Approved Internal Domains (Fail-Closed)
}

// New creates a Router with the given default model name and allowed domains.
func New(defaultModel string, allowedDomains []string) *Router {
	return &Router{
		adapters:       make(map[string]ModelAdapter),
		fallback:       defaultModel,
		allowedDomains: allowedDomains,
	}
}

// Register adds a model adapter to the router.
func (r *Router) Register(adapter ModelAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// Send routes the messages to the named model. If modelName is empty, the
// fallback model is used.
func (r *Router) Send(ctx context.Context, modelName string, messages []Message, opts ModelOpts) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if modelName == "" {
		modelName = r.fallback
	}

	adapter, ok := r.adapters[modelName]
	if !ok {
		return "", fmt.Errorf("router: unknown model %q (registered: %v)", modelName, r.modelNames())
	}

	// ZERO-EGRESS VALIDATION: Fail-Closed Egress Hardening
	if !isApprovedDomain(adapter.Endpoint(), r.allowedDomains) {
		log.Printf("[SECURITY ALERT] Egress Blocked: Model %q attempted to call unapproved domain %q", modelName, adapter.Endpoint())
		return "", fmt.Errorf("OCULTAR Zero-Egress Block: domain %q is not in the approved list (Fail-Closed)", adapter.Endpoint())
	}

	return adapter.Send(ctx, messages, opts)
}

func isApprovedDomain(targetURL string, allowed []string) bool {
	if len(allowed) == 0 {
		return true // If no policy set, default to allow (legacy)
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	host := u.Hostname()

	for _, a := range allowed {
		if host == a {
			return true
		}
	}
	return false
}

// HealthCheckAll pings every registered adapter.
func (r *Router) HealthCheckAll(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error, len(r.adapters))
	for name, adapter := range r.adapters {
		results[name] = adapter.HealthCheck(ctx)
	}
	return results
}

// Models returns the list of registered model names.
func (r *Router) Models() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.modelNames()
}

func (r *Router) modelNames() []string {
	names := make([]string, 0, len(r.adapters))
	for n := range r.adapters {
		names = append(names, n)
	}
	return names
}
