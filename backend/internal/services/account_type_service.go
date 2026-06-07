package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

var (
	// ErrInvalidAccountType is returned when an invalid account type is provided
	ErrInvalidAccountType = errors.New("invalid account type")
	// ErrCannotDowngradeAccountType is returned when trying to downgrade account type
	ErrCannotDowngradeAccountType = errors.New("cannot downgrade account type")
	// ErrBroadcasterVerificationFailed is returned when broadcaster verification fails
	ErrBroadcasterVerificationFailed = errors.New("broadcaster verification failed")
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = repository.ErrUserNotFound
)

// AccountTypeService handles account type conversions and management
type AccountTypeService struct {
	userRepo       *repository.UserRepository
	conversionRepo *repository.AccountTypeConversionRepository
	auditLogRepo   *repository.AuditLogRepository
	mfaService     *MFAService
}

// NewAccountTypeService creates a new account type service
func NewAccountTypeService(
	userRepo *repository.UserRepository,
	conversionRepo *repository.AccountTypeConversionRepository,
	auditLogRepo *repository.AuditLogRepository,
	mfaService *MFAService,
) *AccountTypeService {
	return &AccountTypeService{
		userRepo:       userRepo,
		conversionRepo: conversionRepo,
		auditLogRepo:   auditLogRepo,
		mfaService:     mfaService,
	}
}

// GetUserAccountType retrieves a user's account type with conversion history
func (s *AccountTypeService) GetUserAccountType(ctx context.Context, userID uuid.UUID) (*models.AccountTypeResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get conversion history
	conversions, err := s.conversionRepo.GetByUserID(ctx, userID, 10, 0)
	if err != nil {
		// Don't fail if we can't get history
		conversions = []models.AccountTypeConversion{}
	}

	// Get permissions for the account type
	permissions := models.GetAccountTypePermissions(user.GetAccountType())

	return &models.AccountTypeResponse{
		AccountType:       user.GetAccountType(),
		UpdatedAt:         user.AccountTypeUpdatedAt,
		Permissions:       permissions,
		ConversionHistory: conversions,
	}, nil
}

// ConvertToBroadcaster converts a user to broadcaster account type
func (s *AccountTypeService) ConvertToBroadcaster(ctx context.Context, userID uuid.UUID, reason *string, twitchVerified bool) error {
	// Get current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	currentType := user.GetAccountType()

	// Validate conversion is allowed
	if currentType == models.AccountTypeModerator || currentType == models.AccountTypeAdmin {
		return ErrCannotDowngradeAccountType
	}

	if currentType == models.AccountTypeBroadcaster {
		// Already a broadcaster, nothing to do
		return nil
	}

	// Note: Broadcaster verification is currently disabled to allow self-service conversions.
	// In production, enable strict Twitch broadcaster verification by uncommenting below:
	// if !twitchVerified {
	//     return ErrBroadcasterVerificationFailed
	// }
	// This allows testing and MVP deployment without requiring Twitch API integration.
	// Consider adding a feature flag to toggle verification: cfg.Features.RequireBroadcasterVerification

	// Update account type
	err = s.userRepo.UpdateAccountType(ctx, userID, models.AccountTypeBroadcaster)
	if err != nil {
		return fmt.Errorf("failed to update account type: %w", err)
	}

	// Log the conversion
	conversion := &models.AccountTypeConversion{
		ID:          uuid.New(),
		UserID:      userID,
		OldType:     currentType,
		NewType:     models.AccountTypeBroadcaster,
		Reason:      reason,
		ConvertedBy: nil, // Self-service conversion
		Metadata: map[string]interface{}{
			"twitch_verified": twitchVerified,
			"self_service":    true,
		},
	}

	err = s.conversionRepo.Create(ctx, conversion)
	if err != nil {
		// Log error but don't fail the conversion - audit trail is important but not critical
		utils.Warn("Failed to create conversion audit log", map[string]interface{}{"user_id": userID, "error": err})
	}

	return nil
}

// ConvertToModerator converts a user to moderator account type (admin only)
func (s *AccountTypeService) ConvertToModerator(ctx context.Context, targetUserID, adminUserID uuid.UUID, reason *string) error {
	// Get current user
	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	currentType := user.GetAccountType()

	// Validate conversion is allowed
	if currentType == models.AccountTypeAdmin {
		return ErrCannotDowngradeAccountType
	}

	if currentType == models.AccountTypeModerator {
		// Already a moderator, nothing to do
		return nil
	}

	// Update account type
	err = s.userRepo.UpdateAccountType(ctx, targetUserID, models.AccountTypeModerator)
	if err != nil {
		return fmt.Errorf("failed to update account type: %w", err)
	}

	// Log the conversion
	conversion := &models.AccountTypeConversion{
		ID:          uuid.New(),
		UserID:      targetUserID,
		OldType:     currentType,
		NewType:     models.AccountTypeModerator,
		Reason:      reason,
		ConvertedBy: &adminUserID,
		Metadata: map[string]interface{}{
			"admin_conversion": true,
		},
	}

	err = s.conversionRepo.Create(ctx, conversion)
	if err != nil {
		// Log error but don't fail the conversion - audit trail is important but not critical
		utils.Warn("Failed to create conversion audit log", map[string]interface{}{"user_id": targetUserID, "error": err})
	}

	// Create audit log entry
	if s.auditLogRepo != nil {
		auditLog := &models.ModerationAuditLog{
			ID:          uuid.New(),
			Action:      "account_type_conversion",
			EntityType:  "user",
			EntityID:    targetUserID,
			ModeratorID: adminUserID,
			Reason:      reason,
			Metadata: map[string]interface{}{
				"old_type": currentType,
				"new_type": models.AccountTypeModerator,
			},
			CreatedAt: time.Now(),
		}
		if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
			utils.Warn("Failed to create moderation audit log", map[string]interface{}{"user_id": targetUserID, "error": err})
		}
	}

	// Trigger MFA requirement for moderator role
	if s.mfaService != nil {
		if err := s.mfaService.SetMFARequired(ctx, targetUserID); err != nil {
			// This is a critical security function - fail the operation if MFA cannot be enforced
			return fmt.Errorf("failed to set MFA requirement for user %s after moderator promotion: %w", targetUserID, err)
		}
	} else {
		// MFA service is required for security - this should not happen
		utils.Error("MFA service not available when promoting user to moderator", nil, map[string]interface{}{"user_id": targetUserID})
		return errors.New("MFA service not available when promoting user to moderator")
	}

	return nil
}

