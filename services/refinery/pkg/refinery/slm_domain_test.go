package refinery

import (
	"testing"
)

type DomainMockScanner struct {
	LastDomain string
	IsCalled   bool
}

type MockVault struct{}

func (m *MockVault) StoreToken(hash, token, encrypted string) (bool, error) { return true, nil }
func (m *MockVault) GetToken(hash string) (string, bool)                    { return "", false }
func (m *MockVault) CountAll() int64                                        { return 0 }
func (m *MockVault) Close() error                                           { return nil }

func (d *DomainMockScanner) ScanForPII(text string) (map[string][]string, error) {
	d.IsCalled = true
	// Simulate domain-specific responses
	if d.LastDomain == "clinical" {
		return map[string][]string{"MRN": {"MRN-123456"}}, nil
	}
	if d.LastDomain == "fintech" {
		return map[string][]string{"IBAN": {"DE12 3456 7890"}}, nil
	}
	return map[string][]string{"PERSON": {"John Doe"}}, nil
}

func (d *DomainMockScanner) CheckHealth(host string) {}
func (d *DomainMockScanner) IsAvailable() bool       { return true }
func (d *DomainMockScanner) SetDomain(domain string) { d.LastDomain = domain }

func TestDomainSwapping(t *testing.T) {
	mockData := &DomainMockScanner{}
	vault := &MockVault{}
	masterKey := make([]byte, 32)
	eng := NewRefinery(vault, masterKey)
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
	if mockData.LastDomain != "clinical" {
		t.Errorf("Expected domain clinical, got %s", mockData.LastDomain)
	}
	// Verify that the mock returned the clinical entity which was then tokenized
	// The mock returns MRN: MRN-123456. RefineString should tokenise it.
	// Since MRN-123456 is not one of the default patterns, it only gets tokenised if the AI scanner finds it.
	if !containsToken(res) {
		t.Errorf("Expected clinical entity to be tokenised, got: %s", res)
	}

	// Case 3: Performance Shield (SkipDeepScan)
	eng.SkipDeepScan = true
	mockData.IsCalled = false
	eng.RefineString("Scan this heavy text", "tester", nil)
	if mockData.IsCalled {
		t.Error("Expected scanner to be skipped when SkipDeepScan is true")
	}
}

func containsToken(s string) bool {
	return tokenPattern.MatchString(s)
}
