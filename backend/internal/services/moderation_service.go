package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
)

// Sentinel errors for moderation operations
var (
	ErrModerationPermissionDenied  = errors.New("insufficient permissions: user does not have moderation privileges")
	ErrModerationNotAuthorized     = errors.New("moderator is not authorized to moderate this community")
	ErrModerationCommunityNotFound = errors.New("community not found")
	ErrModerationUserNotFound      = errors.New("user not found")
	ErrModerationNotBanned         = errors.New("user is not banned from this community")
	ErrModerationCannotBanOwner    = errors.New("cannot ban the community owner")
)

// ModerationCommunityRepo defines the methods needed from CommunityRepository
type ModerationCommunityRepo interface {
	GetCommunityByID(ctx context.Context, id uuid.UUID) (*models.Community, error)
	GetMember(ctx context.Context, communityID, userID uuid.UUID) (*models.CommunityMember, error)
	IsBanned(ctx context.Context, communityID, userID uuid.UUID) (bool, error)
	BanMember(ctx context.Context, ban *models.CommunityBan) error
	UnbanMember(ctx context.Context, communityID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, communityID, userID uuid.UUID) error
	ListBans(ctx context.Context, communityID uuid.UUID, limit, offset int) ([]*models.CommunityBan, int, error)
	ListAllBans(ctx context.Context, limit, offset int) ([]*models.CommunityBan, int, error)
	GetBanByID(ctx context.Context, banID uuid.UUID) (*models.CommunityBan, error)
}

// ModerationUserRepo defines the methods needed from UserRepository
type ModerationUserRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// ModerationAuditRepo defines the methods needed from AuditLogRepository
type ModerationAuditRepo interface {
	Create(ctx context.Context, log *models.ModerationAuditLog) error
}

// ModerationService handles core moderation operations including banning, unbanning, and managing bans
type ModerationService struct {
	db            *pgxpool.Pool
	communityRepo ModerationCommunityRepo
	userRepo      ModerationUserRepo
	auditLogRepo  ModerationAuditRepo
}

// NewModerationService creates a new ModerationService
func NewModerationService(
	db *pgxpool.Pool,
	communityRepo *repository.CommunityRepository,
	userRepo *repository.UserRepository,
	auditLogRepo *repository.AuditLogRepository,
) *ModerationService {
	return &ModerationService{
		db:            db,
		communityRepo: communityRepo,
		userRepo:      userRepo,
		auditLogRepo:  auditLogRepo,
	}
}

