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
	"path/filepath"
	"strings"
	"time"

	htmltmpl "html/template"
	texttmpl "text/template"

	"github.com/Edu963/ocultar/pkg/audit"
	"github.com/google/uuid"

	enterpriseai "github.com/Edu963/ocultar-enterprise/pkg/ai"
	enterpriseaudit "github.com/Edu963/ocultar-enterprise/pkg/audit"
	enterpriseconfig "github.com/Edu963/ocultar-enterprise/pkg/config"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/vault"
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

// main is the entry point for the OCULTAR Refinery Refinery Enterprise CLI and HTTP server.
func main() {
	showVersion := flag.Bool("version", false, "Print the OCULTAR refinery version and exit")
	showVersionShort := flag.Bool("v", false, "Print the OCULTAR refinery version and exit (alias)")
	dryRun := flag.Bool("dry-run", false, "Scan for PII without writing to vault; output JSON report")
	report := flag.Bool("report", false, "Produce redacted output AND append a JSON PII report to stderr")
	serve := flag.String("serve", "", "Start local HTTP dashboard on given PORT (e.g. '8080')")
	flag.Parse()

	if *showVersion || *showVersionShort {
		fmt.Printf("OCULTAR Refinery Refinery Enterprise v%s\n", VERSION)
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

	eng := refinery.NewRefinery(vaultProvider, masterKey)
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
func printReport(eng *refinery.Refinery, filesScanned int) {
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

	baseDir := "dist/enterprise/dashboard"
	if _, err := os.Stat("dashboard"); err == nil {
		baseDir = "dashboard"
	}

	for relPath, expectedHash := range manifest.Checksums {
		fullPath := baseDir + "/" + relPath
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

// pilotDataDir returns the absolute path to the pilot_data directory,
// anchored to the current working directory where the binary was launched.
func pilotDataDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "pilot_data"
	}
	return filepath.Join(cwd, "pilot_data")
}

// startServer initializes and starts the local HTTP dashboard and API endpoints.
func startServer(eng *refinery.Refinery, servePort string) {
	if license.Active.Tier == "enterprise" {
		if err := VerifyDashboardIntegrity(); err != nil {
			if eng.PilotMode || eng.DryRun {
				log.Printf("[WARN] Dashboard integrity check failed: %v. Proceeding in Demo/Pilot mode (Safety Bypass ACTIVE).", err)
			} else {
				log.Fatalf("[FATAL] Dashboard integrity check failed: %v", err)
			}
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		baseDir := "apps/web/dist"
		if license.Active.Tier == "enterprise" {
			baseDir = "dist/enterprise/dashboard"
			if _, err := os.Stat("dashboard"); err == nil {
				baseDir = "dashboard"
			}
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

	http.HandleFunc("/api/audit/risk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "active",
				"k_anonymity_threshold": 3,
				"l_diversity_threshold": 2,
				"description": "Risk compliance radar monitoring dataset guarantees.",
				"regulatory_policy": config.Global.RegulatoryPolicy,
			})
			return
		}

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

	http.HandleFunc("/api/pilot/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(10 << 20) // 10MB limit
		if err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("dataset")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Save to uploads
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
		dstPath := filepath.Join(pilotDataDir(), "uploads", filename)
		os.MkdirAll(filepath.Join(pilotDataDir(), "uploads"), 0755)
		dst, _ := os.Create(dstPath)
		if dst != nil {
			io.Copy(dst, file)
			dst.Close()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"filename": filename, "original_name": handler.Filename})
	})

	http.HandleFunc("/api/pilot/riskreport", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DatasetPath string `json:"dataset_path"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if req.DatasetPath == "" {
			http.Error(w, "dataset_path is required", http.StatusBadRequest)
			return
		}

		// Robust file lookup for demo/pilot environments
		datasetFile := req.DatasetPath
		if _, err := os.Stat(datasetFile); os.IsNotExist(err) {
			// Try root-relative if running from services/refinery
			altPath := filepath.Join("../../", req.DatasetPath)
			if _, err := os.Stat(altPath); err == nil {
				datasetFile = altPath
			} else {
				// Try one level up just in case
				altPath = filepath.Join("../", req.DatasetPath)
				if _, err := os.Stat(altPath); err == nil {
					datasetFile = altPath
				}
			}
		}

		// Read dataset from disk
		data, err := os.ReadFile(datasetFile)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read dataset: %v (Checked: %s and alternates)", err, req.DatasetPath), http.StatusInternalServerError)
			return
		}

		var dataset []map[string]interface{}
		if err := json.Unmarshal(data, &dataset); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse JSON: %v", err), http.StatusInternalServerError)
			return
		}

		// Defaults for demo pilot
		qi := []string{"region", "dept"}
		sa := []string{"name", "iban", "email"}

		report := audit.AnalyzeDatasetRisk(dataset, qi, sa)
		
		// Map to full report for template generation
		reportID := strings.ToUpper(uuid.New().String()[:8])
		meta := reportMeta{
			ReportID:           reportID,
			GeneratedAt:        time.Now().UTC().Format("02 January 2006, 15:04 UTC"),
			DatasetScope:       datasetFile,
			MethodologyVersion: reportVersion,
			EngineVersion:      engineVersion,
			TotalRecords:       len(dataset),
		}
		before, after := buildScenarios(report)
		fullRpt := fullReport{Meta: meta, Risk: report, Before: before, After: after}

		// Ensure reports dir exists
		os.MkdirAll(filepath.Join(pilotDataDir(), "reports"), 0755)

		// Generate on-disk Markdown
		mdTmpl := texttmpl.Must(texttmpl.New("md").Parse(mdTemplate))
		mdPath := filepath.Join(pilotDataDir(), "reports", "report_"+reportID+".md")
		mdFile, _ := os.Create(mdPath)
		if mdFile != nil {
			mdTmpl.Execute(mdFile, fullRpt)
			mdFile.Close()
		}

		// Generate on-disk HTML
		funcMap := htmltmpl.FuncMap{ "lower": strings.ToLower, "pct": func(score float64) int { return int(score * 10) } }
		htmlTmpl := htmltmpl.Must(htmltmpl.New("html").Funcs(funcMap).Parse(htmlTemplate))
		htmlPath := filepath.Join(pilotDataDir(), "reports", "report_"+reportID+".html")
		htmlFile, _ := os.Create(htmlPath)
		if htmlFile != nil {
			htmlTmpl.Execute(htmlFile, fullRpt)
			htmlFile.Close()
		}

		// Update History Registry
		type historyItem struct {
			ID           string  `json:"id"`
			Timestamp    string  `json:"timestamp"`
			DatasetName  string  `json:"dataset_name"`
			OverallRisk  string  `json:"overall_risk"`
			RiskScore    float64 `json:"risk_score"`
			TotalRecords int     `json:"total_records"`
		}
		var history []historyItem
		histRaw, _ := os.ReadFile(filepath.Join(pilotDataDir(), "history.json"))
		json.Unmarshal(histRaw, &history)
		
		history = append(history, historyItem{
			ID:           reportID,
			Timestamp:    meta.GeneratedAt,
			DatasetName:  filepath.Base(datasetFile),
			OverallRisk:  report.OverallRiskLevel,
			RiskScore:    report.OverallRiskScore,
			TotalRecords: len(dataset),
		})
		histUpdated, _ := json.MarshalIndent(history, "", "  ")
		os.WriteFile(filepath.Join(pilotDataDir(), "history.json"), histUpdated, 0644)

		response := map[string]interface{}{
			"report":    report,
			"report_id": reportID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/pilot-assessment", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var dataset []map[string]interface{}
		var email, company string

		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			r.ParseMultipartForm(100 * 1024) // 100KB limit
			email = r.FormValue("email")
			company = r.FormValue("company")
			file, _, err := r.FormFile("dataset")
			if err == nil {
				defer file.Close()
				data, _ := io.ReadAll(file)
				json.Unmarshal(data, &dataset)
			}
		} else {
			var req struct {
				Email   string                   `json:"email"`
				Company string                   `json:"company"`
				Dataset []map[string]interface{} `json:"dataset"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				email = req.Email
				company = req.Company
				dataset = req.Dataset
			}
		}

		if email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		// Stateless assessment
		qi := []string{"region", "dept", "age_group"}
		sa := []string{"name", "email", "iban", "salary"}
		report := audit.AnalyzeDatasetRisk(dataset, qi, sa)

		// Store Lead
		type lead struct {
			Email     string    `json:"email"`
			Company   string    `json:"company"`
			Timestamp time.Time `json:"timestamp"`
			RiskLevel string    `json:"risk_level"`
		}
		os.MkdirAll(pilotDataDir(), 0755)
		var leads []lead
		leadRaw, _ := os.ReadFile(filepath.Join(pilotDataDir(), "leads.json"))
		json.Unmarshal(leadRaw, &leads)
		leads = append(leads, lead{Email: email, Company: company, Timestamp: time.Now(), RiskLevel: report.OverallRiskLevel})
		leadUpdated, _ := json.MarshalIndent(leads, "", "  ")
		os.WriteFile(filepath.Join(pilotDataDir(), "leads.json"), leadUpdated, 0644)

		// Map to full report for template generation
		reportID := strings.ToUpper(uuid.New().String()[:8])
		datasetScopeName := company + " Custom Upload"
		meta := reportMeta{
			ReportID:           reportID,
			GeneratedAt:        time.Now().UTC().Format("02 January 2006, 15:04 UTC"),
			DatasetScope:       datasetScopeName,
			MethodologyVersion: reportVersion,
			EngineVersion:      engineVersion,
			TotalRecords:       len(dataset),
		}
		before, after := buildScenarios(report)
		fullRpt := fullReport{Meta: meta, Risk: report, Before: before, After: after}

		// Ensure reports directory exists
		os.MkdirAll(filepath.Join(pilotDataDir(), "reports"), 0755)

		// Generate on-disk Markdown
		mdTmpl := texttmpl.Must(texttmpl.New("md").Parse(mdTemplate))
		mdPath := filepath.Join(pilotDataDir(), "reports", "report_"+reportID+".md")
		if mdFile, _ := os.Create(mdPath); mdFile != nil {
			mdTmpl.Execute(mdFile, fullRpt)
			mdFile.Close()
		}

		// Generate on-disk HTML
		funcMap := htmltmpl.FuncMap{ "lower": strings.ToLower, "pct": func(score float64) int { return int(score * 10) } }
		htmlTmpl := htmltmpl.Must(htmltmpl.New("html").Funcs(funcMap).Parse(htmlTemplate))
		htmlPath := filepath.Join(pilotDataDir(), "reports", "report_"+reportID+".html")
		if htmlFile, _ := os.Create(htmlPath); htmlFile != nil {
			htmlTmpl.Execute(htmlFile, fullRpt)
			htmlFile.Close()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "success",
			"report":      report,
			"report_id":   reportID,
			"full_report": fullRpt,
		})
	})

	http.HandleFunc("/api/pilot/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		content, _ := os.ReadFile(filepath.Join(pilotDataDir(), "history.json"))
		if len(content) == 0 {
			w.Write([]byte(`[]`))
			return
		}
		w.Write(content)
	})

	http.HandleFunc("/api/pilot/report", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Report ID missing", http.StatusBadRequest)
			return
		}

		path := filepath.Join(pilotDataDir(), "reports", "report_"+id+".html")
		content, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "Report not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
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

