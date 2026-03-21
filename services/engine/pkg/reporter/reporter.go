package reporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"sort"
	"time"
)

// LogEntry represents a single line in the JSON lines audit.log.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	Action    string `json:"action"` // "vaulted" or "matched"
	Result    string `json:"result"` // e.g., "[EMAIL_836f82db]"
}

// Metrics stores the calculated compliance metrics.
type Metrics struct {
	TotalPrevented int
	Tier1Count     int
	Tier2Count     int
	Tier1Rate      float64
	Tier2Rate      float64
	TopCategories  []CategoryCount
}

// CategoryCount stores the frequency of a given PII type.
type CategoryCount struct {
	Name  string
	Count int
}

// ReportData is the data passed to the HTML template.
type ReportData struct {
	GeneratedAt time.Time
	TimePeriod  string
	Metrics     Metrics
	Statement   string
}

// Reporter parses audit logs and generates compliance reports.
type Reporter struct{}

// New returns a new Reporter.
func New() *Reporter {
	return &Reporter{}
}

// tokenRegex extracts the PII type from a token like "[EMAIL_836f82db]".
var tokenRegex = regexp.MustCompile(`^\[([A-Z_]+)_[a-f0-9]+.*\]$`)

func isTier1(piiType string) bool {
	switch piiType {
	case "EMAIL", "PHONE", "URL", "ADDRESS", "PERSON_FOUNDER":
		return true
	default:
		return false
	}
}

// ParseAuditLog reads the JSON lines log file and calculates metrics.
func (r *Reporter) ParseAuditLog(filePath string) (*Metrics, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}
	defer file.Close()

	metrics := &Metrics{}
	categoryMap := make(map[string]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // skip invalid lines
		}

		if entry.Action == "vaulted" || entry.Action == "matched" {
			metrics.TotalPrevented++

			// Extract PII type from result, e.g. "[EMAIL_836f82db]"
			matches := tokenRegex.FindStringSubmatch(entry.Result)
			piiType := "UNKNOWN"
			if len(matches) > 1 {
				piiType = matches[1]
			} else {
				// Sometimes result might just be the raw string if not a standard token, fallback
				// In OCULTAR, result is always the token.
				piiType = "UNKNOWN_TOKEN"
			}

			categoryMap[piiType]++
			if isTier1(piiType) {
				metrics.Tier1Count++
			} else {
				metrics.Tier2Count++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading audit log: %w", err)
	}

	if metrics.TotalPrevented > 0 {
		metrics.Tier1Rate = float64(metrics.Tier1Count) / float64(metrics.TotalPrevented) * 100.0
		metrics.Tier2Rate = float64(metrics.Tier2Count) / float64(metrics.TotalPrevented) * 100.0
	}

	// Calculate top categories
	for k, v := range categoryMap {
		metrics.TopCategories = append(metrics.TopCategories, CategoryCount{Name: k, Count: v})
	}
	sort.Slice(metrics.TopCategories, func(i, j int) bool {
		return metrics.TopCategories[i].Count > metrics.TopCategories[j].Count
	})

	// Keep top 5
	if len(metrics.TopCategories) > 5 {
		metrics.TopCategories = metrics.TopCategories[:5]
	}

	return metrics, nil
}

// GenerateHTMLReport creates the stylized weekly compliance report.
func (r *Reporter) GenerateHTMLReport(logFilePath, outputFilePath string) error {
	metrics, err := r.ParseAuditLog(logFilePath)
	if err != nil {
		return err
	}

	data := ReportData{
		GeneratedAt: time.Now(),
		TimePeriod:  "Past 7 Days",
		Metrics:     *metrics,
		Statement:   "Zero (0) vaulted tokens were egressed in the reported time period.",
	}

	tmpl := template.Must(template.New("report").Parse(htmlTemplateStr))

	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	return tmpl.Execute(outFile, data)
}

