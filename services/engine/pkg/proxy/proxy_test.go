package proxy_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/pkg/proxy"
	"github.com/Edu963/ocultar/vault"
)

// setupTestStack initialises an in-memory DuckDB vault + engine + proxy handler
// suitable for unit testing. masterKey is a fixed 32-byte test key.
func setupTestStack(t *testing.T, upstreamURL string) (*proxy.Handler, vault.Provider, []byte) {
	t.Helper()

	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	t.Cleanup(func() { v.Close() })

	masterKey := []byte("01234567890123456789012345678901") // 32 bytes

	config.InitDefaults()
	eng := engine.NewEngine(v, masterKey)
	eng.Serve = "proxy"

	h, err := proxy.NewHandler(eng, v, masterKey, upstreamURL)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h, v, masterKey
}

// TestProxyRedactsRequestPII verifies that PII in the outgoing request body
// is replaced with vault tokens before it reaches the upstream server.
func TestProxyRedactsRequestPII(t *testing.T) {
	var upstreamReceived []byte

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamReceived, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		// Echo back whatever was sent (tokens arrive in response; rehydration will restore PII).
		w.Write(upstreamReceived)
	}))
	defer upstream.Close()

	handler, _, _ := setupTestStack(t, upstream.URL)
	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "My name is John Doe and my email is john@example.com. Call me at +33 6 12 34 56 78."},
		},
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(proxyServer.URL+"/v1/chat/completions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	upstreamStr := string(upstreamReceived)

	// The upstream must NOT see the raw email address.
	if strings.Contains(upstreamStr, "john@example.com") {
		t.Errorf("🚨 EMAIL leaked to upstream!\nUpstream body: %s", upstreamStr)
	} else {
		t.Logf("✅ Email redacted in upstream request")
	}

	// The upstream must NOT see the raw phone number.
	if strings.Contains(upstreamStr, "+33 6 12 34 56 78") {
		t.Errorf("🚨 PHONE leaked to upstream!\nUpstream body: %s", upstreamStr)
	} else {
		t.Logf("✅ Phone redacted in upstream request")
	}

	// Tokens must be present.
	if !strings.Contains(upstreamStr, "[EMAIL_") {
		t.Errorf("🚨 Expected [EMAIL_...] token in upstream body, got: %s", upstreamStr)
	}
}

// TestProxyRehydratesResponseTokens verifies that the proxy restores tokens
// in the upstream response back to the original PII values before returning
// the result to the client.
func TestProxyRehydratesResponseTokens(t *testing.T) {
	// We'll make the proxy redact a request first (which vaults the tokens),
	// then simulate an upstream that echoes back those tokens.
	var capturedToken string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		// Extract the token from the redacted content string (first message content).
		msgs, _ := req["messages"].([]interface{})
		if len(msgs) > 0 {
			if m, ok := msgs[0].(map[string]interface{}); ok {
				capturedToken = m["content"].(string)
			}
		}
		// Reply echoing the token as if the LLM summarised the message.
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"role": "assistant", "content": "The user mentioned " + capturedToken}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstream.Close()

	handler, _, _ := setupTestStack(t, upstream.URL)
	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	payload := map[string]interface{}{
		"model":    "gpt-4o",
		"messages": []map[string]string{{"role": "user", "content": "My email is rehydration@test.io"}},
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(proxyServer.URL+"/v1/chat/completions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	clientBody, _ := io.ReadAll(resp.Body)
	clientStr := string(clientBody)

	// The client must receive the original email, not the token.
	if !strings.Contains(clientStr, "rehydration@test.io") {
		t.Errorf("🚨 Email NOT re-hydrated in client response!\nClient body: %s", clientStr)
	} else {
		t.Logf("✅ Email re-hydrated in client response: %s", clientStr)
	}
}

// TestVaultRehydrateString is a unit test for the vault helper directly.
func TestVaultRehydrateString(t *testing.T) {
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	defer v.Close()

	masterKey := []byte("01234567890123456789012345678901")
	config.InitDefaults()
	eng := engine.NewEngine(v, masterKey)
	eng.Serve = "proxy"

	// Redact a string — this vaults the token.
	refined, err := eng.RefineString("Contact alice@company.org for details.", "test", nil)
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	if !strings.Contains(refined, "[EMAIL_") {
		t.Fatalf("expected redaction, got: %s", refined)
	}

	// Now re-hydrate — we should get the original email back.
	restored, err := proxy.RehydrateString(v, masterKey, refined)
	if err != nil {
		t.Fatalf("unexpected re-hydration error: %v", err)
	}
	if !strings.Contains(restored, "alice@company.org") {
		t.Errorf("🚨 Re-hydration failed!\nRefined:  %s\nRestored: %s", refined, restored)
	} else {
		t.Logf("✅ Re-hydration OK: %s → %s", refined, restored)
	}
}
