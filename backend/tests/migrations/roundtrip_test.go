//go:build integration

package migrations

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationFullRoundTrip applies every migration to an empty database,
// rolls all of them back, and applies them again. Every down migration must
// succeed, the rollback must leave no application objects behind, and the
// re-applied schema must match the first pass.
func TestMigrationFullRoundTrip(t *testing.T) {
	_, databaseURL := testutil.CreateEmptyDatabase(t, "roundtrip")
	ctx := context.Background()
	pool := connectMigrationDatabase(t, databaseURL)

	require.NoError(t, runMigrationOn(databaseURL, "up", 0), "initial up migration")
	initial, err := captureSchemaSnapshot(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, initial.Tables, "up migration should create tables")

	require.NoError(t, runMigrationOn(databaseURL, "down", -1), "full down migration")
	assert.Empty(t, residualApplicationObjects(t, ctx, pool), "down -all should remove every application object")

	require.NoError(t, runMigrationOn(databaseURL, "up", 0), "second up migration")
	final, err := captureSchemaSnapshot(ctx, pool)
	require.NoError(t, err)
	assert.Empty(t, compareSchemaSnapshots(initial, final), "schema should match after a full down/up cycle")
}

// residualApplicationObjects lists public-schema objects left after rolling
// back every migration. golang-migrate's own schema_migrations table and
// extension-owned objects are expected to remain.
func residualApplicationObjects(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []string {
	t.Helper()
	rows, err := pool.Query(ctx, `
		SELECT 'relation ' || c.relname || ' (' || c.relkind::text || ')'
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public'
			AND c.relname NOT IN ('schema_migrations', 'schema_migrations_pkey')
			AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid = c.oid AND d.deptype = 'e')
		UNION ALL
		SELECT 'function ' || p.proname
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = 'public'
			AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid = p.oid AND d.deptype = 'e')
		UNION ALL
		SELECT 'type ' || ty.typname
		FROM pg_type ty
		JOIN pg_namespace n ON n.oid = ty.typnamespace
		WHERE n.nspname = 'public'
			AND ty.typtype IN ('e', 'd', 'c', 'r')
			AND NOT EXISTS (SELECT 1 FROM pg_class c WHERE c.reltype = ty.oid)
			AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid = ty.oid AND d.deptype = 'e')
		ORDER BY 1
	`)
	require.NoError(t, err)
	defer rows.Close()

	var residual []string
	for rows.Next() {
		var object string
		require.NoError(t, rows.Scan(&object))
		residual = append(residual, object)
	}
	require.NoError(t, rows.Err())
	return residual
}
