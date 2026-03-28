// Package main is the entry-point for the OCULTAR transparent HTTP proxy.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/pkg/proxy"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/inference"
	"github.com/Edu963/ocultar/vault"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/hkdf"
)

const VERSION = "1.1.0" // Hardened for Production

const defaultSalt = "ocultar-v112-kdf-salt-fixed-16"

var devMode bool

func init() {
	flag.BoolVar(&devMode, "dev", false, "Enable development mode (allows insecure defaults)")
}

func getSalt() string {
	s := os.Getenv("OCU_SALT")
	if s == "" {
		if !devMode {
			log.Fatalf("[FATAL] OCU_SALT is missing. Production environments MUST define a unique OCU_SALT.")
		}
		log.Printf("[WARN] OCU_SALT is not set — using built-in default salt. (Allowed ONLY in --dev mode)")
		return defaultSalt
	}
	return s
}

func getMasterKey() []byte {
	keyMaterial := os.Getenv("OCU_MASTER_KEY")
	if keyMaterial == "" {
		if !devMode {
			log.Fatalf("[FATAL] OCU_MASTER_KEY is missing. Production environments MUST define a high-entropy master key.")
		}
		log.Printf("[WARN] OCU_MASTER_KEY is not set — using insecure dev key. (Allowed ONLY in --dev mode)")
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

func main() {
	flag.Parse()
	log.Printf("OCULTAR Privacy Proxy v%s starting (DevMode: %v)…", VERSION, devMode)

	// ── Load config ───────────────────────────────────────────────────────────
	cfg := proxy.LoadConfig()
	masterKey := getMasterKey()

	// ── Boot license + config ──────────────────────────────────────────────
	license.Load()
	config.Load()

	// ── Open vault (provider selected by config) ──────────────────────────────
	vaultProvider, err := vault.New(config.Global, cfg.VaultPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to open vault: %v", err)
	}
	defer vaultProvider.Close()

	// ── Boot refinery ───────────────────────────────────────────────────────────
	pilotMode := license.Active.Tier == "community" ||
		os.Getenv("OCU_PILOT_MODE") == "1" ||
		os.Getenv("OCU_PILOT_MODE") == "true"

	eng := refinery.NewRefinery(vaultProvider, masterKey)
	eng.Serve = "proxy"
	eng.PilotMode = pilotMode

	// ── Initialize Enterprise Components ──────────────────────────────────────
	if license.IsEnterprise() || os.Getenv("OCU_FORCE_ENTERPRISE") == "true" {
		log.Printf("[INFO] Initializing Enterprise Security Tiers...")
		// 1. SIEM Auditor
		auditor := &refinery.NoopAuditLogger{} // Replace with real auditor if available
		eng.SetAuditLogger(auditor)

		// 2. Local SLM Scanner
		modelPath := os.Getenv("SLM_MODEL_PATH")
		if modelPath != "" {
			scanner, err := inference.NewLlamaScanner(modelPath)
			if err != nil {
				log.Printf("[WARN] Failed to initialize SLM scanner: %v. AI coverage will be limited.", err)
			} else {
				eng.SetAIScanner(scanner)
				log.Printf("[INFO] Tier 2 AI (llama.cpp) active. Model: %s", modelPath)
			}
		} else {
			log.Printf("[INFO] SLM_MODEL_PATH not set. Tier 2 AI scanning is disabled.")
		}
	}
	
	// ── Build proxy handler ───────────────────────────────────────────────────
	handler, err := proxy.NewHandler(eng, vaultProvider, masterKey, cfg.TargetURL)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	addr := ":" + cfg.Port
	if cfg.Port != "" && cfg.Port[0] == ':' {
		addr = cfg.Port
	}

	fmt.Printf("┌──────────────────────────────────────────────────────┐\n")
	fmt.Printf("│  OCULTAR Privacy Proxy v%-27s │\n", VERSION)
	fmt.Printf("│  Listening : http://localhost%-24s │\n", addr)
	fmt.Printf("│  Target    : %-38s │\n", cfg.TargetURL)
	fmt.Printf("│  Vault     : %-38s │\n", cfg.VaultPath)
	fmt.Printf("│  Pilot     : %-38v │\n", pilotMode)
	fmt.Printf("│  DevMode   : %-38v │\n", devMode)
	fmt.Printf("└──────────────────────────────────────────────────────┘\n")

	mux := http.NewServeMux()
	
	// Health & Metrics
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok", "version":"` + VERSION + `"}`))
	})
	
	if config.Global.PrometheusEnabled {
		log.Printf("[INFO] Metrics enabled on /metrics")
		mux.Handle("/metrics", promhttp.Handler())
	}

	mux.Handle("/", handler)

	log.Fatal(http.ListenAndServe(addr, mux))
}
