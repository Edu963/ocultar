package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Edu963/ocultar-proxy/middleware"
)

// ── mock store ────────────────────────────────────────────────────────────────

type mockStore struct {
	mu     sync.Mutex
	counts map[string]int64
}

func newMockStore() *mockStore {
	return &mockStore{counts: make(map[string]int64)}
}

func (m *mockStore) key(apiKeyHash, month string) string {
	return apiKeyHash + ":" + month
}

func (m *mockStore) Increment(apiKeyHash, month string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[m.key(apiKeyHash, month)]++
	return m.counts[m.key(apiKeyHash, month)], nil
}

func (m *mockStore) Count(apiKeyHash, month string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counts[m.key(apiKeyHash, month)], nil
}

// seed sets a call count directly — used to simulate an over-limit state.
func (m *mockStore) seed(apiKeyHash, month string, count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[m.key(apiKeyHash, month)] = count
}

// ── helpers ───────────────────────────────────────────────────────────────────

// okHandler is a trivial upstream that always returns HTTP 200.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"ok":true}`)
})

// doRequest fires a request at h with the given tier header and Authorization.
func doRequest(h http.Handler, tier, auth string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/refine", strings.NewReader(`{"x":1}`))
	req.Header.Set("Content-Type", "application/json")
	if tier != "" {
		req.Header.Set(middleware.HeaderTier, tier)
	}
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// jsonBody decodes the response body as a map for easy key access.
func jsonBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return m
}

// ── tests ─────────────────────────────────────────────────────────────────────

// TestTier_FreeWithTier2Blocked verifies that a free-tier request is rejected
// with HTTP 403 when the AI scanner is active, and that the response body
// contains the tier_upgrade_required error and upgrade URL.
func TestTier_FreeWithTier2Blocked(t *testing.T) {
	m := middleware.New(okHandler, true /* tier2Available */, newMockStore(), -1)
	rr := doRequest(m, "free", "key-abc")

	if rr.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", rr.Code)
	}
	body := jsonBody(t, rr)
	if body["error"] != "tier_upgrade_required" {
		t.Errorf("want error=tier_upgrade_required, got %v", body["error"])
	}
	if body["upgrade_url"] != "https://ocultar.dev/pricing" {
		t.Errorf("want upgrade_url in body, got %v", body["upgrade_url"])
	}
	if rr.Header().Get(middleware.HeaderTier) != "free" {
		t.Errorf("want X-Ocultar-Tier: free, got %q", rr.Header().Get(middleware.HeaderTier))
	}
	if rr.Header().Get(middleware.HeaderCallsRemaining) != "0" {
		t.Errorf("want X-Ocultar-Calls-Remaining: 0, got %q", rr.Header().Get(middleware.HeaderCallsRemaining))
	}
}

// TestTier_DeveloperWithinLimit verifies that a developer-tier request below
// the 50 000-call monthly cap is forwarded to the upstream (HTTP 200) and that
// the calls-remaining header reflects the decrement.
func TestTier_DeveloperWithinLimit(t *testing.T) {
	store := newMockStore()
	m := middleware.New(okHandler, true, store, -1)
	rr := doRequest(m, "developer", "key-dev-1")

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d — body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get(middleware.HeaderTier) != "developer" {
		t.Errorf("want X-Ocultar-Tier: developer, got %q", rr.Header().Get(middleware.HeaderTier))
	}
	remaining, err := strconv.ParseInt(rr.Header().Get(middleware.HeaderCallsRemaining), 10, 64)
	if err != nil {
		t.Fatalf("parse X-Ocultar-Calls-Remaining: %v", err)
	}
	// After the first call count is 1, so remaining = 50000 - 1 = 49999.
	if remaining != 49_999 {
		t.Errorf("want 49999 calls remaining, got %d", remaining)
	}
}

// TestTier_DeveloperOverLimit verifies that a developer-tier key that has
// exceeded its 50 000-call monthly cap receives HTTP 429 with a Retry-After
// header and no forwarded upstream call.
func TestTier_DeveloperOverLimit(t *testing.T) {
	store := newMockStore()
	// Simulate 50 000 calls already consumed; next Increment returns 50 001.
	// No auth header → HashAPIKey("") hashes the sentinel "anonymous".
	store.seed(middleware.HashAPIKey(""), currentMonth(), 50_000)

	m := middleware.New(okHandler, true, store, -1)
	rr := doRequest(m, "developer", "")

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d — body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Error("want Retry-After header on 429")
	}
	body := jsonBody(t, rr)
	if body["error"] != "monthly_limit_exceeded" {
		t.Errorf("want error=monthly_limit_exceeded, got %v", body["error"])
	}
	if rr.Header().Get(middleware.HeaderCallsRemaining) != "0" {
		t.Errorf("want X-Ocultar-Calls-Remaining: 0, got %q", rr.Header().Get(middleware.HeaderCallsRemaining))
	}
}

