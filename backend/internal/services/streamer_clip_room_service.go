package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrStreamerClipRoomForbidden = errors.New("streamer clip room forbidden")

type streamerClipRoomRepository interface {
	GetOrCreateRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error)
	GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.StreamerClipRoom, error)
	SetRoomActive(ctx context.Context, roomID uuid.UUID, active bool, listenerError *string) error
	CreateItem(ctx context.Context, item *models.StreamerClipRoomItem) error
	ListItems(ctx context.Context, roomID uuid.UUID, status string, limit int) ([]models.StreamerClipRoomItem, error)
	GetItem(ctx context.Context, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error)
	ApproveItem(ctx context.Context, roomID, itemID, approverID uuid.UUID) (*models.StreamerClipRoomItem, error)
	RejectItem(ctx context.Context, roomID, itemID, rejecterID uuid.UUID) (*models.StreamerClipRoomItem, error)
	ReorderApprovedItems(ctx context.Context, roomID uuid.UUID, itemIDs []uuid.UUID) error
}

type clipLookupRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Clip, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Clip, error)
	GetByTwitchClipID(ctx context.Context, twitchClipID string) (*models.Clip, error)
}

type StreamerClipRoomService struct {
	rooms            streamerClipRoomRepository
	clips            clipLookupRepository
	eventBroadcaster func(roomID uuid.UUID, eventType string, data map[string]interface{})
}

func NewStreamerClipRoomService(rooms *repository.StreamerClipRoomRepository, clips *repository.ClipRepository) *StreamerClipRoomService {
	return &StreamerClipRoomService{rooms: rooms, clips: clips}
}

func newStreamerClipRoomService(rooms streamerClipRoomRepository, clips clipLookupRepository) *StreamerClipRoomService {
	return &StreamerClipRoomService{rooms: rooms, clips: clips}
}

func (s *StreamerClipRoomService) SetEventBroadcaster(broadcaster func(roomID uuid.UUID, eventType string, data map[string]interface{})) {
	if s == nil {
		return
	}
	s.eventBroadcaster = broadcaster
}

type TwitchChatClipMessage struct {
	RoomID          uuid.UUID
	TwitchMessageID string
	TwitchUserID    string
	TwitchUsername  string
	MessageText     string
}

func (s *StreamerClipRoomService) GetOrCreateRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoomWithItems, error) {
	room, err := s.rooms.GetOrCreateRoom(ctx, ownerUserID, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("streamer clip room not found")
	}

	items, err := s.rooms.ListItems(ctx, room.ID, "all", math.MaxInt)
	if err != nil {
		return nil, fmt.Errorf("failed to list room items: %w", err)
	}
	items, err = s.hydrateItems(ctx, items)
	if err != nil {
		return nil, err
	}

	result := &models.StreamerClipRoomWithItems{Room: *room}
	for _, item := range items {
		switch item.Status {
		case models.StreamerClipRoomItemStatusApproved:
			result.ApprovedItems = append(result.ApprovedItems, item)
		case models.StreamerClipRoomItemStatusRejected:
			// room detail view only needs pending/approved/skipped, but keep rejected data accessible via skipped bucket? no
			result.SkippedItems = append(result.SkippedItems, item)
		case models.StreamerClipRoomItemStatusSkipped:
			result.SkippedItems = append(result.SkippedItems, item)
		default:
			result.PendingItems = append(result.PendingItems, item)
		}
	}

	return result, nil
}

func (s *StreamerClipRoomService) GetRoomByID(ctx context.Context, actorUserID, roomID uuid.UUID) (*models.StreamerClipRoom, error) {
	room, err := s.requireOwner(ctx, actorUserID, roomID)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (s *StreamerClipRoomService) ListItems(ctx context.Context, actorUserID, roomID uuid.UUID, status string) ([]models.StreamerClipRoomItem, error) {
	if _, err := s.requireOwner(ctx, actorUserID, roomID); err != nil {
		return nil, err
	}
	items, err := s.rooms.ListItems(ctx, roomID, status, math.MaxInt)
	if err != nil {
		return nil, fmt.Errorf("failed to list streamer clip room items: %w", err)
	}
	return s.hydrateItems(ctx, items)
}

func (s *StreamerClipRoomService) GetItem(ctx context.Context, actorUserID, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	if _, err := s.requireOwner(ctx, actorUserID, roomID); err != nil {
		return nil, err
	}
	item, err := s.rooms.GetItem(ctx, roomID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get streamer clip room item: %w", err)
	}
	return s.hydrateItem(ctx, item)
}

func (s *StreamerClipRoomService) StartRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error) {
	room, err := s.rooms.GetOrCreateRoom(ctx, ownerUserID, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("streamer clip room not found")
	}

	now := time.Now().UTC()
	if err := s.rooms.SetRoomActive(ctx, room.ID, true, nil); err != nil {
		return nil, fmt.Errorf("failed to start room: %w", err)
	}
	room.IsActive = true
	room.LastListenerError = nil
	room.ListenerStartedAt = &now
	room.UpdatedAt = now
	return room, nil
}

