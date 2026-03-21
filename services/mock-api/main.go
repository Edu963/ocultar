package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type ScenarioData struct {
	Name         string  `json:"name"`
	Savings      float64 `json:"savings"`
	VaultEntries int     `json:"vault_entries"`
	BlockedPII   int     `json:"blocked_pii"`
	Status       string  `json:"status"`
}

var (
	currentScenario = ScenarioData{
		Name:         "default",
		Savings:      542000,
		VaultEntries: 54200,
		BlockedPII:   1200,
		Status:       "online",
	}
	mu sync.Mutex
)

func main() {
	http.HandleFunc("/api/roi", handleROI)
	http.HandleFunc("/api/scenario", handleScenario)
	http.HandleFunc("/api/status", handleStatus)

	fmt.Println("OCULTAR Production Mock API running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleROI(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(currentScenario)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	json.NewEncoder(w).Encode(map[string]string{"tier": "enterprise", "version": "1.2.0"})
}

func handleScenario(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newScenario struct {
		ID    string `json:"id"`
		Scale string `json:"scale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&newScenario); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	switch newScenario.ID {
	case "finance":
		currentScenario = ScenarioData{
			Name:         "Finance Breach Mitigation",
			Savings:      1250000,
			VaultEntries: 85000,
			BlockedPII:   45000,
			Status:       "active_mitigation",
		}
	case "healthcare":
		currentScenario = ScenarioData{
			Name:         "HIPAA Compliance Shield",
			Savings:      2100000,
			VaultEntries: 120000,
			BlockedPII:   15000,
			Status:       "active_mitigation",
		}
	default:
		currentScenario.Name = "Custom Scenario: " + newScenario.ID
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(currentScenario)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
