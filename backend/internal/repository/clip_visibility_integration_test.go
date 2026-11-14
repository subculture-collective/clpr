//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// insertVisibilityClip inserts an automatically imported clip (no submitter)
// with the given hidden and DMCA flags and removes it when the test ends.
func insertVisibilityClip(t *testing.T, pool *pgxpool.Pool, creator, broadcasterID string, hidden, dmcaRemoved bool) uuid.UUID {
	t.Helper()
	id, twitchID := uuid.New(), uuid.NewString()
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO clips (
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, broadcaster_name, broadcaster_id, created_at, is_hidden, dmca_removed
		) VALUES ($1, $2, $3, $4, 'Visibility clip', $5, $5, $6, NOW(), $7, $8)
	`, id, twitchID, "https://clips.twitch.tv/"+twitchID, "https://clips.twitch.tv/embed?clip="+twitchID,
		creator, broadcasterID, hidden, dmcaRemoved); err != nil {
		t.Fatalf("insert clip: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM clips WHERE id = $1`, id) })
	return id
}

func TestClipRepository_UpdateVisibilityKeepsDMCARemovedClipsHidden(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewClipRepository(pool)
	creator := "visibility-" + uuid.NewString()[:8]
	isHidden := func(id uuid.UUID) bool {
		t.Helper()
		var hidden bool
		if err := pool.QueryRow(ctx, `SELECT is_hidden FROM clips WHERE id = $1`, id).Scan(&hidden); err != nil {
			t.Fatalf("read is_hidden: %v", err)
		}
		return hidden
	}

	dmca := insertVisibilityClip(t, pool, creator, creator, true, true)
	if err := repo.UpdateVisibility(ctx, dmca, false); !errors.Is(err, ErrClipDMCARemoved) {
		t.Fatalf("un-hide of a DMCA-removed clip = %v, want ErrClipDMCARemoved", err)
	}
	if !isHidden(dmca) {
		t.Fatal("DMCA-removed clip was un-hidden")
	}

	shownDMCA := insertVisibilityClip(t, pool, creator, creator, false, true)
	if err := repo.UpdateVisibility(ctx, shownDMCA, true); err != nil {
		t.Fatalf("hide a DMCA-removed clip: %v", err)
	}
	if !isHidden(shownDMCA) {
		t.Fatal("DMCA-removed clip was not hidden")
	}

	normal := insertVisibilityClip(t, pool, creator, creator, true, false)
	if err := repo.UpdateVisibility(ctx, normal, false); err != nil {
		t.Fatalf("un-hide a normal clip: %v", err)
	}
	if isHidden(normal) {
		t.Fatal("normal clip stayed hidden")
	}
}

func TestClipRepository_ListClipsByBroadcasterListsImportedClipsWithoutDMCARemoved(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	broadcaster := "bc-" + uuid.NewString()[:8]

	// None of these clips has a submitter.
	visible := insertVisibilityClip(t, pool, broadcaster, broadcaster, false, false)
	insertVisibilityClip(t, pool, broadcaster, broadcaster, true, false)
	insertVisibilityClip(t, pool, broadcaster, broadcaster, false, true)

	for _, sort := range []string{"recent", "popular", "trending"} {
		clips, total, err := NewClipRepository(pool).ListClipsByBroadcaster(ctx, broadcaster, sort, 10, 0)
		if err != nil {
			t.Fatalf("ListClipsByBroadcaster(%s): %v", sort, err)
		}
		if total != 1 || len(clips) != 1 || clips[0].ID != visible || clips[0].SubmittedByUserID != nil {
			t.Fatalf("sort %s: total=%d clips=%d, want only the visible imported clip %s", sort, total, len(clips), visible)
		}
	}
}
