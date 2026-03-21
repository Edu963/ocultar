package vault

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Edu963/ocultar/pkg/config"
)

// MigrateDuckDBtoPostgres reads all records from the DuckDB provider and inserts them into PostgreSQL.
func MigrateDuckDBtoPostgres(duckdbProv Provider, pgDSN string) error {
	ddb, ok := duckdbProv.(*duckdbProvider)
	if !ok {
		return fmt.Errorf("active provider is not duckdb")
	}

	pg, err := newPostgresProvider(pgDSN)
	if err != nil {
		return fmt.Errorf("failed connecting to postgres: %w", err)
	}
	defer pg.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	rows, err := ddb.db.QueryContext(ctx, "SELECT pii_hash, token, encrypted_pii FROM vault")
	if err != nil {
		return fmt.Errorf("failed reading from duckdb: %w", err)
	}
	defer rows.Close()

	log.Printf("[migration] Starting bulk transfer to PostgreSQL...")
	var count int
	for rows.Next() {
		var hash, token, encrypted string
		if err := rows.Scan(&hash, &token, &encrypted); err != nil {
			return fmt.Errorf("failed scanning row: %w", err)
		}
		_, err = pg.StoreToken(hash, token, encrypted)
		if err != nil {
			return fmt.Errorf("failed storing in postgres: %w", err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	log.Printf("[migration] Successfully transferred %d tokens.", count)

	config.Global.VaultBackend = "postgres"
	config.Global.PostgresDSN = pgDSN
	if err := config.Save(); err != nil {
		return fmt.Errorf("migration succeeded but saving config.yaml failed: %w", err)
	}

	return nil
}
