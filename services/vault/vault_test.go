package vault_test

import (
	"strings"
	"testing"

	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/vault"
)

func TestVaultFactory_PostgresFailsWithInvalidDSN(t *testing.T) {
	cfg := config.Settings{
		VaultBackend: "postgres",
		PostgresDSN:  "postgres://fake:5432",
	}

	_, err := vault.New(cfg, "")
	if err == nil {
		t.Fatal("Expected postgres vault initialization to fail with an unreachable host, but it succeeded")
	}

	// Should be a connection/DNS error, not a license gate.
	if strings.Contains(err.Error(), "does not permit postgres") {
		t.Errorf("Got unexpected license gate error (license system should be a no-op): %v", err)
	}
}
func TestVaultFactory_DuckDBWorks(t *testing.T) {
	cfg := config.Settings{
		VaultBackend: "duckdb",
	}

	v, err := vault.New(cfg, "")
	if err != nil {
		t.Fatalf("Failed to initialize DuckDB vault: %v", err)
	}
	defer v.Close()

	inserted, err := v.StoreToken("hash1", "token1", "enc1")
	if err != nil {
		t.Fatalf("StoreToken failed: %v", err)
	}
	if !inserted {
		t.Fatal("Expected inserted to be true")
	}

	token, found := v.GetToken("hash1")
	if !found || token != "token1" {
		t.Errorf("Expected token1, got %s (found=%v)", token, found)
	}
}
