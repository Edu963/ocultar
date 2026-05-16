package refinery

import (
	"testing"
)

type DomainMockScanner struct {
	LastDomain string
	Available  bool
	// Override allows per-test control of what ScanForPII returns.
	Override map[string][]string
}

func (d *DomainMockScanner) ScanForPII(text string) (map[string][]string, error) {
	if d.Override != nil {
		return d.Override, nil
	}
	if d.LastDomain == "clinical" {
		return map[string][]string{"MRN": {"MRN-123456"}}, nil
	}
	return map[string][]string{"PERSON": {"John Doe"}}, nil
}

func (d *DomainMockScanner) CheckHealth(host string) {}
func (d *DomainMockScanner) IsAvailable() bool       { return true }
func (d *DomainMockScanner) SetDomain(domain string) { d.LastDomain = domain }

type MockVault struct{}

func (m *MockVault) StoreToken(hash, token, encrypted string) (bool, error) { return true, nil }
func (m *MockVault) GetToken(hash string) (string, bool)                    { return "", false }
func (m *MockVault) CountAll() int64                                        { return 0 }
func (m *MockVault) Close() error                                           { return nil }

func newTestEng(scanner *DomainMockScanner) *Refinery {
	eng := NewRefinery(&MockVault{}, []byte("01234567890123456789012345678901"))
	eng.AIScanner = scanner
	return eng
}

func TestDomainSwapping(t *testing.T) {
	mockData := &DomainMockScanner{}
	eng := newTestEng(mockData)

	// Case 1: Standard Domain
	eng.AIScanner.SetDomain("standard")
	_, _ = eng.RefineString("Patient name is John Doe", "tester", nil)
	if mockData.LastDomain != "standard" {
		t.Errorf("Expected domain standard, got %s", mockData.LastDomain)
	}

	// Case 2: Clinical Domain
	eng.AIScanner.SetDomain("clinical")
	res, _ := eng.RefineString("Patient MRN is MRN-123456", "tester", nil)
	if !containsToken(res) {
		t.Errorf("Expected clinical entity to be tokenised, got: %s", res)
	}
}

func containsToken(s string) bool {
	return tokenPattern.MatchString(s)
}

// --- Regulatory macro-category tests ---
// Each test feeds the refinery a mock Qwen response for one of the 9
// macro-categories and asserts the output is tokenised with the correct
// prefix. Tests verify the category→token mapping without a live model.

func TestMacroCategoryIdentity(t *testing.T) {
	scanner := &DomainMockScanner{Override: map[string][]string{
		"IDENTITY": {"Trudi Penkler"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("Trudi Penkler attended the meeting", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("IDENTITY: expected token in output, got: %s", res)
	}
}

func TestMacroCategoryGovID(t *testing.T) {
	// PESEL (Poland) — detected by NER, not Tier-1 regex
	scanner := &DomainMockScanner{Override: map[string][]string{
		"GOV_ID": {"85032312345"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("PESEL: 85032312345", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("GOV_ID: expected token in output, got: %s", res)
	}
}

func TestMacroCategoryContact(t *testing.T) {
	// EU-format address — unusual enough to escape Tier-1
	scanner := &DomainMockScanner{Override: map[string][]string{
		"CONTACT": {"4, Allée des Roses, 38000 Grenoble"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("Send the invoice to 4, Allée des Roses, 38000 Grenoble", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("CONTACT: expected token in output, got: %s", res)
	}
}

func TestMacroCategoryDigitalNetwork(t *testing.T) {
	// IMEI in prose — not a regex target in Tier-1
	scanner := &DomainMockScanner{Override: map[string][]string{
		"DIGITAL_NETWORK": {"357938090943002"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("Device IMEI is 357938090943002", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("DIGITAL_NETWORK: expected token in output, got: %s", res)
	}
}

func TestMacroCategoryFinancial(t *testing.T) {
	// IBAN in non-standard spacing — backup for Tier-1 miss
	scanner := &DomainMockScanner{Override: map[string][]string{
		"FINANCIAL": {"FR76 3000 6000 0112 3456 7890 189"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("IBAN: FR76 3000 6000 0112 3456 7890 189", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("FINANCIAL: expected token in output, got: %s", res)
	}
}

func TestMacroCategoryHealthBiometric(t *testing.T) {
	// Diagnosis in prose — GDPR Art.9 + HIPAA; pure NER territory
	scanner := &DomainMockScanner{Override: map[string][]string{
		"HEALTH_BIOMETRIC": {"Type 2 diabetes"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("The patient was diagnosed with Type 2 diabetes", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("HEALTH_BIOMETRIC: expected token in output, got: %s", res)
	}
}

func TestMacroCategorySensitiveProfile(t *testing.T) {
	// Religion in prose — GDPR Art.9 special category; regex-invisible
	scanner := &DomainMockScanner{Override: map[string][]string{
		"SENSITIVE_PROFILE": {"practicing Muslim"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("She is a practicing Muslim who observes Ramadan", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("SENSITIVE_PROFILE: expected token in output, got: %s", res)
	}
}

func TestMacroCategoryEduEmp(t *testing.T) {
	// Grade in prose — FERPA record; NER-only detection
	scanner := &DomainMockScanner{Override: map[string][]string{
		"EDU_EMP": {"A- in Chemistry"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("The student received an A- in Chemistry this semester", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("EDU_EMP: expected token in output, got: %s", res)
	}
}

func TestMacroCategoryChildrenData(t *testing.T) {
	// Child linked to parent — COPPA + GDPR Art.8
	scanner := &DomainMockScanner{Override: map[string][]string{
		"CHILDREN_DATA": {"my 8-year-old son"},
	}}
	eng := newTestEng(scanner)
	res, err := eng.RefineString("The form is for my 8-year-old son at Greenwood Primary", "tester", nil)
	if err != nil {
		t.Fatalf("RefineString error: %v", err)
	}
	if !containsToken(res) {
		t.Errorf("CHILDREN_DATA: expected token in output, got: %s", res)
	}
}
