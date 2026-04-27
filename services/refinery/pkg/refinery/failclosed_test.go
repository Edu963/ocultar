package refinery_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/proxy"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/vault"
)

// frenchPII is the standard input used across all fail-closed tests.
// It contains three distinct PII tokens we assert never appear in any response body.
const frenchPII = `{"message": "Contact Jean-Pierre Dumont at jean.pierre@societe.fr"}`

var fcMasterKey = []byte("01234567890123456789012345678901") // 32 bytes

// ── Mock vault implementations ────────────────────────────────────────────────

type passingVault struct{}

func (p *passingVault) StoreToken(hash, token, enc string) (bool, error) { return true, nil }
func (p *passingVault) GetToken(hash string) (string, bool)               { return "", false }
func (p *passingVault) CountAll() int64                                    { return 0 }
func (p *passingVault) Close() error                                       { return nil }

type failingVault struct{}

func (f *failingVault) StoreToken(hash, token, enc string) (bool, error) {
	return false, errors.New("vault: write failure simulated")
}
func (f *failingVault) GetToken(hash string) (string, bool) { return "", false }
func (f *failingVault) CountAll() int64                      { return 0 }
func (f *failingVault) Close() error                         { return nil }

type slowVault struct{ delay time.Duration }

func (s *slowVault) StoreToken(hash, token, enc string) (bool, error) {
	time.Sleep(s.delay)
	return true, nil
}
func (s *slowVault) GetToken(hash string) (string, bool) { return "", false }
func (s *slowVault) CountAll() int64                      { return 0 }
func (s *slowVault) Close() error                         { return nil }

// ── Mock AI scanner ───────────────────────────────────────────────────────────

type failingScanner struct{}

func (f *failingScanner) ScanForPII(text string) (map[string][]string, error) {
	return nil, errors.New("slm: connection refused")
}
func (f *failingScanner) CheckHealth(host string) {}
func (f *failingScanner) IsAvailable() bool       { return true } // forces the scan path
func (f *failingScanner) SetDomain(domain string) {}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestProxy(t *testing.T, eng *refinery.Refinery, v vault.Provider, upstreamURL string) *httptest.Server {
	t.Helper()
	h, err := proxy.NewHandler(eng, v, fcMasterKey, upstreamURL)
	if err != nil {
		t.Fatalf("proxy.NewHandler: %v", err)
	}
	return httptest.NewServer(h)
}

func fakeUpstreamServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status":"ok"}`)
	}))
}

// assertNoPII fails the test if any of the three PII tokens appear in body.
func assertNoPII(t *testing.T, body string) {
	t.Helper()
	for _, pii := range []string{"Jean-Pierre", "Dumont", "jean.pierre@societe.fr"} {
		if strings.Contains(body, pii) {
			t.Errorf("PII leaked in response body: %q found", pii)
		}
	}
}

func doPost(t *testing.T, url string, timeout time.Duration) (*http.Response, error) {
	t.Helper()
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(frenchPII))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

// ── Failure scenario tests ────────────────────────────────────────────────────

// TestFailClosed_SLMUnreachable verifies that when the AI scanner returns an
// error the proxy returns HTTP 500, emits no PII, and withholds X-Ocultar-Redacted.
func TestFailClosed_SLMUnreachable(t *testing.T) {
	config.InitDefaults()

	upstream := fakeUpstreamServer()
	defer upstream.Close()

	eng := refinery.NewRefinery(&passingVault{}, fcMasterKey)
	eng.SetAIScanner(&failingScanner{})

	ps := newTestProxy(t, eng, &passingVault{}, upstream.URL)
	defer ps.Close()

	resp, err := doPost(t, ps.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", resp.StatusCode)
	}
	assertNoPII(t, string(body))
	if resp.Header.Get("X-Ocultar-Redacted") != "" {
		t.Error("X-Ocultar-Redacted must be absent when SLM fails")
	}
}

// TestFailClosed_VaultWriteFailure verifies that a vault write error causes
// HTTP 500, no PII in the response body, and no X-Ocultar-Redacted header.
func TestFailClosed_VaultWriteFailure(t *testing.T) {
	config.InitDefaults()

	upstream := fakeUpstreamServer()
	defer upstream.Close()

	eng := refinery.NewRefinery(&failingVault{}, fcMasterKey)

	ps := newTestProxy(t, eng, &failingVault{}, upstream.URL)
	defer ps.Close()

	resp, err := doPost(t, ps.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", resp.StatusCode)
	}
	assertNoPII(t, string(body))
	if resp.Header.Get("X-Ocultar-Redacted") != "" {
		t.Error("X-Ocultar-Redacted must be absent when vault write fails")
	}
}

