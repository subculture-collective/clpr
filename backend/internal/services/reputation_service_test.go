package services

import (
	"testing"
)

func TestGetUserRank(t *testing.T) {
	tests := []struct {
		name  string
		karma int
		want  string
	}{
		{"Newcomer with 0 karma", 0, "Newcomer"},
		{"Newcomer with 50 karma", 50, "Newcomer"},
		{"Member with 100 karma", 100, "Member"},
		{"Member with 300 karma", 300, "Member"},
		{"Regular with 500 karma", 500, "Regular"},
		{"Regular with 800 karma", 800, "Regular"},
		{"Contributor with 1000 karma", 1000, "Contributor"},
		{"Contributor with 3000 karma", 3000, "Contributor"},
		{"Veteran with 5000 karma", 5000, "Veteran"},
		{"Veteran with 8000 karma", 8000, "Veteran"},
		{"Legend with 10000 karma", 10000, "Legend"},
		{"Legend with 50000 karma", 50000, "Legend"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUserRank(tt.karma)
			if got != tt.want {
				t.Errorf("GetUserRank(%d) = %v, want %v", tt.karma, got, tt.want)
			}
		})
	}
}

func TestIsValidBadge(t *testing.T) {
	tests := []struct {
		name    string
		badgeID string
		want    bool
	}{
		{"Valid badge - veteran", "veteran", true},
		{"Valid badge - influencer", "influencer", true},
		{"Valid badge - moderator", "moderator", true},
		{"Invalid badge", "invalid_badge", false},
		{"Empty badge ID", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidBadge(tt.badgeID)
			if got != tt.want {
				t.Errorf("IsValidBadge(%q) = %v, want %v", tt.badgeID, got, tt.want)
			}
		})
	}
}

func TestGetBadgeDefinition(t *testing.T) {
	tests := []struct {
		name     string
		badgeID  string
		wantErr  bool
		wantName string
	}{
		{"Get veteran badge", "veteran", false, "Veteran"},
		{"Get influencer badge", "influencer", false, "Influencer"},
		{"Get invalid badge", "invalid", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBadgeDefinition(tt.badgeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBadgeDefinition(%q) error = %v, wantErr %v", tt.badgeID, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.wantName {
				t.Errorf("GetBadgeDefinition(%q).Name = %v, want %v", tt.badgeID, got.Name, tt.wantName)
			}
		})
	}
}

func TestGetAllBadgeDefinitions(t *testing.T) {
	badges := GetAllBadgeDefinitions()

	if len(badges) == 0 {
		t.Error("GetAllBadgeDefinitions() returned empty slice, want at least one badge")
	}

	// Verify all badges have required fields
	for _, badge := range badges {
		if badge.ID == "" {
			t.Error("Badge has empty ID")
		}
		if badge.Name == "" {
			t.Errorf("Badge %s has empty Name", badge.ID)
		}
		if badge.Category == "" {
			t.Errorf("Badge %s has empty Category", badge.ID)
		}
	}
}

func TestCanUserPerformAction(t *testing.T) {
	tests := []struct {
		name   string
		karma  int
		action string
		want   bool
	}{
		{"Can create tags with 10 karma", 10, "create_tags", true},
		{"Cannot create tags with 5 karma", 5, "create_tags", false},
		{"Can report content with 50 karma", 50, "report_content", true},
		{"Cannot report content with 30 karma", 30, "report_content", false},
		{"Can submit clips with zero uppies", 0, "submit_clips", true},
		{"Can submit clips with negative uppies", -80, "submit_clips", true},
		{"Can nominate featured with 500 karma", 500, "nominate_featured", true},
		{"Cannot nominate featured with 400 karma", 400, "nominate_featured", false},
		{"Unknown action always allowed", 0, "unknown_action", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanUserPerformAction(tt.karma, tt.action)
			if got != tt.want {
				t.Errorf("CanUserPerformAction(%d, %q) = %v, want %v", tt.karma, tt.action, got, tt.want)
			}
		})
	}
}
