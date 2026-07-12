package main

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/handlers"
	"github.com/gin-gonic/gin"
)

func TestIncompleteFeatureRoutesDefaultAbsent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	cfg := &config.Config{}
	infra := &Infrastructure{Config: cfg}
	h := &Handlers{
		Stream:     &handlers.StreamHandler{},
		LiveStatus: &handlers.LiveStatusHandler{},
	}

	registerPlatformRoutes(v1, h, &Services{}, infra)
	registerSocialRoutes(v1, &Handlers{}, &Services{}, infra)

	assertRouteAbsent(t, router, "POST", "/api/v1/streams/:streamer/clips")
	assertRouteAbsent(t, router, "GET", "/api/v1/feed/live")
	assertRouteAbsent(t, router, "POST", "/api/v1/watch-parties")
	assertRouteAbsent(t, router, "GET", "/api/v1/users/:id/watch-party-stats")
}

func TestAccountDeletionRoutesRemainAbsentUntilErasureWorkerExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	registerUserRoutes(v1, &Handlers{}, &Services{}, &Infrastructure{Config: &config.Config{}})

	assertRouteAbsent(t, router, "POST", "/api/v1/users/me/delete")
	assertRouteAbsent(t, router, "POST", "/api/v1/users/me/delete/cancel")
	assertRouteAbsent(t, router, "GET", "/api/v1/users/me/delete/status")
	assertRouteAbsent(t, router, "GET", "/api/v1/users/me/export")
}

func TestIncompleteFeatureRoutesRequireExplicitFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	cfg := &config.Config{FeatureFlags: config.FeatureFlagsConfig{
		StreamClipCreation: true,
		LiveFeed:           true,
		WatchParties:       true,
	}}
	infra := &Infrastructure{Config: cfg}
	h := &Handlers{
		Stream:     &handlers.StreamHandler{},
		LiveStatus: &handlers.LiveStatusHandler{},
	}

	registerPlatformRoutes(v1, h, &Services{}, infra)
	registerSocialRoutes(v1, &Handlers{}, &Services{}, infra)

	assertRoutePresent(t, router, "POST", "/api/v1/streams/:streamer/clips")
	assertRoutePresent(t, router, "GET", "/api/v1/feed/live")
	assertRoutePresent(t, router, "POST", "/api/v1/watch-parties")
	assertRoutePresent(t, router, "GET", "/api/v1/users/:id/watch-party-stats")
}

func TestProfilingRoutesAreDebugOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		mode    string
		present bool
	}{{gin.ReleaseMode, false}, {gin.DebugMode, true}} {
		t.Run(test.mode, func(t *testing.T) {
			router := gin.New()
			cfg := &config.Config{Server: config.ServerConfig{GinMode: test.mode}}
			registerPublicRoutes(router, router.Group("/api/v1"), zeroHandlers(), &Services{}, &Infrastructure{Config: cfg}, cfg)
			if test.present {
				assertRoutePresent(t, router, "GET", "/debug/pprof/profile")
			} else {
				assertRouteAbsent(t, router, "GET", "/debug/pprof/profile")
				assertRouteAbsent(t, router, "GET", "/debug/pprof/heap")
			}
		})
	}
}

func assertRouteAbsent(t *testing.T, router *gin.Engine, method, path string) {
	t.Helper()
	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			t.Fatalf("unexpected route %s %s", method, path)
		}
	}
}

func assertRoutePresent(t *testing.T, router *gin.Engine, method, path string) {
	t.Helper()
	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			return
		}
	}
	t.Fatalf("missing route %s %s", method, path)
}
