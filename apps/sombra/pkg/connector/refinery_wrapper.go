package connector

import (
	"context"
	"fmt"

	refinery_conn "github.com/Edu963/ocultar/pkg/connector"
)

// RefineryWrapper wraps an refinery-level connector to satisfy Sombra's Connector interface.
type RefineryWrapper struct {
	name   string
	inner  refinery_conn.Connector
	policy DataPolicy
}

// NewRefineryWrapper creates a new Sombra connector wrapping an refinery connector.
func NewRefineryWrapper(name string, inner refinery_conn.Connector, policy DataPolicy) *RefineryWrapper {
	return &RefineryWrapper{
		name:   name,
		inner:  inner,
		policy: policy,
	}
}

func (w *RefineryWrapper) Name() string {
	return w.name
}

func (w *RefineryWrapper) Policy() DataPolicy {
	return w.policy
}

func (w *RefineryWrapper) Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error) {
	// Convert Sombra's FetchRequest parameters to refinery's map[string]interface{}
	params := make(map[string]interface{})
	for k, v := range req.Parameters {
		params[k] = v
	}
	// Also include SourceID if present
	if req.SourceID != "" {
		params["source_id"] = req.SourceID
	}

	body, err := w.inner.Fetch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("refinery connector %q fetch failed: %w", w.inner.Type(), err)
	}

	return &FetchResponse{
		Body:        body,
		ContentType: "application/json", // Most refinery connectors return JSON
		Metadata: map[string]string{
			"connector_type": w.inner.Type(),
			"connector_id":   w.inner.ID(),
		},
	}, nil
}
