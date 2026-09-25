//go:build integration

package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestTagTreeLimitsChildrenAndHonoursVisibility(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	t.Cleanup(func() { testutil.CleanupTestDB(t, pool) }) // runs after the row cleanups below
	ctx := context.Background()
	repo := NewTagRepository(pool)

	root := "tree-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM blacklisted_tags WHERE pattern ILIKE $1`, root+"%")
		_, _ = pool.Exec(context.Background(), `DELETE FROM tags WHERE slug = $1 OR parent_slug = $1 OR parent_slug LIKE $2`, root, root+"/%")
	})

	exec(`INSERT INTO tags (name, slug, parent_slug, usage_count) VALUES ($1, $2, NULL, 0)`, "Root "+root, root)
	for i := 1; i <= 6; i++ {
		exec(`INSERT INTO tags (name, slug, parent_slug, usage_count) VALUES ($1, $2, $3, $4)`,
			fmt.Sprintf("Child %s %d", root, i), fmt.Sprintf("%s/c%d", root, i), root, i*10)
	}
	exec(`INSERT INTO tags (name, slug, parent_slug, usage_count) VALUES ($1, $2, $3, 1)`, "Grandchild "+root, root+"/c6/g1", root+"/c6")

	// Hide the two highest-usage children: one by an exact pattern, one by a glob.
	exec(`INSERT INTO blacklisted_tags (pattern) VALUES ($1), ($2)`, strings.ToUpper(root)+"/c6", root+"/c5*")
	var suppressedID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM tags WHERE slug = $1`, root+"/c4").Scan(&suppressedID); err != nil {
		t.Fatalf("find tag: %v", err)
	}
	actorID := uuid.New()
	exec(`INSERT INTO users (id, twitch_id, username) VALUES ($1, $2, $3)`, actorID, uuid.NewString(), root)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM tag_suppressions WHERE suppressed_by = $1`, actorID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, actorID)
	})
	exec(`INSERT INTO tag_suppressions (tag_id, suppressed_by, reason) VALUES ($1, $2, 'test')`, suppressedID, actorID)

	for _, slug := range []string{root + "/c6", root + "/c5", root + "/c5x", root + "/c3"} {
		var inline, function bool
		if err := pool.QueryRow(ctx, `SELECT is_tag_blacklisted($1), NOT (`+visibleTagPredicate("t")+`) FROM (SELECT gen_random_uuid() AS id, $1::text AS slug) t`, slug).Scan(&function, &inline); err != nil {
			t.Fatalf("compare predicates: %v", err)
		}
		if inline != function {
			t.Fatalf("%s: inline predicate hidden=%v, is_tag_blacklisted=%v", slug, inline, function)
		}
	}

	rows, err := repo.GetTagTree(ctx, root, 2)
	if err != nil {
		t.Fatalf("GetTagTree: %v", err)
	}
	var got []string
	for _, row := range rows {
		got = append(got, fmt.Sprintf("%s:%d:%d", strings.TrimPrefix(row.Slug, root), row.Depth, row.ChildCount))
	}
	// c6 (exact pattern, different case), c5 (glob) and c4 (suppressed) are
	// hidden, so c6's grandchild is unreachable. Two of three visible
	// children are returned.
	if want := ":0:3 /c3:1:0 /c2:1:0"; strings.Join(got, " ") != want {
		t.Fatalf("subtree = %q, want %q", strings.Join(got, " "), want)
	}

	forest, err := repo.GetTagForest(ctx, 1)
	if err != nil {
		t.Fatalf("GetTagForest: %v", err)
	}
	var sawRoot, sawChild bool
	for _, row := range forest {
		if row.Slug == root {
			sawRoot = row.Depth == 0 && row.ChildCount == 3
		}
		if row.ParentSlug != nil && *row.ParentSlug == root {
			if sawChild || row.Slug != root+"/c3" {
				t.Fatalf("forest child under %s = %s; want only %s/c3", root, row.Slug, root)
			}
			sawChild = true
		}
		if row.Depth == 0 && row.ChildCount == 0 {
			t.Fatalf("forest root %s has no children", row.Slug)
		}
	}
	if !sawRoot || !sawChild {
		t.Fatalf("forest missing root (%v) or child (%v)", sawRoot, sawChild)
	}
}
