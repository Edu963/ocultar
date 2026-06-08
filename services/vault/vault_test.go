package vault_test

import (
	"strings"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/vault"
)

func TestVaultFactory_CommunityTierBlocksPostgres(t *testing.T) {
	// Set an active license to community tier explicitly
	license.Active = license.Payload{
		Tier: "community",
	}

	cfg := config.Settings{
		VaultBackend: "postgres",
		PostgresDSN:  "postgres://fake:5432",
	}

	_, err := vault.New(cfg, "")
	if err == nil {
		t.Fatal("Expected vault initialization to fail for Community tier requesting postgres, but it succeeded")
	}

	if !strings.Contains(err.Error(), "does not permit postgres High Availability clustering") {
		t.Errorf("Expected license upgrade error message, got: %v", err)
	}
}

func TestVaultFactory_EnterpriseTierAllowsPostgres(t *testing.T) {
	// Set an active license to enterprise tier explicitly
	license.Active = license.Payload{
		Tier: "enterprise",
	}

	cfg := config.Settings{
		VaultBackend: "postgres",
		// Give it a malformed DSN just so we know we passed the license gate
		// and failed at the actual connection attempt
		PostgresDSN: "postgres://fake:5432",
	}

	_, err := vault.New(cfg, "")
	if err == nil {
		// Even for enterprise we expect this to fail because we provided a fake host
		// that isn't actually reachable, but we should NOT fail the license check
	} else if strings.Contains(err.Error(), "does not permit postgres High Availability clustering") {
		t.Errorf("Enterprise license should NOT trigger upgrade block, but got: %v", err)
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
