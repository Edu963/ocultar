package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type InputDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type CommandDef struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Command     string     `json:"command"`
	Args        []string   `json:"args"`
	Inputs      []InputDef `json:"inputs"`
}

type CommandsConfig struct {
	Commands map[string]CommandDef `json:"commands"`
}

var config CommandsConfig

type RunHistory struct {
	RunID     string    `json:"run_id"`
	CommandID string    `json:"command_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"` // "running", "success", "failure"
	Log       string    `json:"log"`
}

var (
	historyList []RunHistory
	historyMu   sync.Mutex
)

func loadConfig() error {
	data, err := os.ReadFile("../../configs/automation_commands.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	log.Printf("Loaded %d commands from config", len(config.Commands))
	return nil
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func commandsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var cmdList []CommandDef
	for _, cmd := range config.Commands {
		cmdList = append(cmdList, cmd)
	}
	json.NewEncoder(w).Encode(cmdList)
}

type RunRequest struct {
	CommandID string            `json:"command_id"`
	Inputs    map[string]string `json:"inputs"`
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cmdDef, exists := config.Commands[req.CommandID]
	if !exists {
		http.Error(w, "Command not found", http.StatusNotFound)
		return
	}

	// Validate inputs
	envVars := os.Environ()
	for _, inDef := range cmdDef.Inputs {
		val, ok := req.Inputs[inDef.Name]
		if inDef.Required && (!ok || val == "") {
			http.Error(w, fmt.Sprintf("Missing required input: %s", inDef.Name), http.StatusBadRequest)
			return
		}
		if ok && val != "" && strings.HasPrefix(inDef.Name, "OCU_") {
			envVars = append(envVars, fmt.Sprintf("%s=%s", inDef.Name, val))
		}
	}

	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())

	historyMu.Lock()
	history := RunHistory{
		RunID:     runID,
		CommandID: cmdDef.ID,
		StartTime: time.Now(),
		Status:    "running",
	}
	historyList = append([]RunHistory{history}, historyList...)
	if len(historyList) > 20 {
		historyList = historyList[:20]
	}
	historyMu.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	flusher, _ := w.(http.Flusher)

	// Security: execute exactly what's in the whitelist
	cmd := exec.Command(cmdDef.Command, cmdDef.Args...)
	cmd.Dir = "/home/edu/ocultar" // Enforce working directory
	cmd.Env = envVars

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(w, "Failed to start command: %v\n", err)
		updateHistory(runID, "failure", fmt.Sprintf("Failed to start command: %v", err))
		return
	}

	multi := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(multi)
	
	var fullLog bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		fullLog.WriteString(line + "\n")
		if flusher != nil {
			flusher.Flush()
		}
	}

	err := cmd.Wait()
	status := "success"
	if err != nil {
		status = "failure"
		fmt.Fprintf(w, "\nCommand exited with error: %v\n", err)
		fullLog.WriteString(fmt.Sprintf("\nCommand exited with error: %v\n", err))
		if flusher != nil {
			flusher.Flush()
		}
	}

	updateHistory(runID, status, fullLog.String())
}

func updateHistory(runID, status, logStr string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	for i, h := range historyList {
		if h.RunID == runID {
			historyList[i].Status = status
			historyList[i].EndTime = time.Now()
			historyList[i].Log = logStr
			break
		}
	}
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	historyMu.Lock()
	defer historyMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(historyList)
}

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load automation config: %v", err)
	}

	http.HandleFunc("/api/automation/commands", corsMiddleware(commandsHandler))
	http.HandleFunc("/api/automation/run", corsMiddleware(runHandler))
	http.HandleFunc("/api/automation/history", corsMiddleware(historyHandler))

	port := "18081"
	log.Printf("CLI Bridge Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
