package services

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
)

type queueStoreStub struct {
	count        int
	item         *models.QueueItem
	convertErr   error
	converted    *models.Playlist
	onlyUnplayed bool
	clearQueue   bool
}

func (s *queueStoreStub) GetUserQueue(context.Context, uuid.UUID, int) ([]models.QueueItemWithClip, error) {
	return nil, nil
}
func (s *queueStoreStub) GetQueueCount(context.Context, uuid.UUID) (int, error) { return s.count, nil }
func (s *queueStoreStub) AddItem(context.Context, *models.QueueItem) error      { return nil }
func (s *queueStoreStub) AddItemAtTop(context.Context, *models.QueueItem) error { return nil }
func (s *queueStoreStub) RemoveItem(context.Context, uuid.UUID, uuid.UUID) error {
	return repository.ErrQueueItemNotFound
}
func (s *queueStoreStub) GetItemByID(context.Context, uuid.UUID, uuid.UUID) (*models.QueueItem, error) {
	return s.item, nil
}
func (s *queueStoreStub) ReorderItem(context.Context, uuid.UUID, uuid.UUID, int) error { return nil }
func (s *queueStoreStub) MarkAsPlayed(context.Context, uuid.UUID, uuid.UUID) error {
	return repository.ErrQueueItemNotFound
}
func (s *queueStoreStub) ClearQueue(context.Context, uuid.UUID) error { return nil }
func (s *queueStoreStub) ConvertToPlaylist(_ context.Context, playlist *models.Playlist, onlyUnplayed, clearQueue bool) error {
	s.converted, s.onlyUnplayed, s.clearQueue = playlist, onlyUnplayed, clearQueue
	return s.convertErr
}

type queueClipStoreStub struct{ clip *models.Clip }

func (s queueClipStoreStub) GetByID(context.Context, uuid.UUID) (*models.Clip, error) {
	return s.clip, nil
}

func TestQueueServiceAddRejectsFullQueue(t *testing.T) {
	store := &queueStoreStub{count: 500}
	service := &QueueService{queueRepo: store, clipRepo: queueClipStoreStub{}}
	_, err := service.AddToQueue(context.Background(), uuid.New(), &models.AddToQueueRequest{ClipID: uuid.NewString()})
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}

func TestQueueServiceMissingItemSentinels(t *testing.T) {
	service := &QueueService{queueRepo: &queueStoreStub{}}
	if err := service.RemoveFromQueue(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, ErrQueueItemNotFound) {
		t.Fatalf("remove: %v", err)
	}
	if err := service.MarkAsPlayed(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, ErrQueueItemNotFound) {
		t.Fatalf("played: %v", err)
	}
	if err := service.ReorderQueue(context.Background(), uuid.New(), &models.ReorderQueueRequest{ItemID: uuid.NewString(), NewPosition: 1}); !errors.Is(err, ErrQueueItemNotFound) {
		t.Fatalf("reorder: %v", err)
	}
}

func TestQueueServiceConvertDelegatesAtomicOptions(t *testing.T) {
	store := &queueStoreStub{}
	service := &QueueService{queueRepo: store}
	description := "saved queue"
	playlist, err := service.ConvertQueueToPlaylist(context.Background(), uuid.New(), &models.ConvertQueueToPlaylistRequest{Title: "  Launch list  ", Description: &description, OnlyUnplayed: true, ClearQueue: true})
	if err != nil {
		t.Fatal(err)
	}
	if playlist != store.converted || playlist.Title != "Launch list" || !store.onlyUnplayed || !store.clearQueue {
		t.Fatalf("conversion was not delegated correctly: %#v", store)
	}
}

func TestQueueServiceConvertMapsEmptySentinel(t *testing.T) {
	service := &QueueService{queueRepo: &queueStoreStub{convertErr: repository.ErrQueueEmpty}}
	_, err := service.ConvertQueueToPlaylist(context.Background(), uuid.New(), &models.ConvertQueueToPlaylistRequest{Title: "Playlist"})
	if !errors.Is(err, ErrQueueEmpty) {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}
}
