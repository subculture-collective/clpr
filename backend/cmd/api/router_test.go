package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/gin-gonic/gin"
)

func TestNamespacedTagSlugsRouteAsOneSegment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter()
	router.Use(func(c *gin.Context) {
		c.Header("X-CLPR-Matched-Route", c.FullPath())
		c.Header("X-CLPR-Slug", c.Param("slug"))
		c.AbortWithStatus(http.StatusNoContent)
	})
	v1 := router.Group("/api/v1")
	infra := &Infrastructure{Config: &config.Config{}}
	registerContentRoutes(v1, zeroHandlers(), &Services{}, infra)
	registerClipRoutes(v1, zeroHandlers(), &Services{}, infra)

	tests := []struct {
		method, path, route, slug string
	}{
		{http.MethodGet, "/api/v1/tags/content%2Ffunny", "/api/v1/tags/:slug", "content/funny"},
		{http.MethodGet, "/api/v1/tags/content%2Ffunny/clips", "/api/v1/tags/:slug/clips", "content/funny"},
		{http.MethodGet, "/api/v1/tags/goosebumps", "/api/v1/tags/:slug", "goosebumps"},
		{http.MethodGet, "/api/v1/tags/tree", "/api/v1/tags/tree", ""},
		{http.MethodDelete, "/api/v1/clips/abc/tags/streamer%2Fenglish", "/api/v1/clips/:id/tags/:slug", "streamer/english"},
	}
	for _, tc := range tests {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if got := rec.Header().Get("X-CLPR-Matched-Route"); got != tc.route {
			t.Errorf("%s %s matched %q, want %q", tc.method, tc.path, got, tc.route)
		}
		if got := rec.Header().Get("X-CLPR-Slug"); got != tc.slug {
			t.Errorf("%s %s slug = %q, want %q", tc.method, tc.path, got, tc.slug)
		}
	}
}
