package vault

// entity_registry.go — implements the five EntityRegistry methods on both
// duckdbProvider and postgresProvider. The registry maps name variants
// (e.g. "John", "Doe", "J. Doe") to a single canonical token (e.g. "[PERSON_1]")
// so that the refinery can produce one stable token for all fragments of the
// same identity across files, prompts, and sessions.

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// DuckDB implementation
// ─────────────────────────────────────────────────────────────────────────────

// RegisterEntity creates (or retrieves) a canonical entity and merges the
// provided variants into entity_variants. Returns the canonical token string.
// Safe for concurrent calls because duckdbProvider uses SetMaxOpenConns(1).
func (p *duckdbProvider) RegisterEntity(entityType, canonicalName string, variants []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entityType = strings.ToUpper(strings.TrimSpace(entityType))
	canonicalName = strings.TrimSpace(canonicalName)
	if entityType == "" || canonicalName == "" {
		return "", fmt.Errorf("entity_type and canonical_name are required")
	}

	// Check if this canonical name is already registered.
	var existingID string
	err := p.db.QueryRowContext(ctx,
		`SELECT id FROM canonical_entities WHERE LOWER(canonical_name) = LOWER(?)`,
		canonicalName,
	).Scan(&existingID)
	if err == nil {
		// Already exists — merge in any new variants and return the token.
		if mergeErr := p.mergeVariantsDuckDB(ctx, existingID, variants); mergeErr != nil {
			return "", mergeErr
		}
		return fmt.Sprintf("[%s]", existingID), nil
	}

	// Generate a new sequential ID: PERSON_1, PERSON_2, etc.
	var count int64
	p.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM canonical_entities WHERE entity_type = ?`, entityType,
	).Scan(&count)
	newID := fmt.Sprintf("%s_%d", entityType, count+1)

	_, err = p.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO canonical_entities (id, entity_type, canonical_name) VALUES (?, ?, ?)`,
		newID, entityType, canonicalName,
	)
	if err != nil {
		return "", fmt.Errorf("[vault/duckdb] RegisterEntity insert: %w", err)
	}

	// Re-query in case another goroutine raced us (INSERT OR IGNORE → no-op).
	var actualID string
	p.db.QueryRowContext(ctx,
		`SELECT id FROM canonical_entities WHERE LOWER(canonical_name) = LOWER(?)`,
		canonicalName,
	).Scan(&actualID)
	if actualID == "" {
		actualID = newID
	}

	if mergeErr := p.mergeVariantsDuckDB(ctx, actualID, variants); mergeErr != nil {
		return "", mergeErr
	}

	log.Printf("[entity-registry] Registered %s → [%s] with %d variant(s)", canonicalName, actualID, len(variants))
	return fmt.Sprintf("[%s]", actualID), nil
}

func (p *duckdbProvider) mergeVariantsDuckDB(ctx context.Context, canonicalID string, variants []string) error {
	for _, v := range variants {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		_, err := p.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO entity_variants (variant_name, canonical_id) VALUES (?, ?)`,
			v, canonicalID,
		)
		if err != nil {
			return fmt.Errorf("[vault/duckdb] mergeVariants(%q): %w", v, err)
		}
	}
	return nil
}

// LookupVariant performs a case-insensitive lookup of variantName against
// entity_variants. Returns the canonical token (e.g. "[PERSON_1]") if found.
func (p *duckdbProvider) LookupVariant(variantName string) (string, bool) {
	variantName = strings.TrimSpace(variantName)
	if len(variantName) < 2 {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var canonicalID string
	err := p.db.QueryRowContext(ctx,
		`SELECT canonical_id FROM entity_variants WHERE LOWER(variant_name) = LOWER(?)`,
		variantName,
	).Scan(&canonicalID)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("[%s]", canonicalID), true
}

// GetEntityByToken looks up the canonical_name for a token like "[PERSON_1]".
// Used during rehydration to expand entity tokens back to the canonical name.
func (p *duckdbProvider) GetEntityByToken(token string) (string, bool) {
	// Strip the surrounding brackets to get the raw ID (e.g. "PERSON_1").
	id := strings.Trim(token, "[]")
	if id == "" {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var name string
	err := p.db.QueryRowContext(ctx,
		`SELECT canonical_name FROM canonical_entities WHERE id = ?`, id,
	).Scan(&name)
	if err != nil {
		return "", false
	}
	return name, true
}

// SeedEntities bulk-registers a list of EntitySeed entries.
func (p *duckdbProvider) SeedEntities(entries []EntitySeed) error {
	for _, e := range entries {
		if _, err := p.RegisterEntity(e.EntityType, e.CanonicalName, e.Variants); err != nil {
			return fmt.Errorf("[vault/duckdb] SeedEntities(%q): %w", e.CanonicalName, err)
		}
	}
	return nil
}

// ListEntities returns all canonical entities with their variant lists.
// Uses a single LEFT JOIN to avoid nested queries on the single DuckDB connection.
func (p *duckdbProvider) ListEntities() ([]EntityRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, `
		SELECT ce.id, ce.entity_type, ce.canonical_name, ev.variant_name
		FROM canonical_entities ce
		LEFT JOIN entity_variants ev ON ev.canonical_id = ce.id
		ORDER BY ce.entity_type, ce.id, ev.variant_name`)
	if err != nil {
		return nil, fmt.Errorf("[vault/duckdb] ListEntities: %w", err)
	}
	defer rows.Close()

	// Build records map keyed by entity ID to collect variants.
	indexed := make(map[string]*EntityRecord)
	var order []string // preserve insertion order

	for rows.Next() {
		var id, entityType, canonicalName string
		var variantName *string // nullable — entity may have no variants
		if err := rows.Scan(&id, &entityType, &canonicalName, &variantName); err != nil {
			return nil, err
		}
		if _, exists := indexed[id]; !exists {
			indexed[id] = &EntityRecord{
				ID:            id,
				EntityType:    entityType,
				CanonicalName: canonicalName,
			}
			order = append(order, id)
		}
		if variantName != nil {
			indexed[id].Variants = append(indexed[id].Variants, *variantName)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	records := make([]EntityRecord, 0, len(order))
	for _, id := range order {
		records = append(records, *indexed[id])
	}
	return records, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Postgres implementation
// ─────────────────────────────────────────────────────────────────────────────

func (p *postgresProvider) RegisterEntity(entityType, canonicalName string, variants []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entityType = strings.ToUpper(strings.TrimSpace(entityType))
	canonicalName = strings.TrimSpace(canonicalName)
	if entityType == "" || canonicalName == "" {
		return "", fmt.Errorf("entity_type and canonical_name are required")
	}

	var existingID string
	err := p.db.QueryRowContext(ctx,
		`SELECT id FROM canonical_entities WHERE LOWER(canonical_name) = LOWER($1)`,
		canonicalName,
	).Scan(&existingID)
	if err == nil {
		if mergeErr := p.mergeVariantsPG(ctx, existingID, variants); mergeErr != nil {
			return "", mergeErr
		}
		return fmt.Sprintf("[%s]", existingID), nil
	}

	var count int64
	p.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM canonical_entities WHERE entity_type = $1`, entityType,
	).Scan(&count)
	newID := fmt.Sprintf("%s_%d", entityType, count+1)

	_, err = p.db.ExecContext(ctx,
		`INSERT INTO canonical_entities (id, entity_type, canonical_name)
		 VALUES ($1, $2, $3) ON CONFLICT (canonical_name) DO NOTHING`,
		newID, entityType, canonicalName,
	)
	if err != nil {
		return "", fmt.Errorf("[vault/postgres] RegisterEntity insert: %w", err)
	}

	var actualID string
	p.db.QueryRowContext(ctx,
		`SELECT id FROM canonical_entities WHERE LOWER(canonical_name) = LOWER($1)`,
		canonicalName,
	).Scan(&actualID)
	if actualID == "" {
		actualID = newID
	}

	if mergeErr := p.mergeVariantsPG(ctx, actualID, variants); mergeErr != nil {
		return "", mergeErr
	}

	log.Printf("[entity-registry] Registered %s → [%s] with %d variant(s)", canonicalName, actualID, len(variants))
	return fmt.Sprintf("[%s]", actualID), nil
}

