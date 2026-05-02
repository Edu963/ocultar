package connector

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// extractText dispatches to the right extractor based on file extension.
// Plain text formats are returned as-is; Office and PDF formats are parsed.
func extractText(name string, data []byte) (string, error) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".docx":
		return extractZipXML(data, func(p string) bool {
			return p == "word/document.xml"
		}, "t") // w:t in OOXML — match by local name only
	case ".pptx":
		return extractZipXML(data, func(p string) bool {
			return strings.HasPrefix(p, "ppt/slides/slide") && strings.HasSuffix(p, ".xml")
		}, "t") // a:t in DrawingML
	case ".xlsx":
		return extractZipXML(data, func(p string) bool {
			return p == "xl/sharedStrings.xml"
		}, "t")
	case ".pdf":
		return extractPDF(data)
	default:
		return string(data), nil
	}
}

// extractZipXML opens an OOXML ZIP archive and concatenates all character data
// from XML elements whose local name matches elemName across every file that
// passes the pathFilter predicate.
func extractZipXML(data []byte, pathFilter func(string) bool, elemName string) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}

	var sb strings.Builder
	for _, f := range r.File {
		if !pathFilter(f.Name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		text, _ := xmlCharData(rc, elemName)
		rc.Close()
		if text != "" {
			sb.WriteString(text)
			sb.WriteByte('\n')
		}
	}
	return strings.TrimSpace(sb.String()), nil
}

// xmlCharData walks an XML stream and returns all character data found inside
// elements whose local name matches localName (namespace prefix ignored).
// Errors mid-stream are silently ignored — we return whatever was collected.
func xmlCharData(r io.Reader, localName string) (string, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false

	var sb strings.Builder
	depth := 0 // tracks nesting inside matching elements
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == localName {
				depth++
			}
		case xml.EndElement:
			if t.Name.Local == localName && depth > 0 {
				depth--
				sb.WriteByte(' ') // word boundary between adjacent runs
			}
		case xml.CharData:
			if depth > 0 {
				sb.Write(t)
			}
		}
	}
	return strings.TrimSpace(sb.String()), nil
}

// extractPDF extracts readable text from a PDF byte slice using a lightweight
// scanner that collects literal string operands from text-rendering operators.
// This handles most standard Latin-character PDFs without an external library.
// CID-keyed fonts and fully-encrypted PDFs produce empty or garbage output.
func extractPDF(data []byte) (string, error) {
	var sb strings.Builder
	i, n := 0, len(data)

	for i < n {
		switch data[i] {
		case '(':
			// Literal string: (text) — may appear in Tj, TJ, or annotation values.
			i++
			start := i
			depth := 1
			for i < n && depth > 0 {
				if data[i] == '\\' {
					i += 2 // skip escaped character
					continue
				}
				if data[i] == '(' {
					depth++
				} else if data[i] == ')' {
					depth--
				}
				i++
			}
			if depth == 0 {
				text := filterPrintable(data[start : i-1])
				if text != "" {
					sb.WriteString(text)
					sb.WriteByte(' ')
				}
			}
		case '<':
			// Hex string: <hexdigits> — used in some Type1 / CIDFont PDFs.
			// Decode pairs of hex digits to ASCII where printable.
			if i+1 < n && data[i+1] != '<' { // << is a dict, not a string
				i++
				for i < n && data[i] != '>' {
					if i+1 < n {
						hi, lo := hexVal(data[i]), hexVal(data[i+1])
						if hi >= 0 && lo >= 0 {
							b := byte(hi<<4 | lo)
							if b >= 0x20 && b < 0x7f {
								sb.WriteByte(b)
							}
							i += 2
							continue
						}
					}
					i++
				}
				sb.WriteByte(' ')
			} else {
				i++
			}
		default:
			i++
		}
	}

	return strings.TrimSpace(sb.String()), nil
}

func filterPrintable(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		if c >= 0x20 && c < 0x7f {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}
