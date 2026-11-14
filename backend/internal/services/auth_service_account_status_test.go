package services

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestActivatesOnTwitchLogin(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"unclaimed", true},
		{"active", false},
		{"pending", false},
		{"merged", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := activatesOnTwitchLogin(&models.User{AccountStatus: tt.status}); got != tt.want {
			t.Errorf("activatesOnTwitchLogin(status %q) = %v, want %v", tt.status, got, tt.want)
		}
	}
}
