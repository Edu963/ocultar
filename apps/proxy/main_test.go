package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/proxy"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/vault"
)

// enableDevMode sets devMode=true for the duration of the test.
// devMode gates log.Fatalf calls in getMasterKey/getSalt, allowing them
// to fall back to defaults instead of killing the process.
func enableDevMode(t *testing.T) {
	t.Helper()
	devMode = true
	t.Cleanup(func() { devMode = false })
}

// setEnv sets environment variables for the test and unsets them after.
func setEnv(t *testing.T, pairs ...string) {
	t.Helper()
	for i := 0; i < len(pairs); i += 2 {
		os.Setenv(pairs[i], pairs[i+1])
		k := pairs[i]
		t.Cleanup(func() { os.Unsetenv(k) })
	}
}

// ── getMasterKey / HKDF ───────────────────────────────────────────────────────

func TestGetMasterKey_HKDF_Determinism(t *testing.T) {
	enableDevMode(t)
	setEnv(t, "OCU_MASTER_KEY", "stable-test-key-material", "OCU_SALT", "stable-test-salt")

	k1 := getMasterKey()
	k2 := getMasterKey()

	if !bytes.Equal(k1, k2) {
		t.Error("HKDF is not deterministic: same inputs produced different keys")
	}
	if len(k1) != 32 {
		t.Errorf("expected 32-byte AES key, got %d bytes", len(k1))
	}
}

func TestGetMasterKey_DifferentKeyMaterial_DifferentOutput(t *testing.T) {
	enableDevMode(t)
	setEnv(t, "OCU_SALT", "shared-salt")

	os.Setenv("OCU_MASTER_KEY", "key-material-alpha")
	k1 := getMasterKey()

	os.Setenv("OCU_MASTER_KEY", "key-material-beta")
	k2 := getMasterKey()

	if bytes.Equal(k1, k2) {
		t.Error("different key material should produce different derived keys")
	}
}

func TestGetMasterKey_DifferentSalt_DifferentOutput(t *testing.T) {
	enableDevMode(t)
	os.Setenv("OCU_MASTER_KEY", "same-key-material")
	t.Cleanup(func() { os.Unsetenv("OCU_MASTER_KEY") })

	os.Setenv("OCU_SALT", "salt-one")
	k1 := getMasterKey()

	os.Setenv("OCU_SALT", "salt-two")
	k2 := getMasterKey()

	t.Cleanup(func() { os.Unsetenv("OCU_SALT") })

	if bytes.Equal(k1, k2) {
		t.Error("different salts should produce different derived keys")
	}
}

func TestGetMasterKey_DevMode_UsesDefaultKey(t *testing.T) {
	enableDevMode(t)
	os.Unsetenv("OCU_MASTER_KEY")
	os.Unsetenv("OCU_SALT")

	key := getMasterKey()
	if len(key) != 32 {
		t.Errorf("expected 32-byte fallback key in dev mode, got %d bytes", len(key))
	}
	// Calling twice with same defaults must be deterministic.
	key2 := getMasterKey()
	if !bytes.Equal(key, key2) {
		t.Error("dev-mode default key is not deterministic")
	}
}

// ── getSalt ───────────────────────────────────────────────────────────────────

func TestGetSalt_FromEnv(t *testing.T) {
	enableDevMode(t)
	setEnv(t, "OCU_SALT", "my-unique-deployment-salt")
	if got := getSalt(); got != "my-unique-deployment-salt" {
		t.Errorf("expected env salt, got %q", got)
	}
}

func TestGetSalt_DevMode_DefaultSalt(t *testing.T) {
	enableDevMode(t)
	os.Unsetenv("OCU_SALT")
	if got := getSalt(); got != defaultSalt {
		t.Errorf("expected default salt %q in dev mode, got %q", defaultSalt, got)
	}
}

// ── /healthz endpoint ─────────────────────────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"` + VERSION + `"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode healthz response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
	if body["version"] != VERSION {
		t.Errorf("expected version=%s, got %q", VERSION, body["version"])
	}
}

// ── E2E proxy flow: PII redaction ─────────────────────────────────────────────

// TestProxyHandler_PII_RedactedBeforeUpstream verifies that the proxy intercepts
// a request containing a plaintext email address and forwards only the tokenised
// form to the upstream target — raw PII must never leave the local boundary.
func TestProxyHandler_PII_RedactedBeforeUpstream(t *testing.T) {
	enableDevMode(t)
	os.Unsetenv("OCU_MASTER_KEY")
	os.Unsetenv("OCU_SALT")

	// Capture what the upstream target actually receives.
	var upstreamBody string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		upstreamBody = string(data)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"choices": "mock response"})
	}))
	defer target.Close()

	masterKey := getMasterKey()
	cfg := config.Settings{VaultBackend: "duckdb"}
	vaultProvider, err := vault.New(cfg, "")
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	defer vaultProvider.Close()

	config.InitDefaults()
	eng := refinery.NewRefinery(vaultProvider, masterKey)

	handler, err := proxy.NewHandler(eng, vaultProvider, masterKey, target.URL)
	if err != nil {
		t.Fatalf("create proxy handler: %v", err)
	}

	proxySrv := httptest.NewServer(handler)
	defer proxySrv.Close()

	reqBody := `{"model":"gpt-4o","messages":[{"role":"user","content":"Hi, my email is secret@example.com"}]}`
	resp, err := http.Post(proxySrv.URL+"/v1/chat/completions", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST to proxy: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("proxy returned %d: %s", resp.StatusCode, body)
	}

	// The upstream must NOT have seen the raw email address.
	if strings.Contains(upstreamBody, "secret@example.com") {
		t.Error("raw email address leaked to upstream — PII was NOT redacted by the proxy")
	}
	// The upstream payload should contain an EMAIL token instead.
	if !strings.Contains(upstreamBody, "[EMAIL_") {
		t.Errorf("expected [EMAIL_...] token in upstream payload; got: %s", upstreamBody)
	}
}
