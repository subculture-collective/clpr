package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// QueueService handles business logic for queue operations
type QueueService struct {
	queueRepo       *repository.QueueRepository
	clipRepo        *repository.ClipRepository
	playlistService *PlaylistService
}

// NewQueueService creates a new QueueService
func NewQueueService(queueRepo *repository.QueueRepository, clipRepo *repository.ClipRepository, playlistService *PlaylistService) *QueueService {
	return &QueueService{
		queueRepo:       queueRepo,
		clipRepo:        clipRepo,
		playlistService: playlistService,
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
		return nil, fmt.Errorf("queue is full (maximum 500 items)")
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
	if err == pgx.ErrNoRows {
		return fmt.Errorf("queue item not found")
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
		return fmt.Errorf("queue item not found")
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
	if err != nil {
		return fmt.Errorf("failed to reorder queue: %w", err)
	}

	return nil
}

// MarkAsPlayed marks a queue item as played
func (s *QueueService) MarkAsPlayed(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	err := s.queueRepo.MarkAsPlayed(ctx, itemID, userID)
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
	// Get all queue items
	queueItems, err := s.queueRepo.GetUserQueue(ctx, userID, 1000) // Max 1000 items
	if err != nil {
		return nil, fmt.Errorf("failed to get queue items: %w", err)
	}

	if len(queueItems) == 0 {
		return nil, fmt.Errorf("queue is empty, nothing to convert")
	}

	// Filter to only include unplayed items if requested
	var itemsToConvert []models.QueueItemWithClip
	if req.OnlyUnplayed {
		for _, item := range queueItems {
			if item.PlayedAt == nil {
				itemsToConvert = append(itemsToConvert, item)
			}
		}
	} else {
		itemsToConvert = queueItems
	}

	if len(itemsToConvert) == 0 {
		return nil, fmt.Errorf("no items to convert (all items have been played)")
	}

	// Create the playlist
	visibility := models.PlaylistVisibilityPrivate
	createReq := &models.CreatePlaylistRequest{
		Title:       req.Title,
		Description: req.Description,
		Visibility:  &visibility,
	}

	playlist, err := s.playlistService.CreatePlaylist(ctx, userID, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	// Add clips to the playlist in queue order
	clipIDs := make([]uuid.UUID, 0, len(itemsToConvert))
	for _, item := range itemsToConvert {
		clipIDs = append(clipIDs, item.ClipID)
	}

	err = s.playlistService.AddClipsToPlaylist(ctx, playlist.ID, userID, clipIDs)
	if err != nil {
		// If adding clips fails, we should ideally delete the playlist
		// but for now just return the error
		return nil, fmt.Errorf("failed to add clips to playlist: %w", err)
	}

	// Clear queue if requested
	if req.ClearQueue {
		err = s.queueRepo.ClearQueue(ctx, userID)
		if err != nil {
			// Log but don't fail - playlist was already created
			fmt.Printf("Warning: failed to clear queue after conversion: %v\n", err)
		}
	}

	return playlist, nil
}
