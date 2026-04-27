package middleware

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// DuckDBUsageStore implements UsageStore using a local DuckDB database.
// It persists the api_usage table alongside the vault.db file.
type DuckDBUsageStore struct {
	db *sql.DB
}

// NewDuckDBUsageStore opens (or creates) a DuckDB database at path and
// initialises the api_usage table.
func NewDuckDBUsageStore(path string) (*DuckDBUsageStore, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // DuckDB does not allow concurrent writers

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS api_usage (
			api_key_hash TEXT    NOT NULL,
			month        TEXT    NOT NULL,
			call_count   INTEGER DEFAULT 0,
			PRIMARY KEY (api_key_hash, month)
		)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &DuckDBUsageStore{db: db}, nil
}

// Increment upserts a row for (apiKeyHash, month), increments call_count by
// one, and returns the new total. The INSERT OR IGNORE + UPDATE pattern
// matches the vault package convention and is safe under SetMaxOpenConns(1).
func (s *DuckDBUsageStore) Increment(apiKeyHash, month string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO api_usage (api_key_hash, month, call_count) VALUES (?, ?, 0)`,
		apiKeyHash, month,
	); err != nil {
		log.Printf("[usage] insert error: %v", err)
		return 0, err
	}

	if _, err := s.db.ExecContext(ctx,
		`UPDATE api_usage SET call_count = call_count + 1 WHERE api_key_hash = ? AND month = ?`,
		apiKeyHash, month,
	); err != nil {
		log.Printf("[usage] update error: %v", err)
		return 0, err
	}

	var count int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT call_count FROM api_usage WHERE api_key_hash = ? AND month = ?`,
		apiKeyHash, month,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Count returns the current call total for (apiKeyHash, month).
// Returns 0 if no row exists yet.
func (s *DuckDBUsageStore) Count(apiKeyHash, month string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	err := s.db.QueryRowContext(ctx,
		`SELECT call_count FROM api_usage WHERE api_key_hash = ? AND month = ?`,
		apiKeyHash, month,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// Close releases the DuckDB connection.
func (s *DuckDBUsageStore) Close() error {
	return s.db.Close()
}
