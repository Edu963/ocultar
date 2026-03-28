package pii_test

import (
	"testing"
	"github.com/Edu963/ocultar/internal/pii"
)

func TestEngineEvasionResistance(t *testing.T) {
	eng := pii.NewRefinery()

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
		{"Valid ES DNI", "DNI: 12345678Z", "12345678Z"},
		{"Valid IT CF", "CF: RSSMRA80A10H501W", "RSSMRA80A10H501W"},
		{"Valid NL BSN", "BSN: 123456782", "123456782"},
		{"Valid PL PESEL", "PESEL: 44051401359", "44051401359"},
		{"Valid DE StId", "StId: 65929970489", "65929970489"},
		{"Base64 Evasion", "Data: REU4OTM3MDQwMDQ0MDUzMjAxMzAwMA==", ""}, // Base64 should be handled by refinery, not raw engine
		{"Multi-line PII", "FR_NIR is\n190017500100112", "190017500100112"},
		{"Valid BR CPF", "CPF: 123.456.789-09", "123.456.789-09"},
		{"Valid CL RUT", "RUT: 12.345.678-5", "12.345.678-5"},
		{"Valid CL RUT with K", "RUT: 16.222.333-K", "16.222.333-K"},
		{"Standard Email", "contact john.doe@example.com for info", "john.doe@example.com"},
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
	eng := pii.NewRefinery()

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

	// Valid NL BSN should pass
	res = eng.Scan("123456782")
	if len(res) == 0 {
		t.Errorf("Expected valid NL BSN to pass, got no hits")
	}

	// Invalid IT CF should fail
	res = eng.Scan("RSSMRA80A01H501X")
	if len(res) > 0 {
		t.Errorf("Expected invalid IT CF to fail, got hit")
	}

	// Valid IT CF should pass
	res = eng.Scan("RSSMRA80A10H501W")
	if len(res) == 0 {
		t.Errorf("Expected valid IT CF to pass, got no hits")
	}

	// Valid PL PESEL should pass
	res = eng.Scan("44051401359")
	if len(res) == 0 {
		t.Errorf("Expected valid PL PESEL to pass, got no hits")
	}

	// Valid DE Steuer-ID should pass
	res = eng.Scan("65929970489")
	if len(res) == 0 {
		t.Errorf("Expected valid DE Steuer-ID to pass, got no hits")
	}

	// Valid Brazil CPF
	res = eng.Scan("12345678909")
	if len(res) == 0 {
		t.Errorf("Expected valid Brazil CPF to pass, got no hits")
	}
	// Invalid Brazil CPF
	res = eng.Scan("12345678900")
	if len(res) > 0 {
		t.Errorf("Expected invalid Brazil CPF to fail, got hit")
	}

	// Valid Chile RUT
	res = eng.Scan("123456785")
	if len(res) == 0 {
		t.Errorf("Expected valid Chile RUT to pass, got no hits")
	}
	// Valid Chile RUT with K
	res = eng.Scan("16222333K")
	if len(res) == 0 {
		t.Errorf("Expected valid Chile RUT with K to pass, got no hits")
	}
	// Invalid Chile RUT
	res = eng.Scan("123456780")
	if len(res) > 0 {
		t.Errorf("Expected invalid Chile RUT to fail, got hit")
	}

	// Valid India Aadhaar
	res = eng.Scan("719543825004")
	if len(res) == 0 {
		t.Errorf("Expected valid India Aadhaar to pass, got no hits")
	}
	// Invalid India Aadhaar
	res = eng.Scan("361153152701")
	if len(res) > 0 {
		t.Errorf("Expected invalid India Aadhaar to fail, got hit")
	}

	// Valid Singapore ID (S)
	res = eng.Scan("S1234567D")
	if len(res) == 0 {
		t.Errorf("Expected valid Singapore ID (S) to pass, got no hits")
	}
	// Valid Singapore ID (F)
	res = eng.Scan("F1234567M")
	if len(res) == 0 {
		t.Errorf("Expected valid Singapore ID (F) to pass, got no hits")
	}
	// Invalid Singapore ID
	res = eng.Scan("S1234567A")
	if len(res) > 0 {
		t.Errorf("Expected invalid Singapore ID to fail, got hit: %+v", res)
	}
}

func BenchmarkEngineScan(b *testing.B) {
	eng := pii.NewRefinery()
	input := "Hello, my name is John Doe and my CPF is 123.456.789-09 and my RUT is 12.345.678-5. My email is john@example.com."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eng.Scan(input)
	}
}
