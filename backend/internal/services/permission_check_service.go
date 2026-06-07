package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// RedisCache defines the interface for Redis caching operations
type RedisCache interface {
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// PermissionCheckService handles permission checks with caching and scope validation
type PermissionCheckService struct {
	communityRepo ModerationCommunityRepo
	userRepo      ModerationUserRepo
	cache         RedisCache
}

// NewPermissionCheckService creates a new PermissionCheckService
func NewPermissionCheckService(
	communityRepo ModerationCommunityRepo,
	userRepo ModerationUserRepo,
	cache RedisCache,
) *PermissionCheckService {
	return &PermissionCheckService{
		communityRepo: communityRepo,
		userRepo:      userRepo,
		cache:         cache,
	}
}

// Cache key constants
const (
	KeyPermissionCanModerate = "permission:can_moderate:%s:%s" // userID, channelID
	KeyPermissionUserScope   = "permission:user_scope:%s"      // userID
)

// Cache TTL constants
const (
	TTLPermissionCheck = 5 * time.Minute
	TTLUserScope       = 10 * time.Minute
)

// PermissionDenialReason provides detailed information about why a permission was denied
type PermissionDenialReason struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (r *PermissionDenialReason) Error() string {
	return r.Message
}

// NewPermissionDenied creates a new permission denial error
func NewPermissionDenied(code, message string, details map[string]interface{}) *PermissionDenialReason {
	return &PermissionDenialReason{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// CanBan checks if an actor can ban a target user from a channel
// Returns nil if allowed, or a PermissionDenialReason if denied
func (s *PermissionCheckService) CanBan(ctx context.Context, actor *models.User, targetUserID, channelID uuid.UUID) error {
	// Check basic moderation permission first
	if err := s.CanModerate(ctx, actor, channelID); err != nil {
		return err
	}

	// Get the channel to check ownership
	channel, err := s.communityRepo.GetCommunityByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Cannot ban the channel owner
	if channel.OwnerID == targetUserID {
		return NewPermissionDenied(
			"CANNOT_BAN_OWNER",
			"cannot ban the channel owner",
			map[string]interface{}{
				"channel_id": channelID.String(),
				"owner_id":   channel.OwnerID.String(),
			},
		)
	}

	// Check if target is already banned
	isBanned, err := s.communityRepo.IsBanned(ctx, channelID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to check ban status: %w", err)
	}
	if isBanned {
		return NewPermissionDenied(
			"ALREADY_BANNED",
			"user is already banned from this channel",
			map[string]interface{}{
				"channel_id":     channelID.String(),
				"target_user_id": targetUserID.String(),
			},
		)
	}

	return nil
}

// CanUnban checks if an actor can unban a user by ban ID
// Returns nil if allowed, or a PermissionDenialReason if denied
func (s *PermissionCheckService) CanUnban(ctx context.Context, actor *models.User, banID uuid.UUID) error {
	// Get the ban record to extract channel and user information
	ban, err := s.getBanByID(ctx, banID)
	if err != nil {
		return fmt.Errorf("failed to get ban record: %w", err)
	}

	// Check if actor has moderation permission for this channel
	if err := s.CanModerate(ctx, actor, ban.CommunityID); err != nil {
		return err
	}

	// Verify the user is actually banned
	isBanned, err := s.communityRepo.IsBanned(ctx, ban.CommunityID, ban.BannedUserID)
	if err != nil {
		return fmt.Errorf("failed to verify ban status: %w", err)
	}
	if !isBanned {
		return NewPermissionDenied(
			"NOT_BANNED",
			"user is not currently banned from this channel",
			map[string]interface{}{
				"ban_id":     banID.String(),
				"channel_id": ban.CommunityID.String(),
				"user_id":    ban.BannedUserID.String(),
			},
		)
	}

	return nil
}

// CanModerate checks if an actor can moderate a specific channel
// Returns nil if allowed, or a PermissionDenialReason if denied
func (s *PermissionCheckService) CanModerate(ctx context.Context, actor *models.User, channelID uuid.UUID) error {
	// Check cache first
	cacheKey := fmt.Sprintf(KeyPermissionCanModerate, actor.ID.String(), channelID.String())
	var canModerate bool
	err := s.cache.GetJSON(ctx, cacheKey, &canModerate)
	if err == nil && canModerate {
		return nil
	}

	var permissionGranted bool

	// Site moderators can moderate anywhere
	if actor.AccountType == models.AccountTypeModerator && actor.ModeratorScope == models.ModeratorScopeSite {
		permissionGranted = true
	} else if actor.AccountType == models.AccountTypeAdmin || actor.Role == models.RoleAdmin {
		// Admins can moderate anywhere
		permissionGranted = true
	} else if actor.AccountType == models.AccountTypeCommunityModerator {
		// Community moderators need to be a mod or admin in the specific community
		member, err := s.communityRepo.GetMember(ctx, channelID, actor.ID)
		if err != nil {
			// Actual database error occurred
			return fmt.Errorf("failed to check community membership: %w", err)
		}
		if member == nil {
			return NewPermissionDenied(
				"NOT_A_MEMBER",
				"moderator is not a member of this channel",
				map[string]interface{}{
					"channel_id": channelID.String(),
					"user_id":    actor.ID.String(),
				},
			)
		}
		if member.Role != models.CommunityRoleMod && member.Role != models.CommunityRoleAdmin {
			return NewPermissionDenied(
				"INSUFFICIENT_PERMISSIONS",
				"insufficient permissions: must be a channel moderator or admin",
				map[string]interface{}{
					"channel_id":  channelID.String(),
					"user_id":     actor.ID.String(),
					"member_role": member.Role,
				},
			)
		}
		permissionGranted = true
	} else {
		// Regular users cannot moderate
		return NewPermissionDenied(
			"NO_MODERATION_PRIVILEGES",
			"user does not have moderation privileges",
			map[string]interface{}{
				"channel_id":   channelID.String(),
				"user_id":      actor.ID.String(),
				"account_type": actor.AccountType,
			},
		)
	}

	// Cache the result if permission was granted
	if permissionGranted {
		s.cache.SetJSON(ctx, cacheKey, true, TTLPermissionCheck)
	}

	return nil
}

// ValidateModeratorScope validates that a moderator has access to the specified channels
// Returns nil if valid, or a PermissionDenialReason if invalid
func (s *PermissionCheckService) ValidateModeratorScope(ctx context.Context, actor *models.User, channelIDs []uuid.UUID) error {
	// Site moderators and admins have no scope restrictions
	if actor.AccountType == models.AccountTypeModerator && actor.ModeratorScope == models.ModeratorScopeSite {
		return nil
	}
	if actor.AccountType == models.AccountTypeAdmin || actor.Role == models.RoleAdmin {
		return nil
	}

	// Community moderators must have the channels in their moderation scope
	if actor.AccountType == models.AccountTypeCommunityModerator {
		if actor.ModeratorScope != models.ModeratorScopeCommunity {
			return NewPermissionDenied(
				"INVALID_MODERATOR_CONFIG",
				"invalid moderator configuration: community moderator without community scope",
				map[string]interface{}{
					"user_id":         actor.ID.String(),
					"account_type":    actor.AccountType,
					"moderator_scope": actor.ModeratorScope,
				},
			)
		}

		// Check cache for user scope
		cacheKey := fmt.Sprintf(KeyPermissionUserScope, actor.ID.String())
		var cachedChannels []uuid.UUID
		err := s.cache.GetJSON(ctx, cacheKey, &cachedChannels)
		if err == nil {
			// Use cached data
			unauthorized := s.findUnauthorizedChannels(channelIDs, cachedChannels)
			if len(unauthorized) > 0 {
				return NewPermissionDenied(
					"SCOPE_VIOLATION",
					"moderator does not have access to all specified channels",
					map[string]interface{}{
						"user_id":               actor.ID.String(),
						"unauthorized_channels": unauthorized,
						"requested_channels":    channelIDs,
						"authorized_channels":   cachedChannels,
					},
				)
			}
			return nil
		}

		// Cache the moderation channels for future checks
		// Cache the moderation channels for future checks
		s.cache.SetJSON(ctx, cacheKey, actor.ModerationChannels, TTLUserScope)

		// Validate each requested channel is in the moderator's scope
		unauthorized := s.findUnauthorizedChannels(channelIDs, actor.ModerationChannels)
		if len(unauthorized) > 0 {
			return NewPermissionDenied(
				"SCOPE_VIOLATION",
				"moderator does not have access to all specified channels",
				map[string]interface{}{
					"user_id":               actor.ID.String(),
					"unauthorized_channels": unauthorized,
					"requested_channels":    channelIDs,
					"authorized_channels":   actor.ModerationChannels,
				},
			)
		}
		return nil
	}

	// Regular users have no moderation scope
	return NewPermissionDenied(
		"NO_MODERATION_PRIVILEGES",
		"user does not have moderation privileges",
		map[string]interface{}{
			"user_id":      actor.ID.String(),
			"account_type": actor.AccountType,
		},
	)
}

// findUnauthorizedChannels returns channel IDs that are not in the authorized list
func (s *PermissionCheckService) findUnauthorizedChannels(requested, authorized []uuid.UUID) []string {
	authorizedMap := make(map[uuid.UUID]bool)
	for _, id := range authorized {
		authorizedMap[id] = true
	}

	unauthorized := []string{}
	for _, id := range requested {
		if !authorizedMap[id] {
			unauthorized = append(unauthorized, id.String())
		}
	}
	return unauthorized
}

// getBanByID retrieves a ban record by ID using the repository
func (s *PermissionCheckService) getBanByID(ctx context.Context, banID uuid.UUID) (*models.CommunityBan, error) {
	return s.communityRepo.GetBanByID(ctx, banID)
}

// InvalidatePermissionCache clears permission cache for a specific user and channel
func (s *PermissionCheckService) InvalidatePermissionCache(ctx context.Context, userID, channelID uuid.UUID) error {
	cacheKey := fmt.Sprintf(KeyPermissionCanModerate, userID.String(), channelID.String())
	return s.cache.Delete(ctx, cacheKey)
}

// InvalidateUserScopeCache clears the scope cache for a specific user
func (s *PermissionCheckService) InvalidateUserScopeCache(ctx context.Context, userID uuid.UUID) error {
	cacheKey := fmt.Sprintf(KeyPermissionUserScope, userID.String())
	return s.cache.Delete(ctx, cacheKey)
}
