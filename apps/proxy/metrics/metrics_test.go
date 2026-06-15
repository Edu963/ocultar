package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/ocultar-dev/ocultar-proxy/metrics"
)

func TestRequestsTotal_Increments(t *testing.T) {
	before := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("/v1/refine"))
	metrics.IncRequest("/v1/refine")
	after := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("/v1/refine"))

	if after-before != 1 {
		t.Errorf("want RequestsTotal delta=1, got %.0f", after-before)
	}
}

func TestRequestsTotal_ByEndpoint(t *testing.T) {
	endpoints := []string{"/v1/refine", "/v1/reveal", "/api/health"}
	befores := make(map[string]float64, len(endpoints))
	for _, ep := range endpoints {
		befores[ep] = testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues(ep))
	}

	for _, ep := range endpoints {
		metrics.IncRequest(ep)
	}

	for _, ep := range endpoints {
		after := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues(ep))
		if after-befores[ep] != 1 {
			t.Errorf("[%s] want delta=1, got %.0f", ep, after-befores[ep])
		}
	}
}