func (s *StreamerClipRoomService) StopRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error) {
	room, err := s.rooms.GetOrCreateRoom(ctx, ownerUserID, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("streamer clip room not found")
	}

	if err := s.rooms.SetRoomActive(ctx, room.ID, false, nil); err != nil {
		return nil, fmt.Errorf("failed to stop room: %w", err)
	}
	room.IsActive = false
	room.LastListenerError = nil
	room.ListenerStartedAt = nil
	room.UpdatedAt = time.Now().UTC()
	return room, nil
}

func (s *StreamerClipRoomService) IngestChatMessage(ctx context.Context, msg TwitchChatClipMessage) ([]models.StreamerClipRoomItem, error) {
	room, err := s.rooms.GetRoomByID(ctx, msg.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to load streamer clip room: %w", err)
	}
	if room == nil {
		return nil, fmt.Errorf("streamer clip room not found")
	}

	links := ExtractClipLinks(msg.MessageText)
	if len(links) == 0 {
		return nil, nil
	}

	items := make([]models.StreamerClipRoomItem, 0, len(links))
	for _, link := range links {
		item, err := s.ingestLink(ctx, room, msg, link)
		if err != nil {
			return nil, err
		}
		item, err = s.hydrateItem(ctx, item)
		if err != nil {
			return nil, err
		}
		s.broadcastItemDetected(item)
		items = append(items, *item)
	}

	return items, nil
}

func (s *StreamerClipRoomService) ApproveItem(ctx context.Context, actorUserID, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	if _, err := s.requireOwner(ctx, actorUserID, roomID); err != nil {
		return nil, err
	}
	item, err := s.rooms.ApproveItem(ctx, roomID, itemID, actorUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("streamer clip room item not found")
		}
		return nil, fmt.Errorf("failed to approve streamer clip room item: %w", err)
	}
	return s.hydrateItem(ctx, item)
}

func (s *StreamerClipRoomService) RejectItem(ctx context.Context, actorUserID, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	if _, err := s.requireOwner(ctx, actorUserID, roomID); err != nil {
		return nil, err
	}
	item, err := s.rooms.RejectItem(ctx, roomID, itemID, actorUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("streamer clip room item not found")
		}
		return nil, fmt.Errorf("failed to reject streamer clip room item: %w", err)
	}
	return s.hydrateItem(ctx, item)
}

func (s *StreamerClipRoomService) MarkListenerStopped(roomID uuid.UUID, err error) {
	if s == nil || err == nil {
		return
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return
	}
	_ = s.rooms.SetRoomActive(context.Background(), roomID, false, &message)
	if s.eventBroadcaster != nil {
		s.eventBroadcaster(roomID, "room_status_changed", map[string]interface{}{
			"room_id":             roomID.String(),
			"is_active":           false,
			"last_listener_error": message,
		})
	}
}

func (s *StreamerClipRoomService) ReorderApprovedItems(ctx context.Context, actorUserID, roomID uuid.UUID, itemIDs []uuid.UUID) error {
	if _, err := s.requireOwner(ctx, actorUserID, roomID); err != nil {
		return err
	}
	if err := s.rooms.ReorderApprovedItems(ctx, roomID, itemIDs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("streamer clip room item not found")
		}
		return fmt.Errorf("failed to reorder streamer clip room items: %w", err)
	}
	return nil
}

func (s *StreamerClipRoomService) requireOwner(ctx context.Context, actorUserID, roomID uuid.UUID) (*models.StreamerClipRoom, error) {
	room, err := s.rooms.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil || room.OwnerUserID != actorUserID {
		return nil, ErrStreamerClipRoomForbidden
	}
	return room, nil
}

func (s *StreamerClipRoomService) ingestLink(ctx context.Context, room *models.StreamerClipRoom, msg TwitchChatClipMessage, link DetectedClipLink) (*models.StreamerClipRoomItem, error) {
	item := s.newItem(room.ID, msg, link)

	clip, skipReason, err := s.resolveClip(ctx, link)
	if err != nil {
		return nil, err
	}
	if clip != nil {
		clipID := clip.ID
		item.ClipID = &clipID
		item.Clip = clip
	}
	if skipReason != nil {
		item.Status = models.StreamerClipRoomItemStatusSkipped
		item.SkipReason = skipReason
	}

	if err := s.rooms.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create streamer clip room item: %w", err)
	}

	if room.ApprovalMode == models.StreamerClipRoomApprovalAuto && item.Status == models.StreamerClipRoomItemStatusPending {
		approved, err := s.rooms.ApproveItem(ctx, room.ID, item.ID, room.OwnerUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-approve streamer clip room item: %w", err)
		}
		return approved, nil
	}

	return item, nil
}

