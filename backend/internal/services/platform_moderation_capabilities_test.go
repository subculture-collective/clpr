package services

import "testing"

func TestDefaultPlatformModerationCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     PlatformModerationCapabilities
	}{
		{
			name:     "twitch",
			platform: "twitch",
			want: PlatformModerationCapabilities{
				CanImportBans:       true,
				CanSyncBansOutbound: true,
				CanImportModerators: true,
				CanVerifyOwnership:  true,
				CanFetchMetadata:    true,
			},
		},
		{
			name:     "kick",
			platform: "kick",
			want: PlatformModerationCapabilities{
				CanVerifyOwnership: true,
				CanFetchMetadata:   true,
			},
		},
		{
			name:     "youtube",
			platform: "youtube",
			want: PlatformModerationCapabilities{
				CanVerifyOwnership: true,
				CanFetchMetadata:   true,
			},
		},
		{
			name:     "tiktok",
			platform: "tiktok",
			want: PlatformModerationCapabilities{
				CanVerifyOwnership: true,
				CanFetchMetadata:   true,
			},
		},
		{
			name:     "unknown platform",
			platform: "x/twitter",
			want:     PlatformModerationCapabilities{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultPlatformModerationCapabilities(tt.platform)
			if got != tt.want {
				t.Fatalf("DefaultPlatformModerationCapabilities(%q) = %+v, want %+v", tt.platform, got, tt.want)
			}
		})
	}
}
