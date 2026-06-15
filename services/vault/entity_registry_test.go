package vault_test

// entity_registry_test.go — tests for the Path 3 Persistent Entity Registry.
// Covers: idempotent registration, variant merging, case-insensitive lookup,
// entity token rehydration, bulk seeding, and ListEntities.

import (
	"path/filepath"
	"testing"

	"github.com/ocultar-dev/ocultar/pkg/config"
	"github.com/ocultar-dev/ocultar/vault"
)

func openRegistryVault(t *testing.T) vault.Provider {
	t.Helper()
	vaultPath := filepath.Join(t.TempDir(), "registry.db")
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, vaultPath)
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	t.Cleanup(func() { v.Close() })
	return v
}

// TestEntityRegistry_RegisterAndLookup verifies that RegisterEntity returns a
// canonical token and that each variant is immediately findable via LookupVariant.
func TestEntityRegistry_RegisterAndLookup(t *testing.T) {
	v := openRegistryVault(t)

	tok, err := v.RegisterEntity("PERSON", "John Doe", []string{"John", "Doe", "J. Doe"})
	if err != nil {
		t.Fatalf("RegisterEntity: %v", err)
	}
	if tok != "[PERSON_1]" {
		t.Errorf("want [PERSON_1], got %s", tok)
	}

	for _, variant := range []string{"John", "Doe", "J. Doe"} {
		got, found := v.LookupVariant(variant)
		if !found {
			t.Errorf("LookupVariant(%q): not found", variant)
			continue
		}
		if got != "[PERSON_1]" {
			t.Errorf("LookupVariant(%q): want [PERSON_1], got %s", variant, got)
		}
	}
	t.Logf("✅ Register + lookup: [PERSON_1]")
}

// TestEntityRegistry_CaseInsensitiveLookup ensures the lookup is
// case-insensitive so "john" finds the same entity as "John".
func TestEntityRegistry_CaseInsensitiveLookup(t *testing.T) {
	v := openRegistryVault(t)

	if _, err := v.RegisterEntity("PERSON", "Jane Smith", []string{"Jane", "Smith"}); err != nil {
		t.Fatalf("RegisterEntity: %v", err)
	}

	for _, variant := range []string{"jane", "JANE", "Jane", "smith", "SMITH"} {
		_, found := v.LookupVariant(variant)
		if !found {
			t.Errorf("LookupVariant(%q): expected found=true (case-insensitive)", variant)
		}
	}
	t.Logf("✅ Case-insensitive lookup confirmed")
}

// TestEntityRegistry_Idempotent ensures that registering the same canonical
// name twice returns the same token and does not create duplicate rows.
func TestEntityRegistry_Idempotent(t *testing.T) {
	v := openRegistryVault(t)

	tok1, err := v.RegisterEntity("PERSON", "Alice Martin", []string{"Alice"})
	if err != nil {
		t.Fatalf("first RegisterEntity: %v", err)
	}

	tok2, err := v.RegisterEntity("PERSON", "Alice Martin", []string{"A. Martin", "Ms. Martin"})
	if err != nil {
		t.Fatalf("second RegisterEntity: %v", err)
	}

	if tok1 != tok2 {
		t.Errorf("idempotent registration returned different tokens: %s vs %s", tok1, tok2)
	}

	// All three variants (original + merged) must resolve to the same token.
	for _, v2 := range []string{"Alice", "A. Martin", "Ms. Martin"} {
		got, found := v.LookupVariant(v2)
		if !found || got != tok1 {
			t.Errorf("LookupVariant(%q): want %s, got %s (found=%v)", v2, tok1, got, found)
		}
	}
	t.Logf("✅ Idempotent: both calls returned %s, all 3 variants resolve correctly", tok1)
}

