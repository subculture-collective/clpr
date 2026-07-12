package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{RoleUser, true},
		{RoleModerator, true},
		{RoleAdmin, true},
		{"invalid", false},
		{"", false},
		{"ADMIN", false}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if got := IsValidRole(tt.role); got != tt.valid {
				t.Errorf("IsValidRole(%q) = %v, want %v", tt.role, got, tt.valid)
			}
		})
	}
}

func TestUserHasRole(t *testing.T) {
	user := &User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     RoleModerator,
	}

	tests := []struct {
		role  string
		hasIt bool
	}{
		{RoleModerator, true},
		{RoleAdmin, false},
		{RoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if got := user.HasRole(tt.role); got != tt.hasIt {
				t.Errorf("HasRole(%q) = %v, want %v", tt.role, got, tt.hasIt)
			}
		})
	}
}

func TestUserHasAnyRole(t *testing.T) {
	user := &User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     RoleModerator,
	}

	tests := []struct {
		name   string
		roles  []string
		hasAny bool
	}{
		{"has moderator", []string{RoleModerator}, true},
		{"has moderator in list", []string{RoleAdmin, RoleModerator}, true},
		{"has no match", []string{RoleAdmin, RoleUser}, false},
		{"empty list", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.HasAnyRole(tt.roles...); got != tt.hasAny {
				t.Errorf("HasAnyRole(%v) = %v, want %v", tt.roles, got, tt.hasAny)
			}
		})
	}
}

func TestUserIsAdmin(t *testing.T) {
	tests := []struct {
		role    string
		isAdmin bool
	}{
		{RoleAdmin, true},
		{RoleModerator, false},
		{RoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			user := &User{
				ID:       uuid.New(),
				Username: "testuser",
				Role:     tt.role,
			}
			if got := user.IsAdmin(); got != tt.isAdmin {
				t.Errorf("IsAdmin() = %v, want %v for role %q", got, tt.isAdmin, tt.role)
			}
		})
	}
}

func TestUserIsModerator(t *testing.T) {
	tests := []struct {
		role        string
		isModerator bool
	}{
		{RoleAdmin, false},
		{RoleModerator, true},
		{RoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			user := &User{
				ID:       uuid.New(),
				Username: "testuser",
				Role:     tt.role,
			}
			if got := user.IsModerator(); got != tt.isModerator {
				t.Errorf("IsModerator() = %v, want %v for role %q", got, tt.isModerator, tt.role)
			}
		})
	}
}

func TestUserIsModeratorOrAdmin(t *testing.T) {
	tests := []struct {
		role               string
		isModeratorOrAdmin bool
	}{
		{RoleAdmin, true},
		{RoleModerator, true},
		{RoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			user := &User{
				ID:       uuid.New(),
				Username: "testuser",
				Role:     tt.role,
			}
			if got := user.IsModeratorOrAdmin(); got != tt.isModeratorOrAdmin {
				t.Errorf("IsModeratorOrAdmin() = %v, want %v for role %q", got, tt.isModeratorOrAdmin, tt.role)
			}
		})
	}
}

func TestRoleConstants(t *testing.T) {
	// Ensure role constants have expected values
	if RoleUser != "user" {
		t.Errorf("RoleUser = %q, want %q", RoleUser, "user")
	}
	if RoleModerator != "moderator" {
		t.Errorf("RoleModerator = %q, want %q", RoleModerator, "moderator")
	}
	if RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %q, want %q", RoleAdmin, "admin")
	}
}

func TestIsValidAccountType(t *testing.T) {
	tests := []struct {
		accountType string
		valid       bool
	}{
		{AccountTypeMember, true},
		{AccountTypeBroadcaster, true},
		{AccountTypeModerator, true},
		{AccountTypeCommunityModerator, true},
		{AccountTypeAdmin, true},
		{"invalid", false},
		{"", false},
		{"MEMBER", false}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.accountType, func(t *testing.T) {
			if got := IsValidAccountType(tt.accountType); got != tt.valid {
				t.Errorf("IsValidAccountType(%q) = %v, want %v", tt.accountType, got, tt.valid)
			}
		})
	}
}

