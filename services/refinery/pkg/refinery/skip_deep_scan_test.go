package refinery

import (
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
)

// SpyScanner records whether ScanForPII was called, for use in SkipDeepScan assertions.
type SpyScanner struct {
	Called bool
	Result map[string][]string
}

func (s *SpyScanner) ScanForPII(_ string) (map[string][]string, error) {
	s.Called = true
	if s.Result != nil {
		return s.Result, nil
	}
	return map[string][]string{}, nil
}
func (s *SpyScanner) CheckHealth(_ string) {}
func (s *SpyScanner) IsAvailable() bool    { return true }
func (s *SpyScanner) SetDomain(_ string)   {}

func newTestRefinery(scanner AIScanner) *Refinery {
	eng := NewRefinery(&MockVault{}, []byte("01234567890123456789012345678901"))
	eng.AIScanner = scanner
	return eng
}

// RefineString — SkipDeepScan=true must not invoke the scanner.
func TestSkipDeepScan_RefineString_SkipsWhenTrue(t *testing.T) {
	spy := &SpyScanner{}
	eng := newTestRefinery(spy)
	eng.SkipDeepScan = true

	_, err := eng.RefineString("Hello, my name is Jean-Michel Dupont and I work at BNP Paribas.", "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spy.Called {
		t.Error("ScanForPII should not be called when SkipDeepScan=true")
	}
}

// RefineString — SkipDeepScan=false (default) must call the scanner for long unredacted text.
func TestSkipDeepScan_RefineString_CallsWhenFalse(t *testing.T) {
	spy := &SpyScanner{}
	eng := newTestRefinery(spy)
	// SkipDeepScan defaults to false

	_, err := eng.RefineString("Hello, my name is Jean-Michel Dupont and I work at BNP Paribas.", "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !spy.Called {
		t.Error("ScanForPII should be called when SkipDeepScan=false and text is long enough")
	}
}

// ProcessInterface — SkipDeepScan=true must not invoke the batch scanner.
func TestSkipDeepScan_ProcessInterface_SkipsWhenTrue(t *testing.T) {
	spy := &SpyScanner{}
	eng := newTestRefinery(spy)
	eng.SkipDeepScan = true

	data := map[string]interface{}{
		"message": "Please review the account of our client Jean-Michel Dupont from Lyon.",
	}
	_, err := eng.ProcessInterface(data, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spy.Called {
		t.Error("ScanForPII should not be called by ProcessInterface when SkipDeepScan=true")
	}
}

// ProcessInterface — SkipDeepScan=false must call the batch scanner.
func TestSkipDeepScan_ProcessInterface_CallsWhenFalse(t *testing.T) {
	spy := &SpyScanner{}
	eng := newTestRefinery(spy)

	data := map[string]interface{}{
		"message": "Please review the account of our client Jean-Michel Dupont from Lyon.",
	}
	_, err := eng.ProcessInterface(data, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !spy.Called {
		t.Error("ScanForPII should be called by ProcessInterface when SkipDeepScan=false")
	}
}

// Domain scanner must also be bypassed when SkipDeepScan=true.
// activeScanner() reads config.Global.DomainSnapshot, so we set it directly.
func TestSkipDeepScan_DomainScanner_AlsoSkipped(t *testing.T) {
	defaultSpy := &SpyScanner{}
	domainSpy := &SpyScanner{}

	eng := newTestRefinery(defaultSpy)
	eng.SetDomainScanner("finance", domainSpy)
	eng.SkipDeepScan = true

	// Activate the finance domain via the global that activeScanner() reads.
	config.Global.DomainSnapshot = "finance"
	t.Cleanup(func() { config.Global.DomainSnapshot = "" })

	_, err := eng.RefineString("Le client Jean-Michel Dupont a un compte chez BNP Paribas.", "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if defaultSpy.Called {
		t.Error("default scanner should not be called when SkipDeepScan=true")
	}
	if domainSpy.Called {
		t.Error("domain scanner should not be called when SkipDeepScan=true")
	}
}

// Sanity-check: domain scanner IS used when SkipDeepScan=false and domain matches.
func TestSkipDeepScan_DomainScanner_UsedWhenEnabled(t *testing.T) {
	defaultSpy := &SpyScanner{}
	domainSpy := &SpyScanner{Result: map[string][]string{"PERSON": {"Jean-Michel Dupont"}}}

	eng := newTestRefinery(defaultSpy)
	eng.SetDomainScanner("finance", domainSpy)

	config.Global.DomainSnapshot = "finance"
	t.Cleanup(func() { config.Global.DomainSnapshot = "" })

	text := "Le client Jean-Michel Dupont a un compte chez BNP Paribas."
	result, err := eng.RefineString(text, "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !domainSpy.Called {
		t.Error("domain scanner should be called when SkipDeepScan=false and domain matches")
	}
	if defaultSpy.Called {
		t.Error("default scanner should not be called when a domain scanner is active")
	}
	// The name returned by the spy should be tokenised in the output.
	if !containsToken(result) {
		t.Errorf("expected PERSON entity to be tokenised in output, got: %s", result)
	}
}
