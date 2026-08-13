package services

import "testing"

func TestDetectClipSource_SupportedPlatforms(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedPlatform SourcePlatform
		expectedType     SourceType
		expectedID       string
		expectedURL      string
	}{
		{
			name:             "Twitch clips subdomain",
			input:            "https://clips.twitch.tv/SomeClip",
			expectedPlatform: SourcePlatformTwitch,
			expectedType:     SourceTypeTwitch,
			expectedID:       "SomeClip",
			expectedURL:      "https://clips.twitch.tv/SomeClip",
		},
		{
			name:             "Twitch creator clip url",
			input:            "https://www.twitch.tv/example/clip/SomeClip",
			expectedPlatform: SourcePlatformTwitch,
			expectedType:     SourceTypeTwitch,
			expectedID:       "SomeClip",
			expectedURL:      "https://clips.twitch.tv/SomeClip",
		},
		{
			name:             "Kick clip query",
			input:            "https://kick.com/example?clip=123",
			expectedPlatform: SourcePlatformKick,
			expectedType:     SourceTypeExternal,
			expectedID:       "123",
			expectedURL:      "https://kick.com/example?clip=123",
		},
		{
			name:             "YouTube watch",
			input:            "https://www.youtube.com/watch?v=abc123",
			expectedPlatform: SourcePlatformYouTube,
			expectedType:     SourceTypeExternal,
			expectedID:       "abc123",
			expectedURL:      "https://www.youtube.com/watch?v=abc123",
		},
		{
			name:             "YouTube short url",
			input:            "https://youtu.be/abc123",
			expectedPlatform: SourcePlatformYouTube,
			expectedType:     SourceTypeExternal,
			expectedID:       "abc123",
			expectedURL:      "https://www.youtube.com/watch?v=abc123",
		},
		{
			name:             "YouTube shorts",
			input:            "https://www.youtube.com/shorts/abc123",
			expectedPlatform: SourcePlatformYouTubeShorts,
			expectedType:     SourceTypeExternal,
			expectedID:       "abc123",
			expectedURL:      "https://www.youtube.com/shorts/abc123",
		},
		{
			name:             "TikTok video",
			input:            "https://www.tiktok.com/@creator/video/123456789",
			expectedPlatform: SourcePlatformTikTok,
			expectedType:     SourceTypeExternal,
			expectedID:       "123456789",
			expectedURL:      "https://www.tiktok.com/@creator/video/123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectClipSource(tt.input)
			if err != nil {
				t.Fatalf("DetectClipSource() error = %v", err)
			}
			if got.Platform != tt.expectedPlatform {
				t.Fatalf("Platform = %q, want %q", got.Platform, tt.expectedPlatform)
			}
			if got.SourceType != tt.expectedType {
				t.Fatalf("SourceType = %q, want %q", got.SourceType, tt.expectedType)
			}
			if got.SourceID != tt.expectedID {
				t.Fatalf("SourceID = %q, want %q", got.SourceID, tt.expectedID)
			}
			if got.NormalizedURL != tt.expectedURL {
				t.Fatalf("NormalizedURL = %q, want %q", got.NormalizedURL, tt.expectedURL)
			}
		})
	}
}

func TestDetectClipSource_UnsupportedPlatforms(t *testing.T) {
	tests := []string{
		"https://x.com/example/status/123",
		"https://twitter.com/example/status/123",
		"https://www.instagram.com/reel/ABC123/",
		"https://clips.twitch.tv/SomeClip/extra",
		"https://www.twitch.tv/example/clip/SomeClip/extra",
		"https://kick.com/example/extra?clip=123",
		"https://www.youtube.com/watch/extra?v=abc123",
		"https://www.youtube.com/shorts/abc123/extra",
		"https://www.tiktok.com/@creator/video/123456789/extra",
		"https://www.twitch.tv/example/clip/%2e%2e",
		"https://www.youtube.com/shorts/%2Fabc123",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := DetectClipSource(input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != "unsupported source/platform" {
				t.Fatalf("error = %q, want unsupported source/platform", err.Error())
			}
		})
	}
}
