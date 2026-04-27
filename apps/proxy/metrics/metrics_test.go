package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/Edu963/ocultar-proxy/metrics"
	"github.com/Edu963/ocultar-proxy/middleware"
)

// okHandler is a trivial upstream that always returns HTTP 200.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func doRequest(h http.Handler, tier, auth string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/refine", strings.NewReader(`{}`))
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

// mockStore satisfies middleware.UsageStore with zero counts (no limits hit).
type mockStore struct{}

func (mockStore) Increment(_, _ string) (int64, error) { return 1, nil }
func (mockStore) Count(_, _ string) (int64, error)     { return 0, nil }

// TestAPICallsTotal_Increments verifies that a successful developer-tier request
// increments ocultar_api_calls_total{tier="developer"} by exactly 1.
func TestAPICallsTotal_Increments(t *testing.T) {
	m := middleware.New(okHandler, false, mockStore{}, -1)

	before := testutil.ToFloat64(metrics.APICallsTotal.WithLabelValues("developer"))
	doRequest(m, "developer", "key-1")
	after := testutil.ToFloat64(metrics.APICallsTotal.WithLabelValues("developer"))

	if after-before != 1 {
		t.Errorf("want APICallsTotal delta=1, got %.0f", after-before)
	}
}

// TestAPICallsTotal_ByTier confirms each tier label is tracked independently.
func TestAPICallsTotal_ByTier(t *testing.T) {
	m := middleware.New(okHandler, false, mockStore{}, -1)

	tiers := []string{"developer", "team", "enterprise"}
	befores := make(map[string]float64, len(tiers))
	for _, tier := range tiers {
		befores[tier] = testutil.ToFloat64(metrics.APICallsTotal.WithLabelValues(tier))
	}

	for _, tier := range tiers {
		doRequest(m, tier, "key-2")
	}

	for _, tier := range tiers {
		after := testutil.ToFloat64(metrics.APICallsTotal.WithLabelValues(tier))
		if after-befores[tier] != 1 {
			t.Errorf("[%s] want delta=1, got %.0f", tier, after-befores[tier])
		}
	}
}

// TestTierUpgradeRequired_Increments verifies that a free-tier block (when
// tier2Available=true) increments ocultar_tier_upgrade_required_total.
func TestTierUpgradeRequired_Increments(t *testing.T) {
	m := middleware.New(okHandler, true /* tier2Available */, mockStore{}, -1)

	before := testutil.ToFloat64(metrics.TierUpgradeRequiredTotal)
	rr := doRequest(m, "free", "key-3")
	after := testutil.ToFloat64(metrics.TierUpgradeRequiredTotal)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rr.Code)
	}
	if after-before != 1 {
		t.Errorf("want TierUpgradeRequiredTotal delta=1, got %.0f", after-before)
	}
}

// TestTierUpgradeRequired_NotIncrementedOnPass confirms the counter stays flat
// when a non-free request passes through normally.
func TestTierUpgradeRequired_NotIncrementedOnPass(t *testing.T) {
	m := middleware.New(okHandler, true, mockStore{}, -1)

	before := testutil.ToFloat64(metrics.TierUpgradeRequiredTotal)
	doRequest(m, "developer", "key-4")
	after := testutil.ToFloat64(metrics.TierUpgradeRequiredTotal)

	if after != before {
		t.Errorf("want no increment on passing request, got delta=%.0f", after-before)
	}
}

// TestAPICallsTotal_NotIncrementedOnBlock confirms that a blocked free-tier
// request does NOT increment APICallsTotal (it never reaches the upstream).
func TestAPICallsTotal_NotIncrementedOnBlock(t *testing.T) {
	m := middleware.New(okHandler, true, mockStore{}, -1)

	before := testutil.ToFloat64(metrics.APICallsTotal.WithLabelValues("free"))
	doRequest(m, "free", "key-5")
	after := testutil.ToFloat64(metrics.APICallsTotal.WithLabelValues("free"))

	if after != before {
		t.Errorf("want no APICallsTotal increment on blocked request, got delta=%.0f", after-before)
	}
}
