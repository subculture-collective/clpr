//go:build integration

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"github.com/google/uuid"
)

// fakeStreamsClient answers Helix Get Streams from an in-memory set of live
// broadcasters and records every requested batch.
type fakeStreamsClient struct {
	live    map[string]twitch.Stream
	batches [][]string
}

func (f *fakeStreamsClient) GetStreams(_ context.Context, userIDs []string) (*twitch.StreamsResponse, error) {
	if len(userIDs) > 100 {
		return nil, fmt.Errorf("helix accepts at most 100 user_id values, got %d", len(userIDs))
	}
	f.batches = append(f.batches, append([]string(nil), userIDs...))
	resp := &twitch.StreamsResponse{Data: []twitch.Stream{}}
	for _, id := range userIDs {
		if stream, ok := f.live[id]; ok {
			resp.Data = append(resp.Data, stream)
		}
	}
	return resp, nil
}

func TestLiveStatusServiceTracksPopularBroadcasters(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "broadcaster_live_status", "broadcaster_sync_status", "broadcaster_sync_log", "clips")
	ctx := context.Background()

	// 150 broadcasters with visible clips; caseoh_ has the most.
	insertClip := func(broadcasterID string, age time.Duration, hidden bool) {
		t.Helper()
		clipID := uuid.NewString()
		if _, err := pool.Exec(ctx, `
			INSERT INTO clips (id, twitch_clip_id, twitch_clip_url, embed_url, title, creator_name,
				broadcaster_name, broadcaster_id, created_at, is_hidden)
			VALUES ($1, $2, $3, $4, 'clip', 'creator', $5, $5, NOW() - $6::interval, $7)`,
			uuid.New(), clipID, "https://clips.twitch.tv/"+clipID, "https://clips.twitch.tv/embed?clip="+clipID,
			broadcasterID, fmt.Sprintf("%d seconds", int(age.Seconds())), hidden); err != nil {
			t.Fatalf("insert clip: %v", err)
		}
	}
	for i := 0; i < 150; i++ {
		insertClip(fmt.Sprintf("b%03d", i), time.Hour, false)
	}
	for i := 0; i < 5; i++ {
		insertClip("caseoh_id", time.Hour, false)
	}
	// Outside the window, or hidden: never candidates on their own.
	for i := 0; i < 10; i++ {
		insertClip("old_id", 90*24*time.Hour, false)
		insertClip("hidden_id", time.Hour, true)
	}

	repo := repository.NewBroadcasterRepository(pool)
	ids, err := repo.GetLiveStatusCandidateBroadcasterIDs(ctx, 120, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("candidates: %v", err)
	}
	if len(ids) != 120 || ids[0] != "caseoh_id" {
		t.Fatalf("got %d candidates starting with %v, want 120 led by caseoh_id", len(ids), ids[:1])
	}
	for _, id := range ids {
		if id == "old_id" || id == "hidden_id" {
			t.Fatalf("candidate list includes %s", id)
		}
	}

	started := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	client := &fakeStreamsClient{live: map[string]twitch.Stream{
		"caseoh_id": {UserID: "caseoh_id", UserLogin: "caseoh_", UserName: "caseoh_", Type: "live", Title: "stream", GameName: "Just Chatting", ViewerCount: 59000, StartedAt: started},
	}}
	svc := &LiveStatusService{broadcasterRepo: repo, twitchClient: client}

	if err := svc.UpdateLiveStatusForBroadcasters(ctx, ids); err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(client.batches) != 2 || len(client.batches[0]) != 100 || len(client.batches[1]) != 20 {
		t.Fatalf("Helix batches = %d (%v sizes), want 100 + 20", len(client.batches), batchSizes(client.batches))
	}

	live, total, err := repo.ListLiveBroadcasters(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list live: %v", err)
	}
	if total != 1 || len(live) != 1 || live[0].BroadcasterID != "caseoh_id" || live[0].ViewerCount != 59000 {
		t.Fatalf("live list = %d %+v, want caseoh_id with 59000 viewers", total, live)
	}
	status, err := svc.GetLiveStatus(ctx, "caseoh_id")
	if err != nil || !status.IsLive {
		t.Fatalf("GetLiveStatus(caseoh_id) = %+v, %v; want live", status, err)
	}

	// Second pass: offline broadcasters already recorded offline are not rewritten.
	var offlineUpdatedBefore time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM broadcaster_sync_status WHERE broadcaster_id = $1`, ids[1]).Scan(&offlineUpdatedBefore); err != nil {
		t.Fatalf("offline broadcaster sync row: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := svc.UpdateLiveStatusForBroadcasters(ctx, ids); err != nil {
		t.Fatalf("second update: %v", err)
	}
	var offlineUpdatedAfter time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM broadcaster_sync_status WHERE broadcaster_id = $1`, ids[1]).Scan(&offlineUpdatedAfter); err != nil {
		t.Fatalf("offline broadcaster sync row: %v", err)
	}
	if !offlineUpdatedAfter.Equal(offlineUpdatedBefore) {
		t.Fatalf("offline->offline broadcaster was rewritten (%v -> %v)", offlineUpdatedBefore, offlineUpdatedAfter)
	}

	// Going offline is still recorded.
	delete(client.live, "caseoh_id")
	if err := svc.UpdateLiveStatusForBroadcasters(ctx, ids); err != nil {
		t.Fatalf("third update: %v", err)
	}
	if _, total, err := repo.ListLiveBroadcasters(ctx, 10, 0); err != nil || total != 0 {
		t.Fatalf("after going offline live total = %d, %v; want 0", total, err)
	}
	var wentOffline int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM broadcaster_sync_log WHERE broadcaster_id = 'caseoh_id' AND status_change = 'went_offline'`).Scan(&wentOffline); err != nil || wentOffline != 1 {
		t.Fatalf("went_offline log count = %d, %v; want 1", wentOffline, err)
	}
}

func batchSizes(batches [][]string) []int {
	sizes := make([]int, len(batches))
	for i, b := range batches {
		sizes[i] = len(b)
	}
	return sizes
}