// TestEntityRegistry_MultipleEntities checks that sequential IDs are assigned
// independently per entity_type.
func TestEntityRegistry_MultipleEntities(t *testing.T) {
	v := openRegistryVault(t)

	tok1, _ := v.RegisterEntity("PERSON", "Bob Brown", []string{"Bob"})
	tok2, _ := v.RegisterEntity("PERSON", "Carol White", []string{"Carol"})
	tok3, _ := v.RegisterEntity("ORGANIZATION", "Acme Corp", []string{"Acme"})

	if tok1 != "[PERSON_1]" {
		t.Errorf("want [PERSON_1], got %s", tok1)
	}
	if tok2 != "[PERSON_2]" {
		t.Errorf("want [PERSON_2], got %s", tok2)
	}
	if tok3 != "[ORGANIZATION_1]" {
		t.Errorf("want [ORGANIZATION_1], got %s", tok3)
	}
	t.Logf("✅ Sequential IDs: %s, %s, %s", tok1, tok2, tok3)
}

// TestEntityRegistry_GetEntityByToken verifies rehydration: a canonical token
// like "[PERSON_1]" must expand back to the canonical_name.
func TestEntityRegistry_GetEntityByToken(t *testing.T) {
	v := openRegistryVault(t)

	if _, err := v.RegisterEntity("PERSON", "David Lee", []string{"David", "Lee"}); err != nil {
		t.Fatalf("RegisterEntity: %v", err)
	}

	name, found := v.GetEntityByToken("[PERSON_1]")
	if !found {
		t.Fatal("GetEntityByToken([PERSON_1]): not found")
	}
	if name != "David Lee" {
		t.Errorf("want %q, got %q", "David Lee", name)
	}

	// Unknown token must return not-found.
	_, found2 := v.GetEntityByToken("[PERSON_999]")
	if found2 {
		t.Error("GetEntityByToken([PERSON_999]): expected not-found for unknown token")
	}
	t.Logf("✅ GetEntityByToken: [PERSON_1] → %q", name)
}

// TestEntityRegistry_SeedEntities exercises the bulk-seed path.
func TestEntityRegistry_SeedEntities(t *testing.T) {
	v := openRegistryVault(t)

	seeds := []vault.EntitySeed{
		{EntityType: "PERSON", CanonicalName: "Eve Adams", Variants: []string{"Eve", "Adams"}},
		{EntityType: "PERSON", CanonicalName: "Frank Baker", Variants: []string{"Frank", "F. Baker"}},
		{EntityType: "ORGANIZATION", CanonicalName: "GlobalCorp", Variants: []string{"Global Corp", "GC"}},
	}

	if err := v.SeedEntities(seeds); err != nil {
		t.Fatalf("SeedEntities: %v", err)
	}

	cases := []struct{ variant, want string }{
		{"Eve", "[PERSON_1]"},
		{"Adams", "[PERSON_1]"},
		{"Frank", "[PERSON_2]"},
		{"F. Baker", "[PERSON_2]"},
		{"GC", "[ORGANIZATION_1]"},
	}
	for _, c := range cases {
		got, found := v.LookupVariant(c.variant)
		if !found || got != c.want {
			t.Errorf("LookupVariant(%q): want %s, got %s (found=%v)", c.variant, c.want, got, found)
		}
	}
	t.Logf("✅ SeedEntities: 3 entities seeded, all variants resolve correctly")
}

// TestEntityRegistry_ListEntities checks that all registered entities and their
// variants are returned by ListEntities.
func TestEntityRegistry_ListEntities(t *testing.T) {
	v := openRegistryVault(t)

	v.RegisterEntity("PERSON", "Grace Hall", []string{"Grace", "G. Hall"})
	v.RegisterEntity("ORGANIZATION", "TechStart", []string{"TS", "TechStart Inc"})

	records, err := v.ListEntities()
	if err != nil {
		t.Fatalf("ListEntities: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("want 2 records, got %d", len(records))
	}

	for _, rec := range records {
		if len(rec.Variants) == 0 {
			t.Errorf("entity %q has no variants", rec.ID)
		}
		t.Logf("  %s (%s): %v", rec.ID, rec.CanonicalName, rec.Variants)
	}
	t.Logf("✅ ListEntities returned %d records", len(records))
}

// TestEntityRegistry_LookupMiss verifies that an unregistered string returns not-found.
func TestEntityRegistry_LookupMiss(t *testing.T) {
	v := openRegistryVault(t)

	_, found := v.LookupVariant("RandomUnknownString")
	if found {
		t.Error("LookupVariant for unregistered string should return not-found")
	}
	t.Logf("✅ Lookup miss: correctly returned not-found")
}
