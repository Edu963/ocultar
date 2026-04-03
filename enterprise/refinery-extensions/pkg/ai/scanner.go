package ai

import (
	"os"

	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/inference"
)

func NewScanner() refinery.AIScanner {
	// --- Local Native Inference Selection ---
	// Per March 2026 Native Transition: SLM_TYPE=native is now the sole mandatory path.
	// Legacy SLM AI Relay (mock_slm.py) and Ollama HTTP fallbacks have been removed.
	
	sidecarURL := os.Getenv("SLM_SIDECAR_URL")
	if sidecarURL == "" {
		sidecarURL = "http://localhost:8085"
	}
	return inference.NewRemoteScanner(sidecarURL)
}
