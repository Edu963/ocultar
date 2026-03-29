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

	"github.com/Edu963/ocultar/pkg/audit"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/connector"
	_ "github.com/Edu963/ocultar/pkg/connector/slack"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/identities"
	"github.com/Edu963/ocultar/pkg/inference"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/pkg/recon"
	"github.com/Edu963/ocultar/pkg/reporter"
	"github.com/Edu963/ocultar/vault"
	"golang.org/x/crypto/hkdf"
)

const VERSION = "1.14"

const defaultSalt = "ocultar-v112-kdf-salt-fixed-16"

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

// main is the entry point for the OCULTAR Refinery Refinery CLI and HTTP server.
func main() {
	showVersion := flag.Bool("version", false, "Print the OCULTAR refinery version and exit")
	showVersionShort := flag.Bool("v", false, "Print the OCULTAR refinery version and exit (alias)")
	dryRun := flag.Bool("dry-run", false, "Scan for PII without writing to vault; output JSON report")
	report := flag.Bool("report", false, "Produce redacted output AND append a JSON PII report to stderr")
	serve := flag.String("serve", "", "Start local HTTP dashboard on given PORT (e.g. '8080')")
	complianceReport := flag.String("compliance-report", "", "Generate HTML compliance report from the given audit log file")
	complianceOutput := flag.String("report-out", "report.html", "Output path for the HTML compliance report")
	reconPath := flag.String("recon", "", "Run the Data Recon Crawler on a local directory")
	flag.Parse()

	if *showVersion || *showVersionShort {
		fmt.Printf("OCULTAR Refinery Refinery v%s\n", VERSION)
		os.Exit(0)
	}

	if *complianceReport != "" {
		rep := reporter.New()
		err := rep.GenerateHTMLReport(*complianceReport, *complianceOutput)
		if err != nil {
			log.Fatalf("Failed to generate compliance report: %v", err)
		}
		fmt.Printf("Successfully generated compliance report at: %s\n", *complianceOutput)
		os.Exit(0)
	}

	masterKey := getMasterKey()

	// Load license and config before opening the vault so that
	// VaultBackend / PostgresDSN from config.yaml are available.
	license.Load()
	config.Load()

	// Determine the vault path (DuckDB only; ignored for postgres backend).
	vaultPath := os.Getenv("OCU_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = "vault.db"
	}
	if *dryRun {
		vaultPath = ":memory:"
	}

	// Open the vault using the provider selected by configuration.
	vaultProvider, err := vault.New(config.Global, vaultPath)
	if err != nil {
		log.Fatal("Failed to open vault: ", err)
	}
	defer vaultProvider.Close()

	pilotMode := (license.Active.Tier == "community" && os.Getenv("OCU_FORCE_ENTERPRISE") != "true") || (os.Getenv("OCU_PILOT_MODE") == "1" || os.Getenv("OCU_PILOT_MODE") == "true")
	if pilotMode {
		fmt.Printf("⚠️  OCULTAR running in COMMUNITY MODE. ⚠️\n")
	} else {
		fmt.Printf("🚀 OCULTAR ENTERPRISE EDITION 🚀\n")
	}

	eng := refinery.NewRefinery(vaultProvider, masterKey)
	eng.DryRun = *dryRun
	eng.Report = *report
	eng.PilotMode = pilotMode

	// Enable Tier 2 AI if forced or licensed
	if !pilotMode {
		modelPath := os.Getenv("SLM_MODEL_PATH")
		if modelPath != "" {
			scanner, err := inference.NewLlamaScanner(modelPath)
			if err != nil {
				log.Printf("[WARN] Failed to initialize SLM scanner: %v", err)
			} else {
				eng.SetAIScanner(scanner)
				log.Printf("[INFO] Tier 2 AI (llama.cpp) active. Model: %s", modelPath)
			}
		}
	}
	eng.AIScanner.SetDomain(config.Global.DomainSnapshot)

	if *reconPath != "" {
		c := recon.NewCrawler(eng)
		heatmap, err := c.CrawlLocalDirectory(*reconPath)
		if err != nil {
			log.Fatalf("[FATAL] Recon Crawler failed: %v", err)
		}
		fmt.Println(heatmap.ToJSON())
		os.Exit(0)
	}

	// Initialize and start connectors (Task #18, #19 & #20)
	cm := connector.NewManager(eng)
	if os.Getenv("SLACK_TOKEN") != "" {
		if license.HasProConnector(license.CapProSlack) {
			slackCfg := map[string]interface{}{
				"id":           "slack-default",
				"token":        os.Getenv("SLACK_TOKEN"),
				"workspace_id": os.Getenv("SLACK_WORKSPACE_ID"),
			}
			if err := cm.LoadAndStart("slack-default", "slack", slackCfg); err != nil {
				log.Printf("[ERROR] Failed to start Slack connector: %v", err)
			}
		} else {
			log.Printf("[WARN] Slack Connector requires a Pro/Enterprise license bitmask. Skipping.")
		}
	}

	if os.Getenv("MS_CLIENT_ID") != "" {
		if license.HasProConnector(license.CapProSharePoint) {
			spCfg := map[string]interface{}{
				"id":            "sharepoint-default",
				"tenant_id":     os.Getenv("MS_TENANT_ID"),
				"client_id":     os.Getenv("MS_CLIENT_ID"),
				"client_secret": os.Getenv("MS_CLIENT_SECRET"),
				"site_id":       os.Getenv("MS_SHAREPOINT_SITE_ID"),
			}
			if err := cm.LoadAndStart("sharepoint-default", "sharepoint-graph", spCfg); err != nil {
				log.Printf("[ERROR] Failed to start SharePoint connector: %v", err)
			}
		} else {
			log.Printf("[WARN] SharePoint Connector requires a Pro/Enterprise license bitmask. Skipping.")
		}
	}

	// Phase B: Start Live CRM/LDAP sync daemon
	identities.StartSyncWorker()

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
		refinedData, err := eng.ProcessInterface(jsonRaw, actor)
		if err != nil {
			log.Fatalf("Refinery failure: %v", err)
		}
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
			refined, err := eng.RefineString(line, actor, nil)
			if err != nil {
				log.Fatalf("Refinery failure: %v", err)
			}
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
func printReport(eng *refinery.Refinery, filesScanned int) {
	rpt := eng.GenerateReport(filesScanned)
	out, _ := json.MarshalIndent(rpt, "", "  ")
	fmt.Fprintln(os.Stderr, string(out))
}

// startServer initializes and starts the local HTTP dashboard and API endpoints.
func startServer(eng *refinery.Refinery, servePort string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index_v2.html" {
			http.ServeFile(w, r, "index_v2.html")
			return
		}
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config.Global)
	})

	http.HandleFunc("/api/config/regex", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var rule config.RegexRule
			if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			config.AddRegexRule(rule)
			config.Save()
			w.WriteHeader(http.StatusCreated)
		} else if r.Method == http.MethodDelete {
			var payload struct {
				Type string `json:"type"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				config.RemoveRegexRule(payload.Type)
				config.Save()
			}
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/config/dictionary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var payload struct {
				Type string `json:"type"`
				Term string `json:"term"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				config.AddDictionaryTerm(payload.Type, payload.Term)
				config.Save()
				w.WriteHeader(http.StatusCreated)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/vault/migrate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			DSN string `json:"dsn"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if err := vault.MigrateDuckDBtoPostgres(eng.Vault, payload.DSN); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	http.HandleFunc("/api/audit/risk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Dataset             []map[string]interface{} `json:"dataset"`
			QuasiIdentifiers    []string                 `json:"quasi_identifiers"`
			SensitiveAttributes []string                 `json:"sensitive_attributes"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		report := audit.AnalyzeDatasetRisk(req.Dataset, req.QuasiIdentifiers, req.SensitiveAttributes)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	})

	http.HandleFunc("/api/reveal", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		auditorToken := os.Getenv("OCU_AUDITOR_TOKEN")
		if auditorToken == "" {
			http.Error(w, "Unauthorized: OCU_AUDITOR_TOKEN is not configured on this server.", http.StatusForbidden)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+auditorToken {
			eng.AuditLogger.Log("UNKNOWN", "failed_reveal_auth", "N/A", "401 Unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var payload struct {
			Tokens []string `json:"tokens"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		results := make(map[string]string)
		for _, t := range payload.Tokens {
			decrypted, err := refinery.DecryptToken(eng.Vault, eng.MasterKey, t)
			if err == nil && decrypted != t {
				results[t] = decrypted
				eng.AuditLogger.Log("auditor", "revealed", t, "N/A")
			} else {
				results[t] = "ERR_NOT_FOUND"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": results,
		})
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
			refinedData, err := eng.ProcessInterface(jsonRaw, actor)
			if err != nil {
				log.Printf("Refinery error: %v", err)
				http.Error(w, "Ocultar Refinery: internal refinery error", http.StatusInternalServerError)
				return
			}
			outBytes, _ := json.MarshalIndent(refinedData, "", "    ")
			refinedOutput = string(outBytes)
		} else {
			var refinedLines []string
			for _, line := range strings.Split(string(inputData), "\n") {
				if strings.TrimSpace(line) == "" {
					refinedLines = append(refinedLines, line)
					continue
				}
				refined, err := eng.RefineString(line, actor, nil)
				if err != nil {
					log.Printf("Refinery error: %v", err)
					http.Error(w, "Ocultar Refinery: internal refinery error", http.StatusInternalServerError)
					return
				}
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
			refinedData, err := eng.ProcessInterface(data, actor)
			if err != nil {
				log.Printf("Refinery error JSON: %v", err)
				http.Error(w, "Ocultar Refinery: internal refinery error", http.StatusInternalServerError)
				return
			}
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

				var refinedRecord []string
				for _, field := range record {
					if strings.TrimSpace(field) == "" {
						refinedRecord = append(refinedRecord, field)
					} else {
						refined, err := eng.RefineString(field, actor, nil)
						if err != nil {
							log.Printf("Refinery error CSV: %v", err)
							http.Error(w, "Ocultar Refinery: internal refinery error", http.StatusInternalServerError)
							return
						}
						refinedRecord = append(refinedRecord, refined)
					}
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
			refined, err := eng.RefineString(line, actor, nil)
			if err != nil {
				log.Printf("Refinery error JSONL: %v", err)
				http.Error(w, "Ocultar Refinery: internal refinery error", http.StatusInternalServerError)
				return
			}
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
