package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/apps/slm-engine/pkg/inference"
)

// mockEngine is a test double for inference.Tier2Engine that returns
// a preset result without making any HTTP calls.
type mockEngine struct {
	result map[string][]string
	err    error
}

func (m *mockEngine) ScanForPII(text string) (map[string][]string, error) {
	return m.result, m.err
}
func (m *mockEngine) Name() string { return "mock-engine" }
func (m *mockEngine) Close()       {}

// injectScanner assigns a test engine and restores the original after the test.
func injectScanner(t *testing.T, eng inference.Tier2Engine) {
	t.Helper()
	orig := scanner
	scanner = eng
	t.Cleanup(func() { scanner = orig })
}

// ── handleHealth ──────────────────────────────────────────────────────────────

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ok") {
		t.Errorf("expected 'ok' in health response, got: %s", rr.Body.String())
	}
}

// ── handleScan ────────────────────────────────────────────────────────────────

func TestHandleScan_RejectsGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/scan", nil)
	rr := httptest.NewRecorder()
	handleScan(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET, got %d", rr.Code)
	}
}

func TestHandleScan_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleScan(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestHandleScan_EmptyText(t *testing.T) {
	// Empty "text" field bypasses the scanner and returns an empty map immediately.
	req := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader(`{"text":""}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleScan(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for empty text, got %d", rr.Code)
	}
	var result map[string][]string
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map for empty text, got: %v", result)
	}
}

func TestHandleScan_ValidText_ReturnsPII(t *testing.T) {
	injectScanner(t, &mockEngine{
		result: map[string][]string{
			"EMAIL":  {"alice@company.org"},
			"PERSON": {"Alice"},
		},
	})

	body := `{"text":"Hello Alice, reach me at alice@company.org"}`
	req := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleScan(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var result map[string][]string
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result["EMAIL"]) == 0 || result["EMAIL"][0] != "alice@company.org" {
		t.Errorf("expected EMAIL in result, got: %v", result)
	}
	if len(result["PERSON"]) == 0 || result["PERSON"][0] != "Alice" {
		t.Errorf("expected PERSON in result, got: %v", result)
	}
}

func TestHandleScan_ScannerError(t *testing.T) {
	injectScanner(t, &mockEngine{
		err: errors.New("backend unavailable"),
	})

	body := `{"text":"some text"}`
	req := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleScan(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on scanner error, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "scan failed") {
		t.Errorf("expected 'scan failed' in error body, got: %s", rr.Body.String())
	}
}

func TestHandleScan_MissingTextField(t *testing.T) {
	// JSON with no "text" key — treated as empty text.
	injectScanner(t, &mockEngine{result: map[string][]string{}})
	body := `{"message":"wrong key"}`
	req := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleScan(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for missing text key, got %d", rr.Code)
	}
}
