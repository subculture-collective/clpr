package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestClientIPIgnoresVisitorSuppliedForwardingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter()
	if err := configureClientIP(router, strings.Split(config.DefaultTrustedProxies, ",")); err != nil {
		t.Fatal(err)
	}
	router.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	tests := []struct {
		name, remote, xff, realIP, cfIP, want string
	}{
		// Cloudflare appends the visitor to any client-sent XFF; Caddy appends cloudflared.
		{"production chain", "172.27.0.12:40000", "154.47.25.168, 172.20.0.249", "", "", "154.47.25.168"},
		{"spoofed XFF prefix", "172.27.0.12:40000", "203.0.113.77, 154.47.25.168, 172.20.0.249", "", "", "154.47.25.168"},
		{"spoofed private XFF prefix", "172.27.0.12:40000", "10.0.0.1, 154.47.25.168, 172.20.0.249", "", "", "154.47.25.168"},
		{"spoofed X-Real-IP and CF header", "172.27.0.12:40000", "154.47.25.168, 172.20.0.249", "203.0.113.78", "203.0.113.79", "154.47.25.168"},
		{"IPv6 visitor", "172.27.0.12:40000", "2001:db8::1, 172.20.0.249", "", "", "2001:db8::1"},
		{"public peer cannot forward", "198.51.100.9:40000", "203.0.113.77", "", "", "198.51.100.9"},
		{"internal caller without header", "172.27.0.20:40000", "", "", "", "172.27.0.20"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ip", nil)
			req.RemoteAddr = tc.remote
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			if tc.realIP != "" {
				req.Header.Set("X-Real-IP", tc.realIP)
			}
			if tc.cfIP != "" {
				req.Header.Set("CF-Connecting-IP", tc.cfIP)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if got := rec.Body.String(); got != tc.want {
				t.Fatalf("ClientIP = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConfigureClientIPRejectsInvalidProxy(t *testing.T) {
	if err := configureClientIP(newRouter(), []string{"not-an-ip"}); err == nil {
		t.Fatal("expected invalid TRUSTED_PROXIES to fail")
	}
}
