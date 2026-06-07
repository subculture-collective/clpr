package services

import "strings"

// PlatformModerationCapabilities describes creator moderation capabilities supported by a platform.
type PlatformModerationCapabilities struct {
	CanImportBans       bool
	CanSyncBansOutbound bool
	CanImportModerators bool
	CanVerifyOwnership  bool
	CanFetchMetadata    bool
}

// DefaultPlatformModerationCapabilities returns the conservative moderation capability defaults for a platform.
func DefaultPlatformModerationCapabilities(platform string) PlatformModerationCapabilities {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "twitch":
		return PlatformModerationCapabilities{
			CanImportBans:       true,
			CanSyncBansOutbound: true,
			CanImportModerators: true,
			CanVerifyOwnership:  true,
			CanFetchMetadata:    true,
		}
	case "kick":
		return PlatformModerationCapabilities{
			CanVerifyOwnership: true,
			CanFetchMetadata:   true,
		}
	case "youtube":
		return PlatformModerationCapabilities{
			CanVerifyOwnership: true,
			CanFetchMetadata:   true,
		}
	case "tiktok":
		return PlatformModerationCapabilities{
			CanVerifyOwnership: true,
			CanFetchMetadata:   true,
		}
	default:
		return PlatformModerationCapabilities{}
	}
}
