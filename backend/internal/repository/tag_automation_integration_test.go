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

func TestTagBlacklistUsesSafeCaseInsensitiveGlobs(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "blacklisted_tags")
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO blacklisted_tags(pattern) VALUES ('spam*'), ('bot?'), ('100%')`); err != nil {
		t.Fatal(err)
	}
	repo := NewTagRepository(pool)
	for _, slug := range []string{"SPAM", "spam-account", "bot1", "100%"} {
		matched, err := repo.IsBlacklisted(ctx, slug)
		if err != nil {
			t.Fatal(err)
		}
		if !matched {
			t.Errorf("expected %q to be blacklisted", slug)
		}
	}
	for _, slug := range []string{"spa", "bot", "1000"} {
		matched, err := repo.IsBlacklisted(ctx, slug)
		if err != nil {
			t.Fatal(err)
		}
		if matched {
			t.Errorf("did not expect %q to be blacklisted", slug)
		}
	}
}

func TestLegacyTagAliasResolvesCanonicalTag(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "tag_aliases", "tags")
	ctx := context.Background()
	repo := NewTagRepository(pool)
	parent := "content"
	canonical := &models.Tag{ID: uuid.New(), Name: "Content: Funny", Slug: "content/funny", ParentSlug: &parent, CreatedAt: time.Now()}
	root := &models.Tag{ID: uuid.New(), Name: "Content", Slug: parent, CreatedAt: time.Now()}
	if err := repo.Create(ctx, root); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, canonical); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tag_aliases(alias_slug,canonical_tag_id) VALUES ('funny',$1)`, canonical.ID); err != nil {
		t.Fatal(err)
	}
	resolved, err := repo.GetBySlug(ctx, "funny")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ID != canonical.ID {
		t.Fatalf("resolved %s, want %s", resolved.ID, canonical.ID)
	}
}

func TestTagListFiltersByLane(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "tag_suppressions", "tag_aliases", "clip_tags", "tags")
	ctx := context.Background()
	repo := NewTagRepository(pool)
	seed := []struct{ slug, parent string }{
		{"content", ""}, {"game", ""}, {"lang", ""}, {"duration", ""}, {"streamer", ""},
		{"content/boss-fight", "content"}, {"content/funny", "content"},
		{"content/communityoriented", "content"}, {"streamer/english", "streamer"},
		{"game/just-chatting", "game"}, {"lang/en", "lang"}, {"duration/short", "duration"},
		{"goosebumps", ""},
	}
	for i, s := range seed {
		tag := &models.Tag{ID: uuid.New(), Name: s.slug, Slug: s.slug, UsageCount: len(seed) - i, CreatedAt: time.Now()}
		if s.parent != "" {
			parent := s.parent
			tag.ParentSlug = &parent
		}
		if err := repo.Create(ctx, tag); err != nil {
			t.Fatal(err)
		}
	}
	want := map[string][]string{
		"detected":  {"content/boss-fight", "content/funny"},
		"streamer":  {"content/communityoriented", "streamer/english"},
		"category":  {"game/just-chatting"},
		"language":  {"lang/en"},
		"duration":  {"duration/short"},
		"community": {"goosebumps"},
		"":          nil,
	}
	for lane, slugs := range want {
		tags, err := repo.List(ctx, "alphabetical", lane, 100, 0)
		if err != nil {
			t.Fatalf("lane %q: %v", lane, err)
		}
		count, err := repo.Count(ctx, lane)
		if err != nil {
			t.Fatalf("count lane %q: %v", lane, err)
		}
		if lane == "" {
			if len(tags) != len(seed) || count != len(seed) {
				t.Errorf("unfiltered list returned %d tags, count %d; want %d", len(tags), count, len(seed))
			}
			continue
		}
		got := make([]string, 0, len(tags))
		for _, tag := range tags {
			got = append(got, tag.Slug)
		}
		if len(got) != len(slugs) || count != len(slugs) {
			t.Errorf("lane %q returned %v (count %d), want %v", lane, got, count, slugs)
			continue
		}
		for i := range slugs {
			if got[i] != slugs[i] {
				t.Errorf("lane %q returned %v, want %v", lane, got, slugs)
				break
			}
		}
	}
	curated, err := repo.List(ctx, "curated", "", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(curated) != 2 {
		t.Errorf("curated sort returned %d tags, want the 2 detected tags", len(curated))
	}
	trending, err := repo.List(ctx, "trending", "detected", 10, 0)
	if err != nil {
		t.Fatalf("trending with lane: %v", err)
	}
	if len(trending) != 0 {
		t.Errorf("trending without clips returned %d tags", len(trending))
	}
}
