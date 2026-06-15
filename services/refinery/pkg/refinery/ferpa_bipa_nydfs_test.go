package refinery

import (
	"strings"
	"testing"

	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/vault"
)

// setupComplianceRefinery creates a temporary vault and refinery for compliance tests.
func setupComplianceRefinery(t *testing.T) *Refinery {
	t.Helper()
	config.InitDefaults()
	v, err := vault.New(config.Global, t.TempDir()+"/vault.db")
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	t.Cleanup(func() { v.Close() })
	return NewRefinery(v, []byte("0123456789abcdef0123456789abcdef"))
}

// assertRedacted verifies that none of the given PII substrings appear in the refined output.
func assertRedacted(t *testing.T, e *Refinery, input string, mustNotContain ...string) {
	t.Helper()
	out, err := e.RefineString(input, "compliance-test", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	for _, pii := range mustNotContain {
		if strings.Contains(out, pii) {
			t.Errorf("PII leak: %q still present in output\ninput:  %q\noutput: %q", pii, input, out)
		}
	}
}

// ── FERPA — Family Educational Rights and Privacy Act ────────────────────────
// FERPA protects student education records. Key PII: student IDs, GPA, grades,
// disciplinary records, and any data that would identify a student.

func TestFERPA_StudentID(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Student ID S-20230147 has a GPA of 3.8 and email jsmith@university.edu",
		"S-20230147", "jsmith@university.edu",
	)
}

func TestFERPA_GradeWithName(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Emily Johnson received an A in CHEM-201. Contact: emily.johnson@uni.edu",
		"emily.johnson@uni.edu",
	)
}

func TestFERPA_EnrollmentRecord(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Student 445-89-1234 enrolled in Spring 2024. DOB: 03/15/2001. Phone: 555-234-5678",
		"445-89-1234", "03/15/2001", "555-234-5678",
	)
}

// ── BIPA — Biometric Information Privacy Act ──────────────────────────────────
// BIPA (Illinois) governs collection and storage of biometric identifiers:
// fingerprints, retina/iris scans, voiceprints, facial geometry, hand geometry.
// Key PII: any data described as biometric alongside identifying information.

func TestBIPA_BiometricWithName(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Employee Carlos Rivera (ID: EMP-00442) enrolled fingerprint scan on 2024-01-15. Email: c.rivera@corp.com",
		"c.rivera@corp.com",
	)
}

func TestBIPA_FacialGeometryRecord(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Facial geometry template for Maria Gonzalez, SSN 567-89-0123, stored in biometric vault.",
		"567-89-0123",
	)
}

func TestBIPA_VoiceprintWithPhone(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Voiceprint registered for David Park, phone 312-867-5309, email d.park@mail.com",
		"312-867-5309", "d.park@mail.com",
	)
}

// ── NYDFS — New York Department of Financial Services (23 NYCRR 500) ─────────
// NYDFS cybersecurity regulation covers financial institutions. Key PII:
// financial account numbers, SSNs in financial contexts, credit card numbers,
// and any NPI (Nonpublic Personal Information) used by covered entities.

func TestNYDFS_AccountNumberWithSSN(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Account 4400-1234-5678-9012 belongs to SSN holder 123-45-6789. Routing: 021000021",
		"123-45-6789",
	)
}

func TestNYDFS_CreditCardInFinancialReport(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Transaction on card 4111-1111-1111-1111 for customer alice@bank.com, IBAN GB29NWBK60161331926819",
		"4111-1111-1111-1111", "alice@bank.com", "GB29NWBK60161331926819",
	)
}

func TestNYDFS_BrokerageAccountWithEmail(t *testing.T) {
	e := setupComplianceRefinery(t)
	assertRedacted(t, e,
		"Brokerage account B-9934421 for Robert Chen, robert.chen@invest.com, SSN 987-65-4321",
		"robert.chen@invest.com", "987-65-4321",
	)
}