func TestGetAccountTypePermissions(t *testing.T) {
	tests := []struct {
		name             string
		accountType      string
		expectedCount    int
		mustHavePerms    []string
		mustNotHavePerms []string
	}{
		{
			name:          "member permissions",
			accountType:   AccountTypeMember,
			expectedCount: 4,
			mustHavePerms: []string{
				PermissionCreateSubmission,
				PermissionCreateComment,
				PermissionCreateVote,
				PermissionCreateFollow,
			},
			mustNotHavePerms: []string{
				PermissionViewBroadcasterAnalytics,
				PermissionModerateContent,
			},
		},
		{
			name:          "broadcaster permissions",
			accountType:   AccountTypeBroadcaster,
			expectedCount: 6,
			mustHavePerms: []string{
				PermissionCreateSubmission,
				PermissionViewBroadcasterAnalytics,
				PermissionClaimBroadcasterProfile,
			},
			mustNotHavePerms: []string{
				PermissionModerateContent,
			},
		},
		{
			name:          "moderator permissions",
			accountType:   AccountTypeModerator,
			expectedCount: 9,
			mustHavePerms: []string{
				PermissionCreateSubmission,
				PermissionViewBroadcasterAnalytics,
				PermissionModerateContent,
				PermissionModerateUsers,
				PermissionCreateDiscoveryLists,
			},
			mustNotHavePerms: []string{
				PermissionManageUsers,
				PermissionManageSystem,
			},
		},
		{
			name:          "community moderator permissions",
			accountType:   AccountTypeCommunityModerator,
			expectedCount: 4,
			mustHavePerms: []string{
				PermissionCommunityModerate,
				PermissionModerateUsers,
				PermissionViewChannelAnalytics,
				PermissionManageModerators,
			},
			mustNotHavePerms: []string{
				PermissionCreateSubmission,
				PermissionCreateComment,
				PermissionViewBroadcasterAnalytics,
				PermissionModerateContent,
				PermissionCreateDiscoveryLists,
				PermissionManageUsers,
				PermissionManageSystem,
			},
		},
		{
			name:          "admin permissions",
			accountType:   AccountTypeAdmin,
			expectedCount: 16,
			mustHavePerms: []string{
				PermissionCreateSubmission,
				PermissionModerateContent,
				PermissionManageUsers,
				PermissionManageSystem,
				PermissionViewAnalyticsDashboard,
				PermissionCommunityModerate,
				PermissionViewChannelAnalytics,
				PermissionManageModerators,
			},
		},
		{
			name:          "unknown defaults to member",
			accountType:   "unknown",
			expectedCount: 4,
			mustHavePerms: []string{
				PermissionCreateSubmission,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := GetAccountTypePermissions(tt.accountType)

			if len(perms) != tt.expectedCount {
				t.Errorf("Expected %d permissions, got %d", tt.expectedCount, len(perms))
			}

			permSet := make(map[string]bool)
			for _, p := range perms {
				permSet[p] = true
			}

			for _, mustHave := range tt.mustHavePerms {
				if !permSet[mustHave] {
					t.Errorf("Expected permission %q to be present", mustHave)
				}
			}

			for _, mustNotHave := range tt.mustNotHavePerms {
				if permSet[mustNotHave] {
					t.Errorf("Expected permission %q to NOT be present", mustNotHave)
				}
			}
		})
	}
}

