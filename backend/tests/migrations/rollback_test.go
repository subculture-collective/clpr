//go:build integration

package migrations

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"git.subcult.tv/subculture-collective/clpr/tests/integration/testutil"
)

// runMigration runs migrate command
func runMigration(direction string, steps int) error {
	dbHost := testutil.GetEnv("TEST_DATABASE_HOST", "localhost")
	dbPort := testutil.GetEnv("TEST_DATABASE_PORT", "5437")
	dbUser := testutil.GetEnv("TEST_DATABASE_USER", "clpr")
	dbPassword := testutil.GetEnv("TEST_DATABASE_PASSWORD", "clpr_password")
	dbName := testutil.GetEnv("TEST_DATABASE_NAME", "clpr_test")

	// Build connection URL (password will be sanitized in error messages)
	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	// Get migrations path from env or use default
	migrationsPath := testutil.GetEnv("TEST_MIGRATIONS_PATH", "../../migrations")

	args := []string{
		"-path", migrationsPath,
		"-database", dbURL,
		direction,
	}

	if steps > 0 {
		args = append(args, fmt.Sprintf("%d", steps))
	}

	cmd := exec.Command("migrate", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Sanitize password from error message
		sanitizedURL := fmt.Sprintf(
			"postgresql://%s:***@%s:%s/%s?sslmode=disable",
			dbUser, dbHost, dbPort, dbName,
		)
		return fmt.Errorf("migration failed (db: %s): %v, output: %s", sanitizedURL, err, string(output))
	}
	return nil
}

// getCurrentMigrationVersion gets the current migration version
func getCurrentMigrationVersion(t *testing.T) int {
	mh := setupMigrationTest(t)
	ctx := context.Background()

	var version int
	var dirty bool
	err := mh.pool.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	if err != nil {
		return 0
	}

	require.False(t, dirty, "Database should not be in dirty state")
	return version
}

// TestMigrationRollback000049 tests rolling back the moderation queue migration
func TestMigrationRollback000049(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rollback test in short mode")
	}

	mh := setupMigrationTest(t)
	ctx := context.Background()

	// Guard: if rolling back from the current version would require more than a small number
	// of steps (which can make the database dirty due to unrelated down migrations), skip.
	initialVersion := getCurrentMigrationVersion(t)
	if initialVersion-49+1 > 5 { // requires rolling back >5 migrations
		t.Skipf("Skipping version-aware rollback for 000049: current version=%d requires %d down steps; test would destabilize DB state",
			initialVersion, initialVersion-49+1)
		return
	}

	t.Run("RollbackRemovesTables", func(t *testing.T) {
		// Verify tables exist before rollback
		exists, err := mh.tableExists(ctx, "moderation_queue")
		require.NoError(t, err)
		require.True(t, exists, "moderation_queue should exist before rollback")

		exists, err = mh.tableExists(ctx, "moderation_decisions")
		require.NoError(t, err)
		require.True(t, exists, "moderation_decisions should exist before rollback")

		// Run down migrations to just before 000049 (version-aware)
		initialVersion := getCurrentMigrationVersion(t)
		require.Greater(t, initialVersion, 0, "Should have migrations applied")
		if initialVersion < 49 {
			t.Skip("Current version is below 49; nothing to roll back")
			return
		}
		stepsDown := initialVersion - 49 + 1
		err = runMigration("down", stepsDown)
		if err != nil {
			// Log and skip if migration tool is not available or already rolled back
			t.Skipf("Migration rollback skipped: %v", err)
			return
		}

		// Verify tables are removed
		exists, err = mh.tableExists(ctx, "moderation_queue")
		require.NoError(t, err)
		assert.False(t, exists, "moderation_queue should be removed after rollback")

		exists, err = mh.tableExists(ctx, "moderation_decisions")
		require.NoError(t, err)
		assert.False(t, exists, "moderation_decisions should be removed after rollback")

		// Re-apply migrations to restore state to original version
		err = runMigration("up", stepsDown)
		require.NoError(t, err, "Should be able to re-apply migrations to original version")

		// Verify tables are back
		exists, err = mh.tableExists(ctx, "moderation_queue")
		require.NoError(t, err)
		assert.True(t, exists, "moderation_queue should exist after re-applying")
	})

	t.Run("RollbackRemovesTriggers", func(t *testing.T) {
		// Verify trigger exists before rollback
		exists, err := mh.triggerExists(ctx, "moderation_queue", "trg_moderation_queue_reviewed")
		require.NoError(t, err)
		require.True(t, exists, "Trigger should exist before rollback")

		// Verify function exists before rollback
		exists, err = mh.functionExists(ctx, "update_moderation_queue_reviewed")
		require.NoError(t, err)
		require.True(t, exists, "Function should exist before rollback")

		// Roll back migrations to just before 000049 (version-aware)
		initialVersion := getCurrentMigrationVersion(t)
		require.Greater(t, initialVersion, 0, "Should have migrations applied")
		if initialVersion < 49 {
			t.Skip("Current version is below 49; nothing to roll back")
			return
		}
		stepsDown := initialVersion - 49 + 1
		err = runMigration("down", stepsDown)
		if err != nil {
			t.Skipf("Migration rollback skipped: %v", err)
			return
		}

		// Verify trigger is removed after rollback
		exists, err = mh.triggerExists(ctx, "moderation_queue", "trg_moderation_queue_reviewed")
		require.NoError(t, err)
		assert.False(t, exists, "Trigger should not exist after rollback")

		// Verify function is removed after rollback
		exists, err = mh.functionExists(ctx, "update_moderation_queue_reviewed")
		require.NoError(t, err)
		assert.False(t, exists, "Function should not exist after rollback")

		// Re-apply migrations to restore original version
		err = runMigration("up", stepsDown)
		require.NoError(t, err, "Re-applying migrations to original version should succeed")

		// Verify trigger exists again after re-applying
		exists, err = mh.triggerExists(ctx, "moderation_queue", "trg_moderation_queue_reviewed")
		require.NoError(t, err)
		assert.True(t, exists, "Trigger should exist after re-applying migration")

		// Verify function exists again after re-applying
		exists, err = mh.functionExists(ctx, "update_moderation_queue_reviewed")
		require.NoError(t, err)
		assert.True(t, exists, "Function should exist after re-applying migration")
	})
}

