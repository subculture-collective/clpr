package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/gin-gonic/gin"
)

func TestReleaseRouterRejectsTestLoginEvenWithDevelopmentEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := &config.Config{Server: config.ServerConfig{Environment: "development", GinMode: "release"}}
	registerAuthRoutes(router.Group("/api/v1"), zeroHandlers(), &Services{}, &Infrastructure{Config: cfg})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth/test-login", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("release test-login returned %d; want 404", w.Code)
	}
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
