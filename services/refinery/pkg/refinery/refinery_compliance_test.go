package refinery

import (
	"strings"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/vault"
)

func TestGDPRComplianceHeuristics(t *testing.T) {
	config.InitDefaults()

	// Use a temporary file for the test vault to avoid DSN issues with :memory:
	tmpVault := "test_compliance_vault.db"
	v, err := vault.New(config.Global, tmpVault)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer v.Close()

	e := NewRefinery(v, []byte("0123456789abcdef0123456789abcdef"))

	testCases := []struct {
		desc     string
		input    string
		requires []string // Substrings that MUST NOT be present in refined output
	}{
		{
			desc:     "Greedy Neighborhood (Surname Proximity)",
			input:    "[PERSON_VIP_849b0f6c] CABEZAS [PERSON_VIP_48666084]",
			requires: []string{"CABEZAS"},
		},
		{
			desc:     "Multilingual Conjunction Linkage (ET)",
			input:    "GARANKA [PERSON_VIP_849b0f6c] CABEZAS ET MULLER",
			requires: []string{"CABEZAS", "MULLER"},
		},
		{
			desc:     "Professional Title Shield (DR)",
			input:    "X0853 DR GRARD ECHIROLLES 28/01",
			requires: []string{"DR", "GRARD", "ECHIROLLES"},
		},
		{
			desc:     "Semantic Scrubbing (Sensitive Life Event)",
			input:    "Translation divorce documents",
			requires: []string{"divorce"},
		},
		{
			desc:     "Possessive Case Names",
			input:    "Eddie's bday and utilities",
			requires: []string{"Eddie"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			refined, err := e.RefineString(tc.input, "test-compliance", nil)
			if err != nil {
				t.Fatalf("RefineString failed: %v", err)
			}
			for _, req := range tc.requires {
				if strings.Contains(strings.ToLower(refined), strings.ToLower(req)) {
					t.Errorf("[%s] Exposure detected: found %q in result %q", tc.desc, req, refined)
				}
			}
		})
	}
}
