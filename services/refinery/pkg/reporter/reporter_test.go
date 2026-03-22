package reporter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAuditLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_audit.log")

	logData := `{"timestamp":"2026-02-25T06:56:42Z","user":"test","action":"vaulted","result":"[EMAIL_123]"}
{"timestamp":"2026-02-25T06:56:42Z","user":"test","action":"matched","result":"[PHONE_456]"}
{"timestamp":"2026-02-25T06:56:42Z","user":"test","action":"vaulted","result":"[ORG_789]"}
{"timestamp":"2026-02-25T06:56:42Z","user":"test","action":"other","result":"ignoreme"}
{"timestamp":"2026-02-25T06:56:42Z","user":"test","action":"vaulted","result":"[EMAIL_abc]"}
`
	if err := os.WriteFile(logPath, []byte(logData), 0644); err != nil {
		t.Fatalf("failed to write test log: %v", err)
	}

	rep := New()
	metrics, err := rep.ParseAuditLog(logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.TotalPrevented != 4 {
		t.Errorf("expected 4 total prevented, got %d", metrics.TotalPrevented)
	}

	if metrics.Tier1Count != 3 {
		t.Errorf("expected 3 tier 1 (2x EMAIL, 1x PHONE), got %d", metrics.Tier1Count)
	}

	if metrics.Tier2Count != 1 {
		t.Errorf("expected 1 tier 2 (ORG), got %d", metrics.Tier2Count)
	}

	if len(metrics.TopCategories) != 3 {
		t.Errorf("expected 3 categories (EMAIL, PHONE, ORG), got %d", len(metrics.TopCategories))
	}

	if metrics.TopCategories[0].Name != "EMAIL" || metrics.TopCategories[0].Count != 2 {
		t.Errorf("expected top category to be EMAIL:2, got %s:%d", metrics.TopCategories[0].Name, metrics.TopCategories[0].Count)
	}
}

func TestGenerateHTMLReport(t *testing.T) {
	// First write a mock audit.log file similar to what we have in the main dir
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_audit.log")

	logData := `{"timestamp":"2026-02-25T06:56:42.100059377Z","user":"audit-tester","action":"vaulted","result":"[EMAIL_836f82db]"}
{"timestamp":"2026-02-25T06:57:42.100059377Z","user":"audit-tester","action":"vaulted","result":"[PHONE_abc]"}
{"timestamp":"2026-02-25T06:58:42.100059377Z","user":"audit-tester","action":"matched","result":"[EMAIL_836f82db]"}
{"timestamp":"2026-02-25T06:59:42.100059377Z","user":"audit-tester","action":"vaulted","result":"[ORG_xyz]"}
`
	if err := os.WriteFile(logPath, []byte(logData), 0644); err != nil {
		t.Fatalf("failed to write test log: %v", err)
	}

	outPath := filepath.Join(tmpDir, "report.html")

	rep := New()
	if err := rep.GenerateHTMLReport(logPath, outPath); err != nil {
		t.Fatalf("GenerateHTMLReport failed: %v", err)
	}

	// Verify file was created
	stat, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("failed to stat report.html: %v", err)
	}
	if stat.Size() == 0 {
		t.Errorf("report.html is empty")
	}

	// Optional check content length to ensure it rendered OK
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read report.html: %v", err)
	}
	if len(b) < 100 {
		t.Errorf("report.html seems too small, len: %d", len(b))
	}
}
