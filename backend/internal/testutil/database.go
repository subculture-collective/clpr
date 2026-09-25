package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Test database defaults match docker-compose.test.yml. They are local-only and
// intentionally cannot target a production hostname.
const (
	defaultDatabaseHost     = "localhost"
	defaultDatabasePort     = "5437"
	defaultDatabaseUser     = "clpr"
	defaultDatabasePassword = "clpr_password" // #nosec G101 -- local disposable test database only.
	defaultDatabaseName     = "clpr_test"

	// adminDatabase is the maintenance database used for CREATE/DROP DATABASE,
	// so provisioning never holds a session on the migrated base database.
	adminDatabase = "postgres"
	// provisionLockKey serializes template refreshes and clones across the
	// concurrently running test binaries of one `go test ./...` invocation.
	provisionLockKey int64 = 0x636c70725f746d70 // "clpr_tmp"
	maxIdentifierLen       = 63
)

var (
	// packageDatabaseName is set once TestMain provisions an isolated database.
	packageDatabaseName string
	// provisionErr records why TestMain could not provision a database.
	provisionErr error
	// errDatabaseUnavailable marks a missing server or base database, which lets
	// unit-mode binaries continue so database-dependent tests can skip.
	errDatabaseUnavailable = errors.New("test database unavailable")
	identifierUnsafe       = regexp.MustCompile(`[^a-z0-9_]+`)
)

// GetEnv returns an environment variable or a local-test fallback when unset.
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// DatabaseConfig returns the test database settings. TEST_DATABASE_URL takes
// precedence over the individual TEST_DATABASE_* variables. After
// RunWithDatabase provisions a package database, the settings point at it.
func DatabaseConfig() *config.DatabaseConfig {
	cfg := &config.DatabaseConfig{
		Host:     GetEnv("TEST_DATABASE_HOST", defaultDatabaseHost),
		Port:     GetEnv("TEST_DATABASE_PORT", defaultDatabasePort),
		User:     GetEnv("TEST_DATABASE_USER", defaultDatabaseUser),
		Password: GetEnv("TEST_DATABASE_PASSWORD", defaultDatabasePassword),
		Name:     GetEnv("TEST_DATABASE_NAME", defaultDatabaseName),
		SSLMode:  "disable",
	}
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		return cfg
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return cfg
	}
	if host := parsed.Hostname(); host != "" {
		cfg.Host = host
	}
	if port := parsed.Port(); port != "" {
		cfg.Port = port
	}
	if parsed.User != nil {
		if user := parsed.User.Username(); user != "" {
			cfg.User = user
		}
		if password, ok := parsed.User.Password(); ok {
			cfg.Password = password
		}
	}
	if name := strings.TrimPrefix(parsed.Path, "/"); name != "" {
		cfg.Name = name
	}
	if mode := parsed.Query().Get("sslmode"); mode != "" {
		cfg.SSLMode = mode
	}
	return cfg
}

// DatabaseURL returns the connection URL for the test database.
func DatabaseURL() string {
	cfg := DatabaseConfig()
	return databaseURL(cfg, cfg.Name)
}

// PackageDatabaseName returns the isolated database provisioned for this test
// binary, or an empty string when RunWithDatabase did not provision one.
func PackageDatabaseName() string {
	return packageDatabaseName
}

func databaseURL(cfg *config.DatabaseConfig, name string) string {
	target := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     name,
		RawQuery: url.Values{"sslmode": {cfg.SSLMode}}.Encode(),
	}
	return target.String()
}

