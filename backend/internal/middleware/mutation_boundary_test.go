package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type boundaryAuth struct{ valid bool }

func (a boundaryAuth) GetUserFromToken(context.Context, string) (*models.User, error) {
	if !a.valid {
		return nil, errors.New("invalid token")
	}
	return &models.User{ID: uuid.New(), Role: models.RoleUser}, nil
}

func TestMutationAuthorizationBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		method string
		path   string
		header string
		auth   boundaryAuth
		want   int
	}{
		{name: "public read", method: http.MethodGet, path: "/resource/1", want: http.StatusNoContent},
		{name: "anonymous mutation denied", method: http.MethodPatch, path: "/resource/1", want: http.StatusUnauthorized},
		{name: "invalid token denied", method: http.MethodDelete, path: "/resource/1", header: "Bearer bad", want: http.StatusUnauthorized},
		{name: "valid token allowed", method: http.MethodPut, path: "/resource/1", header: "Bearer valid", auth: boundaryAuth{valid: true}, want: http.StatusNoContent},
		{name: "reviewed public callback allowed", method: http.MethodPost, path: "/api/v1/webhooks/stripe", want: http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(MutationAuthorizationBoundary(tt.auth))
			r.Handle(tt.method, tt.path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			response := httptest.NewRecorder()
			r.ServeHTTP(response, req)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}

func TestPublicMutationAllowlistHasReasons(t *testing.T) {
	for route, reason := range PublicMutationPaths {
		if route == "" || reason == "" {
			t.Fatalf("public mutation exception lacks route or reason: %q=%q", route, reason)
		}
	}
}
