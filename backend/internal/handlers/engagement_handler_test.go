package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// newEngagementRouter mirrors the production route /users/:id/engagement.
func newEngagementRouter(requester *models.User) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewEngagementHandler(nil, nil)
	router.GET("/users/:id/engagement", func(c *gin.Context) {
		if requester != nil {
			c.Set("user", requester)
		}
		c.Next()
	}, h.GetUserEngagementScore)
	return router
}

func TestGetUserEngagementScore_ReadsIDRouteParam(t *testing.T) {
	requester := &models.User{ID: uuid.New(), Role: models.RoleUser}

	tests := []struct {
		name string
		path string
		want int
	}{
		// Before the fix the handler read :userId, which this route never sets,
		// so every request failed UUID parsing with 400.
		{"other user's valid id reaches authorization", "/users/" + uuid.New().String() + "/engagement", http.StatusForbidden},
		{"malformed id is rejected", "/users/not-a-uuid/engagement", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			newEngagementRouter(requester).ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestGetUserEngagementScore_MeRequiresAuthenticatedUser(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/me/engagement", nil)
	newEngagementRouter(nil).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
