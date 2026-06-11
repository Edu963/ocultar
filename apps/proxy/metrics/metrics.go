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
	// RequestsTotal counts every request processed by the proxy, labelled by endpoint.
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocultar_requests_total",
		Help: "Total requests processed by the proxy, by endpoint.",
	}, []string{"endpoint"})
)

// IncRequest records one request for the given endpoint label.
func IncRequest(endpoint string) {
	RequestsTotal.WithLabelValues(endpoint).Inc()
}