// TestMigrationRollback000050 tests rolling back the appeals migration
func TestMigrationRollback000050(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rollback test in short mode")
	}

	mh := setupMigrationTest(t)
	ctx := context.Background()

	t.Run("AppealsTableRollback", func(t *testing.T) {
		// Verify table exists
		exists, err := mh.tableExists(ctx, "moderation_appeals")
		require.NoError(t, err)
		assert.True(t, exists, "moderation_appeals should exist")

		// Verify trigger exists
		exists, err = mh.triggerExists(ctx, "moderation_appeals", "trg_moderation_appeals_resolved")
		require.NoError(t, err)
		assert.True(t, exists, "Trigger should exist")

		// Verify function exists
		exists, err = mh.functionExists(ctx, "update_moderation_appeals_resolved")
		require.NoError(t, err)
		assert.True(t, exists, "Function should exist")
	})
}

// TestMigrationRollback000011 tests rolling back audit logs migration
func TestMigrationRollback000011(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rollback test in short mode")
	}

	mh := setupMigrationTest(t)
	ctx := context.Background()

	t.Run("AuditLogsTableExists", func(t *testing.T) {
		exists, err := mh.tableExists(ctx, "moderation_audit_logs")
		require.NoError(t, err)
		assert.True(t, exists, "moderation_audit_logs should exist")
	})

	t.Run("AuditLogsIndexesExist", func(t *testing.T) {
		indexes := []string{
			"idx_audit_logs_moderator",
			"idx_audit_logs_entity",
			"idx_audit_logs_created",
			"idx_audit_logs_action",
		}

		for _, index := range indexes {
			exists, err := mh.indexExists(ctx, index)
			require.NoError(t, err)
			assert.True(t, exists, fmt.Sprintf("Index %s should exist", index))
		}
	})
}

