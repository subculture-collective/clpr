//go:build integration

package services

import (
	"os"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
)

// TestMain gives this package's integration tests their own database so they
// can run in parallel with other database-backed packages.
func TestMain(m *testing.M) {
	os.Exit(testutil.RunWithDatabase(m, "services"))
}