// RunWithDatabase gives one test binary (one Go package) its own database and
// runs its tests against it. Call it from TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(testutil.RunWithDatabase(m, "repository")) }
//
// The database is cloned from "<base>_template", which is refreshed from the
// migrated base database (TEST_DATABASE_NAME, default clpr_test) whenever their
// schema_migrations versions differ. Packages therefore share no rows, locks,
// or schema state and can run in parallel. TEST_DATABASE_* and
// TEST_DATABASE_URL are rewritten to the package database for the duration of
// the run. Set TEST_DATABASE_KEEP=1 to keep the database for inspection.
//
// When the server or base database is unreachable, tests still run so that
// database-dependent tests can skip or fail with the recorded reason.
func RunWithDatabase(m *testing.M, pkg string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	base := DatabaseConfig()
	name, err := provisionPackageDatabase(ctx, base, pkg)
	cancel()
	if err != nil {
		provisionErr = err
		if !errors.Is(err, errDatabaseUnavailable) {
			fmt.Fprintf(os.Stderr, "testutil: provisioning %s database failed: %v\n", pkg, err)
			return 1
		}
		return m.Run()
	}

	packageDatabaseName = name
	restore := setDatabaseEnv(base, name)
	code := m.Run()
	restore()

	if os.Getenv("TEST_DATABASE_KEEP") == "1" {
		fmt.Fprintf(os.Stderr, "testutil: kept test database %s\n", name)
		return code
	}
	dropCtx, dropCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dropCancel()
	if err := dropDatabase(dropCtx, base, name); err != nil {
		fmt.Fprintf(os.Stderr, "testutil: dropping %s failed: %v\n", name, err)
	}
	return code
}

// CreateEmptyDatabase creates an unmigrated database on the test server and
// drops it when the test finishes. It returns the database name and URL.
func CreateEmptyDatabase(t *testing.T, purpose string) (string, string) {
	t.Helper()
	if provisionErr != nil {
		t.Fatalf("test database unavailable: %v", provisionErr)
	}
	cfg := DatabaseConfig()
	name, err := uniqueDatabaseName(baseDatabaseName(cfg), purpose)
	if err != nil {
		t.Fatalf("naming test database: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	admin, err := connectAdmin(ctx, cfg)
	if err != nil {
		t.Fatalf("connecting to test server: %v", err)
	}
	defer admin.Close(context.Background())
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatalf("creating %s: %v", name, err)
	}
	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), time.Minute)
		defer dropCancel()
		if err := dropDatabase(dropCtx, cfg, name); err != nil {
			t.Errorf("dropping %s: %v", name, err)
		}
	})
	return name, databaseURL(cfg, name)
}

// baseDatabaseName recovers the migrated base database name while a package
// database is active, so helpers never clone or modify another package's data.
func baseDatabaseName(cfg *config.DatabaseConfig) string {
	if base := os.Getenv("TEST_DATABASE_BASE_NAME"); base != "" {
		return base
	}
	return cfg.Name
}

func setDatabaseEnv(base *config.DatabaseConfig, name string) func() {
	values := map[string]string{
		"TEST_DATABASE_HOST":      base.Host,
		"TEST_DATABASE_PORT":      base.Port,
		"TEST_DATABASE_USER":      base.User,
		"TEST_DATABASE_PASSWORD":  base.Password,
		"TEST_DATABASE_NAME":      name,
		"TEST_DATABASE_BASE_NAME": base.Name,
		"TEST_DATABASE_URL":       databaseURL(base, name),
	}
	previous := make(map[string]*string, len(values))
	for key, value := range values {
		if old, ok := os.LookupEnv(key); ok {
			previous[key] = &old
		} else {
			previous[key] = nil
		}
		_ = os.Setenv(key, value)
	}
	return func() {
		for key, old := range previous {
			if old == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *old)
			}
		}
	}
}

func connectAdmin(ctx context.Context, cfg *config.DatabaseConfig) (*pgx.Conn, error) {
	return pgx.Connect(ctx, databaseURL(cfg, adminDatabase))
}

