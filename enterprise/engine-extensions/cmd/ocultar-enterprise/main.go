package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	enterpriseai "github.com/Edu963/ocultar-enterprise/pkg/ai"
	enterpriseaudit "github.com/Edu963/ocultar-enterprise/pkg/audit"
	enterpriseconfig "github.com/Edu963/ocultar-enterprise/pkg/config"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/pkg/vault"
	"golang.org/x/crypto/hkdf"
)

const VERSION = "1.14-enterprise"

const defaultSalt = "ocultar-v114-kdf-salt-fixed-16"

// getSalt retrieves the cryptographic salt from the environment or falls back to a default value.
func getSalt() string {
	if s := os.Getenv("OCU_SALT"); s != "" {
		return s
	}
	log.Printf("[WARN] OCU_SALT is not set — using built-in default salt. Set OCU_SALT in production.")
	return defaultSalt
}

// getMasterKey derives the 32-byte AES master key using HKDF based on the environment-provided key material.
func getMasterKey() []byte {
	keyMaterial := os.Getenv("OCU_MASTER_KEY")
	if keyMaterial == "" {
		log.Printf("[WARN] OCU_MASTER_KEY is not set — using insecure dev key. Never deploy this to production.")
		keyMaterial = "default-dev-key-32-chars-long-!!!"
	}

	salt := []byte(getSalt())
	info := []byte("ocultar-aes-key")

	r := hkdf.New(sha256.New, []byte(keyMaterial), salt, info)
	derived := make([]byte, 32)
	if _, err := io.ReadFull(r, derived); err != nil {
		log.Fatalf("[FATAL] HKDF key derivation failed: %v", err)
	}
	return derived
}

// main is the entry point for the OCULTAR Refinery Engine Enterprise CLI and HTTP server.
func main() {
	showVersion := flag.Bool("version", false, "Print the OCULTAR engine version and exit")
	showVersionShort := flag.Bool("v", false, "Print the OCULTAR engine version and exit (alias)")
	dryRun := flag.Bool("dry-run", false, "Scan for PII without writing to vault; output JSON report")
	report := flag.Bool("report", false, "Produce redacted output AND append a JSON PII report to stderr")
	serve := flag.String("serve", "", "Start local HTTP dashboard on given PORT (e.g. '8080')")
	flag.Parse()

	if *showVersion || *showVersionShort {
		fmt.Printf("OCULTAR Refinery Engine Enterprise v%s\n", VERSION)
		os.Exit(0)
	}

	masterKey := getMasterKey()

	vaultPath := os.Getenv("OCU_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = "vault.db"
	}
	if *dryRun {
		vaultPath = ""
	}

	license.Load()
	config.Load() // Loads community base

	if license.Active.Tier == "enterprise" {
		enterpriseconfig.Load() // Loads dynamic configs if active (including vault_backend)
	}

	pilotMode := license.Active.Tier == "community" || os.Getenv("OCU_PILOT_MODE") == "1" || os.Getenv("OCU_PILOT_MODE") == "true"
	if pilotMode {
		fmt.Printf("⚠️  OCULTAR running in COMMUNITY MODE. Vault size limits applied. ⚠️\n")
	}

	// Open vault using the provider selected by config (duckdb or postgres).
	vaultProvider, err := vault.New(config.Global, vaultPath)
	if err != nil {
		log.Fatal("Failed to open vault: ", err)
	}
	defer vaultProvider.Close()

	eng := engine.NewEngine(vaultProvider, masterKey)
	eng.DryRun = *dryRun
	eng.Report = *report
	eng.Serve = *serve
	eng.PilotMode = pilotMode

	// Inject Enterprise Modules
	if license.Active.Tier == "enterprise" {
		auditLogger := enterpriseaudit.NewLogger()
		if err := auditLogger.Init("audit.log"); err != nil {
			log.Printf("[WARN] Failed to initialize enterprise SIEM logger: %v", err)
		}
		defer auditLogger.Close()
		eng.AuditLogger = auditLogger

		aiScanner := enterpriseai.NewScanner()
		aiScanner.SetDomain(config.Global.DomainSnapshot)
		eng.AIScanner = aiScanner
	}

	if *serve != "" {
		startServer(eng, *serve)
		return
	}

	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("Failed to read input: ", err)
	}

	if len(inputData) == 0 {
		if *dryRun {
			printReport(eng, 0)
		}
		return
	}

	actor := "cli-user"
	var jsonRaw interface{}
	if err := json.Unmarshal(inputData, &jsonRaw); err == nil {
		inputData = nil
		refinedData, _ := eng.ProcessInterface(jsonRaw, actor)
		if !*dryRun {
			output, _ := json.MarshalIndent(refinedData, "", "    ")
			fmt.Println(string(output))
		}
	} else {
		lines := strings.Split(string(inputData), "\n")
		inputData = nil
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			refined, _ := eng.RefineString(line, actor, nil)
			if !*dryRun {
				fmt.Println(refined)
			}
		}
	}

	if *dryRun || *report {
		printReport(eng, 1)
	}
}

// printReport outputs the PII redaction metadata to standard error.
func printReport(eng *engine.Engine, filesScanned int) {
	rpt := eng.GenerateReport(filesScanned)
	out, _ := json.MarshalIndent(rpt, "", "  ")
	fmt.Fprintln(os.Stderr, string(out))
}

