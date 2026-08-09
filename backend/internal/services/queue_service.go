package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrQueueItemNotFound = repository.ErrQueueItemNotFound
	ErrQueueFull         = repository.ErrQueueFull
	ErrQueueEmpty        = repository.ErrQueueEmpty
)

type queueStore interface {
	GetUserQueue(context.Context, uuid.UUID, int) ([]models.QueueItemWithClip, error)
	GetQueueCount(context.Context, uuid.UUID) (int, error)
	AddItem(context.Context, *models.QueueItem) error
	AddItemAtTop(context.Context, *models.QueueItem) error
	RemoveItem(context.Context, uuid.UUID, uuid.UUID) error
	GetItemByID(context.Context, uuid.UUID, uuid.UUID) (*models.QueueItem, error)
	ReorderItem(context.Context, uuid.UUID, uuid.UUID, int) error
	MarkAsPlayed(context.Context, uuid.UUID, uuid.UUID) error
	ClearQueue(context.Context, uuid.UUID) error
	ConvertToPlaylist(context.Context, *models.Playlist, bool, bool) error
}

type queueClipStore interface {
	GetByID(context.Context, uuid.UUID) (*models.Clip, error)
}

// QueueService handles business logic for queue operations
type QueueService struct {
	queueRepo queueStore
	clipRepo  queueClipStore
}

// NewQueueService creates a new QueueService
func NewQueueService(queueRepo *repository.QueueRepository, clipRepo *repository.ClipRepository, playlistService *PlaylistService) *QueueService {
	_ = playlistService // retained for constructor compatibility while conversion is repository-atomic
	return &QueueService{
		queueRepo: queueRepo,
		clipRepo:  clipRepo,
	}
}

// GetQueue retrieves a user's queue
func (s *QueueService) GetQueue(ctx context.Context, userID uuid.UUID, limit int) (*models.Queue, error) {
	// Default limit is 100, max is 500
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	// Get queue items with clips
	items, err := s.queueRepo.GetUserQueue(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue: %w", err)
	}

	// Get total count
	total, err := s.queueRepo.GetQueueCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue count: %w", err)
	}

	// Determine next clip
	var nextClip *models.Clip
	if len(items) > 0 && items[0].Clip != nil {
		nextClip = items[0].Clip
	}

	queue := &models.Queue{
		Items:    items,
		Total:    total,
		NextClip: nextClip,
	}

	return queue, nil
}

// AddToQueue adds a clip to the user's queue
func (s *QueueService) AddToQueue(ctx context.Context, userID uuid.UUID, req *models.AddToQueueRequest) (*models.QueueItem, error) {
	// Parse clip ID
	clipID, err := uuid.Parse(req.ClipID)
	if err != nil {
		return nil, fmt.Errorf("invalid clip ID: %w", err)
	}

	// Check queue size limit (500 items)
	count, err := s.queueRepo.GetQueueCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check queue size: %w", err)
	}
	if count >= 500 {
		return nil, ErrQueueFull
	}

	// Verify clip exists
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify clip: %w", err)
	}
	if clip == nil {
		return nil, fmt.Errorf("clip not found")
	}

	// Check if clip is removed or hidden
	if clip.IsRemoved || clip.IsHidden {
		return nil, fmt.Errorf("clip is not available")
	}

	// Determine position
	atEnd := true // default is to add at end
	if req.AtEnd != nil {
		atEnd = *req.AtEnd
	}

	// Create queue item
	item := &models.QueueItem{
		ID:     uuid.New(),
		UserID: userID,
		ClipID: clipID,
	}

	if atEnd {
		// Add to end of queue (repository handles position calculation atomically)
		err = s.queueRepo.AddItem(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to add to queue: %w", err)
		}
	} else {
		// Add to beginning - shifts all existing items down
		err = s.queueRepo.AddItemAtTop(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to add to queue: %w", err)
		}
	}

	return item, nil
}

// RemoveFromQueue removes an item from the queue
func (s *QueueService) RemoveFromQueue(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	// Remove the item (repository handles position shifting)
	err := s.queueRepo.RemoveItem(ctx, itemID, userID)
	if errors.Is(err, repository.ErrQueueItemNotFound) {
		return ErrQueueItemNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to remove from queue: %w", err)
	}

	return nil
}

// ReorderQueue reorders a queue item to a new position
func (s *QueueService) ReorderQueue(ctx context.Context, userID uuid.UUID, req *models.ReorderQueueRequest) error {
	// Parse item ID
	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	// Verify item belongs to user
	item, err := s.queueRepo.GetItemByID(ctx, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify item: %w", err)
	}
	if item == nil {
		return ErrQueueItemNotFound
	}

	// Validate new position is >= 1
	if req.NewPosition < 1 {
		return fmt.Errorf("invalid position: must be >= 1")
	}

	// Get queue count to validate position is within bounds
	count, err := s.queueRepo.GetQueueCount(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get queue count: %w", err)
	}

	// Clamp new position to queue length if it exceeds bounds
	if req.NewPosition > count {
		req.NewPosition = count
	}

	// Perform reordering
	err = s.queueRepo.ReorderItem(ctx, itemID, userID, req.NewPosition)
	if errors.Is(err, repository.ErrQueueItemNotFound) {
		return ErrQueueItemNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to reorder queue: %w", err)
	}

	return nil
}

// MarkAsPlayed marks a queue item as played
func (s *QueueService) MarkAsPlayed(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	err := s.queueRepo.MarkAsPlayed(ctx, itemID, userID)
	if errors.Is(err, repository.ErrQueueItemNotFound) {
		return ErrQueueItemNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to mark as played: %w", err)
	}

	return nil
}

// ClearQueue clears all unplayed items from the queue
func (s *QueueService) ClearQueue(ctx context.Context, userID uuid.UUID) error {
	err := s.queueRepo.ClearQueue(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	return nil
}

// GetQueueCount gets the count of items in the queue
func (s *QueueService) GetQueueCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := s.queueRepo.GetQueueCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue count: %w", err)
	}

	return count, nil
}

// ConvertQueueToPlaylist creates a new playlist from the user's queue
func (s *QueueService) ConvertQueueToPlaylist(ctx context.Context, userID uuid.UUID, req *models.ConvertQueueToPlaylistRequest) (*models.Playlist, error) {
	playlist := &models.Playlist{ID: uuid.New(), UserID: userID, Title: strings.TrimSpace(req.Title), Description: req.Description, Visibility: models.PlaylistVisibilityPrivate}
	if playlist.Title == "" {
		return nil, fmt.Errorf("title cannot be blank")
	}
	if err := s.queueRepo.ConvertToPlaylist(ctx, playlist, req.OnlyUnplayed, req.ClearQueue); err != nil {
		if errors.Is(err, repository.ErrQueueEmpty) {
			return nil, ErrQueueEmpty
		}
		return nil, fmt.Errorf("failed to convert queue: %w", err)
	}
	return playlist, nil
}
