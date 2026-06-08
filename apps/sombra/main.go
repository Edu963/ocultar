package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Edu963/ocultar/apps/sombra/pkg/connector"
	"github.com/Edu963/ocultar/apps/sombra/pkg/handler"
	"github.com/Edu963/ocultar/apps/sombra/pkg/router"
	"github.com/Edu963/ocultar/pkg/audit"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/inference"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/vault"
	"golang.org/x/crypto/hkdf"
)

const defaultSalt = "ocultar-v112-kdf-salt-fixed-16"

func getSalt() string {
	if s := os.Getenv("OCU_SALT"); s != "" {
		return s
	}
	return defaultSalt
}

func getMasterKey() []byte {
	keyMaterial := os.Getenv("OCU_MASTER_KEY")
	if keyMaterial == "" {
		log.Printf("[WARN] OCU_MASTER_KEY is not set — using insecure dev key.")
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
	fmt.Println("🚀 Booting Ocultar Sombra Gateway (Server Mode)...")

	// 0. Initialize defaults and infrastructure
	config.InitDefaults()
	config.Load()
	license.Load()

	if config.Global.JWTSecret == "" {
		log.Printf("[WARN] OCU_JWT_SECRET is not set — Sombra is running in insecure dev mode. Any Bearer value is accepted as actor identity. Set OCU_JWT_SECRET in production.")
	}

	masterKey := getMasterKey()

	vaultPath := os.Getenv("OCU_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = "sombra_vault.db"
	}

	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, vaultPath)
	if err != nil {
		log.Fatalf("Failed to initialize vault: %v", err)
	}
	defer v.Close()

	eng := refinery.NewRefinery(v, masterKey)

	// 0.1 Initialize Tier 2 SLM Scanner via Sidecar
	if os.Getenv("OCU_FORCE_ENTERPRISE") == "true" || license.IsEnterprise() {
		sidecarURL := os.Getenv("SLM_SIDECAR_URL")
		if sidecarURL == "" {
			sidecarURL = "http://localhost:8085"
		}
		scanner := inference.NewRemoteScanner(sidecarURL)
		eng.SetAIScanner(scanner)
		log.Printf("[INFO] Tier 2 AI active via SLM sidecar: %s", sidecarURL)
	}

	// 1. Setup Multi-Model Router
	allowedDomains := []string{"generativelanguage.googleapis.com", "api.openai.com", "api.mistral.ai", "api.anthropic.com", "127.0.0.1"}
	r := router.New("gemini-flash-latest", allowedDomains)

	// Register Gemini — router key is "gemini-flash-latest" (client-facing alias);
	// API model ID is the actual Google identifier used in the generateContent URL.
	gemini := router.NewGemini("gemini-flash-latest", "", "GEMINI_API_KEY", "gemini-2.0-flash")
	r.Register(gemini)

	// Register OpenAI
	openai := router.NewOpenAI("gpt-4o", "", "OPENAI_API_KEY")
	r.Register(openai)
	openaiMini := router.NewOpenAI("gpt-4o-mini", "", "OPENAI_API_KEY")
	r.Register(openaiMini)

	// Register Mistral (OpenAI-compatible API)
	mistral := router.NewOpenAI("mistral-large-latest", "https://api.mistral.ai/v1", "MISTRAL_API_KEY")
	r.Register(mistral)

	// Register Claude
	claude := router.NewClaude("claude-sonnet-4-6", "", "ANTHROPIC_API_KEY")
	r.Register(claude)

	// Register local-slm if needed (optional for this context)
	// localSlm := router.NewLocal("local-slm", "http://localhost:8085")
	// r.Register(localSlm)

	// Register mock AI when SOMBRA_MOCK_AI_URL is set (demo / offline testing)
	if mockURL := os.Getenv("SOMBRA_MOCK_AI_URL"); mockURL != "" {
		r.Register(router.NewLocal("mock-ai", mockURL))
		log.Printf("[INFO] Demo mode: mock-ai registered at %s", mockURL)
	}

	// 1.5 Initialize Immutable Audit Logger
	auditor, err := audit.NewImmutableLogger("sombra_audit.log")
	if err != nil {
		log.Printf("[WARN] Failed to initialize Immutable Logger: %v", err)
	} else {
		log.Printf("[INFO] Immutable Audit Log active. PubKey: %s", auditor.PublicKeyHex())
		defer auditor.Close()
	}

	// 2. Initialize the Sombra Gateway handler
	g := handler.NewGateway(eng, v, masterKey, r, auditor)

	// 3. Register Connectors
	// Default file connector with a relaxed policy for testing
	filePolicy := connector.DataPolicy{
		AllowedModels: []string{"gemini-flash-latest", "local-slm", "gpt-4o", "gpt-4o-mini", "mistral-large-latest", "claude-sonnet-4-6"},
		MaxBodyBytes:  10485760, // 10MB
	}
	g.RegisterConnector(connector.NewFileConnector("file", filePolicy))

	// 4. Start HTTP Server
	port := os.Getenv("SOMBRA_PORT")
	if port == "" {
		port = "8086"
	}

	http.HandleFunc("/query", g.HandleQuery)
	http.HandleFunc("/v1/chat/completions", g.HandleV1ChatCompletions)
	http.HandleFunc("/v1/slack/events", g.HandleSlackEvent)

	// Entity Registry API (Path 3)
	http.HandleFunc("/v1/entities", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			g.HandleEntityRegister(w, r)
		case http.MethodGet:
			g.HandleEntityList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/v1/entities/seed", g.HandleEntitySeed)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok"}`)
	})

	log.Printf("[INFO] Sombra Gateway running on http://localhost:%s", port)
	log.Printf("[INFO] Protected Query endpoint: http://localhost:%s/query", port)

	srv := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
