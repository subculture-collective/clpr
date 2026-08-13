package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

const creatorModerationRestrictionMessage = "You cannot interact with this creator's content because of a creator-level moderation restriction."

var allowedCreatorModeratorPermissions = map[string]struct{}{
	"manage_creator_clips":        {},
	"approve_creator_submissions": {},
	"remove_creator_comments":     {},
	"ban_creator_users":           {},
	"sync_platform_bans":          {},
}

var allowedCreatorBanScopes = map[string]struct{}{
	"interact": {},
	"submit":   {},
	"comment":  {},
}

var allowedCreatorModerationSources = map[string]struct{}{
	"manual":   {},
	"imported": {},
}

// CreatorModerationRepository defines the persistence needed by CreatorModerationService.
type CreatorModerationRepository interface {
	CreateModerator(ctx context.Context, moderator *models.CreatorModerator) error
	ListModeratorsByCreator(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorModerator, error)
	CreateBan(ctx context.Context, ban *models.CreatorBan) error
	GetActiveBanForUser(ctx context.Context, creatorID, targetUserID uuid.UUID, scope string) (*models.CreatorBan, error)
	GetActiveBanForPlatformIdentity(ctx context.Context, creatorID uuid.UUID, platform, platformUserID, scope string) (*models.CreatorBan, error)
}

// CreatorModerationService handles creator-scoped bans and moderator assignments.
type CreatorModerationService struct {
	repo CreatorModerationRepository
}

// NewCreatorModerationService creates a new CreatorModerationService.
func NewCreatorModerationService(repo CreatorModerationRepository) *CreatorModerationService {
	return &CreatorModerationService{repo: repo}
}

// CreateCreatorModeratorRequest captures inputs for moderator assignment.
type CreateCreatorModeratorRequest struct {
	UserID         *uuid.UUID
	Platform       *string
	PlatformUserID *string
	Permissions    []string
	Source         string
}

// CreateCreatorBanRequest captures inputs for creator ban creation.
type CreateCreatorBanRequest struct {
	TargetUserID         *uuid.UUID
	TargetPlatform       *string
	TargetPlatformUserID *string
	Source               string
	Reason               *string
	Scopes               []string
	ExpiresAt            *time.Time
	CreatedByUserID      *uuid.UUID
}

func normalizeModerationSource(source string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(source))
	if normalized == "" {
		normalized = "manual"
	}
	if _, ok := allowedCreatorModerationSources[normalized]; !ok {
		return "", fmt.Errorf("invalid moderation source %q", source)
	}
	return normalized, nil
}

func validateCreatorModeratorPermissions(permissions []string) ([]string, error) {
	if len(permissions) == 0 {
		return []string{}, nil
	}

	validated := make([]string, 0, len(permissions))
	seen := map[string]struct{}{}
	for _, permission := range permissions {
		normalized := strings.ToLower(strings.TrimSpace(permission))
		if normalized == "" {
			return nil, fmt.Errorf("permission cannot be empty")
		}
		if _, ok := allowedCreatorModeratorPermissions[normalized]; !ok {
			return nil, fmt.Errorf("invalid creator permission %q", permission)
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		validated = append(validated, normalized)
	}

	return validated, nil
}

func validateCreatorBanScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, fmt.Errorf("at least one ban scope is required")
	}

	validated := make([]string, 0, len(scopes))
	seen := map[string]struct{}{}
	for _, scope := range scopes {
		normalized := strings.ToLower(strings.TrimSpace(scope))
		if normalized == "" {
			return nil, fmt.Errorf("ban scope cannot be empty")
		}
		if _, ok := allowedCreatorBanScopes[normalized]; !ok {
			return nil, fmt.Errorf("invalid creator ban scope %q", scope)
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		validated = append(validated, normalized)
	}

	return validated, nil
}

