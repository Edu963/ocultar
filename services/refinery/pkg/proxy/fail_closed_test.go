package proxy_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/proxy"
)

// MockVault is a stub Vault Provider that can be configured to fail.
type MockVault struct {
	ShouldFail bool
}

func (m *MockVault) StoreToken(hash, token, encrypted string) (bool, error) {
	if m.ShouldFail {
		return false, io.ErrUnexpectedEOF // Simulate a DB/write failure
	}
	return true, nil
}
func (m *MockVault) GetToken(hash string) (string, bool) { return "", false }
func (m *MockVault) CountAll() int64                     { return 0 }
func (m *MockVault) Close() error                        { return nil }

// MockAIScanner is a stub Scanner that can be configured to fail.
type MockAIScanner struct {
	ShouldFail bool
}

func (m *MockAIScanner) ScanForPII(text string) (map[string][]string, error) {
	if m.ShouldFail {
		return nil, io.ErrClosedPipe // Simulate an offline/crashing SLM API
	}
	return map[string][]string{"PERSON": {"John Doe"}}, nil
}
func (m *MockAIScanner) CheckHealth(host string) {}
func (m *MockAIScanner) IsAvailable() bool       { return true } // Always pretend it's on to force a scan
func (m *MockAIScanner) SetDomain(domain string) {}

// setupTestProxy orchestrates the core Refinery + Proxy with the given mocks.
func setupTestProxy(t *testing.T, vaultFails bool, aiFails bool) (*httptest.Server, *httptest.Server) {
	// 1. Mock the upstream Target application
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "received": true}`))
	}))

	// 2. Configure Refinery
	mockVault := &MockVault{ShouldFail: vaultFails}
	masterKey := make([]byte, 32) // dummy key
	eng := refinery.NewRefinery(mockVault, masterKey)
	eng.AIScanner = &MockAIScanner{ShouldFail: aiFails}

	// 3. Configure Proxy
	p, err := proxy.NewHandler(eng, mockVault, masterKey, upstream.URL)
	if err != nil {
		t.Fatalf("Failed to initialize proxy: %v", err)
	}

	proxyServer := httptest.NewServer(p)
	return proxyServer, upstream
}

// executeTestRequest fires a dummy JSON payload at the proxy and returns the status code.
func executeTestRequest(t *testing.T, proxyURL string) int {
	payload := []byte(`{"message": "Hello John Doe, your email is test@example.com."}`)
	req, err := http.NewRequest("POST", proxyURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed creating test request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed executing test request: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// TestFailClosed_VaultError asserts that if the Vault backend goes offline or fails
// to insert tokens, the Proxy returns an HTTP 500 block instead of allowing
// the PII to leak unredacted downstream.
func TestFailClosed_VaultError(t *testing.T) {
	proxyServer, upstream := setupTestProxy(t, true, false)
	defer proxyServer.Close()
	defer upstream.Close()

	statusCode := executeTestRequest(t, proxyServer.URL)
	if statusCode != http.StatusInternalServerError {
		t.Errorf("Expected HTTP 500 Internal Server Error when Vault fails, got %d", statusCode)
	}
}

// TestFailClosed_ModelInferenceError asserts that if the AI Model timeouts or crashes,
// the Proxy returns an HTTP 500 block.
func TestFailClosed_ModelInferenceError(t *testing.T) {
	proxyServer, upstream := setupTestProxy(t, false, true)
	defer proxyServer.Close()
	defer upstream.Close()

	statusCode := executeTestRequest(t, proxyServer.URL)
	if statusCode != http.StatusInternalServerError {
		t.Errorf("Expected HTTP 500 Internal Server Error when AI inference fails, got %d", statusCode)
	}
}

// TestFailClosed_Success asserts the happy path still functions.
func TestFailClosed_Success(t *testing.T) {
	proxyServer, upstream := setupTestProxy(t, false, false)
	defer proxyServer.Close()
	defer upstream.Close()

	statusCode := executeTestRequest(t, proxyServer.URL)
	if statusCode != http.StatusOK {
		t.Errorf("Expected HTTP 200 OK for happy path, got %d", statusCode)
	}
}