func TestUserCan(t *testing.T) {
	tests := []struct {
		name        string
		accountType string
		role        string
		permission  string
		expected    bool
	}{
		// Member tests
		{
			name:        "member can create submission",
			accountType: AccountTypeMember,
			role:        RoleUser,
			permission:  PermissionCreateSubmission,
			expected:    true,
		},
		{
			name:        "member cannot view broadcaster analytics",
			accountType: AccountTypeMember,
			role:        RoleUser,
			permission:  PermissionViewBroadcasterAnalytics,
			expected:    false,
		},
		{
			name:        "member cannot access community:moderate permission",
			accountType: AccountTypeMember,
			role:        RoleUser,
			permission:  PermissionCommunityModerate,
			expected:    false,
		},
		// Broadcaster tests
		{
			name:        "broadcaster can view analytics",
			accountType: AccountTypeBroadcaster,
			role:        RoleUser,
			permission:  PermissionViewBroadcasterAnalytics,
			expected:    true,
		},
		{
			name:        "broadcaster cannot moderate",
			accountType: AccountTypeBroadcaster,
			role:        RoleUser,
			permission:  PermissionModerateContent,
			expected:    false,
		},
		// Moderator tests
		{
			name:        "moderator can moderate content",
			accountType: AccountTypeModerator,
			role:        RoleModerator,
			permission:  PermissionModerateContent,
			expected:    true,
		},
		{
			name:        "moderator cannot manage system",
			accountType: AccountTypeModerator,
			role:        RoleModerator,
			permission:  PermissionManageSystem,
			expected:    false,
		},
		// Community Moderator tests
		{
			name:        "community moderator can use community:moderate permission",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionCommunityModerate,
			expected:    true,
		},
		{
			name:        "community moderator can moderate users",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionModerateUsers,
			expected:    true,
		},
		{
			name:        "community moderator can view channel analytics",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionViewChannelAnalytics,
			expected:    true,
		},
		{
			name:        "community moderator can manage moderators",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionManageModerators,
			expected:    true,
		},
		{
			name:        "community moderator cannot create submissions",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionCreateSubmission,
			expected:    false,
		},
		{
			name:        "community moderator cannot moderate content",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionModerateContent,
			expected:    false,
		},
		{
			name:        "community moderator cannot create discovery lists",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionCreateDiscoveryLists,
			expected:    false,
		},
		{
			name:        "community moderator cannot manage users",
			accountType: AccountTypeCommunityModerator,
			role:        RoleUser,
			permission:  PermissionManageUsers,
			expected:    false,
		},
		// Admin tests
		{
			name:        "admin can do everything - by role",
			accountType: AccountTypeMember,
			role:        RoleAdmin,
			permission:  PermissionManageSystem,
			expected:    true,
		},
		{
			name:        "admin can do everything - by account type",
			accountType: AccountTypeAdmin,
			role:        RoleUser,
			permission:  PermissionManageSystem,
			expected:    true,
		},
		{
			name:        "admin has community:moderate permission",
			accountType: AccountTypeAdmin,
			role:        RoleAdmin,
			permission:  PermissionCommunityModerate,
			expected:    true,
		},
		{
			name:        "admin has view:channel_analytics permission",
			accountType: AccountTypeAdmin,
			role:        RoleAdmin,
			permission:  PermissionViewChannelAnalytics,
			expected:    true,
		},
		{
			name:        "admin has manage:moderators permission",
			accountType: AccountTypeAdmin,
			role:        RoleAdmin,
			permission:  PermissionManageModerators,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:          uuid.New(),
				Username:    "testuser",
				Role:        tt.role,
				AccountType: tt.accountType,
			}

			if got := user.Can(tt.permission); got != tt.expected {
				t.Errorf("Can(%q) = %v, want %v", tt.permission, got, tt.expected)
			}
		})
	}
}

