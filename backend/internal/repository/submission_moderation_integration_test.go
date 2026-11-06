//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

// Sync can import a Twitch clip after it was submitted; approval must link the
// submission to that clip instead of violating clips.twitch_clip_id UNIQUE.
func TestSubmissionModerateAtomically_ClaimsClipImportedAfterSubmission(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	submitter, other, reviewer := uuid.New(), uuid.New(), uuid.New()
	twitchID := "approval-race-" + uuid.NewString()
	otherTwitchID := "approval-race-other-" + uuid.NewString()
	t.Cleanup(func() {
		bg := context.Background()
		_, _ = pool.Exec(bg, `DELETE FROM moderation_audit_logs WHERE moderator_id = $1`, reviewer)
		_, _ = pool.Exec(bg, `DELETE FROM clip_submissions WHERE twitch_clip_id = ANY($1)`, []string{twitchID, otherTwitchID})
		_, _ = pool.Exec(bg, `DELETE FROM clips WHERE twitch_clip_id = ANY($1)`, []string{twitchID, otherTwitchID})
		_, _ = pool.Exec(bg, `DELETE FROM users WHERE id = ANY($1)`, []uuid.UUID{submitter, other, reviewer})
	})
	for _, id := range []uuid.UUID{submitter, other, reviewer} {
		name := "approval-" + id.String()[:18]
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,twitch_id,username) VALUES($1,$2,$2)`, id, name); err != nil {
			t.Fatal(err)
		}
	}
	insertClip := func(twitchClipID string, owner *uuid.UUID) uuid.UUID {
		id := uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO clips
			(id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,created_at,submitted_by_user_id)
			VALUES($1,$2,'https://clips.example/x','https://clips.example/embed','synced title','creator','synced broadcaster',NOW(),$3)`,
			id, twitchClipID, owner); err != nil {
			t.Fatal(err)
		}
		return id
	}
	createSubmission := func(twitchClipID string) *models.ClipSubmission {
		custom := "custom title"
		sub := &models.ClipSubmission{ID: uuid.New(), UserID: submitter, TwitchClipID: twitchClipID,
			TwitchClipURL: "https://clips.twitch.tv/" + twitchClipID, CustomTitle: &custom, IsNSFW: true,
			Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := NewSubmissionRepository(pool).Create(ctx, sub); err != nil {
			t.Fatal(err)
		}
		return sub
	}
	prepared := func(sub *models.ClipSubmission) map[uuid.UUID]*models.Clip {
		now := time.Now()
		return map[uuid.UUID]*models.Clip{sub.ID: {ID: uuid.New(), TwitchClipID: sub.TwitchClipID, TwitchClipURL: sub.TwitchClipURL,
			EmbedURL: "https://clips.twitch.tv/embed?clip=" + sub.TwitchClipID, Title: "custom title", CreatorName: "creator",
			BroadcasterName: "broadcaster", CreatedAt: now, ImportedAt: now, IsNSFW: sub.IsNSFW,
			SubmittedByUserID: &submitter, SubmittedAt: &sub.CreatedAt, SourceType: "twitch", SourcePlatform: "twitch",
			SourceMetadata: []byte(`{}`)}}
	}
	repo := NewSubmissionRepository(pool)

	sub := createSubmission(twitchID)
	syncedID := insertClip(twitchID, nil)
	approved, err := repo.ModerateAtomically(ctx, []uuid.UUID{sub.ID}, reviewer, true, nil, prepared(sub))
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved[0].ClipID == nil || *approved[0].ClipID != syncedID {
		t.Fatalf("submission clip_id=%v, want synced clip %s", approved[0].ClipID, syncedID)
	}
	var clips int
	var owner *uuid.UUID
	var title string
	var nsfw bool
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) OVER (), submitted_by_user_id, title, is_nsfw FROM clips WHERE twitch_clip_id=$1`, twitchID).Scan(&clips, &owner, &title, &nsfw); err != nil {
		t.Fatal(err)
	}
	if clips != 1 || owner == nil || *owner != submitter || title != "custom title" || !nsfw {
		t.Fatalf("clips=%d owner=%v title=%q nsfw=%v", clips, owner, title, nsfw)
	}

	// A clip already claimed by a different user is not silently reassigned.
	sub2 := createSubmission(otherTwitchID)
	insertClip(otherTwitchID, &other)
	if _, err := repo.ModerateAtomically(ctx, []uuid.UUID{sub2.ID}, reviewer, true, nil, prepared(sub2)); !errors.Is(err, ErrSubmissionClipClaimed) {
		t.Fatalf("approve claimed clip err=%v, want ErrSubmissionClipClaimed", err)
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM clip_submissions WHERE id=$1`, sub2.ID).Scan(&status); err != nil || status != "pending" {
		t.Fatalf("status=%q err=%v", status, err)
	}
}
