//go:build integration

package repository

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

// Like vision, transcription stops retrying a clip after three failed attempts.
func TestClipRepository_TranscriptionQueueStopsAfterThreeAttempts(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	broadcasterID := "tq-" + userID.String()[:20]
	retrying, exhausted := uuid.New(), uuid.New()
	t.Cleanup(func() {
		bg := context.Background()
		_, _ = pool.Exec(bg, `DELETE FROM clips WHERE id = ANY($1)`, []uuid.UUID{retrying, exhausted})
		_, _ = pool.Exec(bg, `DELETE FROM users WHERE id = $1`, userID)
	})
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,twitch_id,username) VALUES($1,$2,$2)`, userID, broadcasterID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO twitch_auth(user_id,twitch_user_id,twitch_username,access_token,refresh_token,expires_at,scopes)
		VALUES($1,$2,$2,'test','test',NOW()+interval '1 day','channel:manage:clips')`, userID, broadcasterID); err != nil {
		t.Fatal(err)
	}
	for id, attempts := range map[uuid.UUID]int{retrying: 2, exhausted: 3} {
		if _, err := pool.Exec(ctx, `INSERT INTO clips
			(id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,broadcaster_id,created_at,
			 transcription_attempt_count,transcription_attempted_at)
			VALUES($1,$2,'https://clips.twitch.tv/x','https://clips.twitch.tv/embed?clip=x','t','c','b',$3,NOW(),$4,NOW()-interval '1 hour')`,
			id, "tq-"+id.String(), broadcasterID, attempts); err != nil {
			t.Fatal(err)
		}
	}

	clips, err := NewClipRepository(pool).GetClipsNeedingTranscription(ctx, 10000)
	if err != nil {
		t.Fatal(err)
	}
	found := map[uuid.UUID]bool{}
	for _, c := range clips {
		found[c.ID] = true
	}
	if !found[retrying] || found[exhausted] {
		t.Fatalf("retrying clip queued=%v, exhausted clip queued=%v", found[retrying], found[exhausted])
	}
}