// BanUser bans a user from a community with permission and scope validation
func (s *ModerationService) BanUser(ctx context.Context, communityID, moderatorID, targetUserID uuid.UUID, reason *string) error {
	start := time.Now()
	var err error
	var errorType string

	// Defer metrics recording
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.ModerationBanOperationDuration.WithLabelValues("ban").Observe(duration)
		
		if err == nil {
			metrics.ModerationBanOperationsTotal.WithLabelValues("ban", "success", "").Inc()
		} else {
			// Determine error type
			if errors.Is(err, ErrModerationPermissionDenied) || errors.Is(err, ErrModerationNotAuthorized) {
				errorType = "permission_denied"
			} else if errors.Is(err, ErrModerationCommunityNotFound) {
				errorType = "community_not_found"
			} else if errors.Is(err, ErrModerationCannotBanOwner) {
				errorType = "cannot_ban_owner"
			} else {
				errorType = "database_error"
			}
			metrics.ModerationBanOperationsTotal.WithLabelValues("ban", "failed", errorType).Inc()
		}
	}()

	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return fmt.Errorf("failed to get moderator: %w", err)
	}

	// Validate scope first for better error messages
	if err = s.validateModerationScope(moderator, communityID); err != nil {
		return err
	}

	// Validate permission
	if err = s.validateModerationPermission(ctx, moderator, communityID); err != nil {
		return err
	}

	// Check if target user is the community owner
	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		err = ErrModerationCommunityNotFound
		return err
	}
	if community.OwnerID == targetUserID {
		err = ErrModerationCannotBanOwner
		return err
	}

	// Remove user from community if they are a member
	// Ignore "not found" errors as the user may not be a member
	if removeErr := s.communityRepo.RemoveMember(ctx, communityID, targetUserID); removeErr != nil {
		// Log non-critical errors but continue with ban operation
		// Only return error if it's a critical database failure
		if removeErr.Error() != "member not found" && removeErr.Error() != "no rows affected" {
			// For now, log and continue as this is not critical for banning
		}
	}

	// Create ban record
	ban := &models.CommunityBan{
		ID:             uuid.New(),
		CommunityID:    communityID,
		BannedUserID:   targetUserID,
		BannedByUserID: &moderatorID,
		Reason:         reason,
		BannedAt:       time.Now(),
	}

	if err = s.communityRepo.BanMember(ctx, ban); err != nil {
		return fmt.Errorf("failed to create ban: %w", err)
	}

	// Log audit entry
	auditStart := time.Now()
	metadata := map[string]interface{}{
		"community_id":    communityID.String(),
		"banned_user_id":  targetUserID.String(),
		"moderator_scope": moderator.ModeratorScope,
	}
	if reason != nil {
		metadata["reason"] = *reason
	}

	auditLog := &models.ModerationAuditLog{
		Action:      "ban_user",
		EntityType:  "community_ban",
		EntityID:    ban.ID,
		ModeratorID: moderatorID,
		Reason:      reason,
		Metadata:    metadata,
	}

	if auditErr := s.auditLogRepo.Create(ctx, auditLog); auditErr != nil {
		metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "failed").Inc()
		metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())
		return fmt.Errorf("failed to create audit log: %w", auditErr)
	}
	metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "success").Inc()
	metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())

	return nil
}

// UnbanUser removes a ban from a user with permission and scope validation
func (s *ModerationService) UnbanUser(ctx context.Context, communityID, moderatorID, targetUserID uuid.UUID) error {
	start := time.Now()
	var err error
	var errorType string

	// Defer metrics recording
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.ModerationBanOperationDuration.WithLabelValues("unban").Observe(duration)
		
		if err == nil {
			metrics.ModerationBanOperationsTotal.WithLabelValues("unban", "success", "").Inc()
		} else {
			// Determine error type
			if errors.Is(err, ErrModerationPermissionDenied) || errors.Is(err, ErrModerationNotAuthorized) {
				errorType = "permission_denied"
			} else if errors.Is(err, ErrModerationNotBanned) {
				errorType = "not_banned"
			} else {
				errorType = "database_error"
			}
			metrics.ModerationBanOperationsTotal.WithLabelValues("unban", "failed", errorType).Inc()
		}
	}()

	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return fmt.Errorf("failed to get moderator: %w", err)
	}

	// Validate scope first for better error messages
	if err = s.validateModerationScope(moderator, communityID); err != nil {
		return err
	}

	// Validate permission
	if err = s.validateModerationPermission(ctx, moderator, communityID); err != nil {
		return err
	}

	// Check if user is actually banned
	isBanned, checkErr := s.communityRepo.IsBanned(ctx, communityID, targetUserID)
	if checkErr != nil {
		err = fmt.Errorf("failed to check ban status: %w", checkErr)
		return err
	}
	if !isBanned {
		err = ErrModerationNotBanned
		return err
	}

	// Remove ban
	if err = s.communityRepo.UnbanMember(ctx, communityID, targetUserID); err != nil {
		return fmt.Errorf("failed to remove ban: %w", err)
	}

	// Log audit entry
	auditStart := time.Now()
	metadata := map[string]interface{}{
		"community_id":    communityID.String(),
		"banned_user_id":  targetUserID.String(),
		"moderator_scope": moderator.ModeratorScope,
	}

	auditLog := &models.ModerationAuditLog{
		Action:      "unban_user",
		EntityType:  "community_ban",
		EntityID:    communityID, // Using community ID as entity since ban is deleted
		ModeratorID: moderatorID,
		Metadata:    metadata,
	}

	if auditErr := s.auditLogRepo.Create(ctx, auditLog); auditErr != nil {
		metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "failed").Inc()
		metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())
		return fmt.Errorf("failed to create audit log: %w", auditErr)
	}
	metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "success").Inc()
	metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())

	return nil
}

