package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// unreachablePool returns a lazily connecting pool whose every query fails
// with a connection error, standing in for a database outage.
func unreachablePool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), "postgres://clpr:unused@127.0.0.1:1/unreachable?connect_timeout=1")
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// Database failures in comment writes must not come back as 400s carrying the
// driver's error text.
func TestCommentWritesHideDatabaseErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := unreachablePool(t)
	service := services.NewCommentService(repository.NewCommentRepository(pool), repository.NewClipRepository(pool), repository.NewUserRepository(pool), nil, nil)
	handler := NewCommentHandler(service)
	userID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", userID) })
	router.POST("/clips/:id/comments", handler.CreateComment)
	router.PUT("/comments/:id", handler.UpdateComment)
	router.DELETE("/comments/:id", handler.DeleteComment)
	router.POST("/comments/:id/vote", handler.VoteOnComment)

	tests := []struct {
		name, method, path, body string
	}{
		{"create", http.MethodPost, "/clips/" + uuid.NewString() + "/comments", `{"content":"hello"}`},
		{"update", http.MethodPut, "/comments/" + uuid.NewString(), `{"content":"hello"}`},
		{"delete", http.MethodDelete, "/comments/" + uuid.NewString(), ``},
		{"vote", http.MethodPost, "/comments/" + uuid.NewString() + "/vote", `{"vote":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500: %s", rec.Code, rec.Body.String())
			}
			body := strings.ToLower(rec.Body.String())
			for _, leak := range []string{"dial", "connect", "127.0.0.1", "failed to get", "failed to check"} {
				if strings.Contains(body, leak) {
					t.Fatalf("response leaks %q: %s", leak, rec.Body.String())
				}
			}
		})
	}
}

func TestCommentWritesRejectInvalidInputWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := unreachablePool(t)
	service := services.NewCommentService(repository.NewCommentRepository(pool), repository.NewClipRepository(pool), repository.NewUserRepository(pool), nil, nil)
	handler := NewCommentHandler(service)

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", uuid.New()) })
	router.PUT("/comments/:id", handler.UpdateComment)
	router.POST("/comments/:id/vote", handler.VoteOnComment)

	for _, tt := range []struct{ name, method, path, body, want string }{
		{"empty update", http.MethodPut, "/comments/" + uuid.NewString(), `{"content":"  "}`, "comment must be at least"},
		{"bad vote", http.MethodPost, "/comments/" + uuid.NewString() + "/vote", `{"vote":2}`, "invalid vote type"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), tt.want) {
				t.Fatalf("got %d %s, want 400 containing %q", rec.Code, rec.Body.String(), tt.want)
			}
		})
	}
}
