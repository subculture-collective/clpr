package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type fakeStreamerClipRoomRepository struct {
	roomsByID            map[uuid.UUID]*models.StreamerClipRoom
	roomIDsByOwnerAndKey map[string]uuid.UUID
	itemsByRoomID        map[uuid.UUID]map[uuid.UUID]*models.StreamerClipRoomItem
	createCalls          int
	approveCalls         int
	rejectCalls          int
	reorderCalls         int
}

func newFakeStreamerClipRoomRepository() *fakeStreamerClipRoomRepository {
	return &fakeStreamerClipRoomRepository{
		roomsByID:            map[uuid.UUID]*models.StreamerClipRoom{},
		roomIDsByOwnerAndKey: map[string]uuid.UUID{},
		itemsByRoomID:        map[uuid.UUID]map[uuid.UUID]*models.StreamerClipRoomItem{},
	}
}

func (f *fakeStreamerClipRoomRepository) seedRoom(room *models.StreamerClipRoom) {
	clone := *room
	f.roomsByID[room.ID] = &clone
	f.roomIDsByOwnerAndKey[f.roomKey(room.OwnerUserID, room.TwitchChannel)] = room.ID
	if _, ok := f.itemsByRoomID[room.ID]; !ok {
		f.itemsByRoomID[room.ID] = map[uuid.UUID]*models.StreamerClipRoomItem{}
	}
}

func (f *fakeStreamerClipRoomRepository) roomKey(ownerID uuid.UUID, channel string) string {
	return ownerID.String() + "|" + channel
}

