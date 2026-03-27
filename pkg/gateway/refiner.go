package gateway

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Edu963/ocultar/pkg/refinery"
)

// Refiner bridges the Sombra Gateway to the core OCULTAR Engine.
type Refiner struct {
	// Instead of a remote Endpoint/APIKey, we embed the core Engine directly
	// to ensure zero-latency, local sanitization before egress.
	Engine *refinery.Refinery
}

// Refine processes the payload through OCULTAR's multi-tiered sanitization layer.
func (r *Refiner) Refine(payload string) (string, error) {
	// 1. Enforce "Fail-Closed" - Block if the engine is disconnected
	if r.Engine == nil {
		return "", errors.New("CRITICAL: OCULTAR engine not configured. Blocking request to prevent egress")
	}

	// 2. Prepare the payload for the engine (handle both JSON and raw strings)
	var data interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		// If it's not JSON, treat it as a standard string
		data = payload
	}

	// 3. Route through the Engine's deep traversal logic
	// Note: We use "sombra-gateway" as a temporary actor until we fix handler.go
	refinedData, err := r.Engine.ProcessInterface(data, "sombra-gateway")
	if err != nil {
		// 4. Enforce "Fail-Closed" - Any engine failure terminates the request
		return "", fmt.Errorf("refinery processing failed, blocking payload: %w", err)
	}

	// 5. Package the refined data back into a string for the external API
	switch v := refinedData.(type) {
	case string:
		return v, nil
	default:
		refinedBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal refined payload, blocking egress: %w", err)
		}
		return string(refinedBytes), nil
	}
}
