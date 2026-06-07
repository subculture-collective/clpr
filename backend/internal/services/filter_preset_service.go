package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

type FilterPresetService struct {
	presetRepo *repository.FilterPresetRepository
}

func NewFilterPresetService(presetRepo *repository.FilterPresetRepository) *FilterPresetService {
	return &FilterPresetService{
		presetRepo: presetRepo,
	}
}

// CreatePreset creates a new filter preset for a user
func (s *FilterPresetService) CreatePreset(ctx context.Context, userID uuid.UUID, req *models.CreateFilterPresetRequest) (*models.UserFilterPreset, error) {
	// Convert filters to JSON
	filtersJSON, err := repository.FiltersToJSON(&req.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to convert filters to JSON: %w", err)
	}

	preset := &models.UserFilterPreset{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		FiltersJSON: filtersJSON,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.presetRepo.CreatePreset(ctx, preset)
	if err != nil {
		return nil, err
	}

	return preset, nil
}

// GetPreset retrieves a filter preset by ID
func (s *FilterPresetService) GetPreset(ctx context.Context, presetID uuid.UUID, userID uuid.UUID) (*models.UserFilterPreset, error) {
	preset, err := s.presetRepo.GetPresetByID(ctx, presetID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if preset.UserID != userID {
		return nil, repository.ErrUnauthorizedPresetAccess
	}

	return preset, nil
}

// GetUserPresets retrieves all filter presets for a user
func (s *FilterPresetService) GetUserPresets(ctx context.Context, userID uuid.UUID) ([]*models.UserFilterPreset, error) {
	return s.presetRepo.GetUserPresets(ctx, userID)
}

// UpdatePreset updates a filter preset
func (s *FilterPresetService) UpdatePreset(ctx context.Context, presetID, userID uuid.UUID, req *models.UpdateFilterPresetRequest) (*models.UserFilterPreset, error) {
	// Get existing preset and verify ownership
	preset, err := s.presetRepo.GetPresetByID(ctx, presetID)
	if err != nil {
		return nil, err
	}

	if preset.UserID != userID {
		return nil, repository.ErrUnauthorizedPresetAccess
	}

	// Update fields if provided
	if req.Name != nil {
		preset.Name = *req.Name
	}

	if req.Filters != nil {
		filtersJSON, err := repository.FiltersToJSON(req.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filters to JSON: %w", err)
		}
		preset.FiltersJSON = filtersJSON
	}

	err = s.presetRepo.UpdatePreset(ctx, preset)
	if err != nil {
		return nil, fmt.Errorf("failed to update preset: %w", err)
	}

	return preset, nil
}

// DeletePreset deletes a filter preset
func (s *FilterPresetService) DeletePreset(ctx context.Context, presetID, userID uuid.UUID) error {
	return s.presetRepo.DeletePreset(ctx, presetID, userID)
}
