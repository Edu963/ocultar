package main

// ssrf_test.go — SSRF and DNS-rebinding protection tests for the OCULTAR proxy.
//
// The proxy accepts an Ocultar-Target header that lets callers redirect a
// request to an arbitrary URL. Every target is validated via resolveTarget /
// isPrivateIP before a TCP connection is ever attempted. Tests here verify
// that all RFC 1918 ranges, loopback, link-local, and special addresses are
// blocked with HTTP 403 and that the block message clearly identifies the
// reason. A 403 response arriving in <5 s proves no TCP connection was made
// (a connection attempt to a non-responsive internal host would time out).

import (
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

// newSSRFProxy builds a proxy whose base target is a harmless local server.
// All SSRF tests override the target via the Ocultar-Target request header.
func newSSRFProxy(t *testing.T) *httptest.Server {
	t.Helper()
	enableDevMode(t)

	safe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(safe.Close)

	masterKey := getMasterKey()
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
	if err != nil {
		t.Fatalf("vault: %v", err)
	}
	t.Cleanup(func() { v.Close() })

	config.InitDefaults()
	eng := refinery.NewRefinery(v, masterKey)
	eng.Serve = "proxy"

	h, err := proxy.NewHandler(eng, v, masterKey, safe.URL)
	if err != nil {
		t.Fatalf("proxy.NewHandler: %v", err)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// ssrfDo fires a POST at proxyURL with Ocultar-Target set to targetURL.
// A 5-second client timeout means any response faster than that proves
// no actual TCP dial was attempted toward the blocked address.
func ssrfDo(t *testing.T, proxyURL, targetURL string) *http.Response {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, proxyURL, strings.NewReader(`{"x":1}`))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocultar-Target", targetURL)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request with Ocultar-Target=%q failed: %v", targetURL, err)
	}
	return resp
}

// assertBlocked verifies that the proxy returned 403 and that the body
// contains a recognisable block reason.
func assertBlocked(t *testing.T, resp *http.Response, label string) {
	t.Helper()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bs := strings.ToLower(string(body))

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("[%s] want HTTP 403, got %d — body: %s", label, resp.StatusCode, body)
	}
	if !strings.Contains(bs, "blocked") && !strings.Contains(bs, "private") &&
		!strings.Contains(bs, "ssrf") && !strings.Contains(bs, "forbidden") {
		t.Errorf("[%s] want body containing 'blocked'/'private'/'ssrf', got: %s", label, body)
	}
}

// TestSSRF_BlockedAddresses is the main table-driven SSRF test.
// SECURITY: fail-closed required — every entry must return 403 and MUST NOT
// result in an outbound TCP connection to the listed address.
func TestSSRF_BlockedAddresses(t *testing.T) {
	srv := newSSRFProxy(t)

	cases := []struct {
		name      string
		targetURL string
	}{
		// ── RFC 1918 class A ──────────────────────────────────────────────────
		// SECURITY: fail-closed required
		{"rfc1918_10_lower", "http://10.0.0.1/"},
		// SECURITY: fail-closed required
		{"rfc1918_10_upper", "http://10.255.255.255/"},

		// ── RFC 1918 class B ──────────────────────────────────────────────────
		// SECURITY: fail-closed required
		{"rfc1918_172_lower", "http://172.16.0.1/"},
		// SECURITY: fail-closed required
		{"rfc1918_172_upper", "http://172.31.255.255/"},

		// ── RFC 1918 class C ──────────────────────────────────────────────────
		// SECURITY: fail-closed required
		{"rfc1918_192_lower", "http://192.168.0.1/"},
		// SECURITY: fail-closed required
		{"rfc1918_192_upper", "http://192.168.255.255/"},

		// ── Loopback ──────────────────────────────────────────────────────────
		// SECURITY: fail-closed required
		{"loopback_127_0_0_1", "http://127.0.0.1/"},
		// SECURITY: fail-closed required
		{"loopback_127_upper", "http://127.255.255.255/"},
		// SECURITY: fail-closed required — DNS rebinding via localhost
		// localhost resolves to 127.0.0.1/::1; isPrivateIP catches both.
		{"dns_rebind_localhost", "http://localhost/"},

		// ── Link-local ────────────────────────────────────────────────────────
		// CRITICAL: AWS/GCP/Azure metadata endpoint
		// Blocking this prevents cloud credential theft via SSRF
		// SECURITY: fail-closed required
		{"link_local_imds", "http://169.254.169.254/"},
		// SECURITY: fail-closed required
		{"link_local_lower", "http://169.254.0.1/"},

		// ── Special addresses ─────────────────────────────────────────────────
		// SECURITY: fail-closed required — unspecified address can reach localhost
		{"unspecified_ipv4", "http://0.0.0.0/"},
		// SECURITY: fail-closed required
		{"ipv6_loopback", "http://[::1]/"},
		// SECURITY: fail-closed required — IPv4-mapped IPv6 loopback
		{"ipv4_mapped_ipv6_loopback", "http://[::ffff:127.0.0.1]/"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := ssrfDo(t, srv.URL, tc.targetURL)
			assertBlocked(t, resp, tc.name)
		})
	}
}

// TestSSRF_PublicTarget_NotBlocked verifies that a publicly routable address
// passes the SSRF guard (the proxy may still fail for other reasons — network
// unreachability, no HTTP server — but the SSRF layer must not block it).
func TestSSRF_PublicTarget_NotBlocked(t *testing.T) {
	srv := newSSRFProxy(t)
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(`{"x":1}`))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocultar-Target", "http://8.8.8.8/")

	resp, err := client.Do(req)
	if err != nil {
		// Network unreachable in test environment — SSRF layer was not the issue.
		t.Skipf("8.8.8.8 unreachable in test environment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("8.8.8.8 (public IP) must not be SSRF-blocked, got 403: %s", body)
	}
}
