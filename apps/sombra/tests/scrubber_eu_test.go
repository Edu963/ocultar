package connector_test

import (
	"strings"
	"testing"

	"github.com/Edu963/ocultar/apps/sombra/pkg/scrubber"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/vault"
)

// newTestScrubber creates an in-memory vault-backed scrubber for testing.
func newTestScrubber(t *testing.T) *scrubber.Scrubber {
	t.Helper()
	config.InitDefaults()
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
	if err != nil {
		t.Fatalf("vault init: %v", err)
	}
	t.Cleanup(func() { v.Close() })
	key := make([]byte, 32)
	return scrubber.New(v, key)
}

// assertRedacted checks that original PII is not present in output.
func assertRedacted(t *testing.T, output, pii, caseName string) {
	t.Helper()
	if strings.Contains(output, pii) {
		t.Errorf("[%s] PII not redacted — found %q in:\n%s", caseName, pii, output)
	}
}

// assertTokenPresent checks that a token type is present.
func assertTokenPresent(t *testing.T, output, tokenType, caseName string) {
	t.Helper()
	if !strings.Contains(output, "["+tokenType+"_") {
		t.Errorf("[%s] expected token [%s_...] not found in:\n%s", caseName, tokenType, output)
	}
}

// ─────────────────────────────────────────────────────────────
// Pass 1 — IBAN / BIC / Creditor Refs
// ─────────────────────────────────────────────────────────────

func TestPass1_IBAN(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name  string
		input string
		pii   string
	}{
		{"French IBAN", "IBAN: FR7630006000011234567890189", "FR7630006000011234567890189"},
		{"German IBAN", "Konto IBAN DE89370400440532013000", "DE89370400440532013000"},
		{"Luxembourg IBAN (PayPal)", "LU92ZZZ0000000000000000058", "LU92ZZZ0000000000000000058"},
		{"Spanish IBAN", "IBAN ES9121000418450200051332", "ES9121000418450200051332"},
		{"Italian IBAN", "IT60X0542811101000000123456", "IT60X0542811101000000123456"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.pii, tc.name)
		assertTokenPresent(t, out, "IBAN", tc.name)
	}
}

func TestPass1_CreditorRef(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name  string
		input string
		pii   string
	}{
		{"French creditor ref", "FR83ZZZ459654", "FR83ZZZ459654"},
		{"French tax ref", "FR46ZZZ005002", "FR46ZZZ005002"},
		{"French GEG ref", "FR37ZZZ002933", "FR37ZZZ002933"},
		{"BNP ERE ref", "FR94ERE001228", "FR94ERE001228"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.pii, tc.name)
	}
}

// ─────────────────────────────────────────────────────────────
// Pass 2 — Account numbers with EU label prefixes
// ─────────────────────────────────────────────────────────────

func TestPass2_AccountLabels(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name  string
		input string
		pii   string
	}{
		{"French 'n°' format", "Compte de Dépôt n° 83117245000", "83117245000"},
		{"English account_number:", "account_number: 883726194", "883726194"},
		{"German Konto-Nr (dash)", "Konto-Nr 12345678901", "12345678901"},
		{"Spanish número de cuenta", "número de cuenta 00490001512310001234", "00490001512310001234"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.pii, tc.name)
		assertTokenPresent(t, out, "ACCOUNT_NUMBER", tc.name)
	}
}

// ─────────────────────────────────────────────────────────────
// Pass 3 — Wire-transfer person names
// ─────────────────────────────────────────────────────────────

func TestPass3_WireTransferNames(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name  string
		input string
		pii   string
	}{
		{"French VIR INST vers", "VIR INST vers Emmanuelle Legoff Groceries", "Emmanuelle Legoff"},
		{"French WEB MONSIEUR", "WEB MONSIEUR Trejos Cabezas repayment", "Trejos"},
		{"French DE MADAME", "DE MADAME Muller Fanny birthday gift", "Muller"},
		// German wire transfers use the same VIR/WEB pattern via SEPA, not ÜBERWEISUNG literal
		{"French DE MONSIEUR (German sender)", "DE MONSIEUR Hans Mueller Miete", "Hans Mueller"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.pii, tc.name)
	}
}

// ─────────────────────────────────────────────────────────────
// Pass 4 — Terminal vendor personal names
// ─────────────────────────────────────────────────────────────

func TestPass4_TerminalVendorNames(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name  string
		input string
		pii   string
	}{
		{"SumUp mixed case", "SumUp *Gaspard Pelletier Saint Pierre", "Gaspard Pelletier"},
		{"SUMUP uppercase", "SUMUP *MICHAEL SABATIER", "MICHAEL SABATIER"},
		{"Zettle", "ZETTLE_*LEDDA ALBERTO VENEZIA", "LEDDA ALBERTO"},
		{"Square", "SQ *VINCENT BIJOUX PARIS", "VINCENT BIJOUX"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.pii, tc.name)
	}
}

// ─────────────────────────────────────────────────────────────
// Pass 5 — Article 9 Health data
// ─────────────────────────────────────────────────────────────

