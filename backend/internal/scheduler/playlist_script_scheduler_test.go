package scheduler

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/google/uuid"
)

type playlistScriptSchedulerServiceStub struct {
	scripts            []*models.PlaylistScript
	generationErr      error
	acknowledged       []uuid.UUID
	acknowledgementErr error
}

func (s *playlistScriptSchedulerServiceStub) ListDueForExecution(context.Context) ([]*models.PlaylistScript, error) {
	return s.scripts, nil
}

func (s *playlistScriptSchedulerServiceStub) GeneratePlaylist(context.Context, uuid.UUID) (*models.Playlist, error) {
	return nil, s.generationErr
}

func (s *playlistScriptSchedulerServiceStub) AcknowledgeEmptyGeneration(_ context.Context, scriptID uuid.UUID) error {
	s.acknowledged = append(s.acknowledged, scriptID)
	return s.acknowledgementErr
}

func (s *playlistScriptSchedulerServiceStub) DeleteStaleGeneratedPlaylists(context.Context) (int64, error) {
	return 0, nil
}

func TestRunDueScriptsAcknowledgesEmptyGeneration(t *testing.T) {
	scriptID := uuid.New()
	service := &playlistScriptSchedulerServiceStub{
		scripts:       []*models.PlaylistScript{{ID: scriptID, Name: "Hidden Gems", Strategy: "sleeper_hits"}},
		generationErr: services.ErrPlaylistGenerationEmpty,
	}

	scheduler := NewPlaylistScriptScheduler(service, 1)
	scheduler.runDueScripts(context.Background())

	if len(service.acknowledged) != 1 || service.acknowledged[0] != scriptID {
		t.Fatalf("expected empty generation for %s to be acknowledged, got %v", scriptID, service.acknowledged)
	}
}

func TestRunDueScriptsDoesNotAcknowledgeRealGenerationFailure(t *testing.T) {
	service := &playlistScriptSchedulerServiceStub{
		scripts:       []*models.PlaylistScript{{ID: uuid.New(), Name: "Broken Script"}},
		generationErr: errors.New("database unavailable"),
	}

	scheduler := NewPlaylistScriptScheduler(service, 1)
	scheduler.runDueScripts(context.Background())

	if len(service.acknowledged) != 0 {
		t.Fatalf("expected real failures not to be acknowledged, got %v", service.acknowledged)
	}
}
