package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

var (
	ErrTemplateNotFound     = errors.New("template not found")
	ErrTemplateNameExists   = errors.New("template name already exists")
	ErrCannotDeleteDefault  = errors.New("cannot delete default templates")
	ErrUnauthorizedTemplate = errors.New("unauthorized to modify this template")
)

type BanReasonTemplateService struct {
	repo          *repository.BanReasonTemplateRepository
	communityRepo *repository.CommunityRepository
	logger        *utils.StructuredLogger
}

func NewBanReasonTemplateService(
	repo *repository.BanReasonTemplateRepository,
	communityRepo *repository.CommunityRepository,
	logger *utils.StructuredLogger,
) *BanReasonTemplateService {
	return &BanReasonTemplateService{
		repo:          repo,
		communityRepo: communityRepo,
		logger:        logger,
	}
}

// canManageBanTemplates reports whether caller may read and change the
// templates of a broadcaster's channel: the broadcaster themselves, or an
// admin or platform moderator. A nil broadcasterID means templates that belong
// to no channel, which only admins and platform moderators manage.
func canManageBanTemplates(caller *models.User, broadcasterID *string) bool {
	if caller == nil {
		return false
	}
	if caller.Role == models.RoleAdmin || caller.Role == models.RoleModerator {
		return true
	}
	return broadcasterID != nil && *broadcasterID != "" &&
		caller.TwitchID != nil && *caller.TwitchID == *broadcasterID
}

// GetTemplate retrieves a template by ID. Default templates are open to every
// signed-in user; other templates need rights on the channel they belong to.
func (s *BanReasonTemplateService) GetTemplate(ctx context.Context, caller *models.User, id uuid.UUID) (*models.BanReasonTemplate, error) {
	template, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}
	if !template.IsDefault && !canManageBanTemplates(caller, template.BroadcasterID) {
		return nil, ErrUnauthorizedTemplate
	}
	return template, nil
}

// ListTemplates retrieves templates for a broadcaster or defaults. Listing a
// broadcaster's templates needs rights on that channel.
func (s *BanReasonTemplateService) ListTemplates(ctx context.Context, caller *models.User, broadcasterID *string, includeDefaults bool) ([]models.BanReasonTemplate, error) {
	if (broadcasterID != nil || !includeDefaults) && !canManageBanTemplates(caller, broadcasterID) {
		return nil, ErrUnauthorizedTemplate
	}
	templates, err := s.repo.List(ctx, broadcasterID, includeDefaults)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	return templates, nil
}

// CreateTemplate creates a new template
func (s *BanReasonTemplateService) CreateTemplate(ctx context.Context, caller *models.User, req *models.CreateBanReasonTemplateRequest) (*models.BanReasonTemplate, error) {
	// Only the broadcaster and platform staff may add templates to a channel
	if !canManageBanTemplates(caller, req.BroadcasterID) {
		return nil, ErrUnauthorizedTemplate
	}
	userID := caller.ID

	template := &models.BanReasonTemplate{
		Name:            req.Name,
		Reason:          req.Reason,
		DurationSeconds: req.DurationSeconds,
		IsDefault:       false, // User-created templates are never default
		BroadcasterID:   req.BroadcasterID,
		CreatedBy:       &userID,
	}

	err := s.repo.Create(ctx, template)
	if err != nil {
		// Check if it's a unique constraint violation (PostgreSQL error code 23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrTemplateNameExists
		}
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	s.logger.Info("Created ban reason template", map[string]interface{}{
		"template_id":    template.ID,
		"name":           template.Name,
		"broadcaster_id": template.BroadcasterID,
		"created_by":     userID,
	})

	return template, nil
}

// UpdateTemplate updates an existing template
func (s *BanReasonTemplateService) UpdateTemplate(ctx context.Context, caller *models.User, templateID uuid.UUID, req *models.UpdateBanReasonTemplateRequest) (*models.BanReasonTemplate, error) {
	// Get existing template
	template, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}

	// Check permissions
	if template.IsDefault {
		return nil, ErrCannotDeleteDefault
	}

	if !canManageBanTemplates(caller, template.BroadcasterID) {
		return nil, ErrUnauthorizedTemplate
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Reason != nil {
		updates["reason"] = *req.Reason
	}
	if req.DurationSeconds != nil {
		updates["duration_seconds"] = *req.DurationSeconds
	}

	if len(updates) == 0 {
		return template, nil // No updates
	}

	err = s.repo.Update(ctx, templateID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	s.logger.Info("Updated ban reason template", map[string]interface{}{
		"template_id": templateID,
		"updated_by":  caller.ID,
	})

	// Fetch updated template
	return s.repo.GetByID(ctx, templateID)
}

// DeleteTemplate deletes a template
func (s *BanReasonTemplateService) DeleteTemplate(ctx context.Context, caller *models.User, templateID uuid.UUID) error {
	// Get template to check permissions
	template, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return ErrTemplateNotFound
	}

	// Check permissions
	if template.IsDefault {
		return ErrCannotDeleteDefault
	}

	if !canManageBanTemplates(caller, template.BroadcasterID) {
		return ErrUnauthorizedTemplate
	}

	err = s.repo.Delete(ctx, templateID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTemplateNotFound
		}
		return fmt.Errorf("failed to delete template: %w", err)
	}

	s.logger.Info("Deleted ban reason template", map[string]interface{}{
		"template_id": templateID,
		"deleted_by":  caller.ID,
	})

	return nil
}

// UseTemplate marks a template as used and increments usage count
func (s *BanReasonTemplateService) UseTemplate(ctx context.Context, templateID uuid.UUID) (*models.BanReasonTemplate, error) {
	template, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}

	err = s.repo.IncrementUsage(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment usage: %w", err)
	}

	// Fetch updated template
	return s.repo.GetByID(ctx, templateID)
}

// GetUsageStats retrieves usage statistics for templates. Stats for a
// broadcaster's templates need rights on that channel.
func (s *BanReasonTemplateService) GetUsageStats(ctx context.Context, caller *models.User, broadcasterID *string) ([]models.BanReasonTemplate, error) {
	if broadcasterID != nil && !canManageBanTemplates(caller, broadcasterID) {
		return nil, ErrUnauthorizedTemplate
	}
	stats, err := s.repo.GetUsageStats(ctx, broadcasterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}
	return stats, nil
}
