package handlers

import "testing"

// Regression: BanUser/MuteUser/TimeoutUser only required the caller to be
// staff, so a moderator could ban the channel owner or an admin.
func TestCanModerateChatTarget(t *testing.T) {
	tests := []struct {
		actor, target string
		want          bool
	}{
		{"owner", "admin", true},
		{"owner", "moderator", true},
		{"owner", "member", true},
		{"owner", "", true},
		{"admin", "moderator", true},
		{"admin", "member", true},
		{"admin", "owner", false},
		{"admin", "admin", false},
		{"moderator", "member", true},
		{"moderator", "", true},
		{"moderator", "moderator", false},
		{"moderator", "admin", false},
		{"moderator", "owner", false},
		{"member", "", false},
	}
	for _, tt := range tests {
		if got := canModerateChatTarget(tt.actor, tt.target); got != tt.want {
			t.Errorf("canModerateChatTarget(%q, %q) = %v, want %v", tt.actor, tt.target, got, tt.want)
		}
	}
}
