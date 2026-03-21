// Package proxy implements the OCULTAR transparent HTTP reverse-proxy.
// It intercepts outgoing POST requests, redacts PII via the engine, forwards
// the sanitised payload to the upstream API, then re-hydrates vault tokens in
// the response before returning it to the caller.
package proxy

import (
	"log"
	"os"
	"strings"
)

// Config holds runtime configuration for the OCULTAR proxy.
type Config struct {
	// Port is the local port the proxy server listens on (e.g. "8080").
	Port string

	// TargetURL is the upstream origin to forward requests to
	// (e.g. "https://api.openai.com"). Must include scheme.
	TargetURL string

	// VaultPath is the path to the DuckDB vault file.
	VaultPath string
}

// LoadConfig reads proxy configuration from environment variables.
func LoadConfig() Config {
	port := os.Getenv("OCU_PROXY_PORT")
	if port == "" {
		port = "8080"
	}

	target := os.Getenv("OCU_PROXY_TARGET")
	if target == "" {
		log.Fatal("[FATAL] OCU_PROXY_TARGET is not set. Set it to the upstream API URL (e.g. https://api.openai.com)")
	}
	// Strip trailing slash for clean URL construction.
	target = strings.TrimRight(target, "/")

	vaultPath := os.Getenv("OCU_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = "vault.db"
	}

	return Config{
		Port:      port,
		TargetURL: target,
		VaultPath: vaultPath,
	}
}