func (s *StreamerClipRoomService) hydrateItem(ctx context.Context, item *models.StreamerClipRoomItem) (*models.StreamerClipRoomItem, error) {
	if s == nil || item == nil || item.ClipID == nil || s.clips == nil {
		return item, nil
	}

	clips, err := s.clips.GetByIDs(ctx, []uuid.UUID{*item.ClipID})
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate streamer clip room item: %w", err)
	}
	if len(clips) == 0 {
		return item, nil
	}
	clip := clips[0]
	item.Clip = &clip
	return item, nil
}

func (s *StreamerClipRoomService) hydrateItems(ctx context.Context, items []models.StreamerClipRoomItem) ([]models.StreamerClipRoomItem, error) {
	if s == nil || len(items) == 0 || s.clips == nil {
		return items, nil
	}

	clipIDs := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item.ClipID == nil {
			continue
		}
		if _, ok := seen[*item.ClipID]; ok {
			continue
		}
		seen[*item.ClipID] = struct{}{}
		clipIDs = append(clipIDs, *item.ClipID)
	}

	if len(clipIDs) == 0 {
		return items, nil
	}

	clips, err := s.clips.GetByIDs(ctx, clipIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate streamer clip room items: %w", err)
	}
	clipMap := make(map[uuid.UUID]models.Clip, len(clips))
	for _, clip := range clips {
		clipMap[clip.ID] = clip
	}

	for i := range items {
		if items[i].ClipID == nil {
			continue
		}
		if clip, ok := clipMap[*items[i].ClipID]; ok {
			clipCopy := clip
			items[i].Clip = &clipCopy
		}
	}

	return items, nil
}

func (s *StreamerClipRoomService) broadcastItemDetected(item *models.StreamerClipRoomItem) {
	if s == nil || s.eventBroadcaster == nil || item == nil {
		return
	}
	s.eventBroadcaster(item.RoomID, "item_detected", map[string]interface{}{"item": item})
}

func (s *StreamerClipRoomService) newItem(roomID uuid.UUID, msg TwitchChatClipMessage, link DetectedClipLink) *models.StreamerClipRoomItem {
	item := &models.StreamerClipRoomItem{
		ID:              uuid.New(),
		RoomID:          roomID,
		SourceURL:       link.SourceURL,
		SourceType:      link.SourceType,
		Status:          models.StreamerClipRoomItemStatusPending,
		DetectedAt:      time.Now().UTC(),
		MessageText:     stringPtr(msg.MessageText),
		TwitchMessageID: stringPtr(msg.TwitchMessageID),
		TwitchUserID:    stringPtr(msg.TwitchUserID),
		TwitchUsername:  stringPtr(msg.TwitchUsername),
	}
	return item
}

func (s *StreamerClipRoomService) resolveClip(ctx context.Context, link DetectedClipLink) (*models.Clip, *string, error) {
	const (
		clipNotImported = "clip_not_imported"
		clipUnavailable = "clip_unavailable"
	)

	switch link.SourceType {
	case models.StreamerClipRoomSourceClpr:
		clipID, err := uuid.Parse(strings.TrimSpace(link.ClprClipID))
		if err != nil {
			reason := clipNotImported
			return nil, &reason, nil
		}
		clip, err := s.clips.GetByID(ctx, clipID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				reason := clipNotImported
				return nil, &reason, nil
			}
			return nil, nil, fmt.Errorf("failed to resolve clip %s: %w", link.SourceURL, err)
		}
		if clip == nil {
			reason := clipNotImported
			return nil, &reason, nil
		}
		if isUnavailableClip(clip) {
			reason := clipUnavailable
			return clip, &reason, nil
		}
		return clip, nil, nil
	case models.StreamerClipRoomSourceTwitch:
		clip, err := s.clips.GetByTwitchClipID(ctx, link.TwitchClipID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				reason := clipNotImported
				return nil, &reason, nil
			}
			return nil, nil, fmt.Errorf("failed to resolve clip %s: %w", link.SourceURL, err)
		}
		if clip == nil {
			reason := clipNotImported
			return nil, &reason, nil
		}
		if isUnavailableClip(clip) {
			reason := clipUnavailable
			return clip, &reason, nil
		}
		return clip, nil, nil
	default:
		reason := clipNotImported
		return nil, &reason, nil
	}
}

func isUnavailableClip(clip *models.Clip) bool {
	if clip == nil {
		return false
	}
	if clip.IsRemoved || clip.IsHidden || clip.DMCARemoved {
		return true
	}
	return clip.RemovedReason != nil && strings.TrimSpace(*clip.RemovedReason) != ""
}
