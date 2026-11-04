//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const statementTimeoutCode = "57014"

func requireStatementTimeout(t *testing.T, err error) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, statementTimeoutCode, pgErr.Code)
}

// A pool with a short request budget cancels a slow statement, the same
// statement completes inside the background budget, and the budget does not
// leak back into the pooled connection.
func TestRunWithBackgroundBudgetOutlivesRequestTimeout(t *testing.T) {
	ctx := context.Background()
	dbURL := testutil.GetEnv("TEST_DATABASE_URL", "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable") // #nosec G101 -- local disposable test database only.
	cfg, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err)
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = "100ms"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	slow := "SELECT pg_sleep(0.3)"
	_, err = pool.Exec(ctx, slow)
	requireStatementTimeout(t, err)

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	require.NoError(t, RunWithBackgroundBudget(ctx, conn, func(tx pgx.Tx) error {
		_, execErr := tx.Exec(ctx, slow)
		return execErr
	}))

	_, err = conn.Exec(ctx, slow)
	requireStatementTimeout(t, err)

	sentinel := errors.New("refresh failed")
	require.ErrorIs(t, RunWithBackgroundBudget(ctx, conn, func(pgx.Tx) error { return sentinel }), sentinel)
}

// The concurrent refresh must run inside the budget transaction; this proves
// PostgreSQL accepts it there and that the production path completes.
func TestRefreshHotScoresRunsInsideBackgroundBudget(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer pool.Close()
	repo := NewClipRepository(pool)
	require.NoError(t, repo.RefreshHotScores(context.Background()))
}
