// Package vault defines the storage abstraction layer for OCULTAR's PII vault.
// It exposes a single Provider interface that all vault backends must implement,
// and a New() factory that selects the right backend based on configuration.
package vault

import (
	"fmt"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/license"
)

// Provider is the storage contract every vault backend must satisfy.
// It intentionally knows nothing about SQL, DuckDB, or Postgres.
type Provider interface {
	// StoreToken persists a (hash → token, encrypted_pii) mapping.
	// Returns (true, nil) when a new row was inserted, (false, nil) when
	// the hash already existed (idempotent), or (false, err) on failure.
	StoreToken(hash, token, encryptedPII string) (inserted bool, err error)

	// GetToken looks up the token for a given PII hash.
	// Returns (token, true) on a cache hit, ("", false) on a miss.
	GetToken(hash string) (token string, found bool)

	// CountAll returns the total number of entries in the vault.
	// Used for pilot-mode enforcement.
	CountAll() int64

	// Close releases any open database connections.
	Close() error
}

// New is the factory function that returns the appropriate Provider based on
// the configuration settings. Community / default users get duckdbProvider;
// Enterprise users with vault_backend=postgres get postgresProvider.
//
// vaultPath is passed explicitly so that main.go can override it with
// ":memory:" when running in dry-run mode.
func New(cfg config.Settings, vaultPath string) (Provider, error) {
	switch cfg.VaultBackend {
	case "postgres":
		if !license.IsEnterprise() {
			return nil, fmt.Errorf("[vault] Active license tier '%s' does not permit postgres High Availability clustering. Please upgrade to Enterprise.", license.Active.Tier)
		}
		if cfg.PostgresDSN == "" {
			return nil, fmt.Errorf("[vault] vault_backend is 'postgres' but postgres_dsn is not set in config.yaml")
		}
		return newPostgresProvider(cfg.PostgresDSN)
	default:
		// "duckdb" or empty string → use local DuckDB file
		return newDuckDBProvider(vaultPath)
	}
}