func TestUserGetAccountType(t *testing.T) {
	tests := []struct {
		name        string
		accountType string
		expected    string
	}{
		{
			name:        "has account type set",
			accountType: AccountTypeBroadcaster,
			expected:    AccountTypeBroadcaster,
		},
		{
			name:        "empty defaults to member",
			accountType: "",
			expected:    AccountTypeMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:          uuid.New(),
				Username:    "testuser",
				AccountType: tt.accountType,
			}

			if got := user.GetAccountType(); got != tt.expected {
				t.Errorf("GetAccountType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestUserGetPermissions(t *testing.T) {
	user := &User{
		ID:          uuid.New(),
		Username:    "testuser",
		AccountType: AccountTypeBroadcaster,
	}

	perms := user.GetPermissions()
	if len(perms) != 6 {
		t.Errorf("Expected 6 permissions for broadcaster, got %d", len(perms))
	}

	// Check for at least one broadcaster permission
	found := false
	for _, p := range perms {
		if p == PermissionViewBroadcasterAnalytics {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected broadcaster permissions to include view:broadcaster_analytics")
	}
}

func TestAccountTypeConstants(t *testing.T) {
	// Ensure account type constants have expected values
	if AccountTypeMember != "member" {
		t.Errorf("AccountTypeMember = %q, want %q", AccountTypeMember, "member")
	}
	if AccountTypeBroadcaster != "broadcaster" {
		t.Errorf("AccountTypeBroadcaster = %q, want %q", AccountTypeBroadcaster, "broadcaster")
	}
	if AccountTypeModerator != "moderator" {
		t.Errorf("AccountTypeModerator = %q, want %q", AccountTypeModerator, "moderator")
	}
	if AccountTypeCommunityModerator != "community_moderator" {
		t.Errorf("AccountTypeCommunityModerator = %q, want %q", AccountTypeCommunityModerator, "community_moderator")
	}
	if AccountTypeAdmin != "admin" {
		t.Errorf("AccountTypeAdmin = %q, want %q", AccountTypeAdmin, "admin")
	}
}

func TestModeratorScopeConstants(t *testing.T) {
	// Ensure moderator scope constants have expected values
	if ModeratorScopeSite != "site" {
		t.Errorf("ModeratorScopeSite = %q, want %q", ModeratorScopeSite, "site")
	}
	if ModeratorScopeCommunity != "community" {
		t.Errorf("ModeratorScopeCommunity = %q, want %q", ModeratorScopeCommunity, "community")
	}
}

func TestUserIsValidModerator(t *testing.T) {
	tests := []struct {
		name               string
		moderatorScope     string
		moderationChannels []uuid.UUID
		expected           bool
	}{
		{
			name:               "not a moderator - empty scope",
			moderatorScope:     "",
			moderationChannels: nil,
			expected:           true,
		},
		{
			name:               "site moderator with empty channel list",
			moderatorScope:     ModeratorScopeSite,
			moderationChannels: []uuid.UUID{},
			expected:           true,
		},
		{
			name:               "site moderator with nil channel list",
			moderatorScope:     ModeratorScopeSite,
			moderationChannels: nil,
			expected:           true,
		},
		{
			name:               "site moderator with channels - invalid",
			moderatorScope:     ModeratorScopeSite,
			moderationChannels: []uuid.UUID{uuid.New()},
			expected:           false,
		},
		{
			name:               "community moderator with one channel",
			moderatorScope:     ModeratorScopeCommunity,
			moderationChannels: []uuid.UUID{uuid.New()},
			expected:           true,
		},
		{
			name:               "community moderator with multiple channels",
			moderatorScope:     ModeratorScopeCommunity,
			moderationChannels: []uuid.UUID{uuid.New(), uuid.New()},
			expected:           true,
		},
		{
			name:               "community moderator without channels - invalid",
			moderatorScope:     ModeratorScopeCommunity,
			moderationChannels: []uuid.UUID{},
			expected:           false,
		},
		{
			name:               "community moderator with nil channels - invalid",
			moderatorScope:     ModeratorScopeCommunity,
			moderationChannels: nil,
			expected:           false,
		},
		{
			name:               "invalid scope",
			moderatorScope:     "invalid",
			moderationChannels: []uuid.UUID{},
			expected:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:                 uuid.New(),
				Username:           "testuser",
				ModeratorScope:     tt.moderatorScope,
				ModerationChannels: tt.moderationChannels,
			}

			if got := user.IsValidModerator(); got != tt.expected {
				t.Errorf("IsValidModerator() = %v, want %v", got, tt.expected)
			}
		})
	}
}
