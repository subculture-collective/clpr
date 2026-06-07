package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockRedisClient is a mock implementation of Redis client for permission tests
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockRedisClient) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

// Delete is the preferred method name for cache deletion
func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// TestCanBan_AdminCanBanAnyone tests that admins can ban any user in any channel
func TestCanBan_AdminCanBanAnyone(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	// Create admin user
	adminID := uuid.New()
	admin := &models.User{
		ID:          adminID,
		Username:    "admin",
		Role:        models.RoleAdmin,
		AccountType: models.AccountTypeAdmin,
	}

	channelID := uuid.New()
	targetUserID := uuid.New()
	ownerID := uuid.New()

	// Mock channel
	channel := &models.Community{
		ID:      channelID,
		OwnerID: ownerID,
	}

	// Setup mocks - admin bypasses cache check
	mockRedis.On("GetJSON", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
	mockRedis.On("SetJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, channelID).Return(channel, nil)
	mockCommunityRepo.On("IsBanned", ctx, channelID, targetUserID).Return(false, nil)

	// Test
	err := service.CanBan(ctx, admin, targetUserID, channelID)

	// Assert
	assert.NoError(t, err)
	mockCommunityRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// TestCanBan_SiteModCanBanAcrossChannels tests that site moderators can ban across all channels
func TestCanBan_SiteModCanBanAcrossChannels(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	// Create site moderator
	siteModID := uuid.New()
	siteMod := &models.User{
		ID:             siteModID,
		Username:       "site_mod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	channelID := uuid.New()
	targetUserID := uuid.New()
	ownerID := uuid.New()

	// Mock channel
	channel := &models.Community{
		ID:      channelID,
		OwnerID: ownerID,
	}

	// Setup mocks - site mod bypasses cache and member checks
	mockRedis.On("GetJSON", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
	mockRedis.On("SetJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, channelID).Return(channel, nil)
	mockCommunityRepo.On("IsBanned", ctx, channelID, targetUserID).Return(false, nil)

	// Test
	err := service.CanBan(ctx, siteMod, targetUserID, channelID)

	// Assert
	assert.NoError(t, err)
	mockCommunityRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// TestCanBan_CommunityModScopeLimited tests that community mods are limited to their assigned channels
func TestCanBan_CommunityModScopeLimited(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	communityModID := uuid.New()
	assignedChannelID := uuid.New()
	otherChannelID := uuid.New()
	targetUserID := uuid.New()

	// Create community moderator with access to only one channel
	communityMod := &models.User{
		ID:                 communityModID,
		Username:           "community_mod",
		AccountType:        models.AccountTypeCommunityModerator,
		ModeratorScope:     models.ModeratorScopeCommunity,
		ModerationChannels: []uuid.UUID{assignedChannelID},
	}

	// Test 1: Can ban in assigned channel
	t.Run("CanBanInAssignedChannel", func(t *testing.T) {
		channel := &models.Community{
			ID:      assignedChannelID,
			OwnerID: uuid.New(),
		}

		member := &models.CommunityMember{
			UserID:      communityModID,
			CommunityID: assignedChannelID,
			Role:        models.CommunityRoleMod,
		}

		mockRedis.On("GetJSON", ctx, mock.MatchedBy(func(key string) bool {
			return key != ""
		}), mock.Anything).Return(assert.AnError).Once()
		mockRedis.On("SetJSON", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockCommunityRepo.On("GetMember", ctx, assignedChannelID, communityModID).Return(member, nil)
		mockCommunityRepo.On("GetCommunityByID", ctx, assignedChannelID).Return(channel, nil)
		mockCommunityRepo.On("IsBanned", ctx, assignedChannelID, targetUserID).Return(false, nil)

		err := service.CanBan(ctx, communityMod, targetUserID, assignedChannelID)
		assert.NoError(t, err)
	})

	// Test 2: Cannot ban in unassigned channel
	t.Run("CannotBanInUnassignedChannel", func(t *testing.T) {
		mockRedis.On("GetJSON", ctx, mock.MatchedBy(func(key string) bool {
			return key != ""
		}), mock.Anything).Return(assert.AnError).Once()
		mockCommunityRepo.On("GetMember", ctx, otherChannelID, communityModID).Return(nil, nil)

		err := service.CanBan(ctx, communityMod, targetUserID, otherChannelID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a member")
	})
}

// TestCanBan_RegularUserDenied tests that regular users cannot ban anyone
func TestCanBan_RegularUserDenied(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	// Create regular user
	regularUserID := uuid.New()
	regularUser := &models.User{
		ID:          regularUserID,
		Username:    "regular_user",
		AccountType: models.AccountTypeMember,
	}

	channelID := uuid.New()
	targetUserID := uuid.New()

	// Setup mocks - cache miss
	mockRedis.On("GetJSON", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	// Test
	err := service.CanBan(ctx, regularUser, targetUserID, channelID)

	// Assert
	assert.Error(t, err)
	denialErr, ok := err.(*PermissionDenialReason)
	assert.True(t, ok)
	assert.Equal(t, "NO_MODERATION_PRIVILEGES", denialErr.Code)
	assert.Contains(t, denialErr.Message, "does not have moderation privileges")
}

// TestCanBan_CannotBanOwner tests that no one can ban the channel owner
func TestCanBan_CannotBanOwner(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	adminID := uuid.New()
	admin := &models.User{
		ID:          adminID,
		Username:    "admin",
		Role:        models.RoleAdmin,
		AccountType: models.AccountTypeAdmin,
	}

	channelID := uuid.New()
	ownerID := uuid.New()

	channel := &models.Community{
		ID:      channelID,
		OwnerID: ownerID,
	}

	// Setup mocks
	mockRedis.On("GetJSON", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
	mockRedis.On("SetJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, channelID).Return(channel, nil)

	// Test - trying to ban the owner
	err := service.CanBan(ctx, admin, ownerID, channelID)

	// Assert
	assert.Error(t, err)
	denialErr, ok := err.(*PermissionDenialReason)
	assert.True(t, ok)
	assert.Equal(t, "CANNOT_BAN_OWNER", denialErr.Code)
}

// TestCanBan_AlreadyBanned tests that trying to ban an already banned user returns error
func TestCanBan_AlreadyBanned(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	adminID := uuid.New()
	admin := &models.User{
		ID:          adminID,
		Username:    "admin",
		Role:        models.RoleAdmin,
		AccountType: models.AccountTypeAdmin,
	}

	channelID := uuid.New()
	targetUserID := uuid.New()
	ownerID := uuid.New()

	channel := &models.Community{
		ID:      channelID,
		OwnerID: ownerID,
	}

	// Setup mocks
	mockRedis.On("GetJSON", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
	mockRedis.On("SetJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, channelID).Return(channel, nil)
	mockCommunityRepo.On("IsBanned", ctx, channelID, targetUserID).Return(true, nil)

	// Test
	err := service.CanBan(ctx, admin, targetUserID, channelID)

	// Assert
	assert.Error(t, err)
	denialErr, ok := err.(*PermissionDenialReason)
	assert.True(t, ok)
	assert.Equal(t, "ALREADY_BANNED", denialErr.Code)
}

// TestCanUnban_AdminCanUnban tests that admins can unban any user
func TestCanUnban_AdminCanUnban(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	adminID := uuid.New()
	admin := &models.User{
		ID:          adminID,
		Username:    "admin",
		Role:        models.RoleAdmin,
		AccountType: models.AccountTypeAdmin,
	}

	banID := uuid.New()
	channelID := uuid.New()
	bannedUserID := uuid.New()

	ban := &models.CommunityBan{
		ID:           banID,
		CommunityID:  channelID,
		BannedUserID: bannedUserID,
	}

	// Setup mocks
	mockRedis.On("GetJSON", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
	mockRedis.On("SetJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCommunityRepo.On("GetBanByID", ctx, banID).Return(ban, nil)
	mockCommunityRepo.On("IsBanned", ctx, channelID, bannedUserID).Return(true, nil)

	// Test
	err := service.CanUnban(ctx, admin, banID)

	// Assert
	assert.NoError(t, err)
	mockCommunityRepo.AssertExpectations(t)
}

// TestCanModerate_CachingWorks tests that permission checks are properly cached
func TestCanModerate_CachingWorks(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	adminID := uuid.New()
	admin := &models.User{
		ID:          adminID,
		Username:    "admin",
		Role:        models.RoleAdmin,
		AccountType: models.AccountTypeAdmin,
	}

	channelID := uuid.New()

	// First call - cache miss, should set cache
	mockRedis.On("GetJSON", ctx, mock.Anything, mock.Anything).Return(assert.AnError).Once()
	mockRedis.On("SetJSON", ctx, mock.Anything, true, TTLPermissionCheck).Return(nil).Once()

	err := service.CanModerate(ctx, admin, channelID)
	assert.NoError(t, err)

	// Second call - cache hit
	mockRedis.On("GetJSON", ctx, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// Simulate cache hit by setting the dest to true
		dest := args.Get(2).(*bool)
		*dest = true
	}).Return(nil).Once()

	err = service.CanModerate(ctx, admin, channelID)
	assert.NoError(t, err)

	mockRedis.AssertExpectations(t)
}

// TestValidateModeratorScope_SiteModNoRestrictions tests site mods have no scope restrictions
func TestValidateModeratorScope_SiteModNoRestrictions(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	siteModID := uuid.New()
	siteMod := &models.User{
		ID:             siteModID,
		Username:       "site_mod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	// Test with multiple random channels
	channels := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Test
	err := service.ValidateModeratorScope(ctx, siteMod, channels)

	// Assert
	assert.NoError(t, err)
}

// TestValidateModeratorScope_CommunityModLimited tests community mods are limited to assigned channels
func TestValidateModeratorScope_CommunityModLimited(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	communityModID := uuid.New()
	assignedChannel1 := uuid.New()
	assignedChannel2 := uuid.New()
	unauthorizedChannel := uuid.New()

	communityMod := &models.User{
		ID:                 communityModID,
		Username:           "community_mod",
		AccountType:        models.AccountTypeCommunityModerator,
		ModeratorScope:     models.ModeratorScopeCommunity,
		ModerationChannels: []uuid.UUID{assignedChannel1, assignedChannel2},
	}

	// Test 1: Valid channels - should succeed
	t.Run("ValidChannels", func(t *testing.T) {
		mockRedis.On("GetJSON", ctx, mock.Anything, mock.Anything).Return(assert.AnError).Once()
		mockRedis.On("SetJSON", ctx, mock.Anything, mock.Anything, TTLUserScope).Return(nil).Once()

		err := service.ValidateModeratorScope(ctx, communityMod, []uuid.UUID{assignedChannel1})
		assert.NoError(t, err)
	})

	// Test 2: Invalid channel - should fail
	t.Run("InvalidChannel", func(t *testing.T) {
		mockRedis.On("GetJSON", ctx, mock.Anything, mock.Anything).Return(assert.AnError).Once()
		mockRedis.On("SetJSON", ctx, mock.Anything, mock.Anything, TTLUserScope).Return(nil).Once()

		err := service.ValidateModeratorScope(ctx, communityMod, []uuid.UUID{unauthorizedChannel})
		assert.Error(t, err)
		denialErr, ok := err.(*PermissionDenialReason)
		assert.True(t, ok)
		assert.Equal(t, "SCOPE_VIOLATION", denialErr.Code)
	})

	// Test 3: Mix of valid and invalid - should fail
	t.Run("MixedChannels", func(t *testing.T) {
		mockRedis.On("GetJSON", ctx, mock.Anything, mock.Anything).Return(assert.AnError).Once()
		mockRedis.On("SetJSON", ctx, mock.Anything, mock.Anything, TTLUserScope).Return(nil).Once()

		err := service.ValidateModeratorScope(ctx, communityMod, []uuid.UUID{assignedChannel1, unauthorizedChannel})
		assert.Error(t, err)
		denialErr, ok := err.(*PermissionDenialReason)
		assert.True(t, ok)
		assert.Equal(t, "SCOPE_VIOLATION", denialErr.Code)
	})
}

// TestValidateModeratorScope_InvalidConfig tests that invalid moderator config is caught
func TestValidateModeratorScope_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	// Community moderator with wrong scope
	invalidMod := &models.User{
		ID:             uuid.New(),
		Username:       "invalid_mod",
		AccountType:    models.AccountTypeCommunityModerator,
		ModeratorScope: models.ModeratorScopeSite, // Wrong scope for community mod!
	}

	// Test
	err := service.ValidateModeratorScope(ctx, invalidMod, []uuid.UUID{uuid.New()})

	// Assert
	assert.Error(t, err)
	denialErr, ok := err.(*PermissionDenialReason)
	assert.True(t, ok)
	assert.Equal(t, "INVALID_MODERATOR_CONFIG", denialErr.Code)
}

// TestInvalidatePermissionCache tests that cache invalidation works
func TestInvalidatePermissionCache(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	userID := uuid.New()
	channelID := uuid.New()

	// Setup mock
	expectedKey := fmt.Sprintf(KeyPermissionCanModerate, userID.String(), channelID.String())
	mockRedis.On("Delete", ctx, expectedKey).Return(nil)

	// Test
	err := service.InvalidatePermissionCache(ctx, userID, channelID)

	// Assert
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

// TestInvalidateUserScopeCache tests that user scope cache invalidation works
func TestInvalidateUserScopeCache(t *testing.T) {
	ctx := context.Background()
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockRedis := new(MockRedisClient)

	service := NewPermissionCheckService(mockCommunityRepo, mockUserRepo, mockRedis)

	userID := uuid.New()

	// Setup mock
	expectedKey := fmt.Sprintf(KeyPermissionUserScope, userID.String())
	mockRedis.On("Delete", ctx, expectedKey).Return(nil)

	// Test
	err := service.InvalidateUserScopeCache(ctx, userID)

	// Assert
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}
