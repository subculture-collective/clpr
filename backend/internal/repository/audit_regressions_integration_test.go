//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertAuditTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, username, display_name, email) VALUES ($1, $2, $2, $3)
	`, id, "audit-"+id.String()[:12], id.String()+"@example.test")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}

func createAuditTestClip(t *testing.T, repo *ClipRepository, title, game, creator string, createdAt time.Time, submittedAt *time.Time, submitter *uuid.UUID) models.Clip {
	t.Helper()
	gameID := "game-" + game
	gameName := game
	creatorID := "creator-" + creator
	id := uuid.New()
	clip := models.Clip{
		ID: id, TwitchClipID: id.String(), TwitchClipURL: "https://clips.twitch.tv/" + id.String(),
		EmbedURL: "https://clips.twitch.tv/embed?clip=" + id.String(), Title: title,
		CreatorName: creator, CreatorID: &creatorID, BroadcasterName: creator, BroadcasterID: &creatorID,
		GameID: &gameID, GameName: &gameName, CreatedAt: createdAt, ImportedAt: time.Now(),
		SubmittedAt: submittedAt, SubmittedByUserID: submitter,
	}
	if err := repo.Create(context.Background(), &clip); err != nil {
		t.Fatalf("create clip %q: %v", title, err)
	}
	return clip
}

// Bug 2: the "new" sort orders by COALESCE(submitted_at, created_at); the
// cursor must page over the same key without skipping or repeating clips.
func TestListWithFiltersNewSortCursorUsesSubmittedAt(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	ctx := context.Background()
	repo := NewClipRepository(pool)
	submitter := insertAuditTestUser(t, pool)

	base := time.Now().UTC().Add(-48 * time.Hour).Truncate(time.Microsecond)
	var want []uuid.UUID
	// Created order is the reverse of submitted order, with sub-second
	// offsets, so a created_at cursor would skip and repeat clips.
	for i := 0; i < 5; i++ {
		created := base.Add(time.Duration(i) * time.Hour)
		submitted := base.Add(30*time.Hour - time.Duration(i)*time.Hour + time.Duration(i)*137*time.Millisecond)
		clip := createAuditTestClip(t, repo, fmt.Sprintf("new %d", i), "cursor", "cursor", created, &submitted, &submitter)
		want = append(want, clip.ID)
	}
	// One clip without submitted_at, falling back to created_at between the others.
	mid := base.Add(28*time.Hour + 500*time.Millisecond)
	fallback := createAuditTestClip(t, repo, "new fallback", "cursor", "cursor", mid, nil, nil)
	want = append(want[:2], append([]uuid.UUID{fallback.ID}, want[2:]...)...)

	var got []uuid.UUID
	var cursor *string
	for page := 0; page < 10; page++ {
		clips, _, err := repo.ListWithFilters(ctx, ClipFilters{Sort: "new", Cursor: cursor}, 2, 0)
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		for _, c := range clips {
			got = append(got, c.ID)
		}
		if len(clips) < 2 {
			break
		}
		last := clips[len(clips)-1]
		encoded := utils.EncodeCursor("new", NewSortCursorValue(&last), last.ID, last.CreatedAt.Unix())
		cursor = &encoded
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("paged order\n got  %v\n want %v", got, want)
	}
}

// Bug 3: suggestions previously used SELECT DISTINCT ... ORDER BY COUNT(*) OVER,
// which Postgres rejects, so game and creator suggestions were always empty.
// Bug 4: tsquery punctuation must not turn into a database error.
func TestSearchSuggestionsAndSpecialCharacterQueries(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	ctx := context.Background()
	clipRepo := NewClipRepository(pool)
	submitter := insertAuditTestUser(t, pool)
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		createAuditTestClip(t, clipRepo, "suggest popular", "Suggestable Popular", "suggestcreator", now, &now, &submitter)
	}
	createAuditTestClip(t, clipRepo, "suggest rare", "Suggestable Rare", "suggestcreator", now, &now, &submitter)

	repo := NewSearchRepository(pool)
	suggestions, err := repo.GetSuggestions(ctx, "suggest", 10)
	if err != nil {
		t.Fatalf("GetSuggestions: %v", err)
	}
	var games, creators []string
	for _, s := range suggestions {
		switch s.Type {
		case "game":
			games = append(games, s.Text)
		case "creator":
			creators = append(creators, s.Text)
		}
	}
	if fmt.Sprint(games) != "[Suggestable Popular Suggestable Rare]" {
		t.Fatalf("game suggestions = %v; want popular first", games)
	}
	if fmt.Sprint(creators) != "[suggestcreator]" {
		t.Fatalf("creator suggestions = %v", creators)
	}

	for _, q := range []string{"foo(", "a|", "!", "it's:", "<->", "x & | y", "суп!"} {
		if _, err := repo.Search(ctx, &models.SearchRequest{Query: q, Page: 1, Limit: 10}); err != nil {
			t.Errorf("Search(%q) error: %v", q, err)
		}
	}
}

// Bugs 7 and 8: controversial ranks evenly split comments first; best uses
// float division so the Wilson interval is not truncated.
func TestCommentControversialAndBestOrdering(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	ctx := context.Background()
	author := insertAuditTestUser(t, pool)
	now := time.Now().UTC()
	clip := createAuditTestClip(t, NewClipRepository(pool), "comments", "comments", "comments", now, nil, nil)

	addComment := func(content string, up, down int) uuid.UUID {
		id := uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO comments (id, clip_id, user_id, content) VALUES ($1, $2, $3, $4)`, id, clip.ID, author, content); err != nil {
			t.Fatalf("create comment: %v", err)
		}
		for i := 0; i < up+down; i++ {
			voter := insertAuditTestUser(t, pool)
			vote := 1
			if i >= up {
				vote = -1
			}
			if _, err := pool.Exec(ctx, `INSERT INTO comment_votes (user_id, comment_id, vote_type) VALUES ($1, $2, $3)`, voter, id, vote); err != nil {
				t.Fatalf("vote: %v", err)
			}
		}
		return id
	}
	oneSided := addComment("one-sided", 6, 0)
	split := addComment("split", 3, 3)
	mostlyUp := addComment("mostly up", 5, 1)

	repo := NewCommentRepository(pool)
	ids := func(sort string) []uuid.UUID {
		comments, err := repo.ListByClipID(ctx, clip.ID, sort, 10, 0, nil)
		if err != nil {
			t.Fatalf("ListByClipID(%s): %v", sort, err)
		}
		out := make([]uuid.UUID, len(comments))
		for i, c := range comments {
			out[i] = c.ID
		}
		return out
	}
	if got, want := ids("controversial"), []uuid.UUID{split, mostlyUp, oneSided}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("controversial = %v; want %v", got, want)
	}
	// Wilson lower bounds: 6/0 = 0.61, 5/1 = 0.44, 3/3 = 0.19.
	if got, want := ids("best"), []uuid.UUID{oneSided, mostlyUp, split}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("best = %v; want %v", got, want)
	}
}