func provisionPackageDatabase(ctx context.Context, base *config.DatabaseConfig, pkg string) (string, error) {
	admin, err := connectAdmin(ctx, base)
	if err != nil {
		return "", fmt.Errorf("%w: connecting to %s:%s: %v", errDatabaseUnavailable, base.Host, base.Port, err)
	}
	defer admin.Close(context.Background())

	var baseExists bool
	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", base.Name).Scan(&baseExists); err != nil {
		return "", fmt.Errorf("checking base database: %w", err)
	}
	if !baseExists {
		return "", fmt.Errorf("%w: base database %q does not exist on %s:%s; run task test:setup", errDatabaseUnavailable, base.Name, base.Host, base.Port)
	}

	if _, err := admin.Exec(ctx, "SELECT pg_advisory_lock($1)", provisionLockKey); err != nil {
		return "", fmt.Errorf("acquiring provisioning lock: %w", err)
	}
	defer func() { _, _ = admin.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", provisionLockKey) }()

	template, err := ensureTemplate(ctx, admin, base)
	if err != nil {
		return "", err
	}
	name, err := uniqueDatabaseName(base.Name, pkg)
	if err != nil {
		return "", err
	}
	if _, err := admin.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s",
		pgx.Identifier{name}.Sanitize(), pgx.Identifier{template}.Sanitize())); err != nil {
		return "", fmt.Errorf("creating %s from %s: %w", name, template, err)
	}
	return name, nil
}

// ensureTemplate returns a template database matching the base database's
// migration version, recreating it from the base database when it is stale.
// The caller must hold provisionLockKey.
func ensureTemplate(ctx context.Context, admin *pgx.Conn, base *config.DatabaseConfig) (string, error) {
	baseVersion, err := migrationVersion(ctx, base, base.Name)
	if err != nil {
		return "", fmt.Errorf("reading %s migration version (run task test:setup): %w", base.Name, err)
	}
	template := base.Name + "_template"

	var exists bool
	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", template).Scan(&exists); err != nil {
		return "", fmt.Errorf("checking template database: %w", err)
	}
	if exists {
		templateVersion, err := migrationVersion(ctx, base, template)
		if err == nil && templateVersion == baseVersion {
			return template, nil
		}
		if _, err := admin.Exec(ctx, "DROP DATABASE "+pgx.Identifier{template}.Sanitize()+" WITH (FORCE)"); err != nil {
			return "", fmt.Errorf("dropping stale template %s: %w", template, err)
		}
	}
	// Cloning requires that no other session is connected to the base
	// database, such as an API server started for browser tests.
	if _, err := admin.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s",
		pgx.Identifier{template}.Sanitize(), pgx.Identifier{base.Name}.Sanitize())); err != nil {
		return "", fmt.Errorf("creating template %s from %s (close other sessions on %s): %w", template, base.Name, base.Name, err)
	}
	return template, nil
}

func migrationVersion(ctx context.Context, cfg *config.DatabaseConfig, name string) (int64, error) {
	conn, err := pgx.Connect(ctx, databaseURL(cfg, name))
	if err != nil {
		return 0, err
	}
	defer conn.Close(context.Background())
	var version int64
	var dirty bool
	if err := conn.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		return 0, err
	}
	if dirty {
		return 0, fmt.Errorf("%s is dirty at migration %d", name, version)
	}
	return version, nil
}

func dropDatabase(ctx context.Context, cfg *config.DatabaseConfig, name string) error {
	admin, err := connectAdmin(ctx, cfg)
	if err != nil {
		return err
	}
	defer admin.Close(context.Background())
	_, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
	return err
}

func uniqueDatabaseName(base, purpose string) (string, error) {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	label := strings.Trim(identifierUnsafe.ReplaceAllString(strings.ToLower(purpose), "_"), "_")
	random := hex.EncodeToString(suffix)
	prefix := base + "_" + label
	if limit := maxIdentifierLen - len(random) - 1; len(prefix) > limit {
		prefix = prefix[:limit]
	}
	return prefix + "_" + random, nil
}

// SetupTestDB returns a pool for this package's isolated test database. The
// package must call RunWithDatabase from TestMain.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	RequirePackageDatabase(t)

	pool, err := pgxpool.New(context.Background(), DatabaseURL())
	if err != nil {
		t.Fatalf("Failed to create test database pool: %v", err)
	}
	t.Cleanup(pool.Close)

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return pool
}

// RequirePackageDatabase fails tests that would otherwise write to the shared
// base database, which is what made parallel package runs interfere.
func RequirePackageDatabase(t *testing.T) {
	t.Helper()
	if provisionErr != nil {
		t.Fatalf("Test database unavailable: %v", provisionErr)
	}
	if packageDatabaseName == "" {
		t.Fatal("No isolated test database: call testutil.RunWithDatabase from this package's TestMain")
	}
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
