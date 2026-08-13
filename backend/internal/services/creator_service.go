package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

var creatorSlugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

var creatorPlatformAllowlist = map[string]struct{}{
	"twitch":  {},
	"kick":    {},
	"youtube": {},
	"tiktok":  {},
}

// CreatorService handles creator account lifecycle and platform linking.
type CreatorService struct {
	creatorRepo creatorRepository
}

type creatorRepository interface {
	CreateCreatorAccount(ctx context.Context, account *models.CreatorAccount) error
	GetCreatorAccountByID(ctx context.Context, id uuid.UUID) (*models.CreatorAccount, error)
	GetCreatorAccountBySlug(ctx context.Context, slug string) (*models.CreatorAccount, error)
	ListCreatorAccountsByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]*models.CreatorAccount, error)
	CreateCreatorPlatformAccount(ctx context.Context, account *models.CreatorPlatformAccount) error
	GetCreatorPlatformAccountByPlatformAndUserID(ctx context.Context, platform, platformUserID string) (*models.CreatorPlatformAccount, error)
	ListCreatorPlatformAccounts(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorPlatformAccount, error)
}

// NewCreatorService creates a new CreatorService.
func NewCreatorService(creatorRepo creatorRepository) *CreatorService {
	return &CreatorService{creatorRepo: creatorRepo}
}

// CreateCreatorAccountRequest captures the input needed to create a creator account.
type CreateCreatorAccountRequest struct {
	DisplayName string
	Slug        string
}

// LinkCreatorPlatformAccountRequest captures the input needed to attach a platform account.
type LinkCreatorPlatformAccountRequest struct {
	Platform              string
	PlatformUserID        string
	PlatformDisplayName   string
	ProfileURL            *string
	CanImportBans         *bool
	CanSyncBansOutbound   *bool
	CanImportModerators   *bool
	CanVerifyOwnership    *bool
	CanFetchMetadata      *bool
	AccessTokenEncrypted  *string
	RefreshTokenEncrypted *string
	TokenExpiresAt        *time.Time
}

// normalizeCreatorSlug validates and normalizes a creator slug.
func normalizeCreatorSlug(input string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(input))
	if slug == "" {
		return "", fmt.Errorf("slug is required")
	}
	if len(slug) > 255 {
		return "", fmt.Errorf("slug must be 255 characters or fewer")
	}
	if !creatorSlugPattern.MatchString(slug) {
		return "", fmt.Errorf("slug may only contain lowercase letters, numbers, and hyphens")
	}
	if !strings.ContainsAny(slug, "abcdefghijklmnopqrstuvwxyz0123456789") {
		return "", fmt.Errorf("slug must contain at least one letter or number")
	}
	return slug, nil
}

func normalizeNonEmptyValue(fieldName, value string, maxLen int) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}
	if len(trimmed) > maxLen {
		return "", fmt.Errorf("%s must be %d characters or fewer", fieldName, maxLen)
	}
	return trimmed, nil
}

func validateCreatorPlatform(platform string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(platform))
	if normalized == "" {
		return "", fmt.Errorf("platform is required")
	}
	if len(normalized) > 30 {
		return "", fmt.Errorf("platform must be 30 characters or fewer")
	}
	if _, ok := creatorPlatformAllowlist[normalized]; !ok {
		return "", fmt.Errorf("unsupported platform %q", platform)
	}
	return normalized, nil
}

