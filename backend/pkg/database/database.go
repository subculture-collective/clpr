package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"git.subcult.tv/subculture-collective/clpr/config"
)

// DB holds the database connection pool
type DB struct {
	Pool *pgxpool.Pool
	SQL  *sql.DB
}

// NewDB creates a new database connection pool
func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	return NewDBWithTracing(cfg, false)
}

// NewDBWithTracing creates a new database connection pool with optional tracing
func NewDBWithTracing(cfg *config.DatabaseConfig, enableTracing bool) (*DB, error) {
	ctx := context.Background()

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = 25                               // Maximum number of connections
	poolConfig.MinConns = 5                                // Minimum number of connections
	poolConfig.MaxConnLifetime = time.Hour                 // Maximum connection lifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute          // Maximum idle time
	poolConfig.HealthCheckPeriod = time.Minute             // Health check interval
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second // Connection timeout

	// Add tracing if enabled
	if enableTracing {
		poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()
		log.Println("Database tracing enabled")
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Database connection pool established successfully")

	// Also create a database/sql DB using pgx stdlib for compatibility with
	// code that expects database/sql interfaces (ExecContext, QueryContext, Tx, etc).
	sqlDB, err := sql.Open("pgx", cfg.GetDatabaseURL())
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to open database/sql connection: %w", err)
	}

	// Configure sql.DB pool settings to match pgxpool where reasonable
	sqlDB.SetMaxOpenConns(int(poolConfig.MaxConns))
	sqlDB.SetMaxIdleConns(int(poolConfig.MinConns))
	sqlDB.SetConnMaxLifetime(poolConfig.MaxConnLifetime)

	// Verify sql.DB connection
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		pool.Close()
		return nil, fmt.Errorf("unable to ping database/sql DB: %w", err)
	}

	return &DB{Pool: pool, SQL: sqlDB}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.SQL != nil {
		db.SQL.Close()
	}
	if db.Pool != nil {
		db.Pool.Close()
	}
	log.Println("Database connection pools closed")
}

// HealthCheck checks if the database is accessible
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats returns connection pool statistics
func (db *DB) GetStats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// ExecContext delegates to the database/sql DB to return sql.Result-compatible values.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if db.SQL == nil {
		return nil, fmt.Errorf("sql DB not initialized")
	}
	return db.SQL.ExecContext(ctx, query, args...)
}

// QueryContext delegates to database/sql
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if db.SQL == nil {
		return nil, fmt.Errorf("sql DB not initialized")
	}
	return db.SQL.QueryContext(ctx, query, args...)
}

// QueryRowContext delegates to database/sql
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.SQL.QueryRowContext(ctx, query, args...)
}

// BeginTx delegates to database/sql
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if db.SQL == nil {
		return nil, fmt.Errorf("sql DB not initialized")
	}
	return db.SQL.BeginTx(ctx, opts)
}
