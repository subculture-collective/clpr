//go:build integration

package migrations

import (
	"context"
	"os"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type migrationTestHelper struct{ pool *pgxpool.Pool }

// TestMain gives the drills their own database. They roll the schema back and
// forward, which must never be visible to other packages' tests.
func TestMain(m *testing.M) {
	os.Exit(testutil.RunWithDatabase(m, "migrations"))
}

// Use exactly the database selected by runMigration, so schema assertions and
// CLI migrations cannot accidentally target different databases.
func setupMigrationTest(t *testing.T) *migrationTestHelper {
	t.Helper()
	testutil.RequirePackageDatabase(t)
	database := testutil.DatabaseConfig().Name
	require.Equal(t, testutil.PackageDatabaseName(), database, "migration drills require their disposable package database")
	return &migrationTestHelper{pool: connectMigrationDatabase(t, testutil.DatabaseURL())}
}

func connectMigrationDatabase(t *testing.T, databaseURL string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(context.Background()))
	return pool
}

func (h *migrationTestHelper) exists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := h.pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}
func (h *migrationTestHelper) tableExists(ctx context.Context, table string) (bool, error) {
	return h.exists(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)", table)
}
func (h *migrationTestHelper) columnExists(ctx context.Context, table, column string) (bool, error) {
	return h.exists(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2)", table, column)
}
func (h *migrationTestHelper) indexExists(ctx context.Context, index string) (bool, error) {
	return h.exists(ctx, "SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = $1)", index)
}
func (h *migrationTestHelper) functionExists(ctx context.Context, function string) (bool, error) {
	return h.exists(ctx, "SELECT EXISTS (SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = 'public' AND p.proname = $1)", function)
}
func (h *migrationTestHelper) triggerExists(ctx context.Context, table, trigger string) (bool, error) {
	return h.exists(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.triggers WHERE trigger_schema = 'public' AND event_object_table = $1 AND trigger_name = $2)", table, trigger)
}
