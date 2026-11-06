//go:build integration

package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Missing comments and clips are 404s, acting on someone else's comment is a
// 403, and only caller mistakes stay 400.
func TestCommentWriteErrorStatuses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	ctx := context.Background()

	commentRepo := repository.NewCommentRepository(pool)
	clipRepo := repository.NewClipRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	handler := NewCommentHandler(services.NewCommentService(commentRepo, clipRepo, userRepo, nil, nil))

	suffix := uuid.NewString()[:8]
	owner := testutil.TestUser()
	owner.Username, owner.TwitchID, owner.Email = "cmt_owner_"+suffix, testutil.StringPtr("cmt_owner_"+suffix), testutil.StringPtr("cmt_owner_"+suffix+"@example.com")
	other := testutil.TestUser()
	other.Username, other.TwitchID, other.Email = "cmt_other_"+suffix, testutil.StringPtr("cmt_other_"+suffix), testutil.StringPtr("cmt_other_"+suffix+"@example.com")
	clip := testutil.TestClip()
	clip.TwitchClipID = "cmt_errors_" + suffix
	comment := testutil.TestComment(owner.ID, clip.ID)
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM comments WHERE clip_id = $1`, clip.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM clips WHERE id = $1`, clip.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = ANY($1)`, []uuid.UUID{owner.ID, other.ID})
	})
	if err := userRepo.Create(ctx, owner); err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := userRepo.Create(ctx, other); err != nil {
		t.Fatalf("create other: %v", err)
	}
	if err := clipRepo.Create(ctx, clip); err != nil {
		t.Fatalf("create clip: %v", err)
	}
	if err := commentRepo.Create(ctx, comment); err != nil {
		t.Fatalf("create comment: %v", err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		if id, err := uuid.Parse(c.GetHeader("X-Test-User")); err == nil {
			c.Set("user_id", id)
			c.Set("user_role", "user")
		}
	})
	router.POST("/clips/:id/comments", handler.CreateComment)
	router.PUT("/comments/:id", handler.UpdateComment)
	router.DELETE("/comments/:id", handler.DeleteComment)
	router.POST("/comments/:id/vote", handler.VoteOnComment)

	missing := uuid.NewString()
	tests := []struct {
		name, method, path, body string
		user                     uuid.UUID
		want                     int
	}{
		{"create on missing clip", http.MethodPost, "/clips/" + missing + "/comments", `{"content":"hi"}`, owner.ID, http.StatusNotFound},
		{"reply to missing parent", http.MethodPost, "/clips/" + clip.ID.String() + "/comments", `{"content":"hi","parent_comment_id":"` + missing + `"}`, owner.ID, http.StatusBadRequest},
		{"update missing comment", http.MethodPut, "/comments/" + missing, `{"content":"edit"}`, owner.ID, http.StatusNotFound},
		{"update someone else's comment", http.MethodPut, "/comments/" + comment.ID.String(), `{"content":"edit"}`, other.ID, http.StatusForbidden},
		{"delete missing comment", http.MethodDelete, "/comments/" + missing, ``, owner.ID, http.StatusNotFound},
		{"delete someone else's comment", http.MethodDelete, "/comments/" + comment.ID.String(), ``, other.ID, http.StatusForbidden},
		{"vote on missing comment", http.MethodPost, "/comments/" + missing + "/vote", `{"vote":1}`, owner.ID, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User", tt.user.String())
			router.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
