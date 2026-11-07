package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

type failingClipDownloadURLProvider struct {
	err   error
	calls int
}

func (s *failingClipDownloadURLProvider) GetDownloadURL(context.Context, string, string, string) (string, error) {
	s.calls++
	return "", s.err
}

type failingWhisperTranscriber struct{ err error }

func (s failingWhisperTranscriber) TranscribeAudio(context.Context, string) (*WhisperResult, error) {
	return nil, s.err
}

func authorizedTranscriptionAuth() stubTwitchAuthLookup {
	return stubTwitchAuthLookup{auth: &models.TwitchAuth{
		TwitchUserID: "broadcaster-123", AccessToken: "streamer-token",
		Scopes: "channel:manage:clips", ExpiresAt: time.Now().Add(time.Hour),
	}}
}

func TestClipTranscriptionPausesAfterProviderFailure(t *testing.T) {
	broadcasterID := "broadcaster-123"
	clip := &models.Clip{TwitchClipID: "clip-456", BroadcasterID: &broadcasterID}
	downloads := &failingClipDownloadURLProvider{err: fmt.Errorf("%w: status 503", ErrTwitchClipDownloadUnavailable)}
	service := NewClipTranscriptionService(authorizedTranscriptionAuth(), downloads, stubRemoteAudioExtractor{}, stubWhisperTranscriber{})
	now := time.Now()
	service.now = func() time.Time { return now }

	_, err := service.TranscribeClip(context.Background(), clip)
	require.ErrorIs(t, err, ErrTranscriptionProviderUnavailable)
	assert.False(t, service.Operational())

	// While paused, the provider is not called again.
	_, err = service.TranscribeClip(context.Background(), clip)
	require.ErrorIs(t, err, ErrTranscriptionProviderUnavailable)
	assert.Equal(t, 1, downloads.calls)

	now = now.Add(transcriptionProviderBackoff)
	assert.True(t, service.Operational())
	_, _ = service.TranscribeClip(context.Background(), clip)
	assert.Equal(t, 2, downloads.calls)
}

func TestClipTranscriptionClipFailuresDoNotPause(t *testing.T) {
	broadcasterID := "broadcaster-123"
	clip := &models.Clip{TwitchClipID: "clip-456", BroadcasterID: &broadcasterID}
	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "clip.wav")
	require.NoError(t, os.WriteFile(wavPath, []byte("audio"), 0644))

	for name, service := range map[string]*ClipTranscriptionService{
		"clip not found on Twitch": NewClipTranscriptionService(authorizedTranscriptionAuth(),
			&failingClipDownloadURLProvider{err: errors.New("Twitch clip download API returned status 404")},
			stubRemoteAudioExtractor{}, stubWhisperTranscriber{}),
		"Whisper rejects the audio": NewClipTranscriptionService(authorizedTranscriptionAuth(),
			stubClipDownloadURLProvider{downloadURL: "https://production.assets.clips.twitchcdn.net/clip.mp4"},
			stubRemoteAudioExtractor{wavPath: wavPath},
			failingWhisperTranscriber{err: errors.New("whisper transcription failed: exit status 1")}),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.TranscribeClip(context.Background(), clip)
			require.Error(t, err)
			assert.NotErrorIs(t, err, ErrTranscriptionProviderUnavailable)
			assert.True(t, service.Operational())
		})
	}
}

func TestClipTranscriptionMissingWhisperRuntimePauses(t *testing.T) {
	broadcasterID := "broadcaster-123"
	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "clip.wav")
	require.NoError(t, os.WriteFile(wavPath, []byte("audio"), 0644))
	service := NewClipTranscriptionService(authorizedTranscriptionAuth(),
		stubClipDownloadURLProvider{downloadURL: "https://production.assets.clips.twitchcdn.net/clip.mp4"},
		stubRemoteAudioExtractor{wavPath: wavPath},
		failingWhisperTranscriber{err: fmt.Errorf("whisper transcription failed: %w", &exec.Error{Name: "python3", Err: exec.ErrNotFound})})

	_, err := service.TranscribeClip(context.Background(), &models.Clip{TwitchClipID: "clip-456", BroadcasterID: &broadcasterID})
	require.ErrorIs(t, err, ErrTranscriptionProviderUnavailable)
	assert.False(t, service.Operational())
}
