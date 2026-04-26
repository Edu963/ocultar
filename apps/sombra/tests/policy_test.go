package connector_test

import (
	"testing"

	"github.com/Edu963/ocultar/apps/sombra/pkg/connector"
)

// ── DataPolicy.IsModelAllowed ─────────────────────────────────────────────────

func TestDataPolicy_IsModelAllowed_EmptyList_AllowsAll(t *testing.T) {
	p := connector.DataPolicy{AllowedModels: nil}
	for _, model := range []string{"gpt-4o", "gemini-pro", "claude-3-opus"} {
		if !p.IsModelAllowed(model) {
			t.Errorf("empty AllowedModels should allow %q, but it was denied", model)
		}
	}
}

func TestDataPolicy_IsModelAllowed_AllowsListedModel(t *testing.T) {
	p := connector.DataPolicy{AllowedModels: []string{"gpt-4o", "claude-3-opus"}}
	if !p.IsModelAllowed("gpt-4o") {
		t.Error("expected gpt-4o to be allowed")
	}
	if !p.IsModelAllowed("claude-3-opus") {
		t.Error("expected claude-3-opus to be allowed")
	}
}

func TestDataPolicy_IsModelAllowed_BlocksUnlistedModel(t *testing.T) {
	p := connector.DataPolicy{AllowedModels: []string{"gpt-4o"}}
	if p.IsModelAllowed("gemini-pro") {
		t.Error("expected gemini-pro to be blocked by AllowedModels policy")
	}
}

func TestDataPolicy_IsModelAllowed_CaseSensitive(t *testing.T) {
	p := connector.DataPolicy{AllowedModels: []string{"gpt-4o"}}
	// Model names are exact-match; "GPT-4O" is not "gpt-4o".
	if p.IsModelAllowed("GPT-4O") {
		t.Error("model matching should be case-sensitive")
	}
}

// ── DataPolicy.ShouldStrip ────────────────────────────────────────────────────

func TestDataPolicy_ShouldStrip_EmptyList_StripNothing(t *testing.T) {
	p := connector.DataPolicy{StripCategories: nil}
	for _, cat := range []string{"EMAIL", "SSN", "PERSON"} {
		if p.ShouldStrip(cat) {
			t.Errorf("empty StripCategories should not strip %q", cat)
		}
	}
}

func TestDataPolicy_ShouldStrip_StripsListedCategory(t *testing.T) {
	p := connector.DataPolicy{StripCategories: []string{"SSN", "ACCOUNT_NUMBER"}}
	if !p.ShouldStrip("SSN") {
		t.Error("expected SSN to be stripped")
	}
	if !p.ShouldStrip("ACCOUNT_NUMBER") {
		t.Error("expected ACCOUNT_NUMBER to be stripped")
	}
}

func TestDataPolicy_ShouldStrip_DoesNotStripUnlistedCategory(t *testing.T) {
	p := connector.DataPolicy{StripCategories: []string{"SSN"}}
	if p.ShouldStrip("EMAIL") {
		t.Error("EMAIL should not be stripped when not in StripCategories")
	}
}

func TestDataPolicy_ShouldStrip_CaseSensitive(t *testing.T) {
	p := connector.DataPolicy{StripCategories: []string{"SSN"}}
	if p.ShouldStrip("ssn") {
		t.Error("category stripping should be case-sensitive")
	}
}

// ── MaxBodyBytes field ────────────────────────────────────────────────────────

func TestDataPolicy_MaxBodyBytes_ZeroMeansUnlimited(t *testing.T) {
	p := connector.DataPolicy{MaxBodyBytes: 0}
	// MaxBodyBytes==0 is documented as "unlimited" — nothing to enforce at the
	// struct level, but we verify the zero value is accessible and correctly typed.
	if p.MaxBodyBytes != 0 {
		t.Errorf("expected 0 for unlimited, got %d", p.MaxBodyBytes)
	}
}

func TestDataPolicy_MaxBodyBytes_Positive(t *testing.T) {
	p := connector.DataPolicy{MaxBodyBytes: 1024 * 1024}
	if p.MaxBodyBytes != 1<<20 {
		t.Errorf("expected 1MiB, got %d", p.MaxBodyBytes)
	}
}

// ── Combined policy enforcement ───────────────────────────────────────────────

func TestDataPolicy_FullPolicy(t *testing.T) {
	p := connector.DataPolicy{
		AllowedModels:   []string{"safe-model"},
		StripCategories: []string{"SSN", "CREDIT_CARD"},
		MaxBodyBytes:    512 * 1024,
	}

	if !p.IsModelAllowed("safe-model") {
		t.Error("safe-model should be allowed")
	}
	if p.IsModelAllowed("unsafe-model") {
		t.Error("unsafe-model should be denied")
	}
	if !p.ShouldStrip("SSN") {
		t.Error("SSN should be stripped")
	}
	if p.ShouldStrip("EMAIL") {
		t.Error("EMAIL should not be stripped")
	}
}