// GetBans retrieves bans for a community with filtering and pagination
func (s *ModerationService) GetBans(ctx context.Context, communityID, moderatorID uuid.UUID, page, limit int) ([]*models.CommunityBan, int, error) {
	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get moderator: %w", err)
	}

	// Validate scope first for better error messages
	if err := s.validateModerationScope(moderator, communityID); err != nil {
		return nil, 0, err
	}

	// Validate permission
	if err := s.validateModerationPermission(ctx, moderator, communityID); err != nil {
		return nil, 0, err
	}

	// Normalize pagination parameters to avoid unreasonable values
	if page < 1 {
		page = 1
	}
	const maxLimit = 100
	if limit <= 0 {
		limit = 20
	} else if limit > maxLimit {
		limit = maxLimit
	}

	offset := (page - 1) * limit
	return s.communityRepo.ListBans(ctx, communityID, limit, offset)
}

// GetAllBans retrieves all bans across all communities (admin/site moderator only)
func (s *ModerationService) GetAllBans(ctx context.Context, moderatorID uuid.UUID, page, limit int) ([]*models.CommunityBan, int, error) {
	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get moderator: %w", err)
	}

	// Only admins and site moderators can list all bans
	isAdmin := moderator.AccountType == models.AccountTypeAdmin || moderator.Role == models.RoleAdmin
	isSiteMod := moderator.AccountType == models.AccountTypeModerator && moderator.ModeratorScope == models.ModeratorScopeSite
	if !isAdmin && !isSiteMod {
		return nil, 0, ErrModerationPermissionDenied
	}

	// Normalize pagination parameters
	if page < 1 {
		page = 1
	}
	const maxLimit = 100
	if limit <= 0 {
		limit = 20
	} else if limit > maxLimit {
		limit = maxLimit
	}

	offset := (page - 1) * limit
	return s.communityRepo.ListAllBans(ctx, limit, offset)
}

// UpdateBan updates the reason for an existing ban
// TODO: This method currently deletes and recreates the ban record, which changes the ban ID
// and timestamp. A proper UpdateBanReason method should be added to CommunityRepository to
// preserve the original ban metadata while updating only the reason field.
func (s *ModerationService) UpdateBan(ctx context.Context, communityID, moderatorID, targetUserID uuid.UUID, newReason *string) error {
	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return fmt.Errorf("failed to get moderator: %w", err)
	}

	// Validate scope first for better error messages
	if err := s.validateModerationScope(moderator, communityID); err != nil {
		return err
	}

	// Validate permission
	if err := s.validateModerationPermission(ctx, moderator, communityID); err != nil {
		return err
	}

	// Check if user is actually banned
	isBanned, err := s.communityRepo.IsBanned(ctx, communityID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to check ban status: %w", err)
	}
	if !isBanned {
		return fmt.Errorf("user is not banned from this community")
	}

	// WORKAROUND: Since community_bans table doesn't have an update method in the repository,
	// we delete and recreate the ban. This changes the ban ID and timestamp, which is not ideal.
	// This technical debt should be addressed by adding a proper update method to the repository.
	if err := s.communityRepo.UnbanMember(ctx, communityID, targetUserID); err != nil {
		return fmt.Errorf("failed to remove old ban: %w", err)
	}

	ban := &models.CommunityBan{
		ID:             uuid.New(),
		CommunityID:    communityID,
		BannedUserID:   targetUserID,
		BannedByUserID: &moderatorID,
		Reason:         newReason,
		BannedAt:       time.Now(),
	}

	if err := s.communityRepo.BanMember(ctx, ban); err != nil {
		return fmt.Errorf("failed to create updated ban: %w", err)
	}

	// Log audit entry
	metadata := map[string]interface{}{
		"community_id":    communityID.String(),
		"banned_user_id":  targetUserID.String(),
		"moderator_scope": moderator.ModeratorScope,
		"action":          "update_ban_reason",
		"note":            "Ban timestamp was reset due to repository limitation",
	}
	if newReason != nil {
		metadata["new_reason"] = *newReason
	}

	auditLog := &models.ModerationAuditLog{
		Action:      "update_ban",
		EntityType:  "community_ban",
		EntityID:    ban.ID,
		ModeratorID: moderatorID,
		Reason:      newReason,
		Metadata:    metadata,
	}

	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// validateModerationPermission checks if a user has permission to perform moderation actions
