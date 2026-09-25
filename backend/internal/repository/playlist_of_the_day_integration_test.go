//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

// Production returned 500 for GET /api/v1/playlists/today because the query
// ordered by gp.generated_at without grouping by it (SQLSTATE 42803).
func TestGetPlaylistOfTheDayReturnsNewestDailyPlaylist(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "generated_playlists", "playlist_items", "playlists", "playlist_scripts", "clips", "users")

	ctx := context.Background()
	repo := NewPlaylistRepository(pool)

	if _, err := repo.GetPlaylistOfTheDay(ctx, nil); !errors.Is(err, ErrPlaylistOfTheDayNotFound) {
		t.Fatalf("empty database: err = %v, want ErrPlaylistOfTheDayNotFound", err)
	}

	ownerID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username) VALUES ($1, $2, $3)`, ownerID, uuid.NewString(), "potd-owner"); err != nil {
		t.Fatalf("create owner: %v", err)
	}
	script := &models.PlaylistScript{
		ID: uuid.New(), Name: "Daily Mix", Sort: "top", ClipLimit: 3,
		Visibility: models.PlaylistVisibilityPublic, IsActive: true,
		Schedule: "daily", Strategy: "twitch_discovery", CreatedBy: &ownerID,
	}
	scriptRepo := NewPlaylistScriptRepository(pool)
	if err := scriptRepo.Create(ctx, script); err != nil {
		t.Fatalf("create script: %v", err)
	}

	clipRepo := NewClipRepository(pool)
	clip := models.Clip{
		ID: uuid.New(), TwitchClipID: uuid.NewString(), TwitchClipURL: "https://clips.twitch.tv/potd",
		EmbedURL: "https://clips.twitch.tv/embed?clip=potd", Title: "potd",
		CreatorName: "creator", BroadcasterName: "broadcaster", CreatedAt: time.Now(), ImportedAt: time.Now(),
	}
	if err := clipRepo.Create(ctx, &clip); err != nil {
		t.Fatalf("create clip: %v", err)
	}

	writer := NewPlaylistGenerationWriter(scriptRepo)
	var newest uuid.UUID
	for i, title := range []string{"Yesterday", "Today"} {
		playlist := &models.Playlist{
			ID: uuid.New(), UserID: ownerID, Title: title, Visibility: models.PlaylistVisibilityPublic,
			ScriptID: &script.ID,
		}
		if err := writer.Persist(ctx, script, playlist, []models.Clip{clip}); err != nil {
			t.Fatalf("persist playlist %q: %v", title, err)
		}
		generatedAt := time.Now().Add(time.Duration(i-1) * 24 * time.Hour)
		if _, err := pool.Exec(ctx, `UPDATE generated_playlists SET generated_at = $2 WHERE playlist_id = $1`, playlist.ID, generatedAt); err != nil {
			t.Fatalf("set generated_at: %v", err)
		}
		newest = playlist.ID
	}

	got, err := repo.GetPlaylistOfTheDay(ctx, nil)
	if err != nil {
		t.Fatalf("GetPlaylistOfTheDay: %v", err)
	}
	if got.ID != newest {
		t.Fatalf("playlist = %s (%s), want newest %s", got.ID, got.Title, newest)
	}
	if got.ClipCount != 1 {
		t.Fatalf("clip_count = %d, want 1", got.ClipCount)
	}
}
