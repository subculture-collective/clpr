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

func TestCreatingKnownGamePopulatesItsCuratedCategories(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "games")
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO categories (id, name, slug, position)
		VALUES ($1, 'Just Chatting', 'just-chatting', 1)
		ON CONFLICT (slug) DO NOTHING`, uuid.New()); err != nil {
		t.Fatalf("create category fixture: %v", err)
	}

	gameRepo := NewGameRepository(pool)
	categoryRepo := NewCategoryRepository(pool)
	now := time.Now()
	game := &models.GameEntity{
		ID: uuid.New(), TwitchGameID: "509658", Name: "Just Chatting",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := gameRepo.Create(context.Background(), game); err != nil {
		t.Fatalf("create game: %v", err)
	}
	category, err := categoryRepo.GetBySlug(context.Background(), "just-chatting")
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	games, err := categoryRepo.GetGamesInCategory(context.Background(), category.ID, nil, 10, 0)
	if err != nil {
		t.Fatalf("get category games: %v", err)
	}
	if len(games) != 1 || games[0].TwitchGameID != game.TwitchGameID {
		t.Fatalf("category games = %+v, want %s", games, game.TwitchGameID)
	}
}

func TestCreatingGameWithDuplicateNameKeepsBothCategories(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "games")
	ctx := context.Background()
	gameRepo := NewGameRepository(pool)
	now := time.Now()
	name := "The Legend of Zelda: Ocarina of Time"
	original := &models.GameEntity{ID: uuid.New(), TwitchGameID: "11557", Name: name, CreatedAt: now, UpdatedAt: now}
	duplicate := &models.GameEntity{ID: uuid.New(), TwitchGameID: "2067181260", Name: name, CreatedAt: now, UpdatedAt: now}
	for _, game := range []*models.GameEntity{original, duplicate, original, duplicate} {
		if err := gameRepo.Create(ctx, game); err != nil {
			t.Fatalf("create game %s: %v", game.TwitchGameID, err)
		}
	}
	for twitchID, wantSlug := range map[string]string{
		"11557":      "the-legend-of-zelda-ocarina-of-time",
		"2067181260": "the-legend-of-zelda-ocarina-of-time-2067181260",
	} {
		game, err := gameRepo.GetByTwitchGameID(ctx, twitchID)
		if err != nil {
			t.Fatalf("get game %s: %v", twitchID, err)
		}
		if game.Slug != wantSlug {
			t.Fatalf("game %s slug = %q, want %q", twitchID, game.Slug, wantSlug)
		}
	}
}
