package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type shareClipRepositoryStub struct {
	byID      map[uuid.UUID]*models.Clip
	byTwitch  map[string]*models.Clip
	err       error
	lookedUps []string
}

func (s *shareClipRepositoryStub) GetByID(_ context.Context, id uuid.UUID) (*models.Clip, error) {
	s.lookedUps = append(s.lookedUps, "id:"+id.String())
	if s.err != nil {
		return nil, s.err
	}
	if clip, ok := s.byID[id]; ok {
		return clip, nil
	}
	return nil, fmt.Errorf("failed to get clip by ID: %w", pgx.ErrNoRows)
}

func (s *shareClipRepositoryStub) GetByTwitchClipID(_ context.Context, twitchClipID string) (*models.Clip, error) {
	s.lookedUps = append(s.lookedUps, "twitch:"+twitchClipID)
	if s.err != nil {
		return nil, s.err
	}
	if clip, ok := s.byTwitch[twitchClipID]; ok {
		return clip, nil
	}
	return nil, fmt.Errorf("failed to get clip by twitch ID: %w", pgx.ErrNoRows)
}

func serveSharePreview(t *testing.T, repo ClipLookupForShare, clipID string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("base_url", "https://clpr.example/")
		c.Next()
	})
	router.GET("/api/v1/share/clips/:id", NewShareHandler(repo).GetClipPreview)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/share/clips/"+clipID, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

func assertShareContains(t *testing.T, body string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(body, fragment) {
			t.Errorf("body is missing %q\n%s", fragment, body)
		}
	}
}

func TestShareClipPreviewRendersClipTags(t *testing.T) {
	id := uuid.MustParse("5f0c7a52-8a55-4d3f-9a1e-2f7f5d0f3b11")
	duration := 29.6
	repo := &shareClipRepositoryStub{byID: map[uuid.UUID]*models.Clip{id: {
		ID:              id,
		TwitchClipID:    "AwkwardHelplessSalamander",
		Title:           "Insane 1v5 clutch",
		BroadcasterName: "xQc",
		GameName:        strPtr("Valorant"),
		ThumbnailURL:    strPtr("https://static-cdn.jtvnw.net/twitch-clips/abc-preview-480x272.jpg"),
		ViewCount:       12345,
		Duration:        &duration,
	}}}

	recorder := serveSharePreview(t, repo, id.String())

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=300" {
		t.Errorf("Cache-Control = %q", got)
	}
	pageURL := "https://clpr.example/clip/" + id.String()
	assertShareContains(t, recorder.Body.String(),
		`<title>Insane 1v5 clutch</title>`,
		`<link rel="canonical" href="`+pageURL+`">`,
		`<meta property="og:url" content="`+pageURL+`">`,
		`<meta property="og:type" content="video.other">`,
		`<meta property="og:title" content="Insane 1v5 clutch · xQc">`,
		`<meta name="twitter:title" content="Insane 1v5 clutch · xQc">`,
		`<h1>Insane 1v5 clutch</h1>`,
		`<meta property="og:description" content="Clip from xQc · Valorant · 12,345 views · 30s — on clpr">`,
		`<meta property="og:image" content="https://static-cdn.jtvnw.net/twitch-clips/abc-preview-480x272.jpg">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<meta name="twitter:image" content="https://static-cdn.jtvnw.net/twitch-clips/abc-preview-480x272.jpg">`,
		`<a href="`+pageURL+`">Watch on clpr</a>`,
	)
	if strings.Contains(recorder.Body.String(), "og:image:width") {
		t.Error("clip thumbnails must not claim the default social card dimensions")
	}
}

func TestShareClipPreviewSocialTitle(t *testing.T) {
	cases := []struct {
		title, broadcaster, want string
	}{
		{"re", "caseoh_", "re · caseoh_"},
		{"  Clutch  ", "  xQc ", "Clutch · xQc"},
		{"No broadcaster", "", "No broadcaster"},
		{"", "xQc", "Twitch clip · xQc"},
	}
	for _, tc := range cases {
		preview := buildClipPreview("https://clpr.example", &models.Clip{ID: uuid.New(), Title: tc.title, BroadcasterName: tc.broadcaster})
		if preview.SocialTitle != tc.want {
			t.Errorf("title %q / broadcaster %q: SocialTitle = %q, want %q", tc.title, tc.broadcaster, preview.SocialTitle, tc.want)
		}
	}
}

