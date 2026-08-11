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

func createCurationClip(t *testing.T, repo *ClipRepository, title, gameID, creatorID string, views int, velocity float64, createdAt time.Time) models.Clip {
	t.Helper()
	clip := models.Clip{
		ID: uuid.New(), TwitchClipID: uuid.NewString(), TwitchClipURL: "https://clips.twitch.tv/" + title,
		EmbedURL: "https://clips.twitch.tv/embed?clip=" + title, Title: title,
		CreatorName: creatorID, CreatorID: &creatorID, BroadcasterName: creatorID,
		BroadcasterID: &creatorID, GameID: &gameID, ViewCount: views,
		CreatedAt: createdAt, ImportedAt: time.Now(),
	}
	if err := repo.Create(context.Background(), &clip); err != nil {
		t.Fatalf("create clip %s: %v", title, err)
	}
	if _, err := repo.pool.Exec(context.Background(), `UPDATE clips SET view_velocity = $1 WHERE id = $2`, velocity, clip.ID); err != nil {
		t.Fatalf("set clip velocity: %v", err)
	}
	return clip
}

func TestDiversityRouletteReturnsAtMostOneClipPerGame(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")
	clipRepo := NewClipRepository(pool)
	now := time.Now()
	clips := []models.Clip{
		createCurationClip(t, clipRepo, "a1", "game-a", "creator-a", 1000, 20, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "a2", "game-a", "creator-b", 900, 18, now.Add(-2*time.Hour)),
		createCurationClip(t, clipRepo, "b1", "game-b", "creator-c", 800, 16, now.Add(-time.Hour)),
	}
	gameByClip := map[uuid.UUID]string{}
	for _, clip := range clips {
		gameByClip[clip.ID] = *clip.GameID
	}

	week := "week"
	result, err := NewPlaylistCurationRepository(pool).DiversityRoulette(context.Background(), &models.PlaylistScript{Timeframe: &week, ClipLimit: 10})
	if err != nil {
		t.Fatalf("diversity roulette: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("result count = %d, want one clip for each of 2 games", len(result))
	}
	seen := map[string]bool{}
	for _, clip := range result {
		game := gameByClip[clip.ID]
		if seen[game] {
			t.Fatalf("game %q appeared more than once", game)
		}
		seen[game] = true
	}
}

func TestClipOfTheDayPrioritizesCurrentVelocity(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")
	clipRepo := NewClipRepository(pool)
	now := time.Now()
	createCurationClip(t, clipRepo, "slow", "game-a", "creator-a", 100000, 0.1, now.Add(-20*time.Hour))
	fast := createCurationClip(t, clipRepo, "fast", "game-b", "creator-b", 5000, 500, now.Add(-time.Hour))

	day := "day"
	result, err := NewPlaylistCurationRepository(pool).ClipOfTheDay(context.Background(), &models.PlaylistScript{Timeframe: &day, ClipLimit: 1})
	if err != nil {
		t.Fatalf("clip of the day: %v", err)
	}
	if len(result) != 1 || result[0].ID != fast.ID {
		t.Fatalf("clip of the day = %+v, want fast clip %s", result, fast.ID)
	}
}

func TestWeekendMixEnforcesCreatorAndGameDiversity(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")
	clipRepo := NewClipRepository(pool)
	now := time.Now()
	clips := []models.Clip{
		createCurationClip(t, clipRepo, "a1", "game-a", "creator-a", 1000, 30, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "a2", "game-a", "creator-a", 900, 20, now.Add(-2*time.Hour)),
		createCurationClip(t, clipRepo, "b1", "game-a", "creator-b", 800, 18, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "c1", "game-a", "creator-c", 700, 16, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "d1", "game-b", "creator-d", 600, 14, now.Add(-time.Hour)),
	}
	creatorByClip, gameByClip := map[uuid.UUID]string{}, map[uuid.UUID]string{}
	for _, clip := range clips {
		creatorByClip[clip.ID] = *clip.CreatorID
		gameByClip[clip.ID] = *clip.GameID
	}

	week := "week"
	result, err := NewPlaylistCurationRepository(pool).WeekendMix(context.Background(), &models.PlaylistScript{Timeframe: &week, ClipLimit: 20})
	if err != nil {
		t.Fatalf("weekend mix: %v", err)
	}
	seenCreators, gameCounts := map[string]bool{}, map[string]int{}
	for _, clip := range result {
		creator := creatorByClip[clip.ID]
		if seenCreators[creator] {
			t.Fatalf("creator %q appeared more than once", creator)
		}
		seenCreators[creator] = true
		gameCounts[gameByClip[clip.ID]]++
	}
	if gameCounts["game-a"] > 2 {
		t.Fatalf("game-a appeared %d times, want at most 2", gameCounts["game-a"])
	}
}
