// Package main is the entry-point for the OCULTAR transparent HTTP proxy.
//
// Usage:
//
//	OCU_PROXY_TARGET=https://api.openai.com \
//	OCU_PROXY_PORT=8080 \
//	OCU_MASTER_KEY=your-secret \
//	./proxy
//
// Configure your application to route HTTP traffic through http://localhost:8080
// and all outgoing POST request bodies will be silently redacted.
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/pkg/proxy"
	"github.com/Edu963/ocultar/pkg/vault"
	"golang.org/x/crypto/hkdf"
)

const VERSION = "1.0.0"

const defaultSalt = "ocultar-v112-kdf-salt-fixed-16"

func getSalt() string {
	if s := os.Getenv("OCU_SALT"); s != "" {
		return s
	}
	log.Printf("[WARN] OCU_SALT is not set — using built-in default salt. Set OCU_SALT in production.")
	return defaultSalt
}

func getMasterKey() []byte {
	keyMaterial := os.Getenv("OCU_MASTER_KEY")
	if keyMaterial == "" {
		log.Printf("[WARN] OCU_MASTER_KEY is not set — using insecure dev key. Never deploy to production.")
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
	log.Printf("OCULTAR Privacy Proxy v%s starting…", VERSION)

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

	// ── Boot engine ───────────────────────────────────────────────────────────
	pilotMode := license.Active.Tier == "community" ||
		os.Getenv("OCU_PILOT_MODE") == "1" ||
		os.Getenv("OCU_PILOT_MODE") == "true"

	eng := engine.NewEngine(vaultProvider, masterKey)
	eng.Serve = "proxy" // marks the engine as running in serve mode (enables hit tracking)
	eng.PilotMode = pilotMode

	// ── Build proxy handler ───────────────────────────────────────────────────
	handler, err := proxy.NewHandler(eng, vaultProvider, masterKey, cfg.TargetURL)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	addr := ":" + cfg.Port
	if cfg.Port[0] == ':' {
		addr = cfg.Port
	}

	fmt.Printf("┌──────────────────────────────────────────────────────┐\n")
	fmt.Printf("│  OCULTAR Privacy Proxy v%-27s │\n", VERSION)
	fmt.Printf("│  Listening : http://localhost%-24s │\n", addr)
	fmt.Printf("│  Target    : %-38s │\n", cfg.TargetURL)
	fmt.Printf("│  Vault     : %-38s │\n", cfg.VaultPath)
	fmt.Printf("│  Pilot     : %-38v │\n", pilotMode)
	fmt.Printf("└──────────────────────────────────────────────────────┘\n")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/", handler)

	log.Fatal(http.ListenAndServe(addr, mux))
}
