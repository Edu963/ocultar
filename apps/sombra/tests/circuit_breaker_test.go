package connector_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/connector"
	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/handler"
	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/router"
	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/pkg/refinery"
	"github.com/ocultar-dev/ocultar/vault"
)

// TestCircuitBreaker_SLMUnavailable_TierOneFallback verifies that when the Tier 2
// SLM sidecar is unreachable, the gateway degrades gracefully to Tier 1 only —
// it must still return 200 OK and must not hang or return a 5xx error.
//
// The refinery is initialised without an SLM URL, which simulates the SLM being
// absent. Tier 1 deterministic detection must still run and mask any PII it finds.
func TestCircuitBreaker_SLMUnavailable_TierOneFallback(t *testing.T) {
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	defer v.Close()

	masterKey := make([]byte, 32)
	config.InitDefaults()

	// Refinery with no SLM URL — Tier 2 circuit breaker will be open from the start.
	eng := refinery.NewRefinery(v, masterKey)

	upstream := &mockModelAdapter{name: "mock-model"}
	r := router.New("mock-model", []string{"mock-internal"})
	r.Register(upstream)

	gw := handler.NewGateway(eng, v, masterKey, r, nil)
	gw.RegisterConnector(connector.NewFileConnector("file", connector.DataPolicy{
		AllowedModels: []string{"mock-model"},
	}))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("connector", "file")
	mw.WriteField("model", "mock-model")
	mw.WriteField("prompt", "Summarise this document.")
	part, _ := mw.CreateFormFile("file", "doc.txt")
	part.Write([]byte("Contact john.doe@example.com or call 555-123-4567."))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/query", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-actor")

	rr := httptest.NewRecorder()
	gw.HandleQuery(rr, req)

	// Must succeed — Tier 1 should have handled this without the SLM.
	if rr.Code != http.StatusOK {
		t.Errorf("circuit breaker fallback failed: expected 200 OK, got %d\nbody: %s",
			rr.Code, rr.Body.String())
	}
}
