package recon

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Edu963/ocultar/pkg/refinery"
)

type FileExposure struct {
	TotalPII int                 `json:"total_pii"`
	Types    map[string][]string `json:"types"`
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
	case ".txt", ".json", ".md", ".csv", ".yml", ".yaml", ".xml", ".html", ".log", ".sql":
		return true
	}
	return false
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

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
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
				Types:    rpt.Hits,
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