// TestFailClosed_SlowVault verifies that a vault with a 15-second artificial delay
// does not leak PII: the client times out and receives no successful response body.
func TestFailClosed_SlowVault(t *testing.T) {
	config.InitDefaults()

	upstream := fakeUpstreamServer()
	defer upstream.Close()

	eng := refinery.NewRefinery(&slowVault{delay: 15 * time.Second}, fcMasterKey)

	ps := newTestProxy(t, eng, &slowVault{delay: 15 * time.Second}, upstream.URL)
	defer ps.Close()

	// Client timeout of 3s — much shorter than the 15s vault delay.
	resp, err := doPost(t, ps.URL, 3*time.Second)
	if err != nil {
		// Expected: client deadline exceeded before any response.
		// No response body means no PII could have been returned.
		if !strings.Contains(err.Error(), "deadline") &&
			!strings.Contains(err.Error(), "timeout") &&
			!strings.Contains(err.Error(), "context") {
			t.Errorf("unexpected error type: %v", err)
		}
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// If a response did arrive it must not be 200 with PII.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("got unexpected 200 response while vault was blocked")
	}
	assertNoPII(t, string(body))
}

// TestFailClosed_QueueFull verifies that when the proxy's wait queue is
// saturated it returns HTTP 429, no PII in body, and no X-Ocultar-Redacted.
func TestFailClosed_QueueFull(t *testing.T) {
	config.InitDefaults()
	config.UpdateSystemLimits(1, 1) // concurrency=1, queue=1

	// Slow upstream — holds the first request so the queue stays full.
	gate := make(chan struct{})
	slowUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-gate
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status":"ok"}`)
	}))
	defer slowUpstream.Close()
	defer func() { close(gate) }()

	eng := refinery.NewRefinery(&passingVault{}, fcMasterKey)
	ps := newTestProxy(t, eng, &passingVault{}, slowUpstream.URL)
	defer ps.Close()

	// First request — blocks in the upstream, holds the queue slot.
	go func() {
		doPost(t, ps.URL, 30*time.Second) //nolint:errcheck
	}()
	time.Sleep(80 * time.Millisecond) // let the first request enter and block

	// Second request — queue is full, expect 429.
	resp, err := doPost(t, ps.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", resp.StatusCode)
	}
	assertNoPII(t, string(body))
	if resp.Header.Get("X-Ocultar-Redacted") != "" {
		t.Error("X-Ocultar-Redacted must be absent when queue is full")
	}
}

// TestFailClosed_ProtectedEntitiesBootGuard verifies that after normal
// initialization the PROTECTED_ENTITY dictionary is non-empty. The refinery
// calls log.Fatalf at boot if protected_entities.json is empty or missing —
// this test confirms the guard has entities to enforce.
// (The Fatalf path itself cannot be tested in-process; see config.go:loadProtectedEntities.)
func TestFailClosed_ProtectedEntitiesBootGuard(t *testing.T) {
	config.InitDefaults()

	for _, d := range config.Global.Dictionaries {
		if d.Type == "PROTECTED_ENTITY" && len(d.Terms) > 0 {
			return // guard is in place
		}
	}
	t.Error("PROTECTED_ENTITY dictionary is empty after InitDefaults — boot guard has nothing to enforce")
}

// ── Happy path ────────────────────────────────────────────────────────────────

// TestFailClosed_HappyPath verifies that a normal request returns HTTP 200
// and that raw PII does not reach the upstream — the email is tokenized in transit.
// Tiers 0–1.5 catch the email address; name detection requires Tier 2 (Enterprise).
func TestFailClosed_HappyPath(t *testing.T) {
	config.InitDefaults()

	// Upstream captures what the proxy forwards so we can inspect it.
	var upstreamBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		upstreamBody = string(b)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"answer":"done"}`)
	}))
	defer upstream.Close()

	eng := refinery.NewRefinery(&passingVault{}, fcMasterKey)
	ps := newTestProxy(t, eng, &passingVault{}, upstream.URL)
	defer ps.Close()

	resp, err := doPost(t, ps.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	// The email must not reach the upstream in raw form.
	if strings.Contains(upstreamBody, "jean.pierre@societe.fr") {
		t.Error("raw email leaked to upstream — refinery did not tokenize it")
	}
	// The token placeholder must be present, confirming redaction happened.
	if !strings.Contains(upstreamBody, "[EMAIL_") {
		t.Error("expected EMAIL token in upstream body, refinery may not have run")
	}
}
