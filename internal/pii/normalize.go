package pii

import "strings"

// Normalize removes all spaces, hyphens, and dots, and converts to uppercase.
// It is useful for uniform matching and validation across different PII entities.
func Normalize(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ".", "")
	return s
}
