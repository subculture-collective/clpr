//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

// Comment threads showed the author's karma even when the author had turned
// show_karma_publicly off.
func TestCommentAuthorKarmaFollowsShowKarmaSetting(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	ctx := context.Background()
	author := insertAuditTestUser(t, pool)
	if _, err := pool.Exec(ctx, `UPDATE users SET karma_points = 4321 WHERE id = $1`, author); err != nil {
		t.Fatalf("set karma: %v", err)
	}
	clip := createAuditTestClip(t, NewClipRepository(pool), "karma", "karma", "karma", time.Now().UTC(), nil, nil)
	commentID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO comments (id, clip_id, user_id, content) VALUES ($1, $2, $3, 'hello')`, commentID, clip.ID, author); err != nil {
		t.Fatalf("create comment: %v", err)
	}
	setShowKarma := func(show bool) {
		t.Helper()
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_settings (user_id, show_karma_publicly) VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET show_karma_publicly = $2`, author, show); err != nil {
			t.Fatalf("set show_karma_publicly: %v", err)
		}
	}

	repo := NewCommentRepository(pool)
	karmas := func() []*int {
		t.Helper()
		thread, err := repo.ListByClipID(ctx, clip.ID, "new", 10, 0, nil)
		if err != nil || len(thread) != 1 {
			t.Fatalf("ListByClipID: %d comments, %v", len(thread), err)
		}
		one, err := repo.GetByID(ctx, commentID, nil)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		byUser, _, err := repo.ListByUserID(ctx, author, 10, 0)
		if err != nil || len(byUser) != 1 {
			t.Fatalf("ListByUserID: %d comments, %v", len(byUser), err)
		}
		return []*int{thread[0].AuthorKarma, one.AuthorKarma, byUser[0].AuthorKarma}
	}

	setShowKarma(false)
	for i, k := range karmas() {
		if k != nil {
			t.Errorf("query %d: hidden karma returned as %d", i, *k)
		}
	}
	setShowKarma(true)
	for i, k := range karmas() {
		if k == nil || *k != 4321 {
			t.Errorf("query %d: karma %v, want 4321", i, k)
		}
	}
}
