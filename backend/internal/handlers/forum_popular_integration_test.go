//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Production returned 500 for the popular-discussions and helpful-replies
// dashboards: the timeframe was bound as a Go int into `$2 || ' days'`, which
// PostgreSQL types as text, so pgx could not encode the argument.
func TestForumDashboardsQueryDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "forum_replies", "forum_threads", "users")

	ctx := context.Background()
	userID, threadID := uuid.New(), uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username) VALUES ($1, $2, $3)`, userID, uuid.NewString(), "forum-popular"); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO forum_threads (id, user_id, title, content, reply_count, view_count) VALUES ($1, $2, 'Popular thread', 'body', 4, 100)`, threadID, userID); err != nil {
		t.Fatalf("create thread: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO forum_replies (thread_id, user_id, content, depth, path) VALUES ($1, $2, 'helpful', 0, 'r1')`, threadID, userID); err != nil {
		t.Fatalf("create reply: %v", err)
	}

	handler := NewForumHandler(pool)
	router := gin.New()
	router.GET("/api/v1/forum/popular", handler.GetPopularDiscussions)
	router.GET("/api/v1/forum/helpful-replies", handler.GetMostHelpfulReplies)

	for _, path := range []string{
		"/api/v1/forum/popular",
		"/api/v1/forum/popular?timeframe=day",
		"/api/v1/forum/popular?timeframe=all",
		"/api/v1/forum/helpful-replies",
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200: %s", path, response.Code, response.Body.String())
		}
		var body struct {
			Data []json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: decode: %v", path, err)
		}
		if len(body.Data) != 1 {
			t.Fatalf("%s: data length = %d, want 1: %s", path, len(body.Data), response.Body.String())
		}
	}
}
