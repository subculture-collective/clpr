package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestPlaylistMembershipRejectsDuplicateIDsBeforeRepositoryWork(t *testing.T) {
	service := NewPlaylistService(nil, nil, "")
	id := uuid.New()
	for _, operation := range []func() error{
		func() error {
			return service.AddClipsToPlaylist(context.Background(), uuid.New(), uuid.New(), []uuid.UUID{id, id})
		},
		func() error {
			return service.ReorderPlaylistClips(context.Background(), uuid.New(), uuid.New(), []uuid.UUID{id, id})
		},
	} {
		if err := operation(); !errors.Is(err, ErrPlaylistMembershipMismatch) {
			t.Fatalf("expected membership mismatch, got %v", err)
		}
	}
}
