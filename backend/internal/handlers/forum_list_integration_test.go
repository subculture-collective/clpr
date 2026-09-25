//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// The web client lists threads with sort=newest|most-replied|trending|hot and
// game_id; the backend accepted only recent|popular|replies and game_filter,
// so the forum index always showed "Failed to load threads".
func TestListThreadsAcceptsClientSortAndFilterValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "forum_replies", "forum_threads", "users")

	ctx := context.Background()
	userID, gameID := uuid.New(), uuid.New()
	twitchGameID := fmt.Sprintf("99%d", time.Now().UnixNano()%1_000_000_000)
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username) VALUES ($1, $2, $3)`, userID, uuid.NewString(), "forum-list"); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO games (id, twitch_game_id, name, slug) VALUES ($1, $2::text, 'Forum Game', 'forum-game-' || $2::text)`, gameID, twitchGameID); err != nil {
		t.Fatalf("create game: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM games WHERE id = $1`, gameID) })

	oldBusy, newQuiet, newActive := uuid.New(), uuid.New(), uuid.New()
	threads := []struct {
		id             uuid.UUID
		title          string
		replies, views int
		age            string
		game           *uuid.UUID
		tags           []string
	}{
		// Most lifetime replies, but a month old.
		{oldBusy, "Old busy thread", 40, 5000, "30 days", nil, []string{"discussion"}},
		// Newest, no engagement.
		{newQuiet, "New quiet thread", 0, 0, "1 minute", &gameID, []string{"help"}},
		// Recent with moderate engagement: should lead the hot ordering.
		{newActive, "New active thread", 12, 300, "3 hours", &gameID, []string{"help", "clips"}},
	}
	for _, th := range threads {
		if _, err := pool.Exec(ctx, `INSERT INTO forum_threads (id, user_id, title, content, reply_count, view_count, game_id, tags, created_at, updated_at)
			VALUES ($1, $2, $3, 'body', $4, $5, $6, $7, NOW() - $8::interval, NOW() - $8::interval)`,
			th.id, userID, th.title, th.replies, th.views, th.game, th.tags, th.age); err != nil {
			t.Fatalf("create thread %s: %v", th.title, err)
		}
	}

	handler := NewForumHandler(pool)
	router := gin.New()
	router.GET("/api/v1/forum/threads", handler.ListThreads)

	list := func(t *testing.T, query string) []uuid.UUID {
		t.Helper()
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/forum/threads?"+query, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d: %s", query, rec.Code, rec.Body.String())
		}
		var body struct {
			Data []struct {
				ID uuid.UUID `json:"id"`
			} `json:"data"`
			Meta struct {
				Limit int `json:"limit"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: decode: %v", query, err)
		}
		ids := make([]uuid.UUID, len(body.Data))
		for i, d := range body.Data {
			ids[i] = d.ID
		}
		return ids
	}
	expectOrder := func(t *testing.T, query string, want ...uuid.UUID) {
		t.Helper()
		got := list(t, query)
		if len(got) != len(want) {
			t.Fatalf("%s: got %d threads, want %d", query, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: position %d = %s, want %s (order %v)", query, i, got[i], want[i], got)
			}
		}
	}

	expectOrder(t, "sort=newest", newQuiet, newActive, oldBusy)
	expectOrder(t, "sort=recent", newQuiet, newActive, oldBusy)
	expectOrder(t, "sort=most-replied", oldBusy, newActive, newQuiet)
	expectOrder(t, "sort=replies", oldBusy, newActive, newQuiet)
	expectOrder(t, "sort=popular", oldBusy, newActive, newQuiet)
	expectOrder(t, "sort=hot", newActive, newQuiet, oldBusy)
	expectOrder(t, "sort=trending", newActive, newQuiet, oldBusy)
	expectOrder(t, "sort=newest&page=1&limit=2", newQuiet, newActive)

	expectOrder(t, "sort=newest&game_id="+gameID.String(), newQuiet, newActive)
	expectOrder(t, "sort=newest&game_filter="+gameID.String(), newQuiet, newActive)
	expectOrder(t, "sort=newest&game_id="+twitchGameID, newQuiet, newActive)
	expectOrder(t, "sort=newest&tags=help&tags=clips", newActive)
	expectOrder(t, "sort=newest&tags=help", newQuiet, newActive)
}