func (p *postgresProvider) mergeVariantsPG(ctx context.Context, canonicalID string, variants []string) error {
	for _, v := range variants {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		_, err := p.db.ExecContext(ctx,
			`INSERT INTO entity_variants (variant_name, canonical_id)
			 VALUES ($1, $2) ON CONFLICT (variant_name) DO NOTHING`,
			v, canonicalID,
		)
		if err != nil {
			return fmt.Errorf("[vault/postgres] mergeVariants(%q): %w", v, err)
		}
	}
	return nil
}

func (p *postgresProvider) LookupVariant(variantName string) (string, bool) {
	variantName = strings.TrimSpace(variantName)
	if len(variantName) < 2 {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var canonicalID string
	err := p.db.QueryRowContext(ctx,
		`SELECT canonical_id FROM entity_variants WHERE LOWER(variant_name) = LOWER($1)`,
		variantName,
	).Scan(&canonicalID)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("[%s]", canonicalID), true
}

func (p *postgresProvider) GetEntityByToken(token string) (string, bool) {
	id := strings.Trim(token, "[]")
	if id == "" {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var name string
	err := p.db.QueryRowContext(ctx,
		`SELECT canonical_name FROM canonical_entities WHERE id = $1`, id,
	).Scan(&name)
	if err != nil {
		return "", false
	}
	return name, true
}

func (p *postgresProvider) SeedEntities(entries []EntitySeed) error {
	for _, e := range entries {
		if _, err := p.RegisterEntity(e.EntityType, e.CanonicalName, e.Variants); err != nil {
			return fmt.Errorf("[vault/postgres] SeedEntities(%q): %w", e.CanonicalName, err)
		}
	}
	return nil
}

func (p *postgresProvider) ListEntities() ([]EntityRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, `
		SELECT ce.id, ce.entity_type, ce.canonical_name, ev.variant_name
		FROM canonical_entities ce
		LEFT JOIN entity_variants ev ON ev.canonical_id = ce.id
		ORDER BY ce.entity_type, ce.id, ev.variant_name`)
	if err != nil {
		return nil, fmt.Errorf("[vault/postgres] ListEntities: %w", err)
	}
	defer rows.Close()

	indexed := make(map[string]*EntityRecord)
	var order []string

	for rows.Next() {
		var id, entityType, canonicalName string
		var variantName *string
		if err := rows.Scan(&id, &entityType, &canonicalName, &variantName); err != nil {
			return nil, err
		}
		if _, exists := indexed[id]; !exists {
			indexed[id] = &EntityRecord{
				ID:            id,
				EntityType:    entityType,
				CanonicalName: canonicalName,
			}
			order = append(order, id)
		}
		if variantName != nil {
			indexed[id].Variants = append(indexed[id].Variants, *variantName)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	records := make([]EntityRecord, 0, len(order))
	for _, id := range order {
		records = append(records, *indexed[id])
	}
	return records, nil
}
