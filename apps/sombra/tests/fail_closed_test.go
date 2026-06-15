package connector_test

import (
	"bytes"
	"context"
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

// recordingAdapter tracks whether the upstream AI was ever called.
type recordingAdapter struct {
	called bool
}

func (a *recordingAdapter) Name() string { return "recording-model" }
func (a *recordingAdapter) Send(_ context.Context, _ []router.Message, _ router.ModelOpts) (string, error) {
	a.called = true
	return "upstream reached — fail-closed violated", nil
}
func (a *recordingAdapter) Endpoint() string                        { return "http://mock-upstream" }
func (a *recordingAdapter) HealthCheck(_ context.Context) error     { return nil }

// TestFailClosed_RefineryCrash_NeverForwardsRaw verifies the fail-closed contract:
// if the vault (and therefore the refinery) is unavailable, the gateway must return
// a 5xx error and must never call the upstream AI model.
func TestFailClosed_RefineryCrash_NeverForwardsRaw(t *testing.T) {
	// Create a vault and immediately close it to simulate an unavailable refinery.
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	v.Close() // closed before the request — all StoreToken calls will fail

	masterKey := make([]byte, 32)
	config.InitDefaults()
	eng := refinery.NewRefinery(v, masterKey)

	upstream := &recordingAdapter{}
	r := router.New("recording-model", []string{"http://mock-upstream"})
	r.Register(upstream)

	gw := handler.NewGateway(eng, v, masterKey, r, nil)
	gw.RegisterConnector(connector.NewFileConnector("file", connector.DataPolicy{
		AllowedModels: []string{"recording-model"},
	}))

	// Build a request containing clear PII that the Tier 1 engine will detect.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("connector", "file")
	mw.WriteField("model", "recording-model")
	mw.WriteField("prompt", "What is the patient SSN?")
	part, _ := mw.CreateFormFile("file", "record.txt")
	part.Write([]byte("Patient SSN: 123-45-6789, email: bob@hospital.org"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/query", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-actor")

	rr := httptest.NewRecorder()
	gw.HandleQuery(rr, req)

	if rr.Code == http.StatusOK {
		t.Errorf("fail-closed violation: got 200 OK despite broken refinery; body: %s", rr.Body.String())
	}

	if upstream.called {
		t.Error("fail-closed violation: upstream AI was called despite refinery failure — raw PII may have been forwarded")
	}
}
