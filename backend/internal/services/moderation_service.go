package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors for moderation operations
var (
	ErrModerationPermissionDenied  = errors.New("insufficient permissions: user does not have moderation privileges")
	ErrModerationNotAuthorized     = errors.New("moderator is not authorized to moderate this community")
	ErrModerationCommunityNotFound = errors.New("community not found")
	ErrModerationUserNotFound      = errors.New("user not found")
	ErrModerationNotBanned         = errors.New("user is not banned from this community")
	ErrModerationCannotBanOwner    = errors.New("cannot ban the community owner")
	ErrModerationCannotBanSelf     = errors.New("you cannot ban yourself")
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

type ModerationBanUnitOfWork interface {
	CreateBanWithAudit(context.Context, *models.CommunityBan, *models.ModerationAuditLog) error
	UnbanWithAudit(ctx context.Context, communityID, userID uuid.UUID, audit *models.ModerationAuditLog) error
}

// ModerationService handles core moderation operations including banning, unbanning, and managing bans
type ModerationService struct {
	db            *pgxpool.Pool
	communityRepo ModerationCommunityRepo
	userRepo      ModerationUserRepo
	auditLogRepo  ModerationAuditRepo
	banWriter     ModerationBanUnitOfWork
}

// NewModerationService creates a new ModerationService
func NewModerationService(
	db *pgxpool.Pool,
	communityRepo *repository.CommunityRepository,
	userRepo *repository.UserRepository,
	auditLogRepo *repository.AuditLogRepository,
) *ModerationService {
	var banWriter ModerationBanUnitOfWork
	if db != nil {
		banWriter = repository.NewModerationBanWriter(db)
	}
	return &ModerationService{
		db:            db,
		communityRepo: communityRepo,
		userRepo:      userRepo,
		auditLogRepo:  auditLogRepo,
		banWriter:     banWriter,
	}
}

// BanUser bans a user from a community with permission and scope validation.
func (s *ModerationService) BanUser(ctx context.Context, communityID, moderatorID, targetUserID uuid.UUID, reason *string) error {
	_, err := s.BanUserWithResult(ctx, communityID, moderatorID, targetUserID, reason)
	return err
}

// BanUserWithResult atomically creates a ban and its audit record and returns the committed ban.
func (s *ModerationService) BanUserWithResult(ctx context.Context, communityID, moderatorID, targetUserID uuid.UUID, reason *string) (created *models.CommunityBan, err error) {
	start := time.Now()
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
			} else if errors.Is(err, ErrModerationCannotBanSelf) {
				errorType = "cannot_ban_self"
			} else if errors.Is(err, ErrModerationUserNotFound) {
				errorType = "user_not_found"
			} else {
				errorType = "database_error"
			}
			metrics.ModerationBanOperationsTotal.WithLabelValues("ban", "failed", errorType).Inc()
		}
	}()

	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get moderator: %w", err)
	}

	// Validate permission
	if err = s.validateModerationPermission(ctx, moderator, communityID); err != nil {
		return nil, err
	}

	if targetUserID == moderatorID {
		err = ErrModerationCannotBanSelf
		return nil, err
	}

	// Check if target user is the community owner
	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		err = ErrModerationCommunityNotFound
		return nil, err
	}
	if community.OwnerID == targetUserID {
		err = ErrModerationCannotBanOwner
		return nil, err
	}

	// A community mod or admin can only be banned by the owner or platform staff
	if !isPlatformStaff(moderator) && community.OwnerID != moderatorID {
		var targetMember *models.CommunityMember
		targetMember, err = s.communityRepo.GetMember(ctx, communityID, targetUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get target member: %w", err)
		}
		if targetMember != nil && (targetMember.Role == models.CommunityRoleMod || targetMember.Role == models.CommunityRoleAdmin) {
			err = ErrModerationPermissionDenied
			return nil, err
		}
	}

	// Check the target exists so a bad ID is a 404, not a foreign key error
	if _, err = s.userRepo.GetByID(ctx, targetUserID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			err = ErrModerationUserNotFound
			return nil, err
		}
		return nil, fmt.Errorf("failed to get target user: %w", err)
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
		ID:          uuid.New(),
		Action:      "ban_user",
		EntityType:  "community_ban",
		EntityID:    ban.ID,
		ModeratorID: moderatorID,
		Reason:      reason,
		Metadata:    metadata,
		ChannelID:   &communityID,
		CreatedAt:   time.Now(),
	}

	if s.banWriter != nil {
		err = s.banWriter.CreateBanWithAudit(ctx, ban, auditLog)
	} else {
		// Unit-test fallback for services constructed directly inside this package.
		_ = s.communityRepo.RemoveMember(ctx, communityID, targetUserID)
		if err = s.communityRepo.BanMember(ctx, ban); err == nil {
			err = s.auditLogRepo.Create(ctx, auditLog)
		}
	}
	if err != nil {
		metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "failed").Inc()
		metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())
		return nil, fmt.Errorf("failed to create ban and audit log: %w", err)
	}
	metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "success").Inc()
	metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())

	return ban, nil
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
		ChannelID:   &communityID,
	}

	if s.banWriter != nil {
		// Remove the ban and record its audit entry in one transaction.
		if err = s.banWriter.UnbanWithAudit(ctx, communityID, targetUserID, auditLog); err != nil {
			if errors.Is(err, repository.ErrCommunityBanNotFound) {
				err = ErrModerationNotBanned
				return err
			}
			metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "failed").Inc()
			metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())
			err = fmt.Errorf("failed to remove ban and create audit log: %w", err)
			return err
		}
	} else {
		// Unit-test fallback for services constructed directly inside this package.
		if err = s.communityRepo.UnbanMember(ctx, communityID, targetUserID); err != nil {
			return fmt.Errorf("failed to remove ban: %w", err)
		}
		if auditErr := s.auditLogRepo.Create(ctx, auditLog); auditErr != nil {
			metrics.ModerationAuditLogOperationsTotal.WithLabelValues("create", "failed").Inc()
			metrics.ModerationAuditLogOperationDuration.WithLabelValues("create").Observe(time.Since(auditStart).Seconds())
			return fmt.Errorf("failed to create audit log: %w", auditErr)
		}
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

	// Only admins and platform moderators can list all bans
	if !isPlatformStaff(moderator) {
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

// isPlatformStaff reports whether a user may moderate every community:
// admins and platform moderators (role moderator).
func isPlatformStaff(user *models.User) bool {
	if user.Role == models.RoleAdmin || user.Role == models.RoleModerator || user.AccountType == models.AccountTypeAdmin {
		return true
	}
	return user.AccountType == models.AccountTypeModerator && user.ModeratorScope == models.ModeratorScopeSite
}

// validateModerationPermission checks if a user may moderate a community.
// Platform staff may moderate anywhere. Everyone else must own the community
// or be a mod or admin member of it, as recorded in the community tables.
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

	if isPlatformStaff(moderator) {
		result = "allowed"
		return nil
	}

	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil || community == nil {
		result = "denied"
		reason = "community_not_found"
		return ErrModerationCommunityNotFound
	}
	if community.OwnerID == moderator.ID {
		result = "allowed"
		return nil
	}

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

// HasModerationPermission checks if a user has permission to moderate a community
// without performing any ban-related queries
func (s *ModerationService) HasModerationPermission(ctx context.Context, communityID, moderatorID uuid.UUID) error {
	// Get moderator user
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return fmt.Errorf("failed to get moderator: %w", err)
	}

	return s.validateModerationPermission(ctx, moderator, communityID)
}
