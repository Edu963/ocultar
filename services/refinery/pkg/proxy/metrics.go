package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request Metrics
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocultar_proxy_requests_total",
		Help: "Total number of requests processed by the proxy.",
	}, []string{"method", "status", "redacted"})

	RequestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ocultar_proxy_request_duration_seconds",
		Help:    "Latency of requests processed by the proxy.",
		Buckets: prometheus.DefBuckets,
	}, []string{"tier"})

	// PII Metrics
	PIIHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocultar_pii_hits_total",
		Help: "Total number of PII hits detected by the refinery.",
	}, []string{"type"})

	// Queue Metrics
	QueueLength = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ocultar_proxy_queue_length",
		Help: "Current number of requests waiting in the proxy queue.",
	})

	DroppedRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocultar_proxy_dropped_requests_total",
		Help: "Total number of requests dropped due to concurrency limits or timeouts.",
	}, []string{"reason"})

	// Vault Metrics
	VaultOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocultar_vault_operations_total",
		Help: "Total number of vault operations performed.",
	}, []string{"op", "result"})
)
