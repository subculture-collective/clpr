package services

import "testing"

func TestResolveSubmissionClipInput(t *testing.T) {
	service := &SubmissionService{}

	tests := []struct {
		name         string
		input        string
		wantSourceID string
		wantURL      string
		wantPlatform SourcePlatform
		wantType     SourceType
		wantErrField string
		wantErrMsg   string
	}{
		{
			name:         "supported YouTube URL is detected as external",
			input:        "https://www.youtube.com/watch?v=abc123",
			wantSourceID: "abc123",
			wantURL:      "https://www.youtube.com/watch?v=abc123",
			wantPlatform: SourcePlatformYouTube,
			wantType:     SourceTypeExternal,
		},
		{
			name:         "x URL is rejected through submission path",
			input:        "https://x.com/example/status/123",
			wantErrField: "clip_url",
			wantErrMsg:   "unsupported source/platform",
		},
		{
			name:         "instagram URL is rejected through submission path",
			input:        "https://www.instagram.com/reel/ABC123/",
			wantErrField: "clip_url",
			wantErrMsg:   "unsupported source/platform",
		},
		{
			name:         "direct Twitch ID bypasses detector",
			input:        "AwkwardHelplessSalamanderSwiftRage",
			wantSourceID: "AwkwardHelplessSalamanderSwiftRage",
			wantURL:      "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			wantPlatform: SourcePlatformTwitch,
			wantType:     SourceTypeTwitch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := service.resolveSubmissionClipInput(tt.input)

			if tt.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if valErr.Field != tt.wantErrField {
					t.Fatalf("field = %q, want %q", valErr.Field, tt.wantErrField)
				}
				if valErr.Message != tt.wantErrMsg {
					t.Fatalf("message = %q, want %q", valErr.Message, tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source.SourceID != tt.wantSourceID {
				t.Fatalf("sourceID = %q, want %q", source.SourceID, tt.wantSourceID)
			}
			if source.NormalizedURL != tt.wantURL {
				t.Fatalf("normalizedURL = %q, want %q", source.NormalizedURL, tt.wantURL)
			}
			if source.Platform != tt.wantPlatform {
				t.Fatalf("platform = %q, want %q", source.Platform, tt.wantPlatform)
			}
			if source.SourceType != tt.wantType {
				t.Fatalf("sourceType = %q, want %q", source.SourceType, tt.wantType)
			}
		})
	}
}
