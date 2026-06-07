package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExternalMetadataFetcher_YouTubeOEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			if got := r.URL.Query().Get("url"); got == "" {
				t.Fatalf("expected url query param")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":"Great play","author_name":"StreamerOne","thumbnail_url":"https://cdn.example/thumb.jpg","html":"<iframe src=\"https://www.youtube.com/embed/abc123\"></iframe>"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	fetcher := NewExternalMetadataFetcher(server.Client())
	metadata, err := fetcher.Fetch(context.Background(), DetectedSource{
		RawURL:        server.URL + "/watch?v=abc123",
		NormalizedURL: server.URL + "/watch?v=abc123",
		Platform:      SourcePlatformYouTube,
		SourceType:    SourceTypeExternal,
		SourceID:      "abc123",
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if metadata.Title != "Great play" {
		t.Fatalf("Title = %q, want %q", metadata.Title, "Great play")
	}
	if metadata.AuthorName != "StreamerOne" {
		t.Fatalf("AuthorName = %q, want %q", metadata.AuthorName, "StreamerOne")
	}
	if metadata.ThumbnailURL != "https://cdn.example/thumb.jpg" {
		t.Fatalf("ThumbnailURL = %q, want thumbnail", metadata.ThumbnailURL)
	}
	if metadata.EmbedURL != "https://www.youtube.com/embed/abc123" {
		t.Fatalf("EmbedURL = %q, want embed URL", metadata.EmbedURL)
	}
	if metadata.DurationVerified {
		t.Fatalf("DurationVerified = true, want false")
	}
	if metadata.DurationSeconds != nil {
		t.Fatalf("DurationSeconds = %v, want nil", metadata.DurationSeconds)
	}
	if _, ok := metadata.Raw["oembed"]; !ok {
		t.Fatalf("Raw missing oembed payload: %#v", metadata.Raw)
	}
}

func TestExternalMetadataFetcher_YouTubeShortsFallsBackToOpenGraph(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			w.WriteHeader(http.StatusNotFound)
		case "/shorts/abc123":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html><html><head>
				<title>Shorts title</title>
				<meta property="og:title" content="Shorts title">
				<meta property="og:image" content="/thumb.jpg">
				<meta property="twitter:creator" content="ShortsCreator">
				<link rel="canonical" href="/shorts/abc123">
			</head><body></body></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	fetcher := NewExternalMetadataFetcher(server.Client())
	metadata, err := fetcher.Fetch(t.Context(), DetectedSource{
		RawURL:        server.URL + "/shorts/abc123",
		NormalizedURL: server.URL + "/shorts/abc123",
		Platform:      SourcePlatformYouTubeShorts,
		SourceType:    SourceTypeExternal,
		SourceID:      "abc123",
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if metadata.Title != "Shorts title" {
		t.Fatalf("Title = %q, want shorts title", metadata.Title)
	}
	if metadata.AuthorName != "ShortsCreator" {
		t.Fatalf("AuthorName = %q, want ShortsCreator", metadata.AuthorName)
	}
	if !strings.HasSuffix(metadata.ThumbnailURL, "/thumb.jpg") {
		t.Fatalf("ThumbnailURL = %q, want thumbnail path", metadata.ThumbnailURL)
	}
	if metadata.EmbedURL != server.URL+"/shorts/abc123" {
		t.Fatalf("EmbedURL = %q, want canonical shorts URL", metadata.EmbedURL)
	}
	if metadata.DurationVerified {
		t.Fatalf("DurationVerified = true, want false")
	}
}

func TestExternalMetadataFetcher_KickAndTikTokOpenGraphFallback(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		platform SourcePlatform
	}{
		{name: "kick", path: "/creator?clip=123", platform: SourcePlatformKick},
		{name: "tiktok", path: "/@creator/video/123456789", platform: SourcePlatformTikTok},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(`<!doctype html><html><head>
					<title>OG title</title>
					<meta property="og:title" content="OG title">
					<meta property="og:image" content="https://cdn.example/og.jpg">
					<meta property="og:video:url" content="https://player.example/embed">
					<meta property="twitter:creator" content="OGCreator">
				</head><body></body></html>`))
			}))
			defer server.Close()

			fetcher := NewExternalMetadataFetcher(server.Client())
			metadata, err := fetcher.Fetch(context.Background(), DetectedSource{
				RawURL:        server.URL + tt.path,
				NormalizedURL: server.URL + tt.path,
				Platform:      tt.platform,
				SourceType:    SourceTypeExternal,
				SourceID:      "123",
			})
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}

			if metadata.Title != "OG title" {
				t.Fatalf("Title = %q, want OG title", metadata.Title)
			}
			if metadata.AuthorName != "OGCreator" {
				t.Fatalf("AuthorName = %q, want OGCreator", metadata.AuthorName)
			}
			if metadata.ThumbnailURL != "https://cdn.example/og.jpg" {
				t.Fatalf("ThumbnailURL = %q, want OG image", metadata.ThumbnailURL)
			}
			if metadata.EmbedURL != "https://player.example/embed" {
				t.Fatalf("EmbedURL = %q, want og:video url", metadata.EmbedURL)
			}
			if metadata.DurationVerified {
				t.Fatalf("DurationVerified = true, want false")
			}
			if _, ok := metadata.Raw["open_graph"]; !ok {
				t.Fatalf("Raw missing open_graph payload: %#v", metadata.Raw)
			}
		})
	}
}
