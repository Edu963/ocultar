package connector_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/apps/sombra/pkg/router"
)

// ── router.New ────────────────────────────────────────────────────────────────

func TestRouter_New_RegistersAdapter(t *testing.T) {
	r := router.New("mock-model", nil)
	r.Register(&mockModelAdapter{name: "mock-model"})

	models := r.Models()
	if len(models) != 1 || models[0] != "mock-model" {
		t.Errorf("expected [mock-model], got %v", models)
	}
}

func TestRouter_New_MultipleAdapters(t *testing.T) {
	r := router.New("primary", nil)
	r.Register(&mockModelAdapter{name: "primary"})
	r.Register(&mockModelAdapter{name: "secondary"})

	models := r.Models()
	if len(models) != 2 {
		t.Errorf("expected 2 adapters, got %d: %v", len(models), models)
	}
}

// ── router.Send ───────────────────────────────────────────────────────────────

func TestRouter_Send_UsesNamedAdapter(t *testing.T) {
	r := router.New("mock-model", nil)
	r.Register(&mockModelAdapter{name: "mock-model"})

	msgs := []router.Message{{Role: "user", Content: "hello"}}
	resp, err := r.Send(context.Background(), "mock-model", msgs, router.ModelOpts{})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	// mockModelAdapter echoes the last message content.
	if !strings.Contains(resp, "hello") {
		t.Errorf("expected echo of prompt in response, got: %s", resp)
	}
}

func TestRouter_Send_UnknownModel_ReturnsError(t *testing.T) {
	r := router.New("mock-model", nil)
	r.Register(&mockModelAdapter{name: "mock-model"})

	_, err := r.Send(context.Background(), "nonexistent-model", nil, router.ModelOpts{})
	if err == nil {
		t.Error("expected error for unknown model, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-model") {
		t.Errorf("error should mention model name, got: %v", err)
	}
}

func TestRouter_Send_EmptyModelName_UsesFallback(t *testing.T) {
	r := router.New("fallback-model", nil)
	r.Register(&mockModelAdapter{name: "fallback-model"})

	msgs := []router.Message{{Role: "user", Content: "test fallback"}}
	resp, err := r.Send(context.Background(), "", msgs, router.ModelOpts{})
	if err != nil {
		t.Fatalf("Send with empty model name: %v", err)
	}
	if !strings.Contains(resp, "test fallback") {
		t.Errorf("expected fallback adapter to handle request, got: %s", resp)
	}
}

// ── Zero-egress domain validation ────────────────────────────────────────────

func TestRouter_Send_BlocksUnapprovedDomain(t *testing.T) {
	// allowedDomains is set to only allow "internal.corp" — the mock adapter
	// returns "http://mock-internal" as its endpoint, which doesn't match.
	r := router.New("mock-model", []string{"internal.corp"})
	r.Register(&mockModelAdapter{name: "mock-model"})

	_, err := r.Send(context.Background(), "mock-model", nil, router.ModelOpts{})
	if err == nil {
		t.Fatal("expected zero-egress block for unapproved domain, got nil")
	}
	if !strings.Contains(err.Error(), "Zero-Egress") && !strings.Contains(err.Error(), "approved") {
		t.Errorf("expected zero-egress error message, got: %v", err)
	}
}

func TestRouter_Send_AllowsApprovedDomain(t *testing.T) {
	// The mockModelAdapter returns "http://mock-internal" — hostname is "mock-internal".
	r := router.New("mock-model", []string{"mock-internal"})
	r.Register(&mockModelAdapter{name: "mock-model"})

	msgs := []router.Message{{Role: "user", Content: "approved domain test"}}
	_, err := r.Send(context.Background(), "mock-model", msgs, router.ModelOpts{})
	if err != nil {
		t.Errorf("expected success for approved domain, got: %v", err)
	}
}

func TestRouter_Send_EmptyAllowedDomains_DefaultsToAllow(t *testing.T) {
	// When allowedDomains is nil/empty, the router defaults to allow (legacy mode).
	r := router.New("mock-model", nil)
	r.Register(&mockModelAdapter{name: "mock-model"})

	msgs := []router.Message{{Role: "user", Content: "open policy"}}
	_, err := r.Send(context.Background(), "mock-model", msgs, router.ModelOpts{})
	if err != nil {
		t.Errorf("expected allow with empty domain list, got: %v", err)
	}
}

// ── HealthCheckAll ────────────────────────────────────────────────────────────

func TestRouter_HealthCheckAll(t *testing.T) {
	r := router.New("mock-model", nil)
	r.Register(&mockModelAdapter{name: "mock-model"})
	r.Register(&mockModelAdapter{name: "secondary"})

	results := r.HealthCheckAll(context.Background())
	if len(results) != 2 {
		t.Errorf("expected 2 health check results, got %d", len(results))
	}
	for name, err := range results {
		if err != nil {
			t.Errorf("adapter %q health check failed: %v", name, err)
		}
	}
}
