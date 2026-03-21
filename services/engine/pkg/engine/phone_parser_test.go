package engine

import (
	"testing"
)

func TestParseAndReplacePhones(t *testing.T) {
	replaceFn := func(match string) string {
		return "[PHONE_REDACTED]"
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// EU Formats
		{"French Mobile", "Call me at +33 6 12 34 56 78 please", "Call me at [PHONE_REDACTED] please"},
		{"French Local", "My number is 06 12 34 56 78", "My number is [PHONE_REDACTED]"},
		{"German", "Support: +49 151 23456789", "Support: [PHONE_REDACTED]"},
		{"Spanish", "Contacta al 612 345 678", "Contacta al [PHONE_REDACTED]"},

		// LATAM Formats
		{"Brazilian Mobile", "Zap: +55 11 91234-5678", "Zap: [PHONE_REDACTED]"},
		{"Mexican", "Tel: 55 1234 5678", "Tel: [PHONE_REDACTED]"},
		{"Argentine", "Llamá al 11 4321-1234", "Llamá al [PHONE_REDACTED]"},
		{"Colombian", "Cel: 300 123 4567", "Cel: [PHONE_REDACTED]"},

		// North American Formats
		{"US Standard", "My office: (555) 123-4567", "My office: [PHONE_REDACTED]"},
		{"US Country Code", "Dial +1-800-555-0199", "Dial [PHONE_REDACTED]"},

		// Edge Cases
		{"Already Tokenized", "Contact [EMAIL_12345] at exactly 12345678", "Contact [EMAIL_12345] at exactly 12345678"}, // The 8 digits might not parse as valid in our regions if too short, let's make it a valid one but overlapping isn't checked here directly since tokenPattern wouldn't match 12345678.
		{"Overlapping Token", "Contact [PHONE_abc123]", "Contact [PHONE_abc123]"},
		{"ISO Date Not Phone", "Event on 2026-02-27 and time", "Event on 2026-02-27 and time"},
		{"Embedded in Text", "You can reach methrough+447911123456if needed", "You can reach methrough[PHONE_REDACTED]if needed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseAndReplacePhones(tc.input, replaceFn)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}
