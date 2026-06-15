// Package vault defines the storage abstraction layer for OCULTAR's PII vault.
// It exposes a single Provider interface that all vault backends must implement,
// and a New() factory that selects the right backend based on configuration.
package vault

import (
	"fmt"

	"github.com/ocultar-dev/ocultar/pkg/config"
)

// EntitySeed defines a single entity for bulk-seeding into the registry.
// Each seed maps a canonical identity (e.g. a patient or customer) to all
// known name variants so the refinery can unify them under one token.
type EntitySeed struct {
	EntityType    string   // e.g. "PERSON", "ORGANIZATION"
	CanonicalName string   // e.g. "John Doe"
	Variants      []string // e.g. ["John", "Doe", "J. Doe", "Mr. Doe"]
}

// EntityRecord is the read-side representation of a registered entity,
// returned by ListEntities for API responses and audits.
type EntityRecord struct {
	ID            string   `json:"id"`             // e.g. "PERSON_1"
	EntityType    string   `json:"entity_type"`    // e.g. "PERSON"
	CanonicalName string   `json:"canonical_name"` // e.g. "John Doe"
	Variants      []string `json:"variants"`       // all registered variant strings
}

// Provider is the storage contract every vault backend must satisfy.
// It intentionally knows nothing about SQL, DuckDB, or Postgres.
type Provider interface {
	// StoreToken persists a (hash → token, encrypted_pii) mapping.
	// Returns (true, nil) when a new row was inserted, (false, nil) when
	// the hash already existed (idempotent), or (false, err) on failure.
	StoreToken(hash, token, encryptedPII string) (inserted bool, err error)

	// GetToken looks up the token for a given PII hash.
	// Returns (token, true) on a cache hit, ("", false) on a miss.
	GetToken(hash string) (token string, found bool)

	// CountAll returns the total number of entries in the vault.
	CountAll() int64

	// Close releases any open database connections.
	Close() error

	// --- Entity Registry (Path 3) ---

	// RegisterEntity creates a canonical entity and maps all provided variants
	// to it. Returns the canonical token (e.g. "[PERSON_1]").
	// Idempotent: if the canonical name is already registered, its token is
	// returned and the new variants are merged in.
	RegisterEntity(entityType, canonicalName string, variants []string) (canonicalToken string, err error)

	// LookupVariant performs a fast, case-insensitive lookup of a string
	// against the entity_variants table. Returns the canonical token
	// (e.g. "[PERSON_1]") if the string is a known variant of any registered
	// entity, so the refinery can replace all fragments with one stable token.
	LookupVariant(variantName string) (canonicalToken string, found bool)

	// GetEntityByToken looks up the canonical_name for an entity token such
	// as "[PERSON_1]". Used during stream rehydration so all variant tokens
	// expand to the same coherent canonical name.
	GetEntityByToken(token string) (canonicalName string, found bool)

	// SeedEntities bulk-registers a list of entities. Each seed's variants
	// are merged into any existing canonical entity with the same name.
	SeedEntities(entries []EntitySeed) error

	// ListEntities returns all registered canonical entities with their
	// variant lists, used by the GET /v1/entities management API.
	ListEntities() ([]EntityRecord, error)
}

// New returns the appropriate Provider based on configuration.
// vault_backend=postgres uses PostgreSQL; anything else uses local DuckDB.
//
// vaultPath is passed explicitly so that main.go can override it with
// ":memory:" when running in dry-run mode.
func New(cfg config.Settings, vaultPath string) (Provider, error) {
	switch cfg.VaultBackend {
	case "postgres":
		if cfg.PostgresDSN == "" {
			return nil, fmt.Errorf("[vault] vault_backend is 'postgres' but postgres_dsn is not set in config.yaml")
		}
		return newPostgresProvider(cfg.PostgresDSN)
	default:
		return newDuckDBProvider(vaultPath)
	}
}
