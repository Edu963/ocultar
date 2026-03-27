package refinery

import (
	"testing"
)

type DomainMockScanner struct {
	LastDomain string
	Available  bool
}

func (d *DomainMockScanner) ScanForPII(text string) (map[string][]string, error) {
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

func TestDomainSwapping(t *testing.T) {
	mockData := &DomainMockScanner{}
	eng := NewRefinery(&MockVault{}, []byte("01234567890123456789012345678901"))
	eng.AIScanner = mockData

	// Case 1: Standard Domain
	eng.AIScanner.SetDomain("standard")
	res, _ := eng.RefineString("Patient name is John Doe", "tester", nil)
	if mockData.LastDomain != "standard" {
		t.Errorf("Expected domain standard, got %s", mockData.LastDomain)
	}

	// Case 2: Clinical Domain
	eng.AIScanner.SetDomain("clinical")
	res, _ = eng.RefineString("Patient MRN is MRN-123456", "tester", nil)
	if !containsToken(res) {
		t.Errorf("Expected clinical entity to be tokenised, got: %s", res)
	}
}

func containsToken(s string) bool {
	return tokenPattern.MatchString(s)
}
