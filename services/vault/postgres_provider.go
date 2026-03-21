package vault

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// postgresProvider implements Provider against a centralised PostgreSQL database.
// This is the Enterprise backend that enables multi-server deployments sharing a
// single vault.
type postgresProvider struct {
	db *sql.DB
}

// newPostgresProvider connects to PostgreSQL using dsn, creates the vault table
// if it does not yet exist (idempotent), and returns the ready provider.
func newPostgresProvider(dsn string) (*postgresProvider, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("[vault/postgres] sql.Open: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("[vault/postgres] ping failed: %w", err)
	}

	// HA Clustering: Configure robust connection pooling for concurrent data center access
	db.SetMaxOpenConns(15) // Restricted to 15 to fit within default 100 connection limit of bare-metal PG containers across 5 horizontal nodes
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * 60 * 1000000000) // 5 minutes (in ns)

	// PostgreSQL-compatible DDL — note ON CONFLICT instead of DuckDB's OR IGNORE
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS vault (
			pii_hash     TEXT PRIMARY KEY,
			token        TEXT NOT NULL,
			encrypted_pii TEXT NOT NULL
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("[vault/postgres] CREATE TABLE: %w", err)
	}

	// HA Clustering: Proper indexing strategy for reverse token lookups (O(1) retrieval)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_vault_token ON vault(token)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("[vault/postgres] CREATE INDEX: %w", err)
	}

	p := &postgresProvider{db: db}
	log.Printf("[vault/postgres] Connected to centralized vault.")
	return p, nil
}

// GetToken retrieves the vaulted token for a given PII hash from PostgreSQL.
func (p *postgresProvider) GetToken(hash string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var token string
	err := p.db.QueryRowContext(ctx, `SELECT token FROM vault WHERE pii_hash = $1`, hash).Scan(&token)
	if err != nil {
		return "", false
	}
	return token, true
}

// StoreToken inserts a new PII hash and its secure token into the PostgreSQL vault.
func (p *postgresProvider) StoreToken(hash, token, encryptedPII string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := p.db.ExecContext(ctx,
		`INSERT INTO vault (pii_hash, token, encrypted_pii)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (pii_hash) DO NOTHING`,
		hash, token, encryptedPII,
	)
	if err != nil {
		log.Printf("[vault/postgres] StoreToken error: %v", err)
		return false, err
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		return true, nil
	}
	return false, nil
}

// CountAll returns the total number of records currently stored in the PostgreSQL vault.
func (p *postgresProvider) CountAll() int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	// Fast estimate for large tables to prevent locking
	err := p.db.QueryRowContext(ctx, `SELECT reltuples::bigint FROM pg_class WHERE relname = 'vault'`).Scan(&count)
	if err == nil && count > 1000 {
		return count
	}
	// Fallback for small tables where stats might be zero
	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM vault`).Scan(&count)
	return count
}

// Close terminates the PostgreSQL connection explicitly.
func (p *postgresProvider) Close() error {
	return p.db.Close()
}

// GetEncryptedByToken performs a reverse lookup by token string.
// Satisfies the internal tokenLookup interface used by engine.DecryptToken.
func (p *postgresProvider) GetEncryptedByToken(token string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var enc string
	err := p.db.QueryRowContext(ctx, `SELECT encrypted_pii FROM vault WHERE token = $1`, token).Scan(&enc)
	if err != nil {
		return "", false
	}
	return enc, true
}