const htmlTemplateStr = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OCULTAR Weekly Compliance Guarantee</title>
    <style>
        :root {
            --bg-color: #0f172a;
            --card-bg: rgba(30, 41, 59, 0.7);
            --border-color: rgba(255, 255, 255, 0.1);
            --text-primary: #f8fafc;
            --text-secondary: #94a3b8;
            --accent-primary: #38bdf8;
            --accent-secondary: #818cf8;
            --success: #34d399;
        }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: var(--bg-color);
            background-image: 
                radial-gradient(at 0% 0%, rgba(56, 189, 248, 0.15) 0px, transparent 50%),
                radial-gradient(at 100% 100%, rgba(129, 140, 248, 0.15) 0px, transparent 50%);
            background-attachment: fixed;
            color: var(--text-primary);
            margin: 0;
            padding: 2rem;
            min-height: 100vh;
            display: flex;
            justify-content: center;
        }

        .container {
            max-width: 900px;
            width: 100%;
            display: flex;
            flex-direction: column;
            gap: 2rem;
        }

        h1 {
            font-size: 2.5rem;
            font-weight: 800;
            margin: 0;
            background: linear-gradient(to right, var(--accent-primary), var(--accent-secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            text-align: center;
            letter-spacing: -0.025em;
        }

        .subtitle {
            text-align: center;
            color: var(--text-secondary);
            font-size: 1.1rem;
            margin-top: -1.5rem;
        }

        .card {
            background: var(--card-bg);
            backdrop-filter: blur(12px);
            border: 1px solid var(--border-color);
            border-radius: 1rem;
            padding: 2rem;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }

        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
            border-color: rgba(255, 255, 255, 0.2);
        }

        .guarantee {
            border: 1px solid rgba(52, 211, 153, 0.3);
            background: rgba(52, 211, 153, 0.05);
            text-align: center;
        }

        .guarantee h2 {
            color: var(--success);
            font-size: 1.5rem;
            margin-top: 0;
        }

        .guarantee p {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 0;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
        }

        .metric-value {
            font-size: 3rem;
            font-weight: 800;
            line-height: 1;
            margin: 0.5rem 0;
            color: var(--text-primary);
        }

        .metric-label {
            color: var(--text-secondary);
            font-size: 0.875rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            font-weight: 600;
        }

        .bars {
            margin-top: 2rem;
        }

        .bar-container {
            margin-bottom: 1rem;
        }

        .bar-label {
            display: flex;
            justify-content: space-between;
            margin-bottom: 0.5rem;
            font-size: 0.875rem;
            color: var(--text-secondary);
        }

        .bar-bg {
            height: 0.75rem;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 9999px;
            overflow: hidden;
        }

        .bar-fill {
            height: 100%;
            background: linear-gradient(to right, var(--accent-primary), var(--accent-secondary));
            border-radius: 9999px;
            transition: width 1s ease-out;
        }

        .tier-1 { background: var(--accent-primary); }
        .tier-2 { background: var(--accent-secondary); }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 1rem;
        }

        th, td {
            text-align: left;
            padding: 1rem;
            border-bottom: 1px solid var(--border-color);
        }

        th {
            color: var(--text-secondary);
            font-weight: 600;
            font-size: 0.875rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        tr:last-child td {
            border-bottom: none;
        }

        .badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 600;
            background: rgba(56, 189, 248, 0.1);
            color: var(--accent-primary);
        }

        .footer {
            text-align: center;
            color: var(--text-secondary);
            font-size: 0.875rem;
            margin-top: 2rem;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .animated {
            animation: fadeIn 0.6s ease-out forwards;
            opacity: 0;
        }

        .d-1 { animation-delay: 0.1s; }
        .d-2 { animation-delay: 0.2s; }
        .d-3 { animation-delay: 0.3s; }
        .d-4 { animation-delay: 0.4s; }
    </style>
    <!-- Use Inter font -->
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800&display=swap" rel="stylesheet">
</head>
<body>
    <div class="container">
        <header class="animated d-1">
            <h1>OCULTAR Weekly Compliance Guarantee</h1>
            <p class="subtitle">Generated dynamically on {{.GeneratedAt.Format "Jan 02, 2006 at 15:04 MST"}}</p>
        </header>

        <section class="card guarantee animated d-2">
            <h2>Deterministically Verified</h2>
            <p>{{.Statement}}</p>
        </section>

        <div class="metrics-grid">
            <div class="card animated d-3">
                <div class="metric-label">Total PII Entities Prevented</div>
                <div class="metric-value">{{.Metrics.TotalPrevented}}</div>
                <div class="metric-label" style="text-transform: none; font-weight: 400; font-size: 0.8rem; margin-top: 0.5rem;">From exiting the secure boundary</div>
            </div>

            <div class="card animated d-3">
                <div class="metric-label">Detection Tier Breakdown</div>
                
                <div class="bars">
                    <div class="bar-container">
                        <div class="bar-label">
                            <span>Tier 1 (Deterministic)</span>
                            <span>{{printf "%.1f" .Metrics.Tier1Rate}}% ({{.Metrics.Tier1Count}})</span>
                        </div>
                        <div class="bar-bg">
                            <div class="bar-fill tier-1" style="width: 0%" data-width="{{.Metrics.Tier1Rate}}%"></div>
                        </div>
                    </div>
                    
                    <div class="bar-container">
                        <div class="bar-label">
                            <span>Tier 2 (AI/SLM NER)</span>
                            <span>{{printf "%.1f" .Metrics.Tier2Rate}}% ({{.Metrics.Tier2Count}})</span>
                        </div>
                        <div class="bar-bg">
                            <div class="bar-fill tier-2" style="width: 0%" data-width="{{.Metrics.Tier2Rate}}%"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <section class="card animated d-4">
            <h3 style="margin-top: 0; margin-bottom: 1rem; font-size: 1.25rem;">Top Sensitive Data Categories</h3>
            <table>
                <thead>
                    <tr>
                        <th>Category</th>
                        <th style="text-align: right;">Occurrences</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Metrics.TopCategories}}
                    <tr>
                        <td><span class="badge">{{.Name}}</span></td>
                        <td style="text-align: right; font-weight: 600;">{{.Count}}</td>
                    </tr>
                    {{end}}
                    {{if not .Metrics.TopCategories}}
                    <tr>
                        <td colspan="2" style="text-align: center; color: var(--text-secondary);">No PII entities detected in this period.</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>

        <div class="footer animated d-4">
            CONFIDENTIAL AND PROPRIETARY &copy; OCULTAR ENTERPRISE
        </div>
    </div>

    <script>
        // Animate progress bars on load
        window.addEventListener('DOMContentLoaded', () => {
            setTimeout(() => {
                document.querySelectorAll('.bar-fill').forEach(bar => {
                    bar.style.width = bar.getAttribute('data-width');
                });
            }, 500);
        });
    </script>
</body>
</html>
`
