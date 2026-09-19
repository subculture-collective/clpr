//go:build integration

package repository

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestSuppressedTagsCannotInfluencePlaylistSelection(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "tag_suppressions", "clip_tags", "tags", "clips", "users")
	ctx := context.Background()
	userID, clipID, tagID := uuid.New(), uuid.New(), uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,twitch_id,username) VALUES($1,$2,'suppression-user')`, userID, "suppression-"+userID.String()[:20]); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO clips
		(id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,view_count,created_at)
		VALUES($1,$2,'https://clips.example/test','https://clips.example/embed','suppression','creator','broadcaster',100,NOW()-INTERVAL '1 hour')`, clipID, "suppression-"+clipID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tags(id,name,slug,parent_slug) VALUES($1,'Suppressed','content/suppressed','content')`, tagID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO clip_tags(clip_id,tag_id) VALUES($1,$2)`, clipID, tagID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tag_suppressions(tag_id,suppressed_by,reason) VALUES($1,$2,'test')`, tagID, userID); err != nil {
		t.Fatal(err)
	}

	curation := NewPlaylistCurationRepository(pool)
	include := &models.PlaylistScript{Tags: []string{"content/suppressed"}, TagsLogic: "and", ClipLimit: 10}
	clips, err := curation.ViralVelocity(ctx, include)
	if err != nil || len(clips) != 0 {
		t.Fatalf("suppressed include clips=%v err=%v", clips, err)
	}
	exclude := &models.PlaylistScript{ExcludeTags: []string{"content/suppressed"}, ClipLimit: 10}
	clips, err = curation.ViralVelocity(ctx, exclude)
	if err != nil || len(clips) != 1 || clips[0].ID != clipID {
		t.Fatalf("suppressed exclusion clips=%v err=%v", clips, err)
	}

	clipRepo := NewClipRepository(pool)
	standard, _, err := clipRepo.ListWithFilters(ctx, ClipFilters{Tags: []string{"content/suppressed"}, TagsLogic: "and"}, 10, 0)
	if err != nil || len(standard) != 0 {
		t.Fatalf("standard suppressed include clips=%v err=%v", standard, err)
	}
	standard, _, err = clipRepo.ListWithFilters(ctx, ClipFilters{ExcludeTags: []string{"content/suppressed"}}, 10, 0)
	if err != nil || len(standard) != 1 || standard[0].ID != clipID {
		t.Fatalf("standard suppressed exclusion clips=%v err=%v", standard, err)
	}
}
