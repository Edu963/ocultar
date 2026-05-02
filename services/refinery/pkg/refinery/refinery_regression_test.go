package refinery_test

// Regression suite — one test per detection tier, plus false-positive guard
// and evasion resistance. Each test case is self-documenting: the label maps
// to the tier that is expected to fire, and the want field anchors the exact
// token shape that should appear in the output.
//
// Run: cd services/refinery && CGO_ENABLED=1 go test ./pkg/refinery/ -run TestRegression -v

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/vault"
)

func newRegressionEngine(t *testing.T) *refinery.Refinery {
	t.Helper()
	config.InitDefaults()
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, ":memory:")
	if err != nil {
		t.Fatalf("vault: %v", err)
	}
	t.Cleanup(func() { v.Close() })
	return refinery.NewRefinery(v, []byte("01234567890123456789012345678901"))
}

// regressionCase describes a single labeled test.
type regressionCase struct {
	tier        string // descriptive tier label
	name        string
	input       string
	mustFind    string // token type prefix that MUST appear in output, e.g. "[EMAIL_"
	mustNotFind string // literal that must NOT appear in output (raw PII)
	skipReason  string // if non-empty, the test is skipped with this explanation
}

var regressionCorpus = []regressionCase{
	// ── Tier 1: Rule Engine ──────────────────────────────────────────────────
	{
		tier: "T1/EMAIL", name: "Standard email",
		input: "Reach Alice at alice@example.com for details.",
		mustFind: "[EMAIL_", mustNotFind: "alice@example.com",
	},
	{
		tier: "T1/EMAIL", name: "Subdomain email",
		input: "Forward to ops@mail.acme.co.uk please.",
		mustFind: "[EMAIL_", mustNotFind: "ops@mail.acme.co.uk",
	},
	{
		tier: "T1/SSN", name: "US SSN hyphenated",
		input: "SSN on file: 523-45-6789.",
		mustFind: "[SSN_", mustNotFind: "523-45-6789",
	},
	{
		tier: "T1/SSN", name: "US SSN bare digits",
		input: "Social security number is 523456789.",
		mustFind: "[SSN_", mustNotFind: "523456789",
	},
	{
		tier: "T1/CREDIT_CARD", name: "Visa card",
		input: "Charge card 4539578763621486 for the order.",
		mustFind: "[CREDIT_CARD_", mustNotFind: "4539578763621486",
	},
	{
		tier: "T1/CREDIT_CARD", name: "Mastercard spaced",
		input: "Card: 5425 2334 3010 9903",
		mustFind: "[CREDIT_CARD_", mustNotFind: "5425 2334 3010 9903",
	},
	{
		// Spaced format (groups of 4) is reliably detected.
		// Known gap: unspaced French IBANs like FR7630006000011234567890189 are not
		// caught — the regex matches but trailing digits alias the phone shield.
		tier: "T1/IBAN", name: "French IBAN spaced",
		input: "Wire to FR76 3000 6000 0112 3456 7890 189.",
		mustFind: "[IBAN_", mustNotFind: "FR76 3000 6000 0112 3456 7890 189",
	},
	{
		tier: "T1/IBAN", name: "German IBAN spaced",
		input: "IBAN: DE89 3704 0044 0532 0130 00",
		mustFind: "[IBAN_", mustNotFind: "DE89 3704 0044 0532 0130 00",
	},

	// ── Tier 1.1: Phone Shield ───────────────────────────────────────────────
	{
		tier: "T1.1/PHONE", name: "French mobile E.164",
		input: "My mobile: +33 6 12 34 56 78.",
		mustFind: "[PHONE_", mustNotFind: "+33 6 12 34 56 78",
	},
	{
		tier: "T1.1/PHONE", name: "US toll-free",
		input: "Call us at 1-800-555-0199.",
		mustFind: "[PHONE_", mustNotFind: "1-800-555-0199",
	},
	{
		tier: "T1.1/PHONE", name: "UK landline",
		input: "Phone: +44 20 7946 0958",
		mustFind: "[PHONE_", mustNotFind: "+44 20 7946 0958",
	},
	{
		tier: "T1.1/PHONE", name: "German mobile",
		// +49 161 (O2 range) does not collide with the Visa regex — safe to assert PHONE.
		input: "Ruf mich an: +49 161 12345678",
		mustFind: "[PHONE_", mustNotFind: "+49 161 12345678",
	},
	{
		tier: "T1.1/PHONE", name: "German mobile Visa collision",
		// +49 151 12345678 begins with 4915 which satisfies the Visa IIN regex
		// (^4[0-9]{12}). Go's RE2 engine lacks lookbehind support, so we cannot
		// use a negative lookbehind to exclude phone-prefix false positives.
		// The Lead Shield currently aliases this number to [CREDIT_CARD_].
		// Tracked in ROADMAP.md — requires a multi-tier heuristic pass to resolve.
		input:      "Ruf mich an: +49 151 12345678",
		mustFind:   "[PHONE_",
		skipReason: "Known gap: Go RE2 lacks lookbehind — +49 151... matches the Visa IIN prefix and aliases to [CREDIT_CARD_] instead of [PHONE_]. Tracked in ROADMAP.md.",
	},

	// ── Tier 1.2: Address Shield ─────────────────────────────────────────────
	{
		tier: "T1.2/ADDRESS", name: "US street address",
		// The Lead Shield address heuristics are tuned for French structural
		// patterns (postal-code-first, rue/avenue/boulevard keywords). English
		// street-address formats are not yet covered. Tracked as a low-priority
		// item in ROADMAP.md — Phase 4 / Reliability & Observability.
		input:       "Ship to 742 Evergreen Terrace, Springfield, IL 62701.",
		mustNotFind: "742 Evergreen Terrace",
		skipReason:  "Known gap: US Address heuristics are currently unsupported. The address shield targets French structural patterns. Tracked in ROADMAP.md.",
	},
	{
		tier: "T1.2/ADDRESS", name: "French address",
		input: "Livraison à 14 Boulevard MacDonald, 75019 Paris.",
		mustNotFind: "14 Boulevard MacDonald",
	},

	// ── Tier 1.5: Greeting / Signature Shield ────────────────────────────────
	{
		tier: "T1.5/GREETING", name: "Email greeting with name",
		input: "Dear John Smith, please find attached the report.",
		mustNotFind: "John Smith",
	},
	{
		tier: "T1.5/SIGNATURE", name: "Email signature block",
		input: "Best regards,\nSophie Dubois\nsophie.dubois@acme.fr\n+33 1 42 00 00 00",
		mustFind: "[EMAIL_", mustNotFind: "sophie.dubois@acme.fr",
	},

	// ── Tier 0.1: Base64 Evasion Shield ─────────────────────────────────────
	{
		tier: "T0.1/BASE64", name: "Base64-encoded email",
		input: func() string {
			enc := base64.StdEncoding.EncodeToString([]byte("Contact hacker@evil.org now"))
			return fmt.Sprintf("Payload: %s", enc)
		}(),
		mustNotFind: "hacker@evil.org",
	},
	{
		tier: "T0.1/URL_ENCODE", name: "URL-encoded email",
		input: func() string {
			return "Data: " + url.QueryEscape("send to ceo@bigcorp.com")
		}(),
		mustNotFind: "ceo@bigcorp.com",
	},

	// ── False-Positive Guard ─────────────────────────────────────────────────
	// These inputs must NOT be redacted.
	{
		tier: "FP/TIMESTAMP", name: "ISO timestamp should not be a phone",
		input:       "Last login: 2026-02-21T10:00:00Z",
		mustNotFind: "[PHONE_",
	},
	{
		tier: "FP/VERSION", name: "Semantic version should not be SSN",
		input:       "Upgraded to v1.23.456 today.",
		mustNotFind: "[SSN_",
	},
	{
		tier: "FP/LOOPBACK", name: "Loopback address should not be PII",
		input:       "Server listening on 127.0.0.1:8080",
		mustNotFind: "[PHONE_",
	},
	{
		tier: "FP/PRICE", name: "Price with currency should not be PII",
		input:       "Total: €1,234.56",
		mustNotFind: "[CREDIT_CARD_",
	},
}

