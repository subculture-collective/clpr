//go:build integration

package migrations

import (
	"context"
	"net"
	"net/url"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type migrationTestHelper struct{ pool *pgxpool.Pool }

// Use exactly the database selected by runMigration, so schema assertions and
// CLI migrations cannot accidentally target different databases.
func setupMigrationTest(t *testing.T) *migrationTestHelper {
	t.Helper()
	database := testutil.GetEnv("TEST_DATABASE_NAME", "clpr_test")
	require.Equal(t, "clpr_test", database, "migration drills require the disposable clpr_test database")
	target := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(testutil.GetEnv("TEST_DATABASE_USER", "clpr"), testutil.GetEnv("TEST_DATABASE_PASSWORD", "clpr_password")),
		Host:     net.JoinHostPort(testutil.GetEnv("TEST_DATABASE_HOST", "localhost"), testutil.GetEnv("TEST_DATABASE_PORT", "5437")),
		Path:     database,
		RawQuery: "sslmode=disable",
	}
	pool, err := pgxpool.New(context.Background(), target.String())
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(context.Background()))
	return &migrationTestHelper{pool: pool}
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
