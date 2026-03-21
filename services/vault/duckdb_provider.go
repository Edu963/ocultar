package vault

import (
	"context"
	"database/sql"
	"log"
	"sync/atomic"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// duckdbProvider implements Provider against a local DuckDB file (or :memory:).
// This is the default backend for Community and single-server Enterprise deployments.
type duckdbProvider struct {
	db    *sql.DB
	count atomic.Int64
}

// newDuckDBProvider opens a DuckDB database at vaultPath, creates the vault
// table if it doesn't exist, and seeds the internal counter.
func newDuckDBProvider(vaultPath string) (*duckdbProvider, error) {
	db, err := sql.Open("duckdb", vaultPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS vault (pii_hash VARCHAR PRIMARY KEY, token VARCHAR, encrypted_pii VARCHAR)`,
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Fix DuckDB Concurrency: limit to a single open connection to prevent "database is locked" errors
	db.SetMaxOpenConns(1)

	var count int64
	db.QueryRow(`SELECT COUNT(*) FROM vault`).Scan(&count)

	p := &duckdbProvider{db: db}
	p.count.Store(count)
	return p, nil
}

// GetToken retrieves the vaulted token for a given PII hash from DuckDB.
func (p *duckdbProvider) GetToken(hash string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var token string
	err := p.db.QueryRowContext(ctx, `SELECT token FROM vault WHERE pii_hash = ?`, hash).Scan(&token)
	if err != nil {
		return "", false
	}
	return token, true
}

// StoreToken inserts a new PII hash and its secure token into the DuckDB vault.
func (p *duckdbProvider) StoreToken(hash, token, encryptedPII string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := p.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO vault (pii_hash, token, encrypted_pii) VALUES (?, ?, ?)`,
		hash, token, encryptedPII,
	)
	if err != nil {
		log.Printf("[vault/duckdb] StoreToken error: %v", err)
		return false, err
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		p.count.Add(1)
		return true, nil
	}
	return false, nil
}

// CountAll returns the total number of records currently stored in the DuckDB vault.
func (p *duckdbProvider) CountAll() int64 {
	return p.count.Load()
}

// Close terminates the DuckDB connection explicitly.
func (p *duckdbProvider) Close() error {
	return p.db.Close()
}

// GetEncryptedByToken performs a reverse lookup: given a vault token string
// (e.g. "[EMAIL_a1b2c3d4]"), it returns the encrypted_pii blob.
// This satisfies the internal tokenLookup interface used by engine.DecryptToken.
func (p *duckdbProvider) GetEncryptedByToken(token string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var enc string
	err := p.db.QueryRowContext(ctx, `SELECT encrypted_pii FROM vault WHERE token = ?`, token).Scan(&enc)
	if err != nil {
		return "", false
	}
	return enc, true
}
