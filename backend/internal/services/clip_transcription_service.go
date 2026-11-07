package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

var ErrClipTranscriptionNotAuthorized = errors.New("clip broadcaster has not authorized transcription")

// ErrTranscriptionProviderUnavailable marks a failure that is not about the
// clip: the authorization store, the Twitch download API or the local ffmpeg
// or Whisper runtime is unavailable. The service pauses after one, so queued
// clips are not each charged a failed attempt.
var ErrTranscriptionProviderUnavailable = errors.New("transcription provider is unavailable")

const transcriptionProviderBackoff = 15 * time.Minute

type TwitchAuthLookup interface {
	GetTwitchAuthByTwitchUserID(context.Context, string) (*models.TwitchAuth, error)
}

type ClipDownloadURLProvider interface {
	GetDownloadURL(context.Context, string, string, string) (string, error)
}

type RemoteAudioExtractor interface {
	Extract(context.Context, string) (string, error)
}

type WhisperTranscriber interface {
	TranscribeAudio(context.Context, string) (*WhisperResult, error)
}

// FFmpegRemoteAudioExtractor streams media into a temporary WAV file.
type FFmpegRemoteAudioExtractor struct {
	FFmpegPath string
	WorkDir    string
}

func (e *FFmpegRemoteAudioExtractor) Extract(ctx context.Context, mediaURL string) (string, error) {
	return ExtractAudioFromURL(ctx, e.FFmpegPath, mediaURL, e.WorkDir)
}

// ClipTranscriptionService hides authorization, official media access, audio
// normalization, cleanup, and Whisper behind one operation.
type ClipTranscriptionService struct {
	auth        TwitchAuthLookup
	downloads   ClipDownloadURLProvider
	extractor   RemoteAudioExtractor
	transcriber WhisperTranscriber

	pauseMu     sync.Mutex
	pausedUntil time.Time
	now         func() time.Time
}

func NewClipTranscriptionService(
	auth TwitchAuthLookup,
	downloads ClipDownloadURLProvider,
	extractor RemoteAudioExtractor,
	transcriber WhisperTranscriber,
) *ClipTranscriptionService {
	return &ClipTranscriptionService{
		auth: auth, downloads: downloads, extractor: extractor, transcriber: transcriber,
		now: time.Now,
	}
}

// Operational reports whether the service is configured and not paused after
// a provider-wide failure.
func (s *ClipTranscriptionService) Operational() bool {
	if s == nil || s.auth == nil || s.downloads == nil || s.extractor == nil || s.transcriber == nil {
		return false
	}
	s.pauseMu.Lock()
	defer s.pauseMu.Unlock()
	return !s.now().Before(s.pausedUntil)
}

// providerFailure pauses the service and marks err as provider-wide.
func (s *ClipTranscriptionService) providerFailure(err error) error {
	s.pauseMu.Lock()
	s.pausedUntil = s.now().Add(transcriptionProviderBackoff)
	s.pauseMu.Unlock()
	return fmt.Errorf("%w: %v", ErrTranscriptionProviderUnavailable, err)
}

func (s *ClipTranscriptionService) TranscribeClip(ctx context.Context, clip *models.Clip) (*WhisperResult, error) {
	if s == nil || s.auth == nil || s.downloads == nil || s.extractor == nil || s.transcriber == nil {
		return nil, fmt.Errorf("clip transcription service is not configured")
	}
	if !s.Operational() {
		return nil, fmt.Errorf("%w: paused after a provider failure", ErrTranscriptionProviderUnavailable)
	}
	if clip == nil || clip.BroadcasterID == nil || strings.TrimSpace(*clip.BroadcasterID) == "" {
		return nil, ErrClipTranscriptionNotAuthorized
	}
	broadcasterID := strings.TrimSpace(*clip.BroadcasterID)
	auth, err := s.auth.GetTwitchAuthByTwitchUserID(ctx, broadcasterID)
	if err != nil {
		return nil, s.providerFailure(fmt.Errorf("loading broadcaster Twitch authorization: %w", err))
	}
	if auth == nil || !hasScope(auth.Scopes, "channel:manage:clips") || !auth.ExpiresAt.After(time.Now()) {
		return nil, ErrClipTranscriptionNotAuthorized
	}

	mediaURL, err := s.downloads.GetDownloadURL(ctx, broadcasterID, clip.TwitchClipID, auth.AccessToken)
	if errors.Is(err, ErrTwitchClipDownloadUnavailable) {
		return nil, s.providerFailure(err)
	}
	if err != nil {
		return nil, fmt.Errorf("getting authorized Twitch clip media: %w", err)
	}
	wavPath, err := s.extractor.Extract(ctx, mediaURL)
	if errors.Is(err, exec.ErrNotFound) {
		return nil, s.providerFailure(fmt.Errorf("streaming clip audio: %w", err))
	}
	if err != nil {
		return nil, fmt.Errorf("streaming clip audio: %w", err)
	}
	defer os.Remove(wavPath)

	result, err := s.transcriber.TranscribeAudio(ctx, wavPath)
	if errors.Is(err, exec.ErrNotFound) {
		return nil, s.providerFailure(err)
	}
	if err != nil {
		return nil, err
	}
	if result == nil || strings.TrimSpace(result.FullText) == "" {
		return nil, fmt.Errorf("Whisper returned an empty transcript")
	}
	result.FullText = strings.TrimSpace(result.FullText)
	return result, nil
}

func hasScope(scopes, required string) bool {
	for _, scope := range strings.Fields(scopes) {
		if scope == required {
			return true
		}
	}
	return false
}
