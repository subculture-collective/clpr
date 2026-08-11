package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTwitchAuthLookup struct{ auth *models.TwitchAuth }

func (s stubTwitchAuthLookup) GetTwitchAuthByTwitchUserID(context.Context, string) (*models.TwitchAuth, error) {
	return s.auth, nil
}

type stubClipDownloadURLProvider struct{ downloadURL string }

func (s stubClipDownloadURLProvider) GetDownloadURL(context.Context, string, string, string) (string, error) {
	return s.downloadURL, nil
}

type stubRemoteAudioExtractor struct{ wavPath string }

func (s stubRemoteAudioExtractor) Extract(context.Context, string) (string, error) {
	return s.wavPath, nil
}

type stubWhisperTranscriber struct{ result *WhisperResult }

func (s stubWhisperTranscriber) TranscribeAudio(context.Context, string) (*WhisperResult, error) {
	return s.result, nil
}

func TestClipTranscriptionServiceTranscribesAuthorizedBroadcasterClip(t *testing.T) {
	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "clip.wav")
	require.NoError(t, os.WriteFile(wavPath, []byte("audio"), 0644))
	broadcasterID := "broadcaster-123"

	service := NewClipTranscriptionService(
		stubTwitchAuthLookup{auth: &models.TwitchAuth{
			TwitchUserID: broadcasterID, AccessToken: "streamer-token",
			Scopes: "channel:bot channel:manage:clips", ExpiresAt: time.Now().Add(time.Hour),
		}},
		stubClipDownloadURLProvider{downloadURL: "https://production.assets.clips.twitchcdn.net/clip.mp4"},
		stubRemoteAudioExtractor{wavPath: wavPath},
		stubWhisperTranscriber{result: &WhisperResult{Language: "en", FullText: "what a save"}},
	)

	result, err := service.TranscribeClip(context.Background(), &models.Clip{
		TwitchClipID: "clip-456", BroadcasterID: &broadcasterID,
	})

	require.NoError(t, err)
	assert.Equal(t, "what a save", result.FullText)
	_, statErr := os.Stat(wavPath)
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}
