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

type rankedClipSeed struct {
	age       time.Duration
	voteScore int
	viewCount int
}

// seedRankedClips creates clips whose hot and rising orders differ from their
// created_at order. Two clips share score and created_at so only the id
// tie-breaker separates them.
func seedRankedClips(t *testing.T, pool *pgxpool.Pool, repo *ClipRepository, now time.Time) []models.Clip {
	t.Helper()
	seeds := []rankedClipSeed{
		{age: 1 * time.Hour, voteScore: 1, viewCount: 50},
		{age: 2*time.Hour + 137*time.Millisecond, voteScore: 500, viewCount: 90000},
		{age: 3 * time.Hour, voteScore: -20, viewCount: 400},
		{age: 30 * time.Hour, voteScore: 9000, viewCount: 250000},
		{age: 5*time.Hour + 999*time.Millisecond, voteScore: 0, viewCount: 0},
		{age: 6 * time.Hour, voteScore: 40, viewCount: 1200},
		{age: 6 * time.Hour, voteScore: 40, viewCount: 1200},
		{age: 47 * time.Hour, voteScore: 2, viewCount: 10},
		{age: 7*time.Hour + 1*time.Millisecond, voteScore: 75, viewCount: 30000},
		{age: 90 * time.Minute, voteScore: 12, viewCount: 700},
		{age: 20 * time.Hour, voteScore: 300, viewCount: 5},
		{age: 26*time.Hour + 500*time.Millisecond, voteScore: -3, viewCount: 99999},
	}
	clips := make([]models.Clip, 0, len(seeds))
	for i, seed := range seeds {
		created := now.Add(-seed.age).Truncate(time.Microsecond)
		clip := createAuditTestClip(t, repo, fmt.Sprintf("ranked %d", i), "ranked", "ranked", created, nil, nil)
		if _, err := pool.Exec(context.Background(),
			`UPDATE clips SET vote_score = $2, view_count = $3 WHERE id = $1`,
			clip.ID, seed.voteScore, seed.viewCount); err != nil {
			t.Fatalf("set scores: %v", err)
		}
		clips = append(clips, clip)
	}
	return clips
}

func clipIDs(clips []models.Clip) []uuid.UUID {
	ids := make([]uuid.UUID, len(clips))
	for i, c := range clips {
		ids[i] = c.ID
	}
	return ids
}

// pageThrough walks a sort with a small page size, building each cursor the
// way FeedHandler does, and returns the concatenated clip IDs.
func pageThrough(t *testing.T, repo *ClipRepository, sort string, snapshot time.Time, laterSnapshot *time.Time) []uuid.UUID {
	t.Helper()
	ctx := context.Background()
	const pageSize = 3
	var got []uuid.UUID
	var cursor *string
	for page := 0; page < 20; page++ {
		filters := ClipFilters{Sort: sort, Cursor: cursor}
		if page == 0 {
			filters.SortSnapshot = &snapshot
		} else if laterSnapshot != nil {
			// The cursor's snapshot must win over any later evaluation time.
			filters.SortSnapshot = laterSnapshot
		}
		clips, _, err := repo.ListWithFilters(ctx, filters, pageSize, 0)
		if err != nil {
			t.Fatalf("%s page %d: %v", sort, page, err)
		}
		got = append(got, clipIDs(clips)...)
		if len(clips) < pageSize {
			return got
		}
		last := clips[len(clips)-1]
		next := utils.Cursor{
			SortKey:   sort,
			ClipID:    last.ID.String(),
			CreatedAt: utils.CursorTimeValue(last.CreatedAt),
			Snapshot:  utils.CursorTimeValue(snapshot),
		}
		if sort == "hot" {
			next.SortValue = float64(last.VoteScore)
		} else {
			next.SortValue = RisingCursorBase(&last)
		}
		encoded := utils.EncodeFullCursor(next)
		cursor = &encoded
	}
	t.Fatalf("%s pagination did not terminate", sort)
	return nil
}

