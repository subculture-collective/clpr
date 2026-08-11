//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestPlaylistGenerationDeduplicatesStrategyResults(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "generated_playlists", "playlist_items", "playlists", "playlist_scripts", "clips", "users")

	ctx := context.Background()
	ownerID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username) VALUES ($1, $2, $3)`, ownerID, uuid.NewString(), "playlist-owner"); err != nil {
		t.Fatalf("create owner: %v", err)
	}
	script := &models.PlaylistScript{
		ID: uuid.New(), Name: "Discovery Mix", Sort: "top", ClipLimit: 3,
		Visibility: models.PlaylistVisibilityPublic, IsActive: true,
		Schedule: "daily", Strategy: "twitch_discovery", CreatedBy: &ownerID,
	}
	scriptRepo := NewPlaylistScriptRepository(pool)
	if err := scriptRepo.Create(ctx, script); err != nil {
		t.Fatalf("create script: %v", err)
	}

	clipRepo := NewClipRepository(pool)
	newClip := func(title string) models.Clip {
		return models.Clip{
			ID: uuid.New(), TwitchClipID: uuid.NewString(), TwitchClipURL: "https://clips.twitch.tv/" + title,
			EmbedURL: "https://clips.twitch.tv/embed?clip=" + title, Title: title,
			CreatorName: "creator", BroadcasterName: "broadcaster", CreatedAt: time.Now(), ImportedAt: time.Now(),
		}
	}
	first, second := newClip("first"), newClip("second")
	for _, clip := range []*models.Clip{&first, &second} {
		if err := clipRepo.Create(ctx, clip); err != nil {
			t.Fatalf("create clip: %v", err)
		}
	}

	playlist := &models.Playlist{
		ID: uuid.New(), UserID: ownerID, Title: "Discovery Mix", Visibility: models.PlaylistVisibilityPublic,
		ScriptID: &script.ID,
	}
	writer := NewPlaylistGenerationWriter(scriptRepo)
	if err := writer.Persist(ctx, script, playlist, []models.Clip{first, first, second}); err != nil {
		t.Fatalf("persist playlist with repeated strategy result: %v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM playlist_items WHERE playlist_id = $1`, playlist.ID).Scan(&count); err != nil {
		t.Fatalf("count playlist items: %v", err)
	}
	if count != 2 {
		t.Fatalf("playlist item count = %d, want 2 unique clips", count)
	}
}