// TestTier_TeamOverLimit verifies that a team-tier key over 200 000 calls
// receives HTTP 429 with a Retry-After header.
func TestTier_TeamOverLimit(t *testing.T) {
	store := newMockStore()
	// Seed with 200 000 calls; Increment will return 200 001.
	store.seed(middleware.HashAPIKey(""), currentMonth(), 200_000)

	m := middleware.New(okHandler, true, store, -1)
	rr := doRequest(m, "team", "")

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Error("want Retry-After header on 429")
	}
	body := jsonBody(t, rr)
	if ml, ok := body["monthly_limit"].(float64); !ok || int64(ml) != 200_000 {
		t.Errorf("want monthly_limit=200000, got %v", body["monthly_limit"])
	}
}

// TestTier_EnterpriseUnlimited verifies that enterprise-tier requests always
// pass through regardless of call volume, and that the remaining header is
// set to "unlimited".
func TestTier_EnterpriseUnlimited(t *testing.T) {
	store := newMockStore()
	// Even a ludicrously high pre-existing count must not trigger 429.
	store.seed(middleware.HashAPIKey(""), currentMonth(), 10_000_000)

	m := middleware.New(okHandler, true, store, -1 /* no enterprise cap */)
	rr := doRequest(m, "enterprise", "")

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d — body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get(middleware.HeaderCallsRemaining) != "unlimited" {
		t.Errorf("want X-Ocultar-Calls-Remaining: unlimited, got %q",
			rr.Header().Get(middleware.HeaderCallsRemaining))
	}
}

// TestTier_EnterpriseConfigurableLimit verifies that when OCU_ENTERPRISE_LIMIT
// is passed as enterpriseLimit, it is enforced like any other tier cap.
func TestTier_EnterpriseConfigurableLimit(t *testing.T) {
	store := newMockStore()
	store.seed(middleware.HashAPIKey(""), currentMonth(), 1_000) // seed at cap

	// Enterprise with a 1 000-call limit.
	m := middleware.New(okHandler, true, store, 1_000)
	rr := doRequest(m, "enterprise", "")

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("want 429 for enterprise over configurable limit, got %d", rr.Code)
	}
}

// TestTier_MissingHeaderDefaultsFree verifies that a request with no
// X-Ocultar-Tier header is treated as free tier, and therefore blocked
// when Tier 2 is available.
func TestTier_MissingHeaderDefaultsFree(t *testing.T) {
	m := middleware.New(okHandler, true, newMockStore(), -1)
	rr := doRequest(m, "" /* no tier header */, "key-xyz")

	if rr.Code != http.StatusForbidden {
		t.Errorf("want 403 (free-tier default), got %d", rr.Code)
	}
	body := jsonBody(t, rr)
	if body["error"] != "tier_upgrade_required" {
		t.Errorf("want tier_upgrade_required, got %v", body["error"])
	}
}

// TestTier_InvalidHeaderReturns400 verifies that an unrecognised tier value
// is rejected with HTTP 400 and an informative error body.
func TestTier_InvalidHeaderReturns400(t *testing.T) {
	m := middleware.New(okHandler, true, newMockStore(), -1)
	rr := doRequest(m, "premium", "key-xyz") // "premium" is not a valid tier

	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
	body := jsonBody(t, rr)
	if body["error"] != "invalid_tier" {
		t.Errorf("want error=invalid_tier, got %v", body["error"])
	}
}

// TestTier_ResponseHeadersPresentOnAllPaths verifies that X-Ocultar-Tier and
// X-Ocultar-Calls-Remaining are set on every non-error response path.
func TestTier_ResponseHeadersPresentOnAllPaths(t *testing.T) {
	cases := []struct {
		tier string
		want int
	}{
		{"developer", http.StatusOK},
		{"team", http.StatusOK},
		{"enterprise", http.StatusOK},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.tier, func(t *testing.T) {
			m := middleware.New(okHandler, true, newMockStore(), -1)
			rr := doRequest(m, tc.tier, "key-1")
			if rr.Code != tc.want {
				t.Errorf("[%s] want %d, got %d", tc.tier, tc.want, rr.Code)
			}
			if rr.Header().Get(middleware.HeaderTier) == "" {
				t.Errorf("[%s] missing X-Ocultar-Tier header", tc.tier)
			}
			if rr.Header().Get(middleware.HeaderCallsRemaining) == "" {
				t.Errorf("[%s] missing X-Ocultar-Calls-Remaining header", tc.tier)
			}
		})
	}
}

// currentMonth returns the current month in "2006-01" format, matching the
// format used by TierMiddleware for usage tracking.
func currentMonth() string {
	return "2026-05" // fixed for deterministic seeding in tests
}
