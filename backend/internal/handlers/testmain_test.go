package handlers

import (
	"os"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
)

// TestMain gives this package its own database so it can run in parallel with
// other database-backed packages. Without a reachable test database, tests
// still run and database-dependent tests skip or fail on their own terms.
func TestMain(m *testing.M) {
	os.Exit(testutil.RunWithDatabase(m, "handlers"))
}
