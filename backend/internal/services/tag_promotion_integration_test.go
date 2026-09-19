//go:build integration

package services

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestPromotionDetectionUpdatesCountsAndHonorsModeration(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "tag_promotion_queue", "tag_suppressions", "clip_tags", "blacklisted_tags", "tags", "clips", "users")
	ctx := context.Background()

	users := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for i, id := range users {
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,twitch_id,username) VALUES($1,$2,$3)`, id, "promotion-"+id.String()[:20], "promotion-user-"+string(rune('a'+i))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tags(id,name,slug,parent_slug) VALUES
		(gen_random_uuid(),'Community candidate','candidate','community'),
		(gen_random_uuid(),'Suppressed candidate','suppressed-candidate','community'),
		(gen_random_uuid(),'Blocked candidate','blocked-candidate','community')`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO blacklisted_tags(pattern) VALUES('blocked*')`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tag_suppressions(tag_id,suppressed_by,reason)
		SELECT id,$1,'test' FROM tags WHERE slug='suppressed-candidate'`, users[0]); err != nil {
		t.Fatal(err)
	}

	insertClip := func(index int, submitter uuid.UUID) uuid.UUID {
		id := uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO clips
			(id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,created_at,submitted_by_user_id,submitted_at)
			VALUES($1,$2,'https://clips.example/test','https://clips.example/embed','promotion','creator','broadcaster',NOW(),$3,NOW())`,
			id, "promotion-clip-"+id.String(), submitter); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO clip_tags(clip_id,tag_id)
			SELECT $1,id FROM tags WHERE slug IN ('candidate','suppressed-candidate','blocked-candidate')`, id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	for i := 0; i < 5; i++ {
		insertClip(i, users[i%len(users)])
	}

	service := NewTagPromotionService(pool)
	queued, err := service.CheckPromotionCandidates(ctx)
	if err != nil || len(queued) != 1 || queued[0] != "candidate" {
		t.Fatalf("queued=%v err=%v", queued, err)
	}
	insertClip(5, users[0])
	queued, err = service.CheckPromotionCandidates(ctx)
	if err != nil || len(queued) != 0 {
		t.Fatalf("second detection queued=%v err=%v", queued, err)
	}
	var usage int
	if err := pool.QueryRow(ctx, `SELECT usage_count FROM tag_promotion_queue WHERE tag_slug='candidate' AND status='pending'`).Scan(&usage); err != nil || usage != 6 {
		t.Fatalf("pending usage=%d err=%v", usage, err)
	}
	if err := service.RejectPromotion(ctx, "candidate", users[0]); err != nil {
		t.Fatal(err)
	}
	queued, err = service.CheckPromotionCandidates(ctx)
	if err != nil || len(queued) != 0 {
		t.Fatalf("rejected candidate recreated: queued=%v err=%v", queued, err)
	}
	var pending int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM tag_promotion_queue WHERE tag_slug='candidate' AND status='pending'`).Scan(&pending); err != nil || pending != 0 {
		t.Fatalf("pending after rejection=%d err=%v", pending, err)
	}
}
