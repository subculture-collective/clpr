//go:build integration

package repository

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

// An export interrupted while processing (crash, deploy) becomes eligible
// again once its processing claim is stale; a running export is not re-queued.
func TestExportRepository_PendingIncludesStaleProcessing(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	user := uuid.New()
	stale, running := uuid.New(), uuid.New()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user) })
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,twitch_id,username) VALUES($1,$2,$2)`, user, "export-"+user.String()[:20]); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO export_requests(id,user_id,creator_name,format,status,created_at,updated_at) VALUES
		($1,$3,'creator','csv','processing',NOW()-interval '2 hours',NOW()-interval '2 hours'),
		($2,$3,'creator','csv','processing',NOW()-interval '2 hours',NOW()-interval '1 minute')`, stale, running, user); err != nil {
		t.Fatal(err)
	}
	requests, err := NewExportRepository(pool).GetPendingExportRequests(ctx, 10000)
	if err != nil {
		t.Fatal(err)
	}
	found := map[uuid.UUID]bool{}
	for _, r := range requests {
		found[r.ID] = true
	}
	if !found[stale] || found[running] {
		t.Fatalf("stale processing returned=%v, running processing returned=%v", found[stale], found[running])
	}
}
