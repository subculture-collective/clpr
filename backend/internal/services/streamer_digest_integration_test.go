//go:build integration

package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestStreamerDigestEndToEndInSandbox(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "streamer_digest_sends", "clip_packs", "countdown_series", "countdown_exclusions",
		"broadcaster_feature_optins", "email_notification_logs", "email_unsubscribe_tokens", "notification_preferences", "clips", "users")
	ctx := context.Background()
	countdowns := repository.NewCountdownRepository(pool)
	notifications := repository.NewNotificationRepository(pool)
	week := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

	// alpha: featured, email on. bravo: featured, email off by default. charlie: featured, excluded.
	users := map[string]uuid.UUID{}
	for _, b := range []string{"alpha", "bravo", "charlie"} {
		id := uuid.New()
		users[b] = id
		if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, email, role, account_type)
			VALUES ($1, $2, $2, $3, $4, 'user', 'member')`, id, b, strings.ToUpper(b[:1])+b[1:], b+"@example.com"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO broadcaster_feature_optins (broadcaster_id, user_id, scope, consent_text_version)
			VALUES ($1, $2, 'channel:manage:clips', '2026-11-07')`, b, id); err != nil {
			t.Fatal(err)
		}
		if _, err := notifications.CreateDefaultPreferences(ctx, id); err != nil {
			t.Fatal(err)
		}
	}
	for _, b := range []string{"alpha", "charlie"} {
		if _, err := pool.Exec(ctx, `UPDATE notification_preferences SET email_enabled = true WHERE user_id = $1`, users[b]); err != nil {
			t.Fatal(err)
		}
	}
	charlie := "charlie"
	if _, err := countdowns.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &charlie, Reason: "request"}, "admin", nil); err != nil {
		t.Fatal(err)
	}
	clipID := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO clips (id, twitch_clip_id, twitch_clip_url, embed_url, title, creator_name, broadcaster_name, broadcaster_id,
			duration, view_count, created_at, imported_at, source_type)
		VALUES ($1, 'tw-a', 'https://clips.twitch.tv/a', 'https://clips.twitch.tv/a', 'Alpha <b>ace</b>', 'x', 'alpha', 'alpha',
			30, 4321, $2, NOW(), 'twitch')`, clipID, week.Add(48*time.Hour)); err != nil {
		t.Fatal(err)
	}

	recipients, err := countdowns.ListDigestRecipients(ctx, week)
	if err != nil || len(recipients) != 1 || recipients[0].UserID != users["alpha"] || recipients[0].Name != "Alpha" {
		t.Fatalf("recipients = %+v, %v (bravo has email off, charlie is excluded)", recipients, err)
	}

	email := NewEmailService(&EmailConfig{Enabled: true, SandboxMode: true, BaseURL: "https://clpr.example"},
		repository.NewEmailNotificationRepository(pool), notifications)
	svc := &StreamerDigestService{Store: countdowns, Sender: email, Now: func() time.Time { return week.AddDate(0, 0, 7).Add(10 * time.Hour) }}
	if sent, err := svc.RunOnce(ctx); err != nil || sent != 1 {
		t.Fatalf("sent %d, %v", sent, err)
	}
	if sent, err := svc.RunOnce(ctx); err != nil || sent != 0 {
		t.Fatalf("second run sent %d, %v", sent, err)
	}
	var status, logStatus, subject string
	var notificationID *uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT status FROM streamer_digest_sends WHERE user_id = $1 AND week_start = '2026-11-02'`, users["alpha"]).Scan(&status); err != nil || status != "sent" {
		t.Fatalf("send row = %q, %v", status, err)
	}
	if err := pool.QueryRow(ctx, `SELECT status, subject, notification_id FROM email_notification_logs WHERE user_id = $1 AND notification_type = 'streamer_digest'`,
		users["alpha"]).Scan(&logStatus, &subject, &notificationID); err != nil || logStatus != "sent" || notificationID != nil ||
		!strings.Contains(subject, "Week of November 2, 2026") {
		t.Fatalf("email log = %q %q %v, %v", logStatus, subject, notificationID, err)
	}

	// The digest's unsubscribe link turns off only the digest.
	var token string
	if err := pool.QueryRow(ctx, `SELECT token FROM email_unsubscribe_tokens WHERE user_id = $1 AND notification_type = 'streamer_digest'`,
		users["alpha"]).Scan(&token); err != nil {
		t.Fatal(err)
	}
	if err := email.Unsubscribe(ctx, token); err != nil {
		t.Fatal(err)
	}
	prefs, err := notifications.GetPreferences(ctx, users["alpha"])
	if err != nil || prefs.NotifyStreamerDigest == nil || *prefs.NotifyStreamerDigest || !prefs.EmailEnabled {
		t.Fatalf("prefs after unsubscribe = %+v, %v", prefs, err)
	}
	if recipients, err := countdowns.ListDigestRecipients(ctx, week.AddDate(0, 0, 7)); err != nil || len(recipients) != 0 {
		t.Fatalf("unsubscribed broadcaster still listed: %+v, %v", recipients, err)
	}

	// A preferences update that leaves the field out keeps it.
	prefs.NotifyStreamerDigest = nil
	prefs.NotifyReplies = !prefs.NotifyReplies
	if err := notifications.UpdatePreferences(ctx, prefs); err != nil {
		t.Fatal(err)
	}
	if again, err := notifications.GetPreferences(ctx, users["alpha"]); err != nil || again.NotifyStreamerDigest == nil || *again.NotifyStreamerDigest {
		t.Fatalf("an update without the field changed it: %+v, %v", again, err)
	}

	content, err := countdowns.StreamerDigestContent(ctx, "alpha", week)
	if err != nil || len(content.TopClips) != 1 || content.TopClips[0].ViewCount != 4321 || content.PackStatus != "" {
		t.Fatalf("content = %+v, %v", content, err)
	}
}