func assertSameOrder(t *testing.T, label string, got, want []uuid.UUID) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("%s\n got  %v\n want %v", label, got, want)
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range got {
		if seen[id] {
			t.Fatalf("%s: clip %s returned twice", label, id)
		}
		seen[id] = true
	}
}

func TestHotSortCursorPaginationMatchesSingleQuery(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	ctx := context.Background()
	repo := NewClipRepository(pool)
	now := time.Now().UTC()
	seeded := seedRankedClips(t, pool, repo, now)

	all, _, err := repo.ListWithFilters(ctx, ClipFilters{Sort: "hot"}, 100, 0)
	if err != nil {
		t.Fatalf("full hot list: %v", err)
	}
	if len(all) != len(seeded) {
		t.Fatalf("full hot list returned %d clips, want %d", len(all), len(seeded))
	}

	// The rank key must order clips exactly as calculate_hot_score does.
	rows, err := pool.Query(ctx, `
		SELECT id FROM clips
		WHERE is_removed = false AND is_hidden = false
		ORDER BY calculate_hot_score(vote_score, created_at) DESC, created_at DESC, id DESC`)
	if err != nil {
		t.Fatalf("reference order: %v", err)
	}
	var reference []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan reference: %v", err)
		}
		reference = append(reference, id)
	}
	rows.Close()
	assertSameOrder(t, "hot rank key vs calculate_hot_score", clipIDs(all), reference)

	assertSameOrder(t, "paged hot order", pageThrough(t, repo, "hot", now, nil), clipIDs(all))
}

func TestRisingSortCursorPaginationUsesOneSnapshot(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	ctx := context.Background()
	repo := NewClipRepository(pool)
	now := time.Now().UTC()
	seeded := seedRankedClips(t, pool, repo, now)
	// A clip created after the snapshot (clock skew or arriving mid-session)
	// must not divide by zero.
	future := createAuditTestClip(t, repo, "ranked future", "ranked", "ranked", now.Add(3*time.Hour), nil, nil)
	_ = future

	snapshot := now.Truncate(time.Microsecond)
	all, _, err := repo.ListWithFilters(ctx, ClipFilters{Sort: "rising", SortSnapshot: &snapshot}, 100, 0)
	if err != nil {
		t.Fatalf("full rising list: %v", err)
	}
	if len(all) != len(seeded)+1 {
		t.Fatalf("full rising list returned %d clips, want %d", len(all), len(seeded)+1)
	}

	// Later pages are evaluated much later; ages shift by a week, which
	// reorders clips under a NOW()-based score, but the cursor pins the
	// snapshot so the session order is unchanged.
	later := snapshot.Add(7 * 24 * time.Hour)
	laterOrder, _, err := repo.ListWithFilters(ctx, ClipFilters{Sort: "rising", SortSnapshot: &later}, 100, 0)
	if err != nil {
		t.Fatalf("later rising list: %v", err)
	}
	if fmt.Sprint(clipIDs(laterOrder)) == fmt.Sprint(clipIDs(all)) {
		t.Fatalf("test data does not reorder over time; it cannot detect snapshot drift")
	}
	assertSameOrder(t, "paged rising order", pageThrough(t, repo, "rising", snapshot, &later), clipIDs(all))
}

func TestRankedSortsAcceptLegacyCreatedAtCursors(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	repo := NewClipRepository(pool)
	now := time.Now().UTC()
	seeded := seedRankedClips(t, pool, repo, now)

	for _, sort := range []string{"hot", "rising"} {
		legacy := utils.EncodeCursor(sort, float64(seeded[0].CreatedAt.Unix()), seeded[0].ID, seeded[0].CreatedAt.Unix())
		if _, _, err := repo.ListWithFilters(context.Background(), ClipFilters{Sort: sort, Cursor: &legacy}, 3, 0); err != nil {
			t.Fatalf("%s legacy cursor: %v", sort, err)
		}
	}
}
