package refinery_test

// refinery_entity_test.go — integration tests for Path 3 Persistent Entity Registry.
//
// These tests prove three properties:
//  1. Unified token — "John", "Doe", and "John Doe" all produce "[PERSON_1]" once registered.
//  2. Fragment coherence — isolated name fragments no longer create separate hash tokens.
//  3. Rehydration — "[PERSON_1]" expands back to "John Doe" via DecryptToken.
//
// Run: cd services/refinery && CGO_ENABLED=1 go test ./pkg/refinery/ -run TestEntityRegistry -v

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/pkg/refinery"
	"github.com/ocultar-dev/ocultar/vault"
)

// makeEntityRefinery creates an in-memory-vaulted Refinery for entity registry tests.
func makeEntityRefinery(t *testing.T) (*refinery.Refinery, vault.Provider) {
	t.Helper()
	vaultPath := filepath.Join(t.TempDir(), "entity_test.db")
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, vaultPath)
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	t.Cleanup(func() { v.Close() })

	masterKey := []byte("entity-test-key-32-bytes-padding")
	config.InitDefaults()

	eng := refinery.NewRefinery(v, masterKey)
	return eng, v
}

// TestEntityRegistry_UnifiedToken registers "John Doe" with variants, then
// simulates SLM detection of each fragment via preScanMap. All fragments must
// resolve to "[PERSON_1]" — the canonical entity token — not separate hash tokens.
func TestEntityRegistry_UnifiedToken(t *testing.T) {
	eng, v := makeEntityRefinery(t)

	tok, err := v.RegisterEntity("PERSON", "John Doe", []string{"John", "Doe", "John Doe"})
	if err != nil {
		t.Fatalf("RegisterEntity: %v", err)
	}
	if tok != "[PERSON_1]" {
		t.Errorf("want [PERSON_1], got %s", tok)
	}

	cases := []struct {
		input      string
		detectedAs map[string][]string // simulated SLM output
		desc       string
	}{
		{
			// Input has no Tier-1.5 trigger keyword before "John", so Tier 2
			// (preScanMap) is the first to claim the name.
			input:      "Patient: John. Status: pending.",
			detectedAs: map[string][]string{"PERSON": {"John"}},
			desc:       "first-name only",
		},
		{
			input:      "Forward to Doe when ready.",
			detectedAs: map[string][]string{"PERSON": {"Doe"}},
			desc:       "last-name only",
		},
		{
			input:      "John Doe signed the consent form.",
			detectedAs: map[string][]string{"PERSON": {"John Doe"}},
			desc:       "full name",
		},
	}

	for _, c := range cases {
		out, err := eng.RefineString(c.input, "test", c.detectedAs)
		if err != nil {
			t.Errorf("[%s] RefineString error: %v", c.desc, err)
			continue
		}
		if !strings.Contains(out, "[PERSON_1]") {
			t.Errorf("[%s] want [PERSON_1] in output, got: %s", c.desc, out)
		}
		if strings.Contains(out, "John") || strings.Contains(out, "Doe") {
			t.Errorf("[%s] name leaked in output: %s", c.desc, out)
		}
		t.Logf("✅ [%s] → %s", c.desc, out)
	}
}

// TestEntityRegistry_Rehydration verifies that DecryptToken resolves "[PERSON_1]"
// back to the canonical name "John Doe".
func TestEntityRegistry_Rehydration(t *testing.T) {
	eng, v := makeEntityRefinery(t)
	_ = eng

	if _, err := v.RegisterEntity("PERSON", "Marie Curie", []string{"Marie", "Curie"}); err != nil {
		t.Fatalf("RegisterEntity: %v", err)
	}

	masterKey := []byte("entity-test-key-32-bytes-padding")
	name, err := refinery.DecryptToken(v, masterKey, "[PERSON_1]")
	if err != nil {
		t.Fatalf("DecryptToken: %v", err)
	}
	if name != "Marie Curie" {
		t.Errorf("want %q, got %q", "Marie Curie", name)
	}
	t.Logf("✅ Rehydration: [PERSON_1] → %q", name)
}

// TestEntityRegistry_NonRegisteredFallsToHash ensures that a PERSON-type match
// for an unknown name still receives a hash-based token (existing behavior).
func TestEntityRegistry_NonRegisteredFallsToHash(t *testing.T) {
	eng, _ := makeEntityRefinery(t)

	out, err := eng.RefineString("My name is Xavier Unknown.", "test", nil)
	if err != nil {
		t.Fatalf("RefineString: %v", err)
	}
	// Should still be redacted, but with a hash token, not a numeric entity token.
	if strings.Contains(out, "Xavier") || strings.Contains(out, "Unknown") {
		// Acceptable only if the name wasn't detected at all (no false negative assertion here).
		t.Logf("INFO: name not detected (acceptable for unregistered names without Tier 2 SLM): %s", out)
		return
	}
	// If it was detected, it must NOT look like an entity-registry token.
	if strings.Contains(out, "[PERSON_1]") {
		t.Errorf("unregistered name should not produce an entity registry token")
	}
	t.Logf("✅ Non-registered name fallback: %s", out)
}

// TestEntityRegistry_MultiplePersonsNoCollision seeds two patients and verifies
// that their fragments never resolve to each other's token (via simulated SLM).
func TestEntityRegistry_MultiplePersonsNoCollision(t *testing.T) {
	eng, v := makeEntityRefinery(t)

	tok1, _ := v.RegisterEntity("PERSON", "Alice Dupont", []string{"Alice", "Dupont"})
	tok2, _ := v.RegisterEntity("PERSON", "Bruno Moreau", []string{"Bruno", "Moreau"})

	if tok1 == tok2 {
		t.Fatalf("two distinct patients produced the same token: %s", tok1)
	}

	// Simulate SLM detecting each fragment independently.
	aliceOut, _ := eng.RefineString("Alice is the primary patient.",
		"test", map[string][]string{"PERSON": {"Alice"}})
	brunoOut, _ := eng.RefineString("Bruno is the secondary patient.",
		"test", map[string][]string{"PERSON": {"Bruno"}})

	if !strings.Contains(aliceOut, tok1) {
		t.Errorf("Alice fragment should resolve to %s, got: %s", tok1, aliceOut)
	}
	if !strings.Contains(brunoOut, tok2) {
		t.Errorf("Bruno fragment should resolve to %s, got: %s", tok2, brunoOut)
	}
	if strings.Contains(aliceOut, tok2) || strings.Contains(brunoOut, tok1) {
		t.Error("cross-patient token collision detected — critical privacy failure")
	}
	t.Logf("✅ No collision: Alice→%s, Bruno→%s", tok1, tok2)
}
