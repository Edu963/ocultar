package audit_test

import (
	"os"
	"strings"
	"testing"
	"time"

	enterpriseaudit "github.com/Edu963/ocultar-enterprise/pkg/audit"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/vault"
)

func TestEnterpriseAuditLogger(t *testing.T) {
	// Create in-memory DuckDB vault via vault.Provider interface.
	_ = os.Remove("test_audit.log")
	_ = os.Remove("test_vault_audit.db")

	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "test_vault_audit.db")
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer v.Close()
	defer os.Remove("test_vault_audit.db")

	_ = os.Remove("test_audit.log")

	logger := enterpriseaudit.NewLogger()
	logger.Init("test_audit.log")
	defer logger.Close()

	masterKey := []byte("01234567890123456789012345678901") // 32 bytes

	// Manually inject Enterprise state and custom rules
	license.Active = license.Payload{Tier: "enterprise", CustomerName: "TestCorp"}
	config.Global.Dictionaries = []config.DictRule{
		{
			Type:  "PERSON_VIP",
			Terms: []string{"Carlos Andrés Segura"},
		},
	}

	eng := refinery.NewRefinery(v, masterKey)
	eng.DryRun = false
	eng.AuditLogger = logger

	input := "Hello mr john.doe@example.com"
	eng.RefineString(input, "audit-tester", nil)

	// Test VIP mapping
	vipInput := "Meeting with Carlos Andrés Segura tomorrow."
	eng.RefineString(vipInput, "audit-tester", nil)

	// Wait a moment for async channel to flush
	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile("test_audit.log")
	if err != nil || len(content) == 0 {
		t.Fatalf("Enterprise tier did not write to the audit logger! Log empty or missing.")
	}

	if !strings.Contains(string(content), "audit-tester") || !strings.Contains(string(content), "vaulted") {
		t.Fatalf("Audit log missing expected tokens: %s", string(content))
	}

	t.Logf("✅ Verified Enterprise JSON SIEM Audit: %s", string(content))
}