func (s *ModerationService) validateModerationPermission(ctx context.Context, moderator *models.User, communityID uuid.UUID) error {
	start := time.Now()
	permissionType := "moderation"
	var result string
	var reason string

	defer func() {
		duration := time.Since(start).Seconds()
		metrics.ModerationPermissionCheckDuration.WithLabelValues(permissionType).Observe(duration)
		metrics.ModerationPermissionChecksTotal.WithLabelValues(permissionType, result).Inc()
		if result == "denied" {
			metrics.ModerationPermissionDenialsTotal.WithLabelValues(permissionType, reason).Inc()
		}
	}()

	// Site moderators (AccountType=moderator with ModeratorScope=site) can moderate anywhere
	if moderator.AccountType == models.AccountTypeModerator && moderator.ModeratorScope == models.ModeratorScopeSite {
		result = "allowed"
		return nil
	}

	// Admins can moderate anywhere
	if moderator.AccountType == models.AccountTypeAdmin || moderator.Role == models.RoleAdmin {
		result = "allowed"
		return nil
	}

	// Community moderators need to be checked for specific community authorization
	// The scope check handles which communities they can moderate
	// Here we just check if they have the moderator role in the community
	if moderator.AccountType == models.AccountTypeCommunityModerator {
		member, err := s.communityRepo.GetMember(ctx, communityID, moderator.ID)
		if err != nil {
			result = "denied"
			reason = "database_error"
			return fmt.Errorf("failed to get member: %w", err)
		}
		if member == nil {
			result = "denied"
			reason = "not_member"
			return ErrModerationPermissionDenied
		}
		if member.Role != models.CommunityRoleMod && member.Role != models.CommunityRoleAdmin {
			result = "denied"
			reason = "insufficient_permissions"
			return ErrModerationPermissionDenied
		}
		result = "allowed"
		return nil
	}

	result = "denied"
	reason = "insufficient_permissions"
	return ErrModerationPermissionDenied
}

// validateModerationScope checks if a community moderator is authorized for the specific community
func (s *ModerationService) validateModerationScope(moderator *models.User, communityID uuid.UUID) error {
	// Site moderators and admins have no scope restrictions
	if moderator.AccountType == models.AccountTypeModerator && moderator.ModeratorScope == models.ModeratorScopeSite {
		return nil
	}
	if moderator.AccountType == models.AccountTypeAdmin || moderator.Role == models.RoleAdmin {
		return nil
	}

	// Community moderators must have the community in their moderation channels
	if moderator.AccountType == models.AccountTypeCommunityModerator {
		if moderator.ModeratorScope != models.ModeratorScopeCommunity {
			return ErrModerationPermissionDenied
		}

		// Check if this community is in their authorized moderation scope
		for _, authorizedCommunityID := range moderator.ModerationChannels {
			if authorizedCommunityID == communityID {
				return nil
			}
		}
		return ErrModerationNotAuthorized
	}

	return nil
}

// HasModerationPermission checks if a user has permission to moderate a community
// without performing any ban-related queries
func (s *ModerationService) HasModerationPermission(ctx context.Context, communityID, moderatorID uuid.UUID) error {
	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return fmt.Errorf("failed to get moderator: %w", err)
	}

	// Validate permission
	if err := s.validateModerationPermission(ctx, moderator, communityID); err != nil {
		return err
	}

	// Validate scope for community moderators
	if err := s.validateModerationScope(moderator, communityID); err != nil {
		return err
	}

	return nil
}
