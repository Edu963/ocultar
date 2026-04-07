package audit

import (
	"testing"
)

func TestAnalyzeDatasetRisk(t *testing.T) {
	// A fully compliant dataset (K=3, L=2 for each group)
	dataset := []map[string]interface{}{
		// Group 1
		{"age": "30-40", "zip": "90210", "disease": "flu"},
		{"age": "30-40", "zip": "90210", "disease": "covid"},
		{"age": "30-40", "zip": "90210", "disease": "cold"},
		// Group 2
		{"age": "20-30", "zip": "10001", "disease": "flu"},
		{"age": "20-30", "zip": "10001", "disease": "flu"},
		{"age": "20-30", "zip": "10001", "disease": "covid"},
	}

	qi := []string{"age", "zip"}
	sa := []string{"disease"}

	report := AnalyzeDatasetRisk(dataset, qi, sa)

	if report.KAnonymity != 3 {
		t.Errorf("Expected K=3, got %d", report.KAnonymity)
	}

	if report.LDiversity != 2 {
		t.Errorf("Expected L=2, got %d", report.LDiversity)
	}

	if !report.IsGDPRPseudonymized {
		t.Errorf("Expected dataset to meet pseudonymization thresholds")
	}

	if report.ViolatingRecords != 0 {
		t.Errorf("Expected 0 violating records, got %d", report.ViolatingRecords)
	}

	// Introduce a highly obvious, re-identifiable outlier row
	dataset = append(dataset, map[string]interface{}{
		"age": "80-90", "zip": "99999", "disease": "cancer",
	})

	report2 := AnalyzeDatasetRisk(dataset, qi, sa)
	if report2.KAnonymity != 1 {
		t.Errorf("Expected K=1 after adding outlier, got %d", report2.KAnonymity)
	}
	if report2.IsGDPRPseudonymized {
		t.Errorf("Expected dataset to fail pseudonymization thresholds after adding outlier")
	}
	if report2.ViolatingRecords != 1 {
		t.Errorf("Expected 1 violating record, got %d", report2.ViolatingRecords)
	}
}
