package refinery

import (
	"testing"
)

func TestParseAndReplaceAddresses(t *testing.T) {
	replaceFn := func(match string) string {
		return "[ADDRESS_REDACTED]"
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// LATAM
		{"Colombian Address 1", "We met at Calle 123 # 45-67 yesterday.", "We met at [ADDRESS_REDACTED] yesterday."},
		{"Colombian Address 2", "Shipping to Cra. 15 #10-20, Bogotá on Friday.", "Shipping to [ADDRESS_REDACTED], [ADDRESS_REDACTED] on Friday."}, // Matches address bounds, then city bounds

		// European
		{"French Address", "HQ is located at 12 rue de la paix 75001 Paris.", "HQ is located at [ADDRESS_REDACTED] [ADDRESS_REDACTED]."},
		{"German Address", "Please send to Hauptstraße 15, 10115 Berlin via post.", "Please send to [ADDRESS_REDACTED] [ADDRESS_REDACTED] via post."},
		{"Spanish Address", "Oficina en Calle Atocha 123 Madrid", "Oficina en [ADDRESS_REDACTED] [ADDRESS_REDACTED]"},

		// North American
		{"US Full", "They live at 1600 Pennsylvania Avenue, Washington, DC 20500", "They live at [ADDRESS_REDACTED]"},
		{"US Short", "Drop it off at 101 Main St.", "Drop it off at [ADDRESS_REDACTED]."},

		// Isolated Cities
		{"City Only", "I am traveling to Paris tomorrow.", "I am traveling to [ADDRESS_REDACTED] tomorrow."},
		{"City Only Accented", "Vivo en Bogotá desde ayer.", "Vivo en [ADDRESS_REDACTED] desde ayer."},

		// Edge Cases
		{"Already Tokenized", "Located at [ADDRESS_12345] building", "Located at [ADDRESS_12345] building"},
		{"False Positive Avoidance", "Calle is a nice person.", "Calle is a nice person."}, // Should not match Calle as an address without numbers
		{"False Positive Avoidance 2", "123 is a number.", "123 is a number."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseAndReplaceAddresses(tc.input, replaceFn)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func BenchmarkParseAndReplaceAddresses(b *testing.B) {
	replaceFn := func(match string) string {
		return "[ADDRESS_REDACTED]"
	}
	input := "Envíalo a Calle 11 # 22-33 en Bogotá o a la oficina en 101 Main St. Saludos a Héctor."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseAndReplaceAddresses(input, replaceFn)
	}
}
