package services

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

func TestShouldAutoCurateGeneratedPlaylist(t *testing.T) {
	tests := []struct {
		name     string
		script   *models.PlaylistScript
		ownerID  uuid.UUID
		expected bool
	}{
		{
			name: "bot-owned public playlists are curated",
			script: &models.PlaylistScript{
				Visibility: models.PlaylistVisibilityPublic,
			},
			ownerID:  BotUserID,
			expected: true,
		},
		{
			name: "bot-owned private playlists stay out of curated collections",
			script: &models.PlaylistScript{
				Visibility: models.PlaylistVisibilityPrivate,
			},
			ownerID:  BotUserID,
			expected: false,
		},
		{
			name: "public user-generated playlists are not auto-curated",
			script: &models.PlaylistScript{
				Visibility: models.PlaylistVisibilityPublic,
			},
			ownerID:  uuid.New(),
			expected: false,
		},
		{
			name:     "nil scripts are never curated",
			script:   nil,
			ownerID:  BotUserID,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := shouldAutoCurateGeneratedPlaylist(tt.script, tt.ownerID)
			if actual != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestPlaylistScriptUpdateRejectsEmptyRequestBeforeRepositoryWork(t *testing.T) {
	service := &PlaylistScriptService{}
	updated, err := service.UpdateScript(context.Background(), uuid.New(), &models.UpdatePlaylistScriptRequest{})
	if err == nil || updated != nil {
		t.Fatalf("expected empty update to fail, got script=%v err=%v", updated, err)
	}
}
