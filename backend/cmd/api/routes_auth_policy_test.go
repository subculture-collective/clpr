package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handlerChainRouter registers the public route groups behind a probe that
// reports the matched route's handler chain instead of running it.
func handlerChainRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := newRouter()
	router.Use(func(c *gin.Context) {
		c.Header("X-Handler-Chain", strings.Join(c.HandlerNames(), ","))
		c.AbortWithStatus(http.StatusNoContent)
	})
	v1 := router.Group("/api/v1")
	infra := &Infrastructure{Config: &config.Config{}}
	registerClipRoutes(v1, zeroHandlers(), &Services{}, infra)
	registerContentRoutes(v1, zeroHandlers(), &Services{}, infra)
	registerUserRoutes(v1, zeroHandlers(), &Services{}, infra)
	return router
}

func handlerChain(t *testing.T, router *gin.Engine, method, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	chain := rec.Header().Get("X-Handler-Chain")
	if chain == "" {
		t.Fatalf("%s %s matched no route", method, path)
	}
	return chain
}

// Signed-in users get personal fields (user_vote, is_favorited, comment votes,
// search history) only when these public GET routes resolve the optional
// identity.
func TestPersonalisedPublicRoutesResolveOptionalIdentity(t *testing.T) {
	router := handlerChainRouter()
	id := uuid.NewString()
	for _, path := range []string{
		"/api/v1/clips",
		"/api/v1/clips/" + id,
		"/api/v1/clips/" + id + "/comments",
		"/api/v1/comments/" + id + "/replies",
		"/api/v1/search",
		"/api/v1/search/scores",
	} {
		if chain := handlerChain(t, router, http.MethodGet, path); !strings.Contains(chain, ".OptionalAuthMiddleware.") {
			t.Errorf("GET %s lacks OptionalAuthMiddleware: %s", path, chain)
		}
	}
}

func TestProfileRoutesEnforcePrivacySettings(t *testing.T) {
	router := handlerChainRouter()
	id := uuid.NewString()
	for _, path := range []string{
		"/api/v1/users/by-username/someone",
		"/api/v1/users/" + id,
		"/api/v1/users/" + id + "/karma",
		"/api/v1/users/" + id + "/comments",
		"/api/v1/users/" + id + "/upvoted",
		"/api/v1/users/" + id + "/downvoted",
		"/api/v1/users/" + id + "/activity",
		"/api/v1/users/" + id + "/clips",
		"/api/v1/users/" + id + "/followers",
		"/api/v1/users/" + id + "/following",
		"/api/v1/users/" + id + "/following/broadcasters",
		"/api/v1/users/" + id + "/games/following",
		"/api/v1/users/" + id + "/twitch-categories/following",
		"/api/v1/users/" + id + "/reputation",
		"/api/v1/users/" + id + "/badges",
	} {
		chain := handlerChain(t, router, http.MethodGet, path)
		optional := strings.Index(chain, ".OptionalAuthMiddleware.")
		privacy := strings.Index(chain, ".ProfilePrivacy.")
		if optional < 0 || privacy < 0 || optional > privacy {
			t.Errorf("GET %s must run OptionalAuthMiddleware then ProfilePrivacy: %s", path, chain)
		}
	}
}

// Any signed-in user could strip any tag from any clip; removal is staff-only.
func TestClipTagRemovalRequiresStaffRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, role := range []string{models.RoleUser, models.RoleModerator, models.RoleAdmin} {
		t.Run(role, func(t *testing.T) {
			router := newRouter()
			router.Use(middleware.JSONRecoveryMiddleware())
			// Stand in for an authenticated request: AuthMiddleware reuses an
			// attached user.
			router.Use(func(c *gin.Context) {
				user := &models.User{ID: uuid.New(), Role: role}
				c.Set("user", user)
				c.Set("user_id", user.ID)
				c.Set("user_role", user.Role)
				c.Next()
			})
			registerClipRoutes(router.Group("/api/v1"), zeroHandlers(), &Services{}, &Infrastructure{Config: &config.Config{}})

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/clips/"+uuid.NewString()+"/tags/funny", nil))
			// Staff requests pass the role check and reach the zero-value
			// handler (which fails without a repository); others are refused.
			refused := rec.Code == http.StatusForbidden
			if staff := role != models.RoleUser; refused == staff {
				t.Fatalf("role %s: status %d, body %s", role, rec.Code, rec.Body.String())
			}
		})
	}
}
