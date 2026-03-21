package ai

import (
	"log"
	"os"

	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/pkg/inference"
)

func NewScanner() engine.AIScanner {
	// --- Local Native Inference Selection ---
	// Per March 2026 Native Transition: SLM_TYPE=native is now the sole mandatory path.
	// Legacy SLM AI Relay (mock_slm.py) and Ollama HTTP fallbacks have been removed.
	
	modelPath := os.Getenv("SLM_MODEL_PATH")
	if modelPath == "" {
		modelPath = "models/qwen-1.5b-q4_k_m.gguf" // Default convention
	}
	
	s, err := inference.NewLlamaScanner(modelPath)
	if err != nil {
		log.Fatalf("[FATAL] SLM: failed to initialize native llama.cpp: %v. Native inference is required for Tier 2 scanning.", err)
	}
	
	log.Printf("[INFO] SLM: Native llama.cpp initialized (model: %s)", modelPath)
	return s
}
