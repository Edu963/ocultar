package connector_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/connector"
	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/handler"
	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/router"
	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/pkg/refinery"
	"github.com/ocultar-dev/ocultar/vault"
)

// capturingAdapter records the exact message content sent to the upstream AI.
type capturingAdapter struct {
	capturedPrompt string
}

func (a *capturingAdapter) Name() string { return "capturing-model" }
func (a *capturingAdapter) Send(_ context.Context, messages []router.Message, _ router.ModelOpts) (string, error) {
	if len(messages) > 0 {
		a.capturedPrompt = messages[len(messages)-1].Content
	}
	return "ok", nil
}
func (a *capturingAdapter) Endpoint() string                    { return "http://mock-capture" }
func (a *capturingAdapter) HealthCheck(_ context.Context) error { return nil }

// TestZeroEgress_RawPIINeverReachesUpstream sends a prompt containing clear PII
// through the Sombra gateway and verifies that the upstream AI receives only
// masked tokens — never the original plaintext values.
func TestZeroEgress_RawPIINeverReachesUpstream(t *testing.T) {
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	defer v.Close()

	masterKey := make([]byte, 32)
	config.InitDefaults()
	eng := refinery.NewRefinery(v, masterKey)

	upstream := &capturingAdapter{}
	r := router.New("capturing-model", []string{"mock-capture"})
	r.Register(upstream)

	gw := handler.NewGateway(eng, v, masterKey, r, nil)
	gw.RegisterConnector(connector.NewFileConnector("file", connector.DataPolicy{
		AllowedModels: []string{"capturing-model"},
	}))

	// PII types detected by Tier 1 (deterministic regex). Names require Tier 2 NER
	// and are intentionally excluded from this assertion.
	rawPII := []string{
		"alice@hospital.org",
		"123-45-6789",
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("connector", "file")
	mw.WriteField("model", "capturing-model")
	mw.WriteField("prompt", "Summarise this patient record.")
	part, _ := mw.CreateFormFile("file", "record.txt")
	part.Write([]byte("Patient: Alice Martin, SSN: 123-45-6789, email: alice@hospital.org"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/query", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-actor")

	rr := httptest.NewRecorder()
	gw.HandleQuery(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	// The upstream must have been called.
	if upstream.capturedPrompt == "" {
		t.Fatal("upstream was never called — cannot verify zero-egress")
	}

	// Verify none of the raw PII values appear in what was sent to the upstream.
	for _, pii := range rawPII {
		if strings.Contains(upstream.capturedPrompt, pii) {
			t.Errorf("zero-egress violation: raw PII %q found in upstream prompt\nupstream received: %q",
				pii, upstream.capturedPrompt)
		}
	}
}