func TestPass5_Art9Health(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name    string
		input   string
		keyword string
	}{
		{"French pharmacy", "PAIEMENT PHARMACIE DU CENTRE GRENOBLE 24.00", "PHARMACIE"},
		{"French C.P.A.M.", "C.P.A.M. DE L ISERE 253350004466 10.00", "C.P.A.M."},
		{"English clinic", "CLINIQUE DE CHARTRES reference 258880", "CLINIQUE"},
		{"Anesthesiology", "ANESTHESIE CHART VOIRON 50.00", "ANESTHESIE"},
		{"German Apotheke", "APOTHEKE AM MARKT BERLIN 12.50", "APOTHEKE"},
		{"Spanish Farmacia", "FARMACIA CENTRAL MADRID 8.90", "FARMACIA"},
		{"Mutual health insurer", "MUTUELLE AESIO 10007640 SOIN 89.11", "MUTUELLE"},
		{"Psychologist", "PSYCHOLOGUE CABINET DR MARTIN 75.00", "PSYCHOLOGUE"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.keyword, tc.name)
		if !strings.Contains(out, "[ART9_HEALTH_RECORD]") {
			t.Errorf("[%s] expected [ART9_HEALTH_RECORD] tag in output: %s", tc.name, out)
		}
	}
}

// ─────────────────────────────────────────────────────────────
// Pass 6 — Transfer memo free-text
// ─────────────────────────────────────────────────────────────

func TestPass6_MemoFreeText(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name  string
		input string
		memo  string
	}{
		{"Divorce proceedings", "VIR INST vers Sabine Fournier Translation divorce documents", "divorce documents"},
		{"Personal note", "VIR INST vers Hector Eduardo Tre Ski trip vacation", "Ski trip vacation"},
		// Birthday: WEB MONSIEUR prefix — note 'bday' is word #4 after beneficiary words,
		// captured in the memo tail by the regex
		{"Birthday reference", "VIR INST vers Hector Tre Eddie bday and utilities", "bday"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.memo, tc.name)
	}
}

// ─────────────────────────────────────────────────────────────
// Pass 7 — Tax reference codes
// ─────────────────────────────────────────────────────────────

func TestPass7_TaxRefs(t *testing.T) {
	sc := newTestScrubber(t)
	cases := []struct {
		name  string
		input string
		pii   string
	}{
		{"French withholding ref", "NNFR46ZZZ0050021G76E0C665618PAS2A", "NNFR46ZZZ0050021G76E0C665618PAS2A"},
	}
	for _, tc := range cases {
		out, err := sc.Prescrub(tc.input)
		if err != nil {
			t.Fatalf("[%s] scrub error: %v", tc.name, err)
		}
		assertRedacted(t, out, tc.pii, tc.name)
	}
}

// ─────────────────────────────────────────────────────────────
// Integration — real-world bank statement snippet
// ─────────────────────────────────────────────────────────────

func TestIntegration_RealBankStatement(t *testing.T) {
	sc := newTestScrubber(t)
	input := `
Compte de Dépôt n° 83117245000
VIREMENT EN VOTRE FAVEUR
VATES TREJOS CABEZAS H
: 3,322.00

VIREMENT EMIS
VIR INST vers Emmanuelle Legoff Romain Gentil Campagne
: 100.00

PRELEVEMENT
C.P.A.M. DE L ISERE 253350004466
: 10.00

VIREMENT EMIS
VIR INST vers Sabine Fournier Translation divorce documents.
: 92.00

PRELEVEMENT
NNFR46ZZZ0050021G76E0C665618PAS2A
FR46ZZZ005002
: 92.00

SUMUP *GASPARD PELLETIER SAINT PIERRE
: 9.50

PAIEMENT
LU96ZZZ0000000000000000058
4RVJ224UH4WJC
: 21.99
`
	out, err := sc.Prescrub(input)
	if err != nil {
		t.Fatalf("prescrub error: %v", err)
	}

	noLeaks := []struct {
		pii  string
		desc string
	}{
		{"83117245000", "account number"},
		{"Emmanuelle Legoff", "third-party name"},
		{"divorce documents", "sensitive memo"},
		{"253350004466", "health insurance ref"},
		{"NNFR46ZZZ0050021G76E0C665618PAS2A", "tax reference"},
		{"LU96ZZZ0000000000000000058", "PayPal IBAN"},
		// Note: 4RVJ224UH4WJC is a 13-char PayPal merchant token — not a valid BIC or IBAN,
		// not personally identifiable, acceptable to leave unredacted at scrubber level.
		{"Gaspard Pelletier", "SumUp vendor name"},
	}

	for _, tc := range noLeaks {
		if strings.Contains(out, tc.pii) {
			t.Errorf("LEAK [%s]: %q still present in output", tc.desc, tc.pii)
		}
	}

	// Art. 9 flag must be present
	if !strings.Contains(out, "[ART9_HEALTH_RECORD]") {
		t.Error("Art. 9 health record tag missing from C.P.A.M. transaction")
	}

	t.Logf("Prescrubbed output:\n%s", out)
}
