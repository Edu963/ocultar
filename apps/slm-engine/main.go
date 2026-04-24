package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Edu963/ocultar/apps/slm-engine/pkg/inference"
)

var scanner inference.Tier2Engine

func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	text := req["text"]
	if text == "" {
		json.NewEncoder(w).Encode(map[string][]string{})
		return
	}

	res, err := scanner.ScanForPII(text)
	if err != nil {
		http.Error(w, fmt.Sprintf("scan failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	var err error

	engine := os.Getenv("TIER2_ENGINE")
	switch engine {
	case "privacy-filter", "":
		// default
		sidecarURL := os.Getenv("PYTHON_SIDECAR_URL")
		if old := os.Getenv("PRIVACY_FILTER_URL"); old != "" {
			log.Println("[DEPRECATED] PRIVACY_FILTER_URL renamed to PYTHON_SIDECAR_URL. Please update your config.")
			sidecarURL = old
		}
		if sidecarURL == "" {
			log.Fatal("[FATAL] PYTHON_SIDECAR_URL not set")
		}
		modelPath := os.Getenv("PRIVACY_FILTER_MODEL_PATH")
		if modelPath == "" {
			modelPath = "openai/privacy-filter"
		}
		scanner, err = inference.NewPrivacyFilterEngine(sidecarURL, modelPath)
		if err != nil {
			log.Fatalf("failed to initialize privacy-filter scanner: %v", err)
		}
	case "llama-cpp":
		log.Fatal("[FATAL] llama-cpp engine not yet implemented. See TODO in scanner.go. Use TIER2_ENGINE=privacy-filter")
	default:
		log.Fatalf("[FATAL] Unknown TIER2_ENGINE: %s", engine)
	}
	defer scanner.Close()

	http.HandleFunc("/scan", handleScan)
	http.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}
	log.Printf("SLM sidecar running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
