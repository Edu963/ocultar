package refinery

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Tier Latency Metrics
	TierLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ocultar_refinery_tier_duration_seconds",
		Help:    "Latency of individual PII detection tiers.",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"tier"})

	// PII Detection Metrics
	DetectionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocultar_refinery_detections_total",
		Help: "Total number of PII detections per entity type and tier.",
	}, []string{"entity", "tier"})

	// Evaluation Metrics
	ProcessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ocultar_refinery_process_total",
		Help: "Total number of strings processed by the refinery.",
	})
)
