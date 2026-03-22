package pii_test

import (
	"testing"
	"github.com/Edu963/ocultar/internal/pii"
)

func TestEngineEvasionResistance(t *testing.T) {
	eng := pii.NewEngine()

	cases := []struct {
		name       string
		input      string
		expectHit  string
	}{
		{"Spaced IBAN", "Konto DE89 3704 0044 0532 0130 00 is mine", "DE89 3704 0044 0532 0130 00"},
		{"Mixed Case IBAN", "iban de89370400440532013000.", "de89370400440532013000"},
		{"Credit Card with Dashes", "Card 4111-1111-1111-1111", "4111-1111-1111-1111"}, // Valid test visa
		{"Boundary Enforcement", "NotAnEU_VATGB1234567890Inside", ""}, // Should not match due to boundaries
		{"Valid EU VAT", "VAT is GB123456789", "GB123456789"},
		{"Valid BE VAT", "VAT is BE0123456789", "BE0123456789"}, 
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := eng.Scan(tc.input)
			t.Logf("Scan results for %q: %+v", tc.input, res)
			
			// Let's also print what the regex purely matched for debugging
			for _, def := range pii.Registry {
				match := def.Pattern.FindAllStringIndex(tc.input, -1)
				if len(match) > 0 {
					matchedStr := tc.input[match[0][0]:match[0][1]]
					t.Logf("Regex %s matched: %q", def.Type, matchedStr)
				}
			}

			found := ""
			if len(res) > 0 {
				found = res[0].Value
			}
			if found != tc.expectHit {
				t.Errorf("Expected %q, got %q", tc.expectHit, found)
			}
		})
	}
}

func TestValidationLayer(t *testing.T) {
	eng := pii.NewEngine()

	// Invalid IBAN should hit nothing
	res := eng.Scan("DE89370400440532013001") // Modified digit
	if len(res) > 0 {
		t.Errorf("Expected invalid IBAN to fail checksum, got hit: %v", res[0].Value)
	}

	// Valid IBAN
	res = eng.Scan("DE89370400440532013000")
	if len(res) != 1 {
		t.Fatalf("Expected valid IBAN to pass, got %d hits", len(res))
	}
	if len(res[0].Method) != 2 || res[0].Method[1] != "checksum" {
		t.Errorf("Expected method to include checksum, got %v", res[0].Method)
	}
}
