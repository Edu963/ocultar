package recon

import (
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/internal/pii"
)

type FileExposure struct {
	TotalPII int                   `json:"total_pii"`
	Hits     []pii.DetectionResult `json:"hits"`
}

type HeatmapReport struct {
	TargetDirectory string                  `json:"target_directory"`
	FilesScanned    int                     `json:"files_scanned"`
	TotalExposures  int                     `json:"total_exposures"`
	Files           map[string]FileExposure `json:"files,omitempty"`
}

type Crawler struct {
	Eng *refinery.Refinery
}

func NewCrawler(eng *refinery.Refinery) *Crawler {
	eng.DryRun = true // Passive discovery only; do not pollute the Enterprise Vault
	return &Crawler{Eng: eng}
}

func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	// We only want to crawl text-based data files at rest. (S3/SQL integration comes later)
	case ".txt", ".json", ".md", ".csv", ".yml", ".yaml", ".xml", ".html", ".log", ".sql", ".pdf":
		return true
	}
	return false
}

func extractText(path string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".pdf" {
		// Leverage existing system tool pdftotext for high-fidelity extraction
		cmd := exec.Command("pdftotext", path, "-")
		return cmd.Output()
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

func (c *Crawler) CrawlLocalDirectory(rootPath string) (HeatmapReport, error) {
	report := HeatmapReport{
		TargetDirectory: rootPath,
		Files:           make(map[string]FileExposure),
	}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "venv" || d.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		if !isTextFile(path) {
			return nil
		}

		info, err := d.Info()
		if err != nil || info.Size() > 50*1024*1024 { // Skip parsing files larger than 50MB
			return nil
		}

		content, err := extractText(path)
		if err != nil {
			log.Printf("[RECON-CRAWLER] Failed to extract text from %s: %v", path, err)
			return nil
		}

		c.Eng.ResetHits()
		_, err = c.Eng.RefineString(string(content), "recon-crawler", nil)
		if err != nil {
			return nil
		}

		rpt := c.Eng.GenerateReport(1)
		if rpt.TotalCount > 0 {
			report.TotalExposures += rpt.TotalCount
			report.Files[path] = FileExposure{
				TotalPII: rpt.TotalCount,
				Hits:     rpt.Hits,
			}
		}

		report.FilesScanned++
		return nil
	})

	return report, err
}

func (h HeatmapReport) ToJSON() string {
	b, _ := json.MarshalIndent(h, "", "  ")
	return string(b)
}
