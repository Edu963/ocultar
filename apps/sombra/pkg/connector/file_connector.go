package connector

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// FileConnector handles local file uploads (CSV, JSON, plain text).
// It is the "manual upload" path — users upload a bank statement, medical
// report, or any file, and Sombra scrubs it before sending to AI.
type FileConnector struct {
	name   string
	policy DataPolicy
}

// NewFileConnector creates a FileConnector with the given name and data policy.
func NewFileConnector(name string, policy DataPolicy) *FileConnector {
	return &FileConnector{name: name, policy: policy}
}

func (f *FileConnector) Name() string { return f.name }

func (f *FileConnector) Policy() DataPolicy { return f.policy }

// Fetch reads the file content. If req.RawBody is non-empty, it is used
// directly (in-memory upload). Otherwise req.SourceID is treated as a
// file path on the local filesystem.
func (f *FileConnector) Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error) {
	var body []byte
	var contentType string

	if len(req.RawBody) > 0 {
		body = req.RawBody
		contentType = req.ContentType
		if contentType == "" {
			contentType = detectContentType(body)
		}
	} else if req.SourceID != "" {
		data, err := os.ReadFile(req.SourceID)
		if err != nil {
			return nil, fmt.Errorf("file connector: read %q: %w", req.SourceID, err)
		}
		body = data
		contentType = detectContentType(data)
	} else {
		return nil, fmt.Errorf("file connector: no file body or source_id provided")
	}

	// Enforce size limit.
	if f.policy.MaxBodyBytes > 0 && int64(len(body)) > f.policy.MaxBodyBytes {
		return nil, fmt.Errorf("file connector: file size %d exceeds policy limit %d", len(body), f.policy.MaxBodyBytes)
	}

	// Normalise to plain text for the refinery.
	text, err := normaliseToText(body, contentType)
	if err != nil {
		return nil, fmt.Errorf("file connector: normalise: %w", err)
	}

	return &FetchResponse{
		ContentType: "text/plain",
		Body:        []byte(text),
		Metadata: map[string]string{
			"original_type": contentType,
			"source":        req.SourceID,
		},
	}, nil
}

// detectContentType guesses the MIME type from the first bytes.
func detectContentType(data []byte) string {
	// DOCX (and other Office Open XML) files are ZIP archives starting with PK magic bytes.
	if len(data) > 4 && data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	s := strings.TrimSpace(string(data))
	if len(s) == 0 {
		return "text/plain"
	}
	if s[0] == '{' || s[0] == '[' {
		return "application/json"
	}
	// Very basic CSV heuristic: first line has commas and no braces.
	if first, _, ok := strings.Cut(s, "\n"); ok && strings.Contains(first, ",") {
		return "text/csv"
	}
	return "text/plain"
}

// normaliseToText converts structured formats to plain text suitable for
// the OCULTAR refinery's string-based redaction pipeline.
func normaliseToText(data []byte, contentType string) (string, error) {
	switch {
	case strings.Contains(contentType, "wordprocessingml") || strings.Contains(contentType, "docx"):
		return normaliseDocx(data)
	case strings.Contains(contentType, "json"):
		return normaliseJSON(data)
	case strings.Contains(contentType, "csv"):
		return normaliseCSV(data)
	default:
		return string(data), nil
	}
}

// normaliseDocx extracts clean plain text from a .docx file in memory.
// It strips Word TOC field codes and other XML artifacts that cause
// false-positive PII detections (e.g. sequential TOC page numbers).
func normaliseDocx(data []byte) (string, error) {
	r, err := docx.ReadDocxFromMemory(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("docx parse error: %w", err)
	}
	doc := r.Editable()
	content := doc.GetContent()
	return cleanDocxText(content), nil
}

// cleanDocxText strips XML tags and Word-specific field codes from extracted text.
// Word's Table of Contents embeds field codes like `\o "1-3" \h \z \u` and
// PAGEREF sequences with numeric IDs. These look like phone numbers to the
// OCULTAR refinery and cause dozens of spurious vault writes.
var (
	wordFieldSwitchRe = regexp.MustCompile(`\\[a-zA-Z](?:\s+"[^"]*")?`)
	pageRefRe         = regexp.MustCompile(`(?i)(PAGEREF|HYPERLINK|TOC|_Toc)\s*[\w\d_]*`)
	multiBlankRe      = regexp.MustCompile(`\n{3,}`)
)

func cleanDocxText(s string) string {
	// 1. Strip XML tags
	var out strings.Builder
	inTag := false
	for _, ch := range s {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			out.WriteRune(ch)
		}
	}
	cleaned := out.String()

	// 2. Remove Word field switches (\o "1-3" \h \z \u etc.)
	cleaned = wordFieldSwitchRe.ReplaceAllString(cleaned, "")

	// 3. Remove PAGEREF / TOC / _Toc field codes and their numeric IDs
	cleaned = pageRefRe.ReplaceAllString(cleaned, "")

	// 4. Drop lines that are pure noise leftover from field codes
	lines := strings.Split(cleaned, "\n")
	onlyNoise := regexp.MustCompile(`^[\s\d\\/_\-\.\,\;\:\|\&\*\#\@\!\(\)\"\']+$`)
	var kept []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || onlyNoise.MatchString(trimmed) {
			continue
		}
		kept = append(kept, trimmed)
	}
	cleaned = strings.Join(kept, "\n")

	// 5. Collapse excessive blank lines
	cleaned = multiBlankRe.ReplaceAllString(cleaned, "\n\n")

	return strings.TrimSpace(cleaned)
}


// normaliseJSON pretty-prints JSON so each value appears on its own line,
// giving the refinery maximum context for PII detection.
func normaliseJSON(data []byte) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		// Not valid JSON — treat as plain text.
		return string(data), nil
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return string(data), nil
	}
	return string(pretty), nil
}

// normaliseCSV converts CSV rows into readable "key: value" lines.
func normaliseCSV(data []byte) (string, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		return string(data), nil // fallback to raw text
	}
	if len(records) < 2 {
		return string(data), nil
	}

	headers := records[0]
	var sb strings.Builder
	for i, row := range records[1:] {
		sb.WriteString(fmt.Sprintf("--- Record %d ---\n", i+1))
		for j, cell := range row {
			header := ""
			if j < len(headers) {
				header = headers[j]
			} else {
				header = fmt.Sprintf("Column_%d", j)
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", header, cell))
		}
		sb.WriteByte('\n')
	}
	return sb.String(), nil
}