// --- Risk Report Generator Integration ---

const reportVersion = "3.1"
const engineVersion = "v1.14"

type reportMeta struct {
	ReportID           string
	GeneratedAt        string
	DatasetScope       string
	MethodologyVersion string
	EngineVersion      string
	TotalRecords       int
}

type fullReport struct {
	Meta   reportMeta
	Risk   audit.RiskReport
	Before scenarioSummary
	After  scenarioSummary
}

type scenarioSummary struct {
	Label       string
	RiskLevel   string
	RiskScore   string
	VaRRange    string
	AIStatus    string
	Description string
}

func buildScenarios(r audit.RiskReport) (scenarioSummary, scenarioSummary) {
	before := scenarioSummary{
		Label:     "Scenario A — Current State (No Protection)",
		RiskLevel: r.OverallRiskLevel,
		RiskScore: fmt.Sprintf("%.1f / 10", r.OverallRiskScore),
		VaRRange: fmt.Sprintf("€%.0f – €%.0f (estimated)", r.Exposure.VaRMin, r.Exposure.VaRMax),
		AIStatus:    r.AI.Status,
		Description: "The raw dataset as-is, transmitted directly to an LLM API or stored in a vector database. All PII fields are exposed in plaintext.",
	}

	afterScoreMin := r.OverallRiskScore * 0.05
	afterScoreMax := r.OverallRiskScore * 0.15
	afterVaRMin := r.Exposure.VaRMin * 0.02
	afterVaRMax := r.Exposure.VaRMin * 0.08

	after := scenarioSummary{
		Label:     "Scenario B — After OCULTAR Processing",
		RiskLevel: "LOW",
		RiskScore: fmt.Sprintf("%.1f – %.1f / 10 (projected)", afterScoreMin, afterScoreMax),
		VaRRange: fmt.Sprintf("€%.0f – €%.0f (projected residual)", afterVaRMin, afterVaRMax),
		AIStatus: "ALLOW",
		Description: "After OCULTAR tokenization and format-preserving encryption pipeline. Direct identifiers are removed and re-identification risk is significantly reduced (though not mathematically eliminated).",
	}
	return before, after
}

