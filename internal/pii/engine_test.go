package pii_test

import (
	"testing"
	"github.com/ocultar-dev/ocultar/internal/pii"
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
		// Cost center / internal GL codes
		{"Cost Center CORP",    "Transfert interne vers 6837-CORP-891 pour un montant de 78709.65 EUR.", "6837-CORP-891"},
		{"Cost Center EMEA",    "Transfert interne vers 1076-EMEA-824 pour un montant de 56198.09 EUR.", "1076-EMEA-824"},
		{"Cost Center US",      "Note de frais de Jean Martin, centre de coût 5866-US-432.", "5866-US-432"},
		{"Cost Center no match", "Ref 12-XY-3 is too short", ""},
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

func TestP0TierOnePatches(t *testing.T) {
	eng := pii.NewRefinery()

	hits := []struct{ name, input string }{
		// IPv6
		{"IPv6 full 8-group", "Device IP: 2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"IPv6 keyword compressed", "The ipv6: fe80:0000:0000:0000:0000:0000:0000:0001"},
		// MAC address
		{"MAC colons", "Interface MAC: AA:BB:CC:DD:EE:FF"},
		{"MAC hyphens", "MAC-48: 00-1A-2B-3C-4D-5E"},
		// VIN
		{"VIN keyword", "VIN 1HGBH41JXMN109186"},
		{"VIN vehicle id", "vehicle identification number 1HGBH41JXMN109186"},
		// License plate
		{"License plate", "license plate: ABC1234"},
		// Vehicle registration
		{"Vehicle reg no", "registration no. AB12CDE"},
		// Card CVV/expiry
		{"Card CVV", "CVV: 123"},
		{"Card CVC2", "cvc2: 4567"},
		{"Card expiry slash", "expiry: 09/26"},
		{"Card expiry valid thru", "valid thru 12/2028"},
		// Crypto wallets
		{"ETH wallet", "Send to 0x742d35Cc6634C0532925a3b844Bc454e4438f44e"},
		{"BTC wallet keyword", "bitcoin: 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"},
		// SSH key
		{"SSH RSA private key", "-----BEGIN RSA PRIVATE KEY-----"},
		{"SSH OPENSSH private key", "-----BEGIN OPENSSH PRIVATE KEY-----"},
		// Device advertising ID
		{"IDFA keyword", "idfa: 550e8400-e29b-41d4-a716-446655440000"},
		{"GAID keyword", "gaid: 550e8400-e29b-41d4-a716-446655440000"},
		// Date of birth
		{"DOB ISO", "DOB: 1985-06-15"},
		{"born on DMY", "born on 15/06/1985"},
		{"date of birth", "date of birth: June 15, 1985"},
		// Clinical dates
		{"Admitted date", "admitted: 2024-03-01"},
		{"Discharge date", "discharge date: 2024-03-10"},
		{"Date of death", "date of death: 2024-03-15"},
		{"DOD", "DOD: 2024-03-15"},
		// GPS coordinates
		{"GPS degree symbol", "Location: 48° N, 2° E"},
		{"GPS keyword decimal", "gps: 48.8566, 2.3522"},
		{"GPS lat keyword", "lat: 48.8566, 2.3522"},
		// ZIP / postal code
		{"ZIP code 5-digit", "zip code: 90210"},
		{"ZIP code 9-digit", "zip: 10001-4321"},
		{"Postal code", "postal code: 10001"},
		// Employee ID
		{"Employee ID alphanumeric", "employee id: EMP-12345"},
		{"Badge number", "badge #: 98765"},
		// Case / docket number
		{"Case no alphanumeric", "case no. CV-2024-001"},
		{"Docket number", "docket no.: 24-CR-456"},
	}

	misses := []struct{ name, input string }{
		{"Plain 5-digit invalid FR dept", "Reference: 00210"}, // dept 00 is not a valid French department
		{"Short ETH hex not wallet", "Color: 0x742d35"},
		{"VIN without keyword", "code 1HGBH41JXMN109186 has no keyword"},
	}

	for _, tc := range hits {
		t.Run(tc.name, func(t *testing.T) {
			res := eng.Scan(tc.input)
			if len(res) == 0 {
				t.Errorf("MISS: expected hit in %q, got none", tc.input)
			}
		})
	}

	for _, tc := range misses {
		t.Run("no_match/"+tc.name, func(t *testing.T) {
			res := eng.Scan(tc.input)
			if len(res) != 0 {
				t.Errorf("FALSE POSITIVE: expected no hit in %q, got %+v", tc.input, res)
			}
		})
	}
}

// TestFreeinvoiceLeakRegression pins the five PII leaks found during the Freebox invoice
// test (2026-05-22). Each case must produce at least one detection hit.
func TestFreeinvoiceLeakRegression(t *testing.T) {
	eng := pii.NewRefinery()

	hits := []struct {
		name      string
		input     string
		wantType  string
	}{
		// Bug B — partial IBAN (bank already starred last 6 digits); keyword-gated rule fires
		{
			"partial IBAN after compte bancaire",
			"prélevé sur le compte bancaire FR76 1390 6001 0083 1172 45****** à partir du",
			"IBAN",
		},
		// Bug C — French subscriber ID must use FR_SUBSCRIBER_ID, not NL_BSN
		{
			"FR subscriber ID keyword-gated",
			"Identifiant abonné : 123456789",
			"FR_SUBSCRIBER_ID",
		},
		// Bug D — French VAT with spaces must match FRANCE_VAT
		{
			"FR VAT spaced format",
			"No de TVA intracommunautaire : FR 12 345 678 901",
			"FRANCE_VAT",
		},
		// Bug E — French postal code with "code postal" keyword
		{
			"FR postal code keyword",
			"code postal: 38100",
			"FR_POSTAL_CODE",
		},
		// Bug F — fiber reference number must be caught as FIBER_REF, not misclassified as PHONE
		{
			"FR fiber ref keyword-gated",
			"Référence prise fibre : 0123456789012",
			"FIBER_REF",
		},
		// Bug H — standalone French postal codes in invoice footers (dept-range gate, no keyword)
		{
			"FR postal code bare 38100",
			"Free Service Abonné 38100 Grenoble",
			"FR_POSTAL_CODE",
		},
		{
			"FR postal code bare 75008",
			"Free SAS 75008 Paris",
			"FR_POSTAL_CODE",
		},
	}

	for _, tc := range hits {
		t.Run(tc.name, func(t *testing.T) {
			res := eng.Scan(tc.input)
			found := false
			for _, r := range res {
				if r.Entity == tc.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("MISS entity=%s in %q — got %+v", tc.wantType, tc.input, res)
			}
		})
	}

	// FR_SUBSCRIBER_ID must NOT fire without keyword context (avoid over-broad NL_BSN behaviour)
	t.Run("no_match/bare 9-digit without subscriber keyword", func(t *testing.T) {
		res := eng.Scan("reference: 123456789")
		for _, r := range res {
			if r.Entity == "FR_SUBSCRIBER_ID" {
				t.Errorf("FALSE POSITIVE: FR_SUBSCRIBER_ID should not fire without subscriber keyword, got %+v", r)
			}
		}
	})
	// FR_POSTAL_CODE dept-range gate must NOT fire on invalid French department prefix
	t.Run("no_match/postal code invalid dept 99001", func(t *testing.T) {
		res := eng.Scan("Reference: 99001")
		for _, r := range res {
			if r.Entity == "FR_POSTAL_CODE" {
				t.Errorf("FALSE POSITIVE: FR_POSTAL_CODE should not fire on invalid dept prefix 99, got %+v", r)
			}
		}
	})
}

func BenchmarkEngineScan(b *testing.B) {
	eng := pii.NewRefinery()
	input := "Hello, my name is John Doe and my CPF is 123.456.789-09 and my RUT is 12.345.678-5. My email is john@example.com."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eng.Scan(input)
	}
}
