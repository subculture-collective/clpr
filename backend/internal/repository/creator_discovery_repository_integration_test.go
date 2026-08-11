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

func createCreatorDiscoveryClip(
	t *testing.T,
	repo *ClipRepository,
	broadcasterID string,
	views int,
	velocity float64,
	createdAt time.Time,
	importedAt time.Time,
) {
	t.Helper()
	gameName := "Just Chatting"
	thumbnailURL := "https://example.com/" + broadcasterID + ".jpg"
	clip := models.Clip{
		ID: uuid.New(), TwitchClipID: uuid.NewString(),
		TwitchClipURL: "https://clips.twitch.tv/" + broadcasterID,
		EmbedURL:      "https://clips.twitch.tv/embed?clip=" + broadcasterID,
		Title:         "Latest from " + broadcasterID, CreatorName: broadcasterID,
		BroadcasterName: broadcasterID, BroadcasterID: &broadcasterID,
		GameName: &gameName, ThumbnailURL: &thumbnailURL, ViewCount: views,
		CreatedAt: createdAt, ImportedAt: importedAt,
	}
	if err := repo.Create(context.Background(), &clip); err != nil {
		t.Fatalf("create clip for %s: %v", broadcasterID, err)
	}
	if _, err := repo.pool.Exec(
		context.Background(),
		`UPDATE clips SET view_velocity = $1 WHERE id = $2`,
		velocity,
		clip.ID,
	); err != nil {
		t.Fatalf("set velocity for %s: %v", broadcasterID, err)
	}
}

func TestListCreatorDiscoveryBuildsMomentumAndFreshnessRails(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "broadcaster_follows")

	clipRepo := NewClipRepository(pool)
	now := time.Now()
	createCreatorDiscoveryClip(t, clipRepo, "established", 100_000, 200, now.Add(-time.Hour), now.Add(-45*24*time.Hour))
	createCreatorDiscoveryClip(t, clipRepo, "rising", 8_000, 2_000, now.Add(-30*time.Minute), now.Add(-14*24*time.Hour))
	createCreatorDiscoveryClip(t, clipRepo, "new-face", 1_000, 100, now.Add(-15*time.Minute), now.Add(-time.Hour))

	rails, err := NewBroadcasterRepository(pool).ListCreatorDiscovery(context.Background(), 2)
	if err != nil {
		t.Fatalf("list creator discovery: %v", err)
	}
	if len(rails.Trending) != 2 || len(rails.Rising) != 2 || len(rails.New) != 1 {
		t.Fatalf(
			"rail sizes = trending:%d rising:%d new:%d, want 2, 2, 1",
			len(rails.Trending), len(rails.Rising), len(rails.New),
		)
	}
	if rails.Trending[0].BroadcasterID != "established" {
		t.Fatalf("top trending creator = %q, want established", rails.Trending[0].BroadcasterID)
	}
	if rails.Rising[0].BroadcasterID != "rising" {
		t.Fatalf("top rising creator = %q, want rising", rails.Rising[0].BroadcasterID)
	}
	if rails.New[0].BroadcasterID != "new-face" {
		t.Fatalf("new creator = %q, want new-face", rails.New[0].BroadcasterID)
	}
	if rails.New[0].LatestClipThumbnail == nil || rails.New[0].TwitchCategoryName == nil {
		t.Fatalf("new creator metadata is incomplete: %+v", rails.New[0])
	}
}
