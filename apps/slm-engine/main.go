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
	endpoint := os.Getenv("PRIVACY_FILTER_URL")
	if endpoint == "" {
		log.Fatal("[FATAL] PRIVACY_FILTER_URL is required for SLM sidecar")
	}

	var err error
	scanner, err = inference.NewPrivacyFilterEngine(endpoint)
	if err != nil {
		log.Fatalf("failed to initialize privacy-filter scanner: %v", err)
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
