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
