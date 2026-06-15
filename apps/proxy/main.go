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
	"path/filepath"

	"github.com/ocultar-dev/ocultar/pkg/audit"
	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/pkg/proxy"
	"github.com/ocultar-dev/ocultar/pkg/refinery"
	"github.com/ocultar-dev/ocultar/pkg/inference"
	"github.com/ocultar-dev/ocultar/vault"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/hkdf"
)

const VERSION = "1.1.0"

const defaultSalt = "ocultar-v112-kdf-salt-fixed-16"

var devMode bool

func init() {
	flag.BoolVar(&devMode, "dev", false, "Enable development mode (allows insecure defaults)")
}

// auditAdapter bridges audit.ImmutableLogger to the refinery.AuditLogger interface.
type auditAdapter struct {
	logger *audit.ImmutableLogger
}

func (a *auditAdapter) Init(_ string) error { return nil }
func (a *auditAdapter) Close()              { a.logger.Close() }
func (a *auditAdapter) Log(actor, action, resource, mapping string) {
	if err := a.logger.Log(actor, action, resource, "ALLOW", mapping); err != nil {
		log.Printf("[WARN] audit write failed: %v", err)
	}
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

	cfg := proxy.LoadConfig()
	masterKey := getMasterKey()
	config.Load()

	vaultProvider, err := vault.New(config.Global, cfg.VaultPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to open vault: %v", err)
	}
	defer vaultProvider.Close()

	eng := refinery.NewRefinery(vaultProvider, masterKey)
	eng.Serve = "proxy"

	// Immutable audit log — active when OCU_AUDIT_PRIVATE_KEY is set.
	auditActive := false
	if keyHex := os.Getenv("OCU_AUDIT_PRIVATE_KEY"); keyHex != "" {
		privKey, err := audit.LoadPrivateKeyFromHex(keyHex)
		if err != nil {
			log.Fatalf("[FATAL] %v", err)
		}
		logPath := os.Getenv("OCU_AUDIT_LOG_PATH")
		if logPath == "" {
			logPath = filepath.Join(filepath.Dir(cfg.VaultPath), "audit.log")
		}
		immutableLog, err := audit.NewImmutableLoggerWithKey(logPath, privKey)
		if err != nil {
			log.Fatalf("[FATAL] Failed to open audit log at %s: %v", logPath, err)
		}
		defer immutableLog.Close()
		eng.SetAuditLogger(&auditAdapter{logger: immutableLog})
		auditActive = true
		log.Printf("[INFO] Immutable audit log active: %s (public key: %s)", logPath, immutableLog.PublicKeyHex())
	} else {
		eng.SetAuditLogger(&refinery.NoopAuditLogger{})
	}

	// Tier 2 AI NER — active when SLM_SIDECAR_URL is set (defaults to localhost:8085).
	sidecarURL := os.Getenv("SLM_SIDECAR_URL")
	if sidecarURL == "" {
		sidecarURL = "http://localhost:8085"
	}
	scanner := inference.NewRemoteScanner(sidecarURL)
	eng.SetAIScanner(scanner)
	log.Printf("[INFO] Tier 2 AI (Remote Sidecar) configured at %s", sidecarURL)

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
	fmt.Printf("│  Metrics   : http://localhost%-23s │\n", addr+"/metrics")
	fmt.Printf("│  Audit log : %-38v │\n", auditActive)
	fmt.Printf("│  DevMode   : %-38v │\n", devMode)
	fmt.Printf("└──────────────────────────────────────────────────────┘\n")

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok", "version":"` + VERSION + `"}`))
	})

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", handler)

	log.Fatal(http.ListenAndServe(addr, mux))
}
