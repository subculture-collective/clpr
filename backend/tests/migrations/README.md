# Migration Tests

This directory contains comprehensive tests for database migrations, specifically focusing on moderation-related migrations.

## Overview

These tests verify that database migrations are applied correctly and can be rolled back cleanly. They test:

- **Schema Creation**: Tables, columns, and data types are created correctly
- **Constraints**: CHECK, UNIQUE, FOREIGN KEY, and NOT NULL constraints are enforced
- **Indexes**: Indexes are created and used efficiently
- **Triggers & Functions**: PostgreSQL triggers and functions work as expected
- **Data Integrity**: Data is preserved through migration operations
- **Rollback Safety**: Migrations can be rolled back cleanly

## Running Tests

### Prerequisites

1. Ensure test environment is set up:
   ```bash
   make test-setup
   ```

2. Install golang-migrate if not already installed:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

### Run Migration Tests

```bash
# Run all migration tests
go test -v -tags=integration ./tests/migrations/...

# Run specific test
go test -v -tags=integration ./tests/migrations/... -run TestModerationQueueMigration

# Run with coverage
go test -tags=integration ./tests/migrations/... -coverprofile=coverage.out
```

## Test Coverage

The tests cover the following migrations:

### Migration 000011 - Moderation Audit Logs
- ✅ Table creation
- ✅ Column schema
- ✅ Indexes for efficient querying
- ✅ Constraint enforcement

### Migration 000049 - Moderation Queue System
- ✅ moderation_queue table
- ✅ moderation_decisions table
- ✅ CHECK constraints (content_type, status, priority, confidence)
- ✅ UNIQUE constraint on pending queue items
- ✅ Foreign key relationships
- ✅ Indexes for queue filtering
- ✅ Automatic reviewed_at trigger

### Migration 000050 - Moderation Appeals
- ✅ moderation_appeals table
- ✅ Status constraints
- ✅ UNIQUE constraint on pending appeals
- ✅ Foreign key to moderation_decisions
- ✅ Automatic resolved_at trigger

### Migration 000069 - Forum Moderation
- ✅ forum_threads table
- ✅ forum_replies table
- ✅ moderation_actions table
- ✅ user_bans table
- ✅ content_flags table
- ✅ Reply count trigger
- ✅ Flag count trigger

### Migration 000097 - Updated Moderation Audit Logs
- ✅ New columns (actor_id, target_user_id, channel_id, ip_address, user_agent)
- ✅ New indexes
- ✅ Backward compatibility with old columns

## Test Structure

### Helper Functions

- `tableExists()` - Check if table exists
- `columnExists()` - Check if column exists in table
- `indexExists()` - Check if index exists
- `constraintExists()` - Check if constraint exists
- `functionExists()` - Check if PostgreSQL function exists
- `triggerExists()` - Check if trigger exists
- `getColumnType()` - Get data type of column
- `isColumnNullable()` - Check if column allows NULL

### Test Categories

1. **Schema Tests**: Verify tables, columns, and types match migration definitions
2. **Constraint Tests**: Verify all constraints are properly enforced
3. **Index Tests**: Verify indexes exist and are used by query planner
4. **Trigger Tests**: Verify triggers fire correctly and enforce business logic
5. **Data Integrity Tests**: Verify data is preserved and cascade deletes work
6. **Type Tests**: Verify column data types and nullability

## Example Test

```go
t.Run("ModerationQueueStatusConstraint", func(t *testing.T) {
    // Valid status should succeed
    queueID := uuid.New()
    _, err := mh.pool.Exec(ctx, `
        INSERT INTO moderation_queue (id, content_type, content_id, reason, status)
        VALUES ($1, 'comment', $2, 'spam', 'pending')
    `, queueID, uuid.New())
    require.NoError(t, err)

    // Invalid status should fail
    _, err = mh.pool.Exec(ctx, `
        INSERT INTO moderation_queue (id, content_type, content_id, reason, status)
        VALUES ($1, 'comment', $2, 'spam', 'invalid_status')
    `, uuid.New(), uuid.New())
    assert.Error(t, err, "Invalid status should fail constraint check")
})
```

## Migration Rollback Drills

### Shadow Database Testing

The migration rollback drills (`shadow_db_drills_test.go`) provide comprehensive automated testing of database migrations in a shadow database environment. These tests ensure migrations can be safely applied and rolled back without leaving residual objects or causing schema drift.

#### What It Tests

1. **Full Migration Cycle**
   - Captures initial schema snapshot
   - Rolls back one migration
   - Re-applies the migration
   - Verifies final schema matches initial state
   - No residual objects (tables, indexes, triggers, constraints, functions)

2. **Integrity Validation**
   - Referential integrity (foreign key constraints)
   - Index integrity (no orphaned indexes)
   - Constraint integrity (no orphaned constraints)
   - Trigger integrity (no orphaned triggers)
   - Data corruption checks with test fixtures

3. **Performance Baseline Recording**
   - Tracks migration execution time (forward and backward)
   - Reports stored in `test-reports/migration-drills/`
   - Thresholds: 30 seconds for up/down migrations
   - JSON format for easy parsing and trending

4. **Drift Detection**
   - Detects unexpected schema changes
   - Compares snapshots to identify drift
   - Validates migrations don't leave residual state

#### Running Drills Locally

```bash
# Run all migration drills
go test -v -tags=integration -run="TestShadowDatabaseMigrationDrills" ./tests/migrations/...

# Run specific drill tests
go test -v -tags=integration -run="TestMigrationDriftDetection" ./tests/migrations/...
go test -v -tags=integration -run="TestResidualObjectsDetection" ./tests/migrations/...

# Run with timeout for long-running drills
go test -v -tags=integration -timeout=30m ./tests/migrations/...
```

#### Performance Reports

Performance baselines are saved to `test-reports/migration-drills/`:
- `perf-up-<timestamp>.json` - Forward migration timing
- `perf-down-<timestamp>.json` - Rollback migration timing
- `initial-snapshot.json` - Schema before operations
- `after-rollback-snapshot.json` - Schema after rollback
- `final-snapshot.json` - Schema after re-applying

#### Updating Performance Thresholds

Edit `shadow_db_drills_test.go`:
```go
const (
    PerformanceThresholdUp = 30.0   // seconds for forward migrations
    PerformanceThresholdDown = 30.0 // seconds for rollback migrations
)
```

#### CI Integration

The drills run automatically in CI on:
- Push to main/develop branches (when migrations change)
- Pull requests affecting migrations
- Manual workflow dispatch

CI fails on:
- Schema drift after rollback cycle
- Residual objects after rollback
- Integrity check failures

CI warns on:
- Performance threshold violations (non-blocking)

## Adding New Migration Tests

When adding a new migration test:

1. Create a new test function following the naming convention: `TestXXXMigration000NNN`
2. Use the `setupMigrationTest()` helper to get database connections
3. Test schema creation (tables, columns, constraints, indexes)
4. Test constraint enforcement with valid and invalid data
5. Test triggers and functions if applicable
6. Clean up test data properly

## Notes

- Tests use the `//go:build integration` build tag
- Test database is isolated (clpr_test on port 5437)
- Coverage metrics are not applicable as these are schema tests
- All tests should clean up after themselves
- Tests should be idempotent and can run in any order
- Migration drills use shadow database to test rollback safety