func normalizeProfileURL(profileURL *string) *string {
	if profileURL == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*profileURL)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func boolPtrValue(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// CreateCreatorAccount creates a creator account for the specified owner.
func (s *CreatorService) CreateCreatorAccount(ctx context.Context, ownerUserID uuid.UUID, req *CreateCreatorAccountRequest) (*models.CreatorAccount, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	displayName, err := normalizeNonEmptyValue("display_name", req.DisplayName, 255)
	if err != nil {
		return nil, err
	}
	slug, err := normalizeCreatorSlug(req.Slug)
	if err != nil {
		return nil, err
	}

	account := &models.CreatorAccount{
		ID:          uuid.New(),
		OwnerUserID: ownerUserID,
		DisplayName: displayName,
		Slug:        slug,
	}

	if s.creatorRepo == nil {
		return nil, fmt.Errorf("creator repository is not configured")
	}
	if err := s.creatorRepo.CreateCreatorAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetCreatorAccountByID retrieves a creator account by its ID.
func (s *CreatorService) GetCreatorAccountByID(ctx context.Context, id uuid.UUID) (*models.CreatorAccount, error) {
	if s.creatorRepo == nil {
		return nil, fmt.Errorf("creator repository is not configured")
	}
	return s.creatorRepo.GetCreatorAccountByID(ctx, id)
}

// GetCreatorAccountBySlug retrieves a creator account by slug.
func (s *CreatorService) GetCreatorAccountBySlug(ctx context.Context, slug string) (*models.CreatorAccount, error) {
	normalized, err := normalizeCreatorSlug(slug)
	if err != nil {
		return nil, err
	}
	if s.creatorRepo == nil {
		return nil, fmt.Errorf("creator repository is not configured")
	}
	return s.creatorRepo.GetCreatorAccountBySlug(ctx, normalized)
}

// ListCreatorAccountsByOwner lists creator accounts owned by a user.
func (s *CreatorService) ListCreatorAccountsByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]*models.CreatorAccount, error) {
	if s.creatorRepo == nil {
		return nil, fmt.Errorf("creator repository is not configured")
	}
	return s.creatorRepo.ListCreatorAccountsByOwner(ctx, ownerUserID)
}

// LinkPlatformAccount attaches a platform account to a creator account.
func (s *CreatorService) LinkPlatformAccount(ctx context.Context, creatorID uuid.UUID, req *LinkCreatorPlatformAccountRequest) (*models.CreatorPlatformAccount, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	platform, err := validateCreatorPlatform(req.Platform)
	if err != nil {
		return nil, err
	}
	platformUserID, err := normalizeNonEmptyValue("platform_user_id", req.PlatformUserID, 255)
	if err != nil {
		return nil, err
	}
	platformDisplayName, err := normalizeNonEmptyValue("platform_display_name", req.PlatformDisplayName, 255)
	if err != nil {
		return nil, err
	}
	defaultCapabilities := DefaultPlatformModerationCapabilities(platform)

	account := &models.CreatorPlatformAccount{
		ID:                    uuid.New(),
		CreatorID:             creatorID,
		Platform:              platform,
		PlatformUserID:        platformUserID,
		PlatformDisplayName:   platformDisplayName,
		ProfileURL:            normalizeProfileURL(req.ProfileURL),
		CanImportBans:         boolPtrValue(req.CanImportBans, defaultCapabilities.CanImportBans),
		CanSyncBansOutbound:   boolPtrValue(req.CanSyncBansOutbound, defaultCapabilities.CanSyncBansOutbound),
		CanImportModerators:   boolPtrValue(req.CanImportModerators, defaultCapabilities.CanImportModerators),
		CanVerifyOwnership:    boolPtrValue(req.CanVerifyOwnership, defaultCapabilities.CanVerifyOwnership),
		CanFetchMetadata:      boolPtrValue(req.CanFetchMetadata, defaultCapabilities.CanFetchMetadata),
		AccessTokenEncrypted:  req.AccessTokenEncrypted,
		RefreshTokenEncrypted: req.RefreshTokenEncrypted,
		TokenExpiresAt:        req.TokenExpiresAt,
	}

	if s.creatorRepo == nil {
		return nil, fmt.Errorf("creator repository is not configured")
	}
	if err := s.creatorRepo.CreateCreatorPlatformAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetCreatorPlatformAccount retrieves a linked platform account by platform identity.
func (s *CreatorService) GetCreatorPlatformAccount(ctx context.Context, platform, platformUserID string) (*models.CreatorPlatformAccount, error) {
	platform, err := validateCreatorPlatform(platform)
	if err != nil {
		return nil, err
	}
	platformUserID, err = normalizeNonEmptyValue("platform_user_id", platformUserID, 255)
	if err != nil {
		return nil, err
	}
	if s.creatorRepo == nil {
		return nil, fmt.Errorf("creator repository is not configured")
	}
	return s.creatorRepo.GetCreatorPlatformAccountByPlatformAndUserID(ctx, platform, platformUserID)
}

// ListCreatorPlatformAccounts lists all linked platform accounts for a creator.
func (s *CreatorService) ListCreatorPlatformAccounts(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorPlatformAccount, error) {
	if s.creatorRepo == nil {
		return nil, fmt.Errorf("creator repository is not configured")
	}
	return s.creatorRepo.ListCreatorPlatformAccounts(ctx, creatorID)
}