func validateCreatorIdentity(userID *uuid.UUID, platform, platformUserID *string) error {
	if userID == nil && (platform == nil || platformUserID == nil) {
		return fmt.Errorf("either user_id or platform identity is required")
	}
	if (platform == nil) != (platformUserID == nil) {
		return fmt.Errorf("platform and platform_user_id must both be set together")
	}
	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// CreateModerator stores a creator-scoped moderator assignment.
func (s *CreatorModerationService) CreateModerator(ctx context.Context, creatorID uuid.UUID, req *CreateCreatorModeratorRequest) (*models.CreatorModerator, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	normalizedPlatform := normalizeOptionalString(req.Platform)
	normalizedPlatformUserID := normalizeOptionalString(req.PlatformUserID)
	if err := validateCreatorIdentity(req.UserID, normalizedPlatform, normalizedPlatformUserID); err != nil {
		return nil, err
	}
	permissions, err := validateCreatorModeratorPermissions(req.Permissions)
	if err != nil {
		return nil, err
	}
	source, err := normalizeModerationSource(req.Source)
	if err != nil {
		return nil, err
	}

	moderator := &models.CreatorModerator{
		ID:             uuid.New(),
		CreatorID:      creatorID,
		UserID:         req.UserID,
		Platform:       normalizedPlatform,
		PlatformUserID: normalizedPlatformUserID,
		Permissions:    permissions,
		Source:         source,
	}

	if s.repo == nil {
		return nil, fmt.Errorf("creator moderation repository is not configured")
	}
	if err := s.repo.CreateModerator(ctx, moderator); err != nil {
		return nil, err
	}

	return moderator, nil
}

// ListModeratorsByCreator lists creator moderators.
func (s *CreatorModerationService) ListModeratorsByCreator(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorModerator, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("creator moderation repository is not configured")
	}
	return s.repo.ListModeratorsByCreator(ctx, creatorID)
}

// CreateBan stores a creator-level ban.
func (s *CreatorModerationService) CreateBan(ctx context.Context, creatorID uuid.UUID, req *CreateCreatorBanRequest) (*models.CreatorBan, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	normalizedTargetPlatform := normalizeOptionalString(req.TargetPlatform)
	normalizedTargetPlatformUserID := normalizeOptionalString(req.TargetPlatformUserID)
	if err := validateCreatorIdentity(req.TargetUserID, normalizedTargetPlatform, normalizedTargetPlatformUserID); err != nil {
		return nil, err
	}
	scopes, err := validateCreatorBanScopes(req.Scopes)
	if err != nil {
		return nil, err
	}
	source, err := normalizeModerationSource(req.Source)
	if err != nil {
		return nil, err
	}

	ban := &models.CreatorBan{
		ID:                   uuid.New(),
		CreatorID:            creatorID,
		TargetUserID:         req.TargetUserID,
		TargetPlatform:       normalizedTargetPlatform,
		TargetPlatformUserID: normalizedTargetPlatformUserID,
		Source:               source,
		Reason:               req.Reason,
		Scopes:               scopes,
		ExpiresAt:            req.ExpiresAt,
		CreatedByUserID:      req.CreatedByUserID,
		SyncStatus:           "local_only",
	}

	if s.repo == nil {
		return nil, fmt.Errorf("creator moderation repository is not configured")
	}
	if err := s.repo.CreateBan(ctx, ban); err != nil {
		return nil, err
	}

	return ban, nil
}

func (s *CreatorModerationService) checkCreatorBan(ctx context.Context, creatorID, userID uuid.UUID, scope string) (bool, string, error) {
	if s.repo == nil {
		return false, "", fmt.Errorf("creator moderation repository is not configured")
	}
	ban, err := s.repo.GetActiveBanForUser(ctx, creatorID, userID, scope)
	if err != nil {
		return false, "", err
	}
	if ban != nil {
		return false, creatorModerationRestrictionMessage, nil
	}
	return true, "", nil
}

// CanInteract checks whether a user may interact with creator content.
func (s *CreatorModerationService) CanInteract(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error) {
	return s.checkCreatorBan(ctx, creatorID, userID, "interact")
}

// CanSubmit checks whether a user may submit content for a creator.
func (s *CreatorModerationService) CanSubmit(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error) {
	return s.checkCreatorBan(ctx, creatorID, userID, "submit")
}

// CanComment checks whether a user may comment on creator content.
func (s *CreatorModerationService) CanComment(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error) {
	return s.checkCreatorBan(ctx, creatorID, userID, "comment")
}