const mdTemplate = `# OCULTAR Data Risk Assessment Report

> **CONFIDENTIAL — For Authorised Recipients Only**
> This report constitutes a technical risk and privacy assessment based on automated analysis. It is informational in nature and does not constitute legal advice or a regulatory compliance determination. Distribution is restricted to named stakeholders.

---

## Report Metadata

| Field | Value |
| :--- | :--- |
| **Report ID** | OCU-{{.Meta.ReportID}} |
| **Generated** | {{.Meta.GeneratedAt}} |
| **Dataset Scope** | ` + "`" + `{{.Meta.DatasetScope}}` + "`" + ` |
| **Records Analysed** | {{.Meta.TotalRecords}} |
| **Methodology Version** | v{{.Meta.MethodologyVersion}} |
| **Engine** | OCULTAR Enterprise {{.Meta.EngineVersion}} |

---

## Executive Risk Summary

{{if eq .Risk.OverallRiskLevel "CRITICAL"}}> [!CAUTION]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: {{if .Risk.IsGDPRPseudonymized}}✅ Meets Common Pseudonymization Thresholds{{else}}⚠️ High Non-Compliance Likelihood (External Processing Scenarios){{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "HIGH"}}> [!WARNING]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: {{if .Risk.IsGDPRPseudonymized}}✅ Meets Common Pseudonymization Thresholds{{else}}⚠️ High Non-Compliance Likelihood (External Processing Scenarios){{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "MEDIUM"}}> [!IMPORTANT]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: {{if .Risk.IsGDPRPseudonymized}}✅ Meets Common Pseudonymization Thresholds{{else}}⚠️ Elevated Risk — Review Recommended{{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "LOW"}}> [!NOTE]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: ✅ Meets Common Pseudonymization Thresholds**{{end}}

The dataset identified in this report contains an estimated **{{.Risk.ViolatingRecords}} records** that fall below commonly cited EU pseudonymization thresholds. In its current state, this data **{{if .Risk.IsGDPRPseudonymized}}appears to satisfy commonly cited thresholds for use{{else}}presents elevated technical risk for use{{end}} with external AI systems and LLM APIs** without prior sanitisation.

The estimated financial exposure associated with unauthorised disclosure of this dataset is in the range of **€{{printf "%.0f" .Risk.Exposure.VaRMin}} – €{{printf "%.0f" .Risk.Exposure.VaRMax}}**. This is a **simulated estimate** grounded in the OCULTAR Three-Pillar VaR model, incorporating regulatory simulation anchors, operational incident benchmarks, and a risk multiplier. Actual impact is subject to contextual factors and organisational mitigating controls.

---

## Risk Scorecard

| Category | Score | Level | Business Implication |
| :--- | :---: | :---: | :--- |
| **Identifiability Risk** | {{printf "%.1f" .Risk.Identifiability.Score}}/10 | {{.Risk.Identifiability.Label}} | {{.Risk.Identifiability.Implication}} |
| **Financial Sensitivity** | {{printf "%.1f" .Risk.FinancialSensitivity.Score}}/10 | {{.Risk.FinancialSensitivity.Label}} | {{.Risk.FinancialSensitivity.Implication}} |
| **Re-identification Risk** | {{printf "%.1f" .Risk.ReidentificationRisk.Score}}/10 | {{.Risk.ReidentificationRisk.Label}} | {{.Risk.ReidentificationRisk.Implication}} |
| **Compliance Readiness** | {{printf "%.1f" .Risk.ComplianceReadiness.Score}}/10 | {{.Risk.ComplianceReadiness.Label}} | {{.Risk.ComplianceReadiness.Implication}} |
| **Overall** | **{{printf "%.1f" .Risk.OverallRiskScore}}/10** | **{{.Risk.OverallRiskLevel}}** | Weighted composite score (Identifiability 35%, Financial 25%, Re-id 25%, Compliance 15%) |

---

## Technical Metrics — Interpreted

### K-Anonymity
**Raw Score:** {{.Risk.KAnonymity}}

{{.Risk.KAnonymityInterpretation}}

> **Industry Benchmark:** Common industry frameworks suggest a minimum K-score of 3–5 for basic pseudonymization. This is a technical benchmark, not a mandatory legal threshold—contextual factors, processing purpose, and applicable exemptions determine actual compliance obligations.

### L-Diversity
**Raw Score:** {{.Risk.LDiversity}}

{{.Risk.LDiversityInterpretation}}

> **Industry Benchmark:** An L-Diversity score of ≥2 is commonly recommended to mitigate homogeneity attacks, as referenced in ISO/IEC 29101 (Privacy Architecture Framework). This is an industry guideline; applicable legal thresholds depend on jurisdictional context.

---

## Financial Exposure Model

The **Value at Risk (VaR)** range below is computed using a three-component methodology anchored to industry breach cost benchmarks. All figures are **simulated estimates** and should not be interpreted as predicted fine amounts or contractual commitments.

### VaR Components

| Component | Methodology | Min Estimate | Max Estimate |
| :--- | :--- | ---: | ---: |
| **Regulatory Exposure** | Simulation anchor (€10k–€100k base) × Dataset Risk Score ({{printf "%.2f" .Risk.DatasetRiskScore}}) | **€{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMin}}** | **€{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMax}}** |
| **Operational Cost** | Industry benchmark (€100–€300/record) × {{.Risk.TotalRecords}} records | **€{{printf "%.0f" .Risk.Exposure.OperationalCostMin}}** | **€{{printf "%.0f" .Risk.Exposure.OperationalCostMax}}** |
| **Risk Multiplier** | Profile-driven tiering (K={{.Risk.KAnonymity}}, L={{.Risk.LDiversity}}) | **{{printf "%.1f" .Risk.Exposure.RiskMultiplierMin}}×** | **{{printf "%.1f" .Risk.Exposure.RiskMultiplierMax}}×** |
| | | | |
| **Value at Risk (Estimated)** | **(Regulatory + Operational) × Multiplier** | **€{{printf "%.0f" .Risk.Exposure.VaRMin}}** | **€{{printf "%.0f" .Risk.Exposure.VaRMax}}** |

> **Assumptions & Methodology Note:**
> {{.Risk.Exposure.AssumptionsNote}}

---

## AI & LLM Exposure Assessment

### Decision: {{.Risk.AI.Status}}

| Parameter | Assessment |
| :--- | :--- |
| **External LLM API Safety** | {{.Risk.AI.LLMExposure}} risk |
| **Internal Copilot Safety** | {{if eq .Risk.AI.Status "ALLOW"}}✅ Permitted with monitoring{{else if eq .Risk.AI.Status "SANITIZE_FIRST"}}⚠️ Permitted after OCULTAR processing{{else}}🚫 Not recommended without sanitisation{{end}} |
| **Vector DB / RAG Indexing** | {{if .Risk.AI.RAGSafe}}✅ Estimated safe for indexing{{else}}🚫 Not recommended without prior processing{{end}} |

**RAG & Vector Database Guidance:**
{{.Risk.AI.RAGGuidance}}

**Recommended Action:**
{{.Risk.AI.Recommendation}}

---

## Before / After Simulation

This section demonstrates the modelled impact of the OCULTAR Enterprise pipeline on your dataset's risk profile. Figures are projected estimates based on typical processing outcomes.

| Metric | {{.Before.Label}} | {{.After.Label}} |
| :--- | :--- | :--- |
| **Risk Level** | 🔴 {{.Before.RiskLevel}} | 🟢 {{.After.RiskLevel}} |
| **Risk Score** | {{.Before.RiskScore}} | {{.After.RiskScore}} |
| **Financial Exposure (VaR)** | {{.Before.VaRRange}} | {{.After.VaRRange}} |
| **AI / LLM Status** | {{.Before.AIStatus}} | {{.After.AIStatus}} |

**What changes:**
- **Before:** {{.Before.Description}}
- **After:** {{.After.Description}}

---

## Assumptions

The following assumptions underpin all quantitative estimates in this report:

| Assumption | Value / Range | Basis |
| :--- | :--- | :--- |
| **Regulatory anchor (low)** | €10,000 | Simulation baseline |
| **Regulatory anchor (high)** | €100,000 | Simulation ceiling |
| **Operational cost per record** | €100–€300 | Industry study range |
| **Pseudonymization threshold** | K≥3, L≥2 | Common benchmark |

---

## Remediation Plan

{{.Risk.Recommendation}}

---

## Appendix: Methodology & Standards

This report applies the following analytical frameworks:

- **K-Anonymity** (Sweeney, 2002)
- **L-Diversity** (Machanavajjhala et al., 2006)
- **GDPR Article 5(1)(f)**
- **ISO/IEC 29101**

> This report was generated automatically by OCULTAR Enterprise {{.Meta.EngineVersion}}. technical assessment only.

---

*OCULTAR Enterprise {{.Meta.EngineVersion}} | Methodology v{{.Meta.MethodologyVersion}} | Report ID: OCU-{{.Meta.ReportID}}*
*Generated: {{.Meta.GeneratedAt}}*
`

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>OCULTAR Risk Assessment — OCU-{{.Meta.ReportID}}</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  :root {
    --critical: #dc2626; --high: #ea580c; --medium: #d97706; --low: #16a34a;
    --bg: #f8fafc; --surface: #ffffff; --border: #e2e8f0;
    --text: #0f172a; --muted: #64748b; --accent: #1e40af;
  }
  body { font-family: 'Inter', sans-serif; background: var(--bg); color: var(--text); font-size: 14px; line-height: 1.6; padding: 40px 24px; }
  .container { max-width: 960px; margin: 0 auto; }
  .report-header { background: var(--text); color: white; padding: 40px; border-radius: 12px; margin-bottom: 32px; position: relative; overflow: hidden; }
  .report-header h1 { font-size: 22px; font-weight: 700; margin-bottom: 4px; }
  .meta-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-top: 24px; }
  .meta-item label { display: block; font-size: 10px; text-transform: uppercase; opacity: 0.5; }
  .meta-item span { font-size: 13px; font-weight: 500; }
  .risk-banner { border-radius: 10px; padding: 24px 28px; margin-bottom: 28px; display: flex; align-items: center; gap: 20px; border: 1px solid var(--border); }
  .risk-banner.CRITICAL { background: #fef2f2; border-color: #fecaca; }
  .risk-banner.HIGH { background: #fff7ed; border-color: #fed7aa; }
  .risk-banner.MEDIUM { background: #fffbeb; border-color: #fde68a; }
  .risk-banner.LOW { background: #f0fdf4; border-color: #bbf7d0; }
  .risk-dial { width: 64px; height: 64px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 700; color: white; flex-shrink: 0; }
  .CRITICAL .risk-dial { background: var(--critical); }
  .HIGH .risk-dial { background: var(--high); }
  .MEDIUM .risk-dial { background: var(--medium); }
  .LOW .risk-dial { background: var(--low); }
  .section { background: var(--surface); border: 1px solid var(--border); border-radius: 10px; padding: 28px; margin-bottom: 20px; }
  .section h2 { font-size: 14px; font-weight: 700; margin-bottom: 16px; padding-bottom: 10px; border-bottom: 1px solid var(--border); color: var(--accent); text-transform: uppercase; }
  table { width: 100%; border-collapse: collapse; }
  th { text-align: left; padding: 10px; background: var(--bg); font-size: 11px; text-transform: uppercase; color: var(--muted); }
  td { padding: 12px; border-bottom: 1px solid #f1f5f9; font-size: 13px; }
  .badge { padding: 2px 8px; border-radius: 100px; font-size: 10px; font-weight: 600; }
  .badge-critical { background: #fef2f2; color: var(--critical); }
  .badge-low { background: #f0fdf4; color: var(--low); }
  .footer { text-align: center; margin-top: 40px; font-size: 11px; color: var(--muted); }
</style>
</head>
<body>
  <div class="container">
    <div class="report-header">
      <h1>OCULTAR Risk Assessment</h1>
      <div class="meta-grid">
        <div class="meta-item"><label>Report ID</label><span>OCU-{{.Meta.ReportID}}</span></div>
        <div class="meta-item"><label>Generated</label><span>{{.Meta.GeneratedAt}}</span></div>
        <div class="meta-item"><label>Engine</label><span>{{.Meta.EngineVersion}}</span></div>
      </div>
    </div>

    <div class="risk-banner {{.Risk.OverallRiskLevel}}">
      <div class="risk-dial">{{printf "%.1f" .Risk.OverallRiskScore}}</div>
      <div>
        <h2 style="font-size:18px; margin-bottom:4px;">{{.Risk.OverallRiskLevel}} Risk — {{if .Risk.IsGDPRPseudonymized}}Pseudonymized (Heuristic Assessment){{else}}Elevated Technical Risk Level{{end}}</h2>
        <p style="font-size:13px; opacity:0.7;">Estimated Var Range: <strong>€{{printf "%.0f" .Risk.Exposure.VaRMin}} - €{{printf "%.0f" .Risk.Exposure.VaRMax}}</strong></p>
      </div>
    </div>

    <div class="section">
      <h2>Risk Scorecard</h2>
      <table>
        <thead><tr><th>Category</th><th>Score</th><th>Level</th><th>Business Implication</th></tr></thead>
        <tbody>
          <tr>
            <td>Identifiability</td>
            <td>{{printf "%.1f" .Risk.Identifiability.Score}}</td>
            <td><span class="badge badge-{{lower .Risk.Identifiability.Label}}">{{.Risk.Identifiability.Label}}</span></td>
            <td>{{.Risk.Identifiability.Implication}}</td>
          </tr>
          <tr>
            <td>Financial Exposure</td>
            <td>{{printf "%.1f" .Risk.FinancialSensitivity.Score}}</td>
            <td><span class="badge badge-{{lower .Risk.FinancialSensitivity.Label}}">{{.Risk.FinancialSensitivity.Label}}</span></td>
            <td>{{.Risk.FinancialSensitivity.Implication}}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="section">
      <h2>Technical Metrics — Interpreted</h2>
      <div style="margin-bottom:20px;">
        <strong>K-Anonymity Score: {{.Risk.KAnonymity}}</strong><br>
        <p style="font-size:12px; color:var(--muted); margin-top:4px;">{{.Risk.KAnonymityInterpretation}}</p>
      </div>
      <div>
        <strong>L-Diversity Score: {{.Risk.LDiversity}}</strong><br>
        <p style="font-size:12px; color:var(--muted); margin-top:4px;">{{.Risk.LDiversityInterpretation}}</p>
      </div>
    </div>

    <div class="section">
      <h2>Financial Exposure — Three-Pillar VaR Model</h2>
      <p style="font-size:12px; color:var(--muted); margin-bottom:16px;">This model anchors technical risk scores to industry breach benchmarks (IBM/Ponemon) to simulate potential Value at Risk (VaR). All figures are projected ranges.</p>
      <table>
        <thead>
          <tr><th>Pillar / Component</th><th>Methodology</th><th style="text-align:right">Min Est. (€)</th><th style="text-align:right">Max Est. (€)</th></tr>
        </thead>
        <tbody>
          <tr>
            <td><strong>1. Regulatory Exposure</strong></td>
            <td>Simulation anchors (€10k-€100k) × Score</td>
            <td style="text-align:right">{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMin}}</td>
            <td style="text-align:right">{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMax}}</td>
          </tr>
          <tr>
            <td><strong>2. Operational Cost</strong></td>
            <td>Industry benchmarks (€100-€300/record)</td>
            <td style="text-align:right">{{printf "%.0f" .Risk.Exposure.OperationalCostMin}}</td>
            <td style="text-align:right">{{printf "%.0f" .Risk.Exposure.OperationalCostMax}}</td>
          </tr>
          <tr>
            <td><strong>3. Risk Multiplier</strong></td>
            <td>Profile-driven tiering (K/L profile)</td>
            <td style="text-align:right">{{printf "%.1f" .Risk.Exposure.RiskMultiplierMin}}×</td>
            <td style="text-align:right">{{printf "%.1f" .Risk.Exposure.RiskMultiplierMax}}×</td>
          </tr>
          <tr style="background:var(--bg); font-weight:700;">
            <td colspan="2">TOTAL VALUE AT RISK (SIMULATED RANGE)</td>
            <td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.VaRMin}}</td>
            <td style="text-align:right; color:var(--critical);">€{{printf "%.0f" .Risk.Exposure.VaRMax}}</td>
          </tr>
        </tbody>
      </table>
      <p style="font-size:11px; color:var(--muted); margin-top:12px; line-height:1.4;">{{.Risk.Exposure.AssumptionsNote}}</p>
    </div>

    <div class="section">
      <h2>AI & LLM Exposure Assessment</h2>
      <table>
        <thead><tr><th>Parameter</th><th>Assessment / Guidance</th></tr></thead>
        <tbody>
          <tr><td><strong>Decision</strong></td><td style="font-weight:700; color:{{if eq .Risk.AI.Status "ALLOW"}}var(--low){{else}}var(--critical){{end}};">{{.Risk.AI.Status}}</td></tr>
          <tr><td><strong>External LLM API Safety</strong></td><td>{{.Risk.AI.LLMExposure}} Risk Profile</td></tr>
          <tr><td><strong>Vector DB / RAG Indexing</strong></td><td>{{if .Risk.AI.RAGSafe}}✅ Estimated safe for indexing{{else}}🚫 Sanitisation required before indexing{{end}}</td></tr>
        </tbody>
      </table>
      <div style="margin-top:16px; font-size:12px; border-left:4px solid var(--accent); padding-left:16px; color:var(--muted);">
        <strong>RAG Guidance:</strong> {{.Risk.AI.RAGGuidance}}
      </div>
    </div>

    <div class="section">
      <h2>Before / After Impact Simulation</h2>
      <table>
        <thead><tr><th>Metric</th><th>{{.Before.Label}}</th><th>{{.After.Label}}</th></tr></thead>
        <tbody>
          <tr><td><strong>Risk Level</strong></td><td><span class="badge badge-critical">{{.Before.RiskLevel}}</span></td><td><span class="badge badge-low">{{.After.RiskLevel}}</span></td></tr>
          <tr><td><strong>Risk Score</strong></td><td>{{.Before.RiskScore}}</td><td>{{.After.RiskScore}}</td></tr>
          <tr><td><strong>VaR Range</strong></td><td>{{.Before.VaRRange}}</td><td>{{.After.VaRRange}}</td></tr>
        </tbody>
      </table>
    </div>

    <div class="section">
      <h2>Structured Remediation Plan</h2>
      <div style="font-size:13px; color:var(--text); white-space:pre-wrap; line-height:1.6;">{{.Risk.Recommendation}}</div>
    </div>

    <div class="footer">
      Generated automatically by OCULTAR Enterprise. Methodology v{{.Meta.MethodologyVersion}}<br>
      © 2026 OCULTAR Security. All rights reserved.
    </div>
  </div>
</body>
</html>
`
