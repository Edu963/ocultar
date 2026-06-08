package scrubber

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/Edu963/ocultar/internal/pii"
	"github.com/Edu963/ocultar/vault"
)

// art9Tag is prepended to lines containing health-related data.
const art9Tag = "[ART9_HEALTH_RECORD]"

// Scrubber holds a reference to the vault and the encryption key.
type Scrubber struct {
	vault     vault.Provider
	masterKey []byte
	engine    *pii.Refinery
}

// New creates a Scrubber backed by the same vault as the gateway.
func New(v vault.Provider, masterKey []byte) *Scrubber {
	return &Scrubber{
		vault:     v,
		masterKey: masterKey,
		engine:    pii.NewRefinery(),
	}
}

// Prescrub runs the rigorous deterministic EU+UK detection engine over the text.
func (s *Scrubber) Prescrub(text string) (string, error) {
	// Find all detections
	detections := s.engine.Scan(text)
	if len(detections) == 0 {
		return text, nil
	}

	// Identify lines containing health-related data for Article 9 tagging
	healthLines := make(map[int]bool)
	var lineOffsets []int
	lineOffsets = append(lineOffsets, 0)
	for i, ch := range text {
		if ch == '\n' {
			lineOffsets = append(lineOffsets, i+1)
		}
	}

	for _, d := range detections {
		if d.Entity == "HEALTH_ENTITY" {
			lineIdx := 0
			for i, off := range lineOffsets {
				if d.Range.Start >= off {
					lineIdx = i
				} else {
					break
				}
			}
			healthLines[lineIdx] = true
		}
	}

	// Redact using the centralized engine
	redactedText, err := s.engine.Redact(text, func(d pii.DetectionResult) (string, error) {
		return s.storeToken(d.Value, d.Entity)
	})
	if err != nil {
		return "", fmt.Errorf("engine redact: %w", err)
	}

	// Prepend `art9Tag` to lines that had a HEALTH_ENTITY
	if len(healthLines) > 0 {
		lines := strings.Split(redactedText, "\n")
		for i := range lines {
			if healthLines[i] {
				if !strings.HasPrefix(lines[i], art9Tag) {
					lines[i] = art9Tag + " " + lines[i]
				}
			}
		}
		redactedText = strings.Join(lines, "\n")
	}

	return redactedText, nil
}

func (s *Scrubber) storeToken(value string, piiType string) (string, error) {
	h := fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
	token := fmt.Sprintf("[%s_%s]", piiType, h[:8])
	encrypted, err := encrypt([]byte(value), s.masterKey)
	if err != nil {
		return "", err
	}
	if _, err := s.vault.StoreToken(h, token, encrypted); err != nil {
		return "", err
	}
	return token, nil
}