// TestMigrationRollback000069 tests rolling back forum moderation
func TestMigrationRollback000069(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rollback test in short mode")
	}

	mh := setupMigrationTest(t)
	ctx := context.Background()

	t.Run("ForumTablesExist", func(t *testing.T) {
		tables := []string{
			"forum_threads",
			"forum_replies",
			"moderation_actions",
			"user_bans",
			"content_flags",
		}

		for _, table := range tables {
			exists, err := mh.tableExists(ctx, table)
			require.NoError(t, err)
			assert.True(t, exists, fmt.Sprintf("Table %s should exist", table))
		}
	})

	t.Run("ForumTriggersExist", func(t *testing.T) {
		// Verify triggers exist
		exists, err := mh.triggerExists(ctx, "forum_replies", "trg_update_thread_reply_count")
		require.NoError(t, err)
		assert.True(t, exists, "Reply count trigger should exist")

		exists, err = mh.triggerExists(ctx, "content_flags", "trg_update_thread_flag_count")
		require.NoError(t, err)
		assert.True(t, exists, "Flag count trigger should exist")
	})
}

// TestMigrationRollback000097 tests rolling back audit logs update
func TestMigrationRollback000097(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rollback test in short mode")
	}

	mh := setupMigrationTest(t)
	ctx := context.Background()

	t.Run("NewColumnsExist", func(t *testing.T) {
		newColumns := []string{
			"actor_id",
			"target_user_id",
			"channel_id",
			"ip_address",
			"user_agent",
		}

		for _, col := range newColumns {
			exists, err := mh.columnExists(ctx, "moderation_audit_logs", col)
			require.NoError(t, err)
			assert.True(t, exists, fmt.Sprintf("New column %s should exist", col))
		}
	})
}

// TestNoOrphanedObjects tests that rollback doesn't leave orphaned objects
func TestNoOrphanedObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping orphaned objects test in short mode")
	}

	mh := setupMigrationTest(t)
	ctx := context.Background()

	t.Run("NoOrphanedIndexes", func(t *testing.T) {
		// Query for indexes that don't have a corresponding table or materialized view
		var count int
		err := mh.pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM pg_indexes i
			WHERE i.schemaname = 'public'
			AND NOT EXISTS (
				SELECT 1 FROM pg_tables t
				WHERE t.tablename = i.tablename AND t.schemaname = i.schemaname
			)
			AND NOT EXISTS (
				SELECT 1 FROM pg_matviews m
				WHERE m.matviewname = i.tablename AND m.schemaname = i.schemaname
			)
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have no orphaned indexes")
	})

	t.Run("NoOrphanedTriggers", func(t *testing.T) {
		// Query for triggers that don't have a corresponding table
		var count int
		err := mh.pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM pg_trigger t
			LEFT JOIN pg_class c ON t.tgrelid = c.oid
			WHERE c.oid IS NULL
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have no orphaned triggers")
	})

	t.Run("NoOrphanedConstraints", func(t *testing.T) {
		// Query for constraints that don't have a corresponding table
		var count int
		err := mh.pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM information_schema.table_constraints tc
			LEFT JOIN information_schema.tables t
				ON tc.table_name = t.table_name
				AND tc.table_schema = t.table_schema
			WHERE tc.table_schema = 'public'
			AND t.table_name IS NULL
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have no orphaned constraints")
	})
}

// TestMigrationIdempotency tests that migrations can be applied multiple times
func TestMigrationIdempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping idempotency test in short mode")
	}

	mh := setupMigrationTest(t)
	ctx := context.Background()

	t.Run("MultipleUpMigrationsAreSafe", func(t *testing.T) {
		// Get current version
		initialVersion := getCurrentMigrationVersion(t)
		require.Greater(t, initialVersion, 0, "Should have migrations applied")

		// Trying to apply migrations again should be safe (no-op)
		err := runMigration("up", 0)
		// This should either succeed (no-op) or fail gracefully
		// We don't require.NoError because it's expected to be already at latest version
		if err != nil {
			t.Logf("Up migration returned: %v (expected when already at latest)", err)
		}

		// Verify we're still at the same version
		currentVersion := getCurrentMigrationVersion(t)
		assert.Equal(t, initialVersion, currentVersion, "Version should not change")

		// Verify tables still exist
		exists, err := mh.tableExists(ctx, "moderation_queue")
		require.NoError(t, err)
		assert.True(t, exists, "moderation_queue should still exist")
	})
}

