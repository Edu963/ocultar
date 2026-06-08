package inference

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockSidecar builds a test HTTP server that responds to /health and /scan
// with the supplied handlers. Passing nil for a handler installs a default
// that returns 200 OK on /health and an empty JSON object on /scan.
func mockSidecar(t *testing.T, healthStatus int, scanResponse map[string][]string, scanStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(healthStatus)
		case "/scan":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(scanStatus)
			json.NewEncoder(w).Encode(scanResponse)
		default:
			http.NotFound(w, r)
		}
	}))
}

// ── NewPrivacyFilterEngine ────────────────────────────────────────────────────

func TestNewPrivacyFilterEngine_Success(t *testing.T) {
	srv := mockSidecar(t, http.StatusOK, nil, http.StatusOK)
	defer srv.Close()

	eng, err := NewPrivacyFilterEngine(srv.URL, "test-model")
	if err != nil {
		t.Fatalf("expected successful creation, got: %v", err)
	}
	if eng == nil {
		t.Fatal("expected non-nil engine")
	}
	if eng.Name() != "privacy-filter (test-model)" {
		t.Errorf("unexpected Name(): %s", eng.Name())
	}
}

func TestNewPrivacyFilterEngine_HealthCheckFails(t *testing.T) {
	srv := mockSidecar(t, http.StatusServiceUnavailable, nil, http.StatusOK)
	defer srv.Close()

	_, err := NewPrivacyFilterEngine(srv.URL, "test-model")
	if err == nil {
		t.Error("expected error when health check returns non-200")
	}
}

func TestNewPrivacyFilterEngine_ServerUnreachable(t *testing.T) {
	// Point to a port that has no listener.
	_, err := NewPrivacyFilterEngine("http://127.0.0.1:19999", "test-model")
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

// ── ScanForPII ────────────────────────────────────────────────────────────────

func TestScanForPII_ReturnsEntities(t *testing.T) {
	expected := map[string][]string{
		"EMAIL":  {"alice@example.com"},
		"PERSON": {"Alice"},
	}
	srv := mockSidecar(t, http.StatusOK, expected, http.StatusOK)
	defer srv.Close()

	eng, err := NewPrivacyFilterEngine(srv.URL, "test-model")
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	result, err := eng.ScanForPII("Hello Alice, contact alice@example.com")
	if err != nil {
		t.Fatalf("ScanForPII: %v", err)
	}
	if len(result["EMAIL"]) == 0 || result["EMAIL"][0] != "alice@example.com" {
		t.Errorf("expected EMAIL entity, got: %v", result)
	}
	if len(result["PERSON"]) == 0 || result["PERSON"][0] != "Alice" {
		t.Errorf("expected PERSON entity, got: %v", result)
	}
}

func TestScanForPII_EmptyResponse(t *testing.T) {
	srv := mockSidecar(t, http.StatusOK, map[string][]string{}, http.StatusOK)
	defer srv.Close()

	eng, err := NewPrivacyFilterEngine(srv.URL, "model")
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	result, err := eng.ScanForPII("no PII here")
	if err != nil {
		t.Fatalf("ScanForPII: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got: %v", result)
	}
}

func TestScanForPII_ScanEndpointNonOK(t *testing.T) {
	srv := mockSidecar(t, http.StatusOK, nil, http.StatusInternalServerError)
	defer srv.Close()

	eng, err := NewPrivacyFilterEngine(srv.URL, "model")
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	_, err = eng.ScanForPII("some text")
	if err == nil {
		t.Error("expected error when /scan returns 500")
	}
}

func TestScanForPII_MalformedResponse(t *testing.T) {
	// Server returns invalid JSON on /scan.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
		case "/scan":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not-json-at-all"))
		}
	}))
	defer srv.Close()

	eng, err := NewPrivacyFilterEngine(srv.URL, "model")
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	_, err = eng.ScanForPII("some text")
	if err == nil {
		t.Error("expected error for malformed JSON response from /scan")
	}
}

func TestPrivacyFilterEngine_Close(t *testing.T) {
	srv := mockSidecar(t, http.StatusOK, nil, http.StatusOK)
	defer srv.Close()

	eng, err := NewPrivacyFilterEngine(srv.URL, "model")
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	// Close is a no-op but must not panic.
	eng.Close()
}
