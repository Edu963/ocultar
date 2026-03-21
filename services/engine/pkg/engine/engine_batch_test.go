package engine_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/vault"
)

func TestRefineBatchConcurrency(t *testing.T) {
	_ = os.Remove("test_batch_vault.db")
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "test_batch_vault.db")
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer v.Close()
	defer os.Remove("test_batch_vault.db")

	masterKey := []byte("01234567890123456789012345678901")

	config.InitDefaults()
	eng := engine.NewEngine(v, masterKey)
	eng.DryRun = false
	eng.Report = false

	// Generate a moderate slice simulating CSV load to avoid long test times but ensure concurrency works
	numItems := 1000
	items := make([]interface{}, numItems)
	for i := 0; i < numItems; i++ {
		// Mock CSV row
		items[i] = map[string]interface{}{
			"id":     i,
			"name":   fmt.Sprintf("User%d", i),
			"email":  fmt.Sprintf("user%d@example.com", i),
			"phone":  "+1-800-555-0199",
			"active": i%2 == 0,
		}
	}

	actor := "batch_tester"
	results, err := eng.RefineBatch(items, actor)
	if err != nil {
		t.Fatalf("RefineBatch failed: %v", err)
	}

	if len(results) != numItems {
		t.Fatalf("Expected %d results, got %d", numItems, len(results))
	}

	// Verify sample result
	sample, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Result item is not map[string]interface{}")
	}

	emailStr, ok := sample["email"].(string)
	if !ok {
		t.Fatalf("Email field is not string")
	}

	if strings.Contains(emailStr, "@example.com") {
		t.Errorf("Failed to redact email in batch processing: %s", emailStr)
	}

	if !strings.HasPrefix(emailStr, "[EMAIL_") {
		t.Errorf("Expected tokenized email, got: %s", emailStr)
	}
}
