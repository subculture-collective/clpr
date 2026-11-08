//go:build integration

package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

type fakeDailySource struct {
	calls int
	clips []models.DailyTopClip
}

func (f *fakeDailySource) ListDailyTopClips(context.Context, time.Time, time.Time, int) ([]models.DailyTopClip, error) {
	f.calls++
	return f.clips, nil
}

func TestDiscordDailyTopClipsDelivery(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "webhook_deliveries", "webhook_subscriptions", "users")
	ctx := context.Background()

	var mu sync.Mutex
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		mu.Lock()
		bodies = append(bodies, body)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent) // what Discord answers
	}))
	defer server.Close()

	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, 'tw-hook', 'hook_owner', 'Hook Owner', 'user', 'member')`, userID); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewOutboundWebhookRepository(pool)
	// The test server is on loopback, which CreateSubscription refuses, so
	// the row goes in through the repository.
	sub := &models.WebhookSubscription{ID: uuid.New(), UserID: userID, URL: server.URL, Secret: "s", IsActive: true,
		Events: []string{models.WebhookEventDailyTopClips}, TargetType: models.WebhookTargetDiscord}
	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.GetSubscriptionByID(ctx, sub.ID); err != nil || got.TargetType != models.WebhookTargetDiscord {
		t.Fatalf("subscription = %+v, %v", got, err)
	}

	thumb := "https://clips-media.example/1.jpg"
	source := &fakeDailySource{clips: []models.DailyTopClip{{ClipID: uuid.New(), Title: "ace", BroadcasterName: "alpha", ThumbnailURL: &thumb, ViewCount: 900}}}
	svc := NewOutboundWebhookService(repo)
	svc.SetDailyTopClips(source, "https://clpr.tv", 16)

	before := time.Date(2026, 11, 7, 15, 59, 0, 0, time.UTC)
	if err := svc.QueueDailyTopClips(ctx, before); err != nil || source.calls != 0 {
		t.Fatalf("queued before the hour: calls %d, %v", source.calls, err)
	}
	due := time.Date(2026, 11, 7, 16, 5, 0, 0, time.UTC)
	for i := 0; i < 2; i++ {
		if err := svc.QueueDailyTopClips(ctx, due); err != nil {
			t.Fatal(err)
		}
	}
	if source.calls != 1 {
		t.Fatalf("source called %d times, want once a day", source.calls)
	}
	// A second instance, or a restart, queues the same day again: the
	// delivery must not double.
	other := NewOutboundWebhookService(repo)
	other.SetDailyTopClips(source, "https://clpr.tv", 16)
	if err := other.QueueDailyTopClips(ctx, due); err != nil {
		t.Fatal(err)
	}
	var queued int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM webhook_deliveries WHERE subscription_id = $1`, sub.ID).Scan(&queued); err != nil || queued != 1 {
		t.Fatalf("queued deliveries = %d, %v", queued, err)
	}

	if err := svc.ProcessPendingDeliveries(ctx, 10); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 1 {
		t.Fatalf("Discord received %d posts", len(bodies))
	}
	if bodies[0]["content"] != "**Top Twitch clips of Nov 7, 2026**" || bodies[0]["username"] != "clpr" {
		t.Fatalf("body = %v", bodies[0])
	}
	embeds, _ := bodies[0]["embeds"].([]any)
	if len(embeds) != 1 {
		t.Fatalf("embeds = %v", bodies[0]["embeds"])
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM webhook_deliveries WHERE subscription_id = $1`, sub.ID).Scan(&status); err != nil || status != "delivered" {
		t.Fatalf("delivery status = %q, %v", status, err)
	}
}
