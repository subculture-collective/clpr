//go:build integration

package repository

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestStreamerClipRoomRepository_ListActiveRoomsAndHasItemForClip(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewStreamerClipRoomRepository(pool)

	owner := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username) VALUES ($1, $2, $2)`, owner, "room-"+owner.String()[:18]); err != nil {
		t.Fatalf("insert owner: %v", err)
	}
	// Deleting the owner cascades to the rooms and their items.
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, owner) })

	active, inactive := uuid.New(), uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO streamer_clip_rooms (id, owner_user_id, twitch_channel, is_active, last_listener_error, listener_started_at) VALUES
		($1, $3, 'active_channel', true, 'earlier failure', NOW()),
		($2, $3, 'inactive_channel', false, NULL, NULL)`, active, inactive, owner); err != nil {
		t.Fatalf("insert rooms: %v", err)
	}

	rooms, err := repo.ListActiveRooms(ctx)
	if err != nil {
		t.Fatalf("ListActiveRooms: %v", err)
	}
	found := false
	for _, room := range rooms {
		if room.ID == inactive {
			t.Fatal("ListActiveRooms returned an inactive room")
		}
		if room.ID == active {
			found = true
			if room.OwnerUserID != owner || room.TwitchChannel != "active_channel" || !room.IsActive ||
				room.ApprovalMode != "manual" || room.LastListenerError == nil || room.ListenerStartedAt == nil {
				t.Fatalf("active room = %+v", room)
			}
		}
	}
	if !found {
		t.Fatalf("ListActiveRooms missed the active room among %d rooms", len(rooms))
	}

	clipID := insertVisibilityClip(t, pool, "room-creator", "room-broadcaster", false, false)
	const clipURL, bareURL = "https://clpr.tv/clip/queued", "https://clips.twitch.tv/BareUrl"
	if _, err := pool.Exec(ctx, `
		INSERT INTO streamer_clip_room_items (room_id, clip_id, source_url, source_type, status) VALUES
		($1, $2, $3, 'clpr', 'pending'),
		($1, NULL, $4, 'twitch', 'rejected')`, active, clipID, clipURL, bareURL); err != nil {
		t.Fatalf("insert items: %v", err)
	}

	otherClip := uuid.New()
	cases := []struct {
		name     string
		room     uuid.UUID
		clipID   *uuid.UUID
		url      string
		statuses []string
		want     bool
	}{
		{"same clip under another URL", active, &clipID, "https://clpr.tv/clip/other", []string{"pending", "approved"}, true},
		{"same URL without a clip", active, nil, clipURL, []string{"pending"}, true},
		{"same URL with another clip", active, &otherClip, bareURL, []string{"rejected"}, true},
		{"status not asked for", active, &clipID, clipURL, []string{"approved"}, false},
		{"unknown clip and URL", active, &otherClip, "https://clpr.tv/clip/other", []string{"pending", "rejected"}, false},
		{"another room", inactive, &clipID, clipURL, []string{"pending"}, false},
	}
	for _, tc := range cases {
		got, err := repo.HasItemForClip(ctx, tc.room, tc.clipID, tc.url, tc.statuses)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s: HasItemForClip = %v, want %v", tc.name, got, tc.want)
		}
	}
}
