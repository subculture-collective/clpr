package database

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestApplyConnectionSafetySetsDatabaseTimeouts(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig("postgresql://user:pass@localhost:5432/database?sslmode=disable")
	require.NoError(t, err)

	applyConnectionSafety(poolConfig)

	require.Equal(t, databaseStatementTimeout, poolConfig.ConnConfig.RuntimeParams["statement_timeout"])
	require.Equal(t, databaseLockTimeout, poolConfig.ConnConfig.RuntimeParams["lock_timeout"])
}
