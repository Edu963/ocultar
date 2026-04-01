package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
)

func runCommandHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmdStr := r.FormValue("cmd")
	if cmdStr == "" {
		http.Error(w, "Command required", http.StatusBadRequest)
		return
	}

	log.Printf("Executing: %s", cmdStr)

	cmd := exec.Command("bash", "-c", cmdStr)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")

	multi := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(multi)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	cmd.Wait()
}

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/run", runCommandHandler)

	log.Println("Demo Bridge running on http://localhost:9999")
	log.Fatal(http.ListenAndServe(":9999", nil))
}