// ConvertToAdmin converts a user to admin account type (admin only)
func (s *AccountTypeService) ConvertToAdmin(ctx context.Context, targetUserID, adminUserID uuid.UUID, reason *string) error {
	// Get current user
	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	currentType := user.GetAccountType()

	if currentType == models.AccountTypeAdmin {
		// Already an admin, nothing to do
		return nil
	}

	// Update account type
	err = s.userRepo.UpdateAccountType(ctx, targetUserID, models.AccountTypeAdmin)
	if err != nil {
		return fmt.Errorf("failed to update account type: %w", err)
	}

	// Log the conversion
	conversion := &models.AccountTypeConversion{
		ID:          uuid.New(),
		UserID:      targetUserID,
		OldType:     currentType,
		NewType:     models.AccountTypeAdmin,
		Reason:      reason,
		ConvertedBy: &adminUserID,
		Metadata: map[string]interface{}{
			"admin_conversion": true,
		},
	}

	err = s.conversionRepo.Create(ctx, conversion)
	if err != nil {
		// Log error but don't fail the conversion - audit trail is important but not critical
		utils.Warn("Failed to create conversion audit log", map[string]interface{}{"user_id": targetUserID, "error": err})
	}

	// Create audit log entry
	if s.auditLogRepo != nil {
		auditLog := &models.ModerationAuditLog{
			ID:          uuid.New(),
			Action:      "account_type_conversion",
			EntityType:  "user",
			EntityID:    targetUserID,
			ModeratorID: adminUserID,
			Reason:      reason,
			Metadata: map[string]interface{}{
				"old_type": currentType,
				"new_type": models.AccountTypeAdmin,
			},
			CreatedAt: time.Now(),
		}
		if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
			utils.Warn("Failed to create moderation audit log", map[string]interface{}{"user_id": targetUserID, "error": err})
		}
	}

	// Trigger MFA requirement for admin role
	if s.mfaService != nil {
		if err := s.mfaService.SetMFARequired(ctx, targetUserID); err != nil {
			// This is a critical security function - log and return error to prevent admin promotion without MFA
			utils.Error("SECURITY: Failed to set MFA requirement for user after admin promotion", err, map[string]interface{}{"user_id": targetUserID})
			utils.Warn("SECURITY: Manually verify MFA requirement for user", map[string]interface{}{"user_id": targetUserID})
			return fmt.Errorf("failed to set MFA requirement for user %s after admin promotion: %w", targetUserID, err)
		}
	} else {
		// MFA service is required for security - this should not happen
		utils.Error("MFA service not available when promoting user to admin", nil, map[string]interface{}{"user_id": targetUserID})
		return errors.New("MFA service not available when promoting user to admin")
	}

	return nil
}

// GetConversionHistory retrieves conversion history for a user
func (s *AccountTypeService) GetConversionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AccountTypeConversion, int, error) {
	conversions, err := s.conversionRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.conversionRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return conversions, total, nil
}

// GetRecentConversions retrieves recent conversions across all users (admin only)
func (s *AccountTypeService) GetRecentConversions(ctx context.Context, limit, offset int) ([]models.AccountTypeConversion, int, error) {
	conversions, err := s.conversionRepo.GetRecentConversions(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.conversionRepo.CountTotal(ctx)
	if err != nil {
		return nil, 0, err
	}

	return conversions, total, nil
}

// GetAccountTypeStats returns statistics about account type distribution
func (s *AccountTypeService) GetAccountTypeStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	accountTypes := []string{
		models.AccountTypeMember,
		models.AccountTypeBroadcaster,
		models.AccountTypeModerator,
		models.AccountTypeAdmin,
	}

	for _, accountType := range accountTypes {
		count, err := s.conversionRepo.CountByAccountType(ctx, accountType)
		if err != nil {
			return nil, err
		}
		stats[accountType] = count
	}

	return stats, nil
}

// ValidateAccountTypeConversion checks if a conversion is valid
func (s *AccountTypeService) ValidateAccountTypeConversion(ctx context.Context, userID uuid.UUID, targetType string) error {
	// Validate target type
	if !models.IsValidAccountType(targetType) {
		return ErrInvalidAccountType
	}

	// Get current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	currentType := user.GetAccountType()

	// Define upgrade paths
	// member -> broadcaster (allowed)
	// member -> moderator (admin only)
	// broadcaster -> moderator (admin only)
	// any -> admin (admin only)

	// Prevent downgrades
	typeHierarchy := map[string]int{
		models.AccountTypeMember:      1,
		models.AccountTypeBroadcaster: 2,
		models.AccountTypeModerator:   3,
		models.AccountTypeAdmin:       4,
	}

	if typeHierarchy[targetType] < typeHierarchy[currentType] {
		return ErrCannotDowngradeAccountType
	}

	return nil
}