func TestShareClipPreviewEscapesClipFields(t *testing.T) {
	repo := &shareClipRepositoryStub{byTwitch: map[string]*models.Clip{"Evil_Slug-1": {
		ID:              uuid.MustParse("0b9d3c1e-58f5-4a4e-8b53-6a1f0f7b8a01"),
		TwitchClipID:    "Evil_Slug-1",
		Title:           `"><script>alert(1)</script> & 'quotes'`,
		BroadcasterName: `<img src=x onerror=alert(2)>`,
		GameName:        strPtr(`Game "One"`),
		ThumbnailURL:    strPtr(`javascript:alert(3)`),
		ViewCount:       1,
	}}}

	recorder := serveSharePreview(t, repo, "Evil_Slug-1")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	body := recorder.Body.String()
	for _, raw := range []string{"<script>", "<img src=x", `"><`, "javascript:"} {
		if strings.Contains(body, raw) {
			t.Errorf("body contains unescaped %q\n%s", raw, body)
		}
	}
	assertShareContains(t, body,
		`&lt;script&gt;alert(1)&lt;/script&gt; &amp; &#39;quotes&#39;`,
		`<meta property="og:title" content="&#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt; &amp; &#39;quotes&#39; · &lt;img src=x onerror=alert(2)&gt;">`,
		`Clip from &lt;img src=x onerror=alert(2)&gt; · Game &#34;One&#34; · 1 view`,
		// Unsafe thumbnail schemes fall back to the site social card.
		`<meta property="og:image" content="https://clpr.example/social-card.png">`,
		`<meta property="og:image:width" content="1200">`,
	)
}

func TestShareClipPreviewNotFoundReturnsGenericTags(t *testing.T) {
	cases := map[string]struct {
		repo   *shareClipRepositoryStub
		clipID string
	}{
		"unknown uuid":   {repo: &shareClipRepositoryStub{}, clipID: uuid.NewString()},
		"unknown slug":   {repo: &shareClipRepositoryStub{}, clipID: "MissingClip"},
		"invalid id":     {repo: &shareClipRepositoryStub{}, clipID: "bad.id%3Cscript%3E"},
		"removed clip":   {repo: &shareClipRepositoryStub{byTwitch: map[string]*models.Clip{"Removed": {ID: uuid.New(), Title: "gone", IsRemoved: true}}}, clipID: "Removed"},
		"hidden clip":    {repo: &shareClipRepositoryStub{byTwitch: map[string]*models.Clip{"Hidden": {ID: uuid.New(), Title: "hidden", IsHidden: true}}}, clipID: "Hidden"},
		"dmca clip":      {repo: &shareClipRepositoryStub{byTwitch: map[string]*models.Clip{"Dmca": {ID: uuid.New(), Title: "dmca", DMCARemoved: true}}}, clipID: "Dmca"},
		"overlong input": {repo: &shareClipRepositoryStub{}, clipID: strings.Repeat("a", 129)},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recorder := serveSharePreview(t, tc.repo, tc.clipID)
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
				t.Errorf("Content-Type = %q", got)
			}
			assertShareContains(t, recorder.Body.String(),
				`<link rel="canonical" href="https://clpr.example/">`,
				`<meta property="og:type" content="website">`,
				`<meta property="og:title" content="clpr - Discover Creators and Live Moments">`,
				`<meta property="og:image" content="https://clpr.example/social-card.png">`,
			)
		})
	}
}

func TestShareClipPreviewInvalidIDSkipsLookup(t *testing.T) {
	repo := &shareClipRepositoryStub{}
	serveSharePreview(t, repo, "bad.id")
	if len(repo.lookedUps) != 0 {
		t.Fatalf("invalid id reached the repository: %v", repo.lookedUps)
	}
}

func TestShareClipPreviewRepositoryFailureIsNotCached(t *testing.T) {
	repo := &shareClipRepositoryStub{err: errors.New("connection refused")}

	recorder := serveSharePreview(t, repo, "SomeClip")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	assertShareContains(t, recorder.Body.String(), `<meta property="og:title" content="clpr - Discover Creators and Live Moments">`)
}

func TestShareClipPreviewDescriptionFormatting(t *testing.T) {
	long := 125.2
	clip := &models.Clip{BroadcasterName: "pokimane", ViewCount: 1234567, Duration: &long}
	if got, want := clipPreviewDescription(clip), "Clip from pokimane · 1,234,567 views · 2:05 — on clpr"; got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}
}

func TestShareClipPreviewResolvesRootRelativeThumbnails(t *testing.T) {
	if got, want := absoluteImageURL("https://clpr.example", strPtr("/media/thumb.jpg")), "https://clpr.example/media/thumb.jpg"; got != want {
		t.Fatalf("absoluteImageURL = %q, want %q", got, want)
	}
	if got := absoluteImageURL("https://clpr.example", strPtr("//evil.example/x.jpg")); got != "" {
		t.Fatalf("protocol-relative thumbnail = %q, want it rejected", got)
	}
}
