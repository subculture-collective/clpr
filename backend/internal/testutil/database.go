package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupTestDB creates a test database connection pool for integration tests
// It expects PostgreSQL to be running (e.g., via Docker Compose)
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Use test database configuration
	dbURL := "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Failed to create test database pool: %v", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return pool
}

// CleanupTestDB closes the database pool and cleans up test data
func CleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if pool != nil {
		pool.Close()
	}
}

// TruncateTables removes all data from test tables
func TruncateTables(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()

	ctx := context.Background()
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		_, err := pool.Exec(ctx, query)
		if err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}
