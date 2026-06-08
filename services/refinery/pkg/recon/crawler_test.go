package recon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/vault"
)

func TestCrawlLocalDirectory(t *testing.T) {
	// Setup generic refinery mapping
	os.Chdir("../..")
	config.InitDefaults()
	v, err := vault.New(config.Global, "")
	if err != nil {
		t.Fatalf("Failed to init vault: %v", err)
	}
	eng := refinery.NewRefinery(v, []byte("01234567890123456789012345678901"))

	crawler := NewCrawler(eng)

	// Create temporary directory for crawling
	tempDir := t.TempDir()

	// 1. Create a safe file
	safePath := filepath.Join(tempDir, "safe.txt")
	os.WriteFile(safePath, []byte("This is a completely normal system log with no sensitive information."), 0644)

	// 2. Create a high-risk file
	riskPath := filepath.Join(tempDir, "customer_data.txt")
	os.WriteFile(riskPath, []byte(`Client: client@example.com Phone: +33 6 12 34 56 78 Note: Send to Paris.`), 0644)

	// Execute crawl
	report, err := crawler.CrawlLocalDirectory(tempDir)
	if err != nil {
		t.Fatalf("Crawl error: %v", err)
	}

	if report.FilesScanned != 2 {
		t.Errorf("Expected 2 files scanned, got %d", report.FilesScanned)
	}

	// We expect the high-risk file to be populated, but the safe file to be omitted.
	t.Logf("Full scan report: %+v", report)
	if len(report.Files) != 1 {
		t.Errorf("Expected exactly 1 file to have PII exposures, got %d", len(report.Files))
	}

	if exp, ok := report.Files[riskPath]; !ok {
		t.Errorf("crawler failed to flag riskPath: %s", riskPath)
	} else {
		if exp.TotalPII < 2 {
			t.Errorf("Expected at least 2 PII elements (email, phone), got %d", exp.TotalPII)
		}
	}
}