// VerifyDashboardIntegrity checks the SHA256 hashes of served assets against the manifest.
func VerifyDashboardIntegrity() error {
	path := "security/dashboard_integrity.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("integrity manifest missing: %v", err)
	}

	var manifest struct {
		Checksums map[string]string `json:"checksums"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %v", err)
	}

	for relPath, expectedHash := range manifest.Checksums {
		fullPath := "dist/enterprise/dashboard/" + relPath
		f, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("asset missing: %s", relPath)
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		actualHash := fmt.Sprintf("%x", h.Sum(nil))

		if actualHash != expectedHash {
			return fmt.Errorf("INTEGRITY_VIOLATION: %s (expected %s, got %s)", relPath, expectedHash, actualHash)
		}
	}

	log.Printf("[INFO] Dashboard integrity verified successfully (%d assets).", len(manifest.Checksums))
	return nil
}

// startServer initializes and starts the local HTTP dashboard and API endpoints.
func startServer(eng *engine.Engine, servePort string) {
	if license.Active.Tier == "enterprise" {
		if err := VerifyDashboardIntegrity(); err != nil {
			log.Fatalf("[FATAL] Dashboard integrity check failed: %v", err)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		baseDir := "apps/web/dist"
		if license.Active.Tier == "enterprise" {
			baseDir = "dist/enterprise/dashboard"
		}

		// Static Asset Routing
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			http.ServeFile(w, r, baseDir+r.URL.Path)
			return
		}
		if r.URL.Path == "/vite.svg" {
			http.ServeFile(w, r, baseDir+"/vite.svg")
			return
		}

		// SPA Entrypoint
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, baseDir+"/index.html")
			return
		}

		if r.URL.Path == "/roi_calc.html" {
			http.ServeFile(w, r, "apps/web/roi_calc.html")
			return
		}

		if !strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
	})

	http.HandleFunc("/api/refine", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		eng.ResetHits()

		inputData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var refinedOutput string
		var jsonRaw interface{}
		actor := r.RemoteAddr

		if err := json.Unmarshal(inputData, &jsonRaw); err == nil {
			refinedData, _ := eng.ProcessInterface(jsonRaw, actor)
			outBytes, _ := json.MarshalIndent(refinedData, "", "    ")
			refinedOutput = string(outBytes)
		} else {
			var refinedLines []string
			for _, line := range strings.Split(string(inputData), "\n") {
				if strings.TrimSpace(line) == "" {
					refinedLines = append(refinedLines, line)
					continue
				}
				refined, _ := eng.RefineString(line, actor, nil)
				refinedLines = append(refinedLines, refined)
			}
			refinedOutput = strings.Join(refinedLines, "\n")
		}

		rpt := eng.GenerateReport(1)

		response := map[string]interface{}{
			"refined": refinedOutput,
			"report":  rpt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vaultEntries := int64(0)
		if eng.Vault != nil {
			vaultEntries = eng.Vault.CountAll()
		}

		slmStatus := "Offline"
		if eng.AIScanner != nil && eng.AIScanner.IsAvailable() {
			slmStatus = "Online"
		}

		slmHost := os.Getenv("SLM_HOST")
		if slmHost == "" {
			slmHost = "http://localhost:8080"
		}

		response := map[string]interface{}{
			"vault_entries": vaultEntries,
			"slm_status":    slmStatus,
			"slm_host":      slmHost,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		response := map[string]interface{}{
			"tier": license.Active.Tier,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/vault", func(w http.ResponseWriter, r *http.Request) {
		if license.Active.Tier != "enterprise" {
			http.Error(w, "Enterprise license required", http.StatusForbidden)
			return
		}
		vaultEntries := int64(0)
		if eng.Vault != nil {
			vaultEntries = eng.Vault.CountAll()
		}
		response := map[string]interface{}{
			"status":  "online",
			"entries": vaultEntries,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/roi", func(w http.ResponseWriter, r *http.Request) {
		if license.Active.Tier != "enterprise" {
			http.Error(w, "Enterprise license required", http.StatusForbidden)
			return
		}
		vaultEntries := int64(0)
		if eng.Vault != nil {
			vaultEntries = eng.Vault.CountAll()
		}
		// Calculate ROI: Example €10 per scrubbed entity
		savings := vaultEntries * 10
		response := map[string]interface{}{
			"status":  "online",
			"savings": savings,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/refine/file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file upload", http.StatusBadRequest)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=cleaned_%s", handler.Filename))

		eng.ResetHits()
		actor := r.RemoteAddr
		if strings.HasSuffix(strings.ToLower(handler.Filename), ".json") {
			w.Header().Set("Content-Type", "application/json")
			var data interface{}
			if err := json.NewDecoder(file).Decode(&data); err != nil {
				http.Error(w, "Invalid JSON file", http.StatusBadRequest)
				return
			}
			refinedData, _ := eng.ProcessInterface(data, actor)
			json.NewEncoder(w).Encode(refinedData)
			return
		}

		if strings.HasSuffix(strings.ToLower(handler.Filename), ".csv") {
			w.Header().Set("Content-Type", "text/csv")
			reader := csv.NewReader(file)
			reader.FieldsPerRecord = -1
			writer := csv.NewWriter(w)
			defer writer.Flush()

			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error reading CSV record: %v", err)
					continue
				}

				// Parallelize field processing within the record using RefineBatch
				items := make([]interface{}, len(record))
				for i, field := range record {
					items[i] = field
				}
				refinedItems, err := eng.RefineBatch(items, actor)
				if err != nil {
					log.Printf("RefineBatch error: %v", err)
					refinedItems = items
				}

				refinedRecord := make([]string, len(refinedItems))
				for i, ri := range refinedItems {
					refinedRecord[i] = ri.(string)
				}
				writer.Write(refinedRecord)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				fmt.Fprintln(w, line)
				continue
			}
			refined, _ := eng.RefineString(line, actor, nil)
			fmt.Fprintln(w, refined)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error scanning file: %v", err)
		}
	})

	port := servePort
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	fmt.Printf("OCULTAR REST API running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
