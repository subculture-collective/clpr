package services

import (
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestMarkStaleLiveStatusFailsOffline(t *testing.T) {
	now := time.Now()
	title, game := "Live now", "Game"
	started := now.Add(-time.Hour)
	status := &models.BroadcasterLiveStatus{
		BroadcasterID: "123", IsLive: true, ViewerCount: 99, StreamTitle: &title,
		GameName: &game, StartedAt: &started, LastChecked: now.Add(-3 * time.Minute),
	}
	markStaleLiveStatus(status, now)
	if status.IsLive || !status.IsStale || status.ViewerCount != 0 || status.StreamTitle != nil || status.GameName != nil || status.StartedAt != nil {
		t.Fatalf("stale status was not safely normalized: %+v", status)
	}
}

func TestMarkStaleLiveStatusPreservesFreshData(t *testing.T) {
	now := time.Now()
	status := &models.BroadcasterLiveStatus{BroadcasterID: "123", IsLive: true, ViewerCount: 99, LastChecked: now.Add(-time.Minute)}
	markStaleLiveStatus(status, now)
	if !status.IsLive || status.IsStale || status.ViewerCount != 99 {
		t.Fatalf("fresh status was modified: %+v", status)
	}
}
