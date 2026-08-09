package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func TestAmbiguousOpenAPIRoutesResolveByStaticSegmentAndMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Header("X-CLPR-Matched-Route", c.FullPath())
		c.AbortWithStatus(http.StatusNoContent)
	})

	v1 := router.Group("/api/v1")
	cfg := &config.Config{}
	infra := &Infrastructure{Config: cfg}
	handlers := zeroHandlers()
	services := &Services{}
	registerUserRoutes(v1, handlers, services, infra)
	registerSocialRoutes(v1, handlers, services, infra)
	registerAdminRoutes(v1, handlers, services, infra)

	tests := []struct {
		method, path, expected string
	}{
		{http.MethodGet, "/api/v1/users/by-username/reputation", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/karma", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/badges", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/comments", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/clips", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/activity", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/upvoted", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/downvoted", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/followers", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/following", "/api/v1/users/by-username/:username"},
		{http.MethodPost, "/api/v1/users/by-username/follow", "/api/v1/users/:id/follow"},
		{http.MethodPost, "/api/v1/users/by-username/block", "/api/v1/users/:id/block"},
		{http.MethodGet, "/api/v1/users/by-username/engagement", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/account-type", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/feeds", "/api/v1/users/by-username/:username"},
		{http.MethodGet, "/api/v1/users/by-username/filter-presets", "/api/v1/users/by-username/:username"},
		{http.MethodPost, "/api/v1/playlists/share/bookmark", "/api/v1/playlists/:id/bookmark"},
		{http.MethodPost, "/api/v1/playlists/share/clips", "/api/v1/playlists/:id/clips"},
		{http.MethodGet, "/api/v1/playlists/share/collaborators", "/api/v1/playlists/share/:token"},
		{http.MethodPost, "/api/v1/playlists/share/copy", "/api/v1/playlists/:id/copy"},
		{http.MethodPost, "/api/v1/playlists/share/like", "/api/v1/playlists/:id/like"},
		{http.MethodGet, "/api/v1/playlists/share/share-link", "/api/v1/playlists/share/:token"},
		{http.MethodPost, "/api/v1/playlists/share/track-share", "/api/v1/playlists/:id/track-share"},
		{http.MethodPost, "/api/v1/admin/moderation/abuse/approve", "/api/v1/admin/moderation/:id/approve"},
		{http.MethodPost, "/api/v1/admin/moderation/abuse/reject", "/api/v1/admin/moderation/:id/reject"},
	}
	if len(tests) != 25 {
		t.Fatalf("route-precedence execution cases = %d, want 25 owned exceptions", len(tests))
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if got := response.Header().Get("X-CLPR-Matched-Route"); got != test.expected {
				t.Fatalf("matched route %q, want %q", got, test.expected)
			}
		})
	}
}

func TestOpenAPIAmbiguousRouteIgnoreBudget(t *testing.T) {
	contents, err := os.ReadFile("../../../.redocly.lint-ignore.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]map[string][]string
	if err := yaml.Unmarshal(contents, &document); err != nil {
		t.Fatal(err)
	}
	if len(document) != 1 {
		t.Fatalf("ignore file must contain exactly one API, got %d", len(document))
	}
	rules, ok := document["docs/openapi/openapi.yaml"]
	if !ok {
		t.Fatal("ignore file must target docs/openapi/openapi.yaml")
	}
	if len(rules) != 1 {
		t.Fatalf("only no-ambiguous-paths may be ignored, got %d rules", len(rules))
	}
	actual, ok := rules["no-ambiguous-paths"]
	if !ok {
		t.Fatal("no-ambiguous-paths exception list is missing")
	}

	const users = "#/paths/~1api~1v1~1users~1{id}~1"
	const playlists = "#/paths/~1api~1v1~1playlists~1{id}~1"
	const moderation = "#/paths/~1api~1v1~1admin~1moderation~1{id}~1"
	expected := make([]string, 0, 25)
	for _, suffix := range []string{
		"reputation", "karma", "badges", "comments", "clips", "activity", "upvoted", "downvoted",
		"followers", "following", "follow", "block", "engagement", "account-type", "feeds", "filter-presets",
	} {
		expected = append(expected, users+suffix)
	}
	for _, suffix := range []string{"approve", "reject"} {
		expected = append(expected, moderation+suffix)
	}
	for _, suffix := range []string{"bookmark", "clips", "collaborators", "copy", "like", "share-link", "track-share"} {
		expected = append(expected, playlists+suffix)
	}

	sort.Strings(actual)
	sort.Strings(expected)
	if len(actual) != len(expected) {
		t.Fatalf("ambiguous-route exception count is %d, want the fixed budget %d", len(actual), len(expected))
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("ambiguous-route exception mismatch at %d: got %q, want %q", index, actual[index], expected[index])
		}
	}
}