// Bug 6: every OFFSET-paginated sort needs a unique tie-breaker. Equal
// timestamps and counts must still page through every clip exactly once.
func TestOffsetPaginatedListsHaveUniqueTieBreakers(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "discovery_clips")
	ctx := context.Background()
	clipRepo := NewClipRepository(pool)
	discoveryRepo := NewDiscoveryClipRepository(pool)
	created := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)

	const n = 7
	for i := 0; i < n; i++ {
		createAuditTestClip(t, clipRepo, fmt.Sprintf("tie %d", i), "tie", "tie", created, nil, nil)
		id := uuid.New()
		if err := discoveryRepo.Create(ctx, &models.DiscoveryClip{
			ID: id, TwitchClipID: id.String(), TwitchClipURL: "https://clips.twitch.tv/" + id.String(),
			EmbedURL: "https://clips.twitch.tv/embed?clip=" + id.String(), Title: "tie",
			CreatorName: "tie", BroadcasterName: "tie", CreatedAt: created, ImportedAt: time.Now(),
		}); err != nil {
			t.Fatalf("create discovery clip: %v", err)
		}
	}

	for _, sort := range []string{"hot", "new", "top", "views", "trending", "rising", "discussed", ""} {
		seen := map[uuid.UUID]bool{}
		for offset := 0; offset < n; offset += 2 {
			clips, _, err := clipRepo.ListScrapedClipsWithFilters(ctx, ClipFilters{Sort: sort}, 2, offset)
			if err != nil {
				t.Fatalf("scraped sort %q: %v", sort, err)
			}
			for _, c := range clips {
				seen[c.ID] = true
			}
		}
		if len(seen) != n {
			t.Errorf("scraped sort %q paged %d distinct clips; want %d", sort, len(seen), n)
		}
	}
	for _, sort := range []string{"views", "trending", "new", ""} {
		seen := map[uuid.UUID]bool{}
		for offset := 0; offset < n; offset += 2 {
			clips, _, err := discoveryRepo.ListWithFilters(ctx, DiscoveryClipFilters{Sort: sort}, 2, offset)
			if err != nil {
				t.Fatalf("discovery sort %q: %v", sort, err)
			}
			for _, c := range clips {
				seen[c.ID] = true
			}
		}
		if len(seen) != n {
			t.Errorf("discovery sort %q paged %d distinct clips; want %d", sort, len(seen), n)
		}
	}
}
