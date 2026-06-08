// Package metrics registers the Ocultar proxy-layer Prometheus counters.
// Refinery-internal metrics (latency, PII hits, vault ops) live in
// services/refinery/pkg/proxy/metrics.go and are exported from the same
// default registry, so they appear together at /metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// APICallsTotal counts every request that reaches the tier middleware,
	// labelled by the resolved tier. Use this to track revenue-weighted traffic.
	APICallsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocultar_api_calls_total",
		Help: "Total API calls processed, by tier.",
	}, []string{"tier"})

	// TierUpgradeRequiredTotal counts free-tier requests blocked because the
	// Tier 2 AI scanner is active. A rising rate signals conversion opportunity.
	TierUpgradeRequiredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ocultar_tier_upgrade_required_total",
		Help: "Free-tier requests blocked because Tier 2 AI is active.",
	})
)

// IncAPICall records one call for the given tier label.
func IncAPICall(tier string) {
	APICallsTotal.WithLabelValues(tier).Inc()
}

// IncTierUpgradeRequired records one free-tier block event.
func IncTierUpgradeRequired() {
	TierUpgradeRequiredTotal.Inc()
}