func (f *fakeStreamerClipRoomRepository) GetOrCreateRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error) {
	key := f.roomKey(ownerUserID, channel)
	if id, ok := f.roomIDsByOwnerAndKey[key]; ok {
		return f.cloneRoom(f.roomsByID[id]), nil
	}
	room := &models.StreamerClipRoom{
		ID:            uuid.New(),
		OwnerUserID:   ownerUserID,
		TwitchChannel: channel,
		ApprovalMode:  models.StreamerClipRoomApprovalManual,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	f.seedRoom(room)
	return f.cloneRoom(room), nil
}

func (f *fakeStreamerClipRoomRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.StreamerClipRoom, error) {
	room, ok := f.roomsByID[roomID]
	if !ok {
		return nil, nil
	}
	return f.cloneRoom(room), nil
}

func (f *fakeStreamerClipRoomRepository) SetRoomActive(ctx context.Context, roomID uuid.UUID, active bool, listenerError *string) error {
	room, ok := f.roomsByID[roomID]
	if !ok {
		return nil
	}
	room.IsActive = active
	room.LastListenerError = listenerError
	if active {
		now := time.Now().UTC()
		room.ListenerStartedAt = &now
	} else {
		room.ListenerStartedAt = nil
	}
	room.UpdatedAt = time.Now().UTC()
	return nil
}

func (f *fakeStreamerClipRoomRepository) CreateItem(ctx context.Context, item *models.StreamerClipRoomItem) error {
	f.createCalls++
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.DetectedAt.IsZero() {
		item.DetectedAt = time.Now().UTC()
	}
	if _, ok := f.itemsByRoomID[item.RoomID]; !ok {
		f.itemsByRoomID[item.RoomID] = map[uuid.UUID]*models.StreamerClipRoomItem{}
	}
	clone := *item
	cloneCopy := clone
	f.itemsByRoomID[item.RoomID][item.ID] = &cloneCopy
	item.CreatedAt = time.Now().UTC()
	item.UpdatedAt = item.CreatedAt
	return nil
}

func (f *fakeStreamerClipRoomRepository) ListItems(ctx context.Context, roomID uuid.UUID, status string, limit int) ([]models.StreamerClipRoomItem, error) {
	items := f.itemsByRoomID[roomID]
	results := make([]models.StreamerClipRoomItem, 0, len(items))
	for _, item := range items {
		if status != "" && status != "all" && item.Status != status {
			continue
		}
		results = append(results, *f.cloneItem(item))
	}
	return results, nil
}

func (f *fakeStreamerClipRoomRepository) GetItem(ctx context.Context, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	items := f.itemsByRoomID[roomID]
	item, ok := items[itemID]
	if !ok {
		return nil, nil
	}
	return f.cloneItem(item), nil
}

func (f *fakeStreamerClipRoomRepository) ApproveItem(ctx context.Context, roomID, itemID, approverID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	f.approveCalls++
	item, ok := f.itemsByRoomID[roomID][itemID]
	if !ok {
		return nil, errors.New("not found")
	}
	nextPosition := 1
	for _, existing := range f.itemsByRoomID[roomID] {
		if existing.Status == models.StreamerClipRoomItemStatusApproved && existing.Position != nil && *existing.Position >= nextPosition {
			nextPosition = *existing.Position + 1
		}
	}
	pos := nextPosition
	now := time.Now().UTC()
	item.Status = models.StreamerClipRoomItemStatusApproved
	item.Position = &pos
	item.ApprovedAt = &now
	item.ApprovedByUserID = &approverID
	item.RejectedAt = nil
	item.RejectedByUserID = nil
	item.UpdatedAt = now
	return f.cloneItem(item), nil
}

func (f *fakeStreamerClipRoomRepository) RejectItem(ctx context.Context, roomID, itemID, rejecterID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	f.rejectCalls++
	item, ok := f.itemsByRoomID[roomID][itemID]
	if !ok {
		return nil, errors.New("not found")
	}
	now := time.Now().UTC()
	item.Status = models.StreamerClipRoomItemStatusRejected
	item.Position = nil
	item.ApprovedAt = nil
	item.ApprovedByUserID = nil
	item.RejectedAt = &now
	item.RejectedByUserID = &rejecterID
	item.UpdatedAt = now
	return f.cloneItem(item), nil
}

func (f *fakeStreamerClipRoomRepository) ReorderApprovedItems(ctx context.Context, roomID uuid.UUID, itemIDs []uuid.UUID) error {
	f.reorderCalls++
	for idx, itemID := range itemIDs {
		item, ok := f.itemsByRoomID[roomID][itemID]
		if !ok {
			return errors.New("not found")
		}
		pos := idx + 1
		item.Position = &pos
	}
	return nil
}

func (f *fakeStreamerClipRoomRepository) cloneRoom(room *models.StreamerClipRoom) *models.StreamerClipRoom {
	if room == nil {
		return nil
	}
	clone := *room
	return &clone
}

func (f *fakeStreamerClipRoomRepository) cloneItem(item *models.StreamerClipRoomItem) *models.StreamerClipRoomItem {
	if item == nil {
		return nil
	}
	clone := *item
	return &clone
}

type fakeClipLookupRepository struct {
	clipsByID          map[uuid.UUID]*models.Clip
	clipsByTwitchID    map[string]*models.Clip
	getByIDCalls       int
	getByIDsCalls      int
	getByTwitchIDCalls int
}

func newFakeClipLookupRepository() *fakeClipLookupRepository {
	return &fakeClipLookupRepository{
		clipsByID:       map[uuid.UUID]*models.Clip{},
		clipsByTwitchID: map[string]*models.Clip{},
	}
}

func (f *fakeClipLookupRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Clip, error) {
	f.getByIDCalls++
	clip, ok := f.clipsByID[id]
	if !ok {
		return nil, nil
	}
	clone := *clip
	return &clone, nil
}

func (f *fakeClipLookupRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Clip, error) {
	f.getByIDsCalls++
	clips := make([]models.Clip, 0, len(ids))
	for _, id := range ids {
		clip, ok := f.clipsByID[id]
		if !ok {
			continue
		}
		clone := *clip
		clips = append(clips, clone)
	}
	return clips, nil
}

func (f *fakeClipLookupRepository) GetByTwitchClipID(ctx context.Context, twitchClipID string) (*models.Clip, error) {
	f.getByTwitchIDCalls++
	clip, ok := f.clipsByTwitchID[twitchClipID]
	if !ok {
		return nil, nil
	}
	clone := *clip
	return &clone, nil
}

func TestStreamerClipRoomServiceIngestsClprClipAsPending(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: ownerID, TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalManual, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clipID := uuid.New()

	rooms := newFakeStreamerClipRoomRepository()
	rooms.seedRoom(room)
	clips := newFakeClipLookupRepository()
	clips.clipsByID[clipID] = &models.Clip{ID: clipID, TwitchClipID: "clpr-1", TwitchClipURL: "https://clpr.tv/clip/clpr-1", Title: "clip"}

	svc := newStreamerClipRoomService(rooms, clips)
	items, err := svc.IngestChatMessage(ctx, TwitchChatClipMessage{
		RoomID:          room.ID,
		TwitchMessageID: "msg-1",
		TwitchUserID:    "123",
		TwitchUsername:  "viewer",
		MessageText:     "watch https://clpr.tv/clip/" + clipID.String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.Status != models.StreamerClipRoomItemStatusPending {
		t.Fatalf("expected pending status, got %s", item.Status)
	}
	if item.ClipID == nil || *item.ClipID != clipID {
		t.Fatalf("expected clip id %s, got %#v", clipID, item.ClipID)
	}
	if item.ID == uuid.Nil {
		t.Fatal("expected non-zero item ID")
	}
	if item.DetectedAt.IsZero() {
		t.Fatal("expected detected_at to be set")
	}
	if item.SkipReason != nil {
		t.Fatalf("expected no skip reason, got %q", *item.SkipReason)
	}
}

func TestStreamerClipRoomServiceHydratesListedItemsWithClip(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: ownerID, TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalManual, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clipID := uuid.New()

	rooms := newFakeStreamerClipRoomRepository()
	rooms.seedRoom(room)
	itemID := uuid.New()
	rooms.itemsByRoomID[room.ID][itemID] = &models.StreamerClipRoomItem{ID: itemID, RoomID: room.ID, ClipID: &clipID, Status: models.StreamerClipRoomItemStatusApproved, DetectedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clips := newFakeClipLookupRepository()
	clips.clipsByID[clipID] = &models.Clip{ID: clipID, TwitchClipID: "clpr-3", TwitchClipURL: "https://clpr.tv/clip/clpr-3", Title: "hydrated clip"}

	svc := newStreamerClipRoomService(rooms, clips)
	items, err := svc.ListItems(ctx, ownerID, room.ID, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Clip == nil || items[0].Clip.ID != clipID {
		t.Fatalf("expected hydrated clip %s, got %#v", clipID, items[0].Clip)
	}
	if clips.getByIDsCalls == 0 {
		t.Fatal("expected batch clip lookup to be used")
	}
}

func TestStreamerClipRoomServiceBroadcastsDetectedItems(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: ownerID, TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalManual, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clipID := uuid.New()

	rooms := newFakeStreamerClipRoomRepository()
	rooms.seedRoom(room)
	clips := newFakeClipLookupRepository()
	clips.clipsByID[clipID] = &models.Clip{ID: clipID, TwitchClipID: "clpr-4", TwitchClipURL: "https://clpr.tv/clip/clpr-4", Title: "clip"}

	svc := newStreamerClipRoomService(rooms, clips)
	var gotType string
	var gotRoomID uuid.UUID
	var gotItem *models.StreamerClipRoomItem
	svc.SetEventBroadcaster(func(roomID uuid.UUID, eventType string, data map[string]interface{}) {
		gotRoomID = roomID
		gotType = eventType
		if item, ok := data["item"].(*models.StreamerClipRoomItem); ok {
			gotItem = item
		}
	})

	items, err := svc.IngestChatMessage(ctx, TwitchChatClipMessage{
		RoomID:          room.ID,
		TwitchMessageID: "msg-4",
		TwitchUserID:    "123",
		TwitchUsername:  "viewer",
		MessageText:     "https://clpr.tv/clip/" + clipID.String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotType != "item_detected" || gotRoomID != room.ID {
		t.Fatalf("unexpected broadcast: type=%s room=%s", gotType, gotRoomID)
	}
	if gotItem == nil || gotItem.Clip == nil || gotItem.Clip.ID != clipID {
		t.Fatalf("expected broadcast item with hydrated clip, got %#v", gotItem)
	}
	if len(items) != 1 || items[0].Clip == nil || items[0].Clip.ID != clipID {
		t.Fatalf("expected returned item to remain hydrated, got %#v", items)
	}
}

func TestStreamerClipRoomServiceHydratesApprovedItem(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: ownerID, TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalManual, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clipID := uuid.New()
	itemID := uuid.New()

	rooms := newFakeStreamerClipRoomRepository()
	rooms.seedRoom(room)
	rooms.itemsByRoomID[room.ID][itemID] = &models.StreamerClipRoomItem{ID: itemID, RoomID: room.ID, ClipID: &clipID, Status: models.StreamerClipRoomItemStatusPending, DetectedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clips := newFakeClipLookupRepository()
	clips.clipsByID[clipID] = &models.Clip{ID: clipID, TwitchClipID: "clpr-5", TwitchClipURL: "https://clpr.tv/clip/clpr-5", Title: "approved clip"}

	svc := newStreamerClipRoomService(rooms, clips)
	item, err := svc.ApproveItem(ctx, ownerID, room.ID, itemID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Clip == nil || item.Clip.ID != clipID {
		t.Fatalf("expected hydrated approved item clip, got %#v", item.Clip)
	}
}

func TestStreamerClipRoomServiceMarksListenerStopped(t *testing.T) {
	rooms := newFakeStreamerClipRoomRepository()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: uuid.New(), TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalManual, IsActive: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	rooms.seedRoom(room)

	svc := newStreamerClipRoomService(rooms, newFakeClipLookupRepository())
	svc.MarkListenerStopped(room.ID, errors.New("irc disconnected"))
	updated := rooms.roomsByID[room.ID]
	if updated.IsActive {
		t.Fatal("expected room to be marked inactive")
	}
	if updated.LastListenerError == nil || *updated.LastListenerError != "irc disconnected" {
		t.Fatalf("expected listener error to be stored, got %#v", updated.LastListenerError)
	}
}

func TestStreamerClipRoomServiceSkipsUnknownTwitchClip(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: ownerID, TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalManual, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}

	rooms := newFakeStreamerClipRoomRepository()
	rooms.seedRoom(room)
	clips := newFakeClipLookupRepository()

	svc := newStreamerClipRoomService(rooms, clips)
	items, err := svc.IngestChatMessage(ctx, TwitchChatClipMessage{
		RoomID:          room.ID,
		TwitchMessageID: "msg-2",
		TwitchUserID:    "123",
		TwitchUsername:  "viewer",
		MessageText:     "https://clips.twitch.tv/UnknownSlug",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.Status != models.StreamerClipRoomItemStatusSkipped {
		t.Fatalf("expected skipped status, got %s", item.Status)
	}
	if item.SkipReason == nil || *item.SkipReason != "clip_not_imported" {
		t.Fatalf("expected clip_not_imported skip reason, got %#v", item.SkipReason)
	}
	if item.ClipID != nil {
		t.Fatalf("expected no clip id, got %#v", item.ClipID)
	}
}

func TestStreamerClipRoomServiceAutoApprovesWhenConfigured(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: ownerID, TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalAuto, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clipID := uuid.New()

	rooms := newFakeStreamerClipRoomRepository()
	rooms.seedRoom(room)
	clips := newFakeClipLookupRepository()
	clips.clipsByID[clipID] = &models.Clip{ID: clipID, TwitchClipID: "clpr-2", TwitchClipURL: "https://clpr.tv/clip/clpr-2", Title: "clip"}

	svc := newStreamerClipRoomService(rooms, clips)
	items, err := svc.IngestChatMessage(ctx, TwitchChatClipMessage{
		RoomID:          room.ID,
		TwitchMessageID: "msg-3",
		TwitchUserID:    "123",
		TwitchUsername:  "viewer",
		MessageText:     "https://clpr.tv/clip/" + clipID.String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.Status != models.StreamerClipRoomItemStatusApproved {
		t.Fatalf("expected approved status, got %s", item.Status)
	}
	if item.Position == nil || *item.Position != 1 {
		t.Fatalf("expected position 1, got %#v", item.Position)
	}
	if rooms.createCalls != 1 {
		t.Fatalf("expected 1 create call, got %d", rooms.createCalls)
	}
	if rooms.approveCalls != 1 {
		t.Fatalf("expected 1 approve call, got %d", rooms.approveCalls)
	}
}

func TestStreamerClipRoomServiceRejectsNonOwnerApproval(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	room := &models.StreamerClipRoom{ID: uuid.New(), OwnerUserID: ownerID, TwitchChannel: "moonmoon", ApprovalMode: models.StreamerClipRoomApprovalManual, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	itemID := uuid.New()

	rooms := newFakeStreamerClipRoomRepository()
	rooms.seedRoom(room)
	rooms.itemsByRoomID[room.ID][itemID] = &models.StreamerClipRoomItem{ID: itemID, RoomID: room.ID, Status: models.StreamerClipRoomItemStatusPending, DetectedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	clips := newFakeClipLookupRepository()

	svc := newStreamerClipRoomService(rooms, clips)
	_, err := svc.ApproveItem(ctx, uuid.New(), room.ID, itemID)
	if !errors.Is(err, ErrStreamerClipRoomForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	if rooms.approveCalls != 0 {
		t.Fatalf("expected no approve calls, got %d", rooms.approveCalls)
	}
}