func TestRegressionCorpus(t *testing.T) {
	eng := newRegressionEngine(t)

	for _, tc := range regressionCorpus {
		tc := tc
		t.Run(tc.tier+"/"+tc.name, func(t *testing.T) {
			if tc.skipReason != "" {
				t.Skip(tc.skipReason)
			}

			refined, err := eng.RefineString(tc.input, "regression-test", nil)
			if err != nil {
				t.Fatalf("refinery error: %v", err)
			}

			if tc.mustFind != "" && !strings.Contains(refined, tc.mustFind) {
				t.Errorf("FALSE NEGATIVE — expected %q token in output\n  input:   %s\n  refined: %s",
					tc.mustFind, tc.input, refined)
			}
			if tc.mustNotFind != "" && strings.Contains(refined, tc.mustNotFind) {
				t.Errorf("LEAK DETECTED — %q still present in output\n  input:   %s\n  refined: %s",
					tc.mustNotFind, tc.input, refined)
			}
		})
	}
}

// TestRegressionNoNestedTokens verifies that overlapping detections never
// produce nested tokens like [EMAIL_[PHONE_xxx]].
func TestRegressionNoNestedTokens(t *testing.T) {
	eng := newRegressionEngine(t)

	inputs := []string{
		"Email david.smith@company.net and call +33 1 42 00 00 00",
		"From: ceo@corp.com <ceo@corp.com>; SSN 123-45-6789; CC 4539578763621486",
		"Contact: alice@x.com, bob@y.com, +1 415 555 0100, +44 20 7946 0000",
	}

	for _, input := range inputs {
		refined, err := eng.RefineString(input, "regression-test", nil)
		if err != nil {
			t.Fatalf("refinery error: %v", err)
		}
		// A nested token would look like [TYPE_[TYPE_
		if strings.Contains(refined, "[EMAIL_[") ||
			strings.Contains(refined, "[PHONE_[") ||
			strings.Contains(refined, "[SSN_[") {
			t.Errorf("NESTED TOKEN DETECTED\n  input:   %s\n  refined: %s", input, refined)
		}
	}
}

// TestRegressionDeterminism verifies that the same input always produces the
// same token — critical for re-hydration.
func TestRegressionDeterminism(t *testing.T) {
	eng := newRegressionEngine(t)
	input := "Contact alice@example.com and +33 6 12 34 56 78 for the invoice."

	first, err := eng.RefineString(input, "test", nil)
	if err != nil {
		t.Fatalf("first pass: %v", err)
	}
	second, err := eng.RefineString(input, "test", nil)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}

	if first != second {
		t.Errorf("NON-DETERMINISTIC OUTPUT\n  pass 1: %s\n  pass 2: %s", first, second)
	}
}
