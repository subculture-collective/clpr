package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

var ErrClipTranscriptionNotAuthorized = errors.New("clip broadcaster has not authorized transcription")

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
}

func NewClipTranscriptionService(
	auth TwitchAuthLookup,
	downloads ClipDownloadURLProvider,
	extractor RemoteAudioExtractor,
	transcriber WhisperTranscriber,
) *ClipTranscriptionService {
	return &ClipTranscriptionService{
		auth: auth, downloads: downloads, extractor: extractor, transcriber: transcriber,
	}
}

func (s *ClipTranscriptionService) TranscribeClip(ctx context.Context, clip *models.Clip) (*WhisperResult, error) {
	if s == nil || s.auth == nil || s.downloads == nil || s.extractor == nil || s.transcriber == nil {
		return nil, fmt.Errorf("clip transcription service is not configured")
	}
	if clip == nil || clip.BroadcasterID == nil || strings.TrimSpace(*clip.BroadcasterID) == "" {
		return nil, ErrClipTranscriptionNotAuthorized
	}
	broadcasterID := strings.TrimSpace(*clip.BroadcasterID)
	auth, err := s.auth.GetTwitchAuthByTwitchUserID(ctx, broadcasterID)
	if err != nil {
		return nil, fmt.Errorf("loading broadcaster Twitch authorization: %w", err)
	}
	if auth == nil || !hasScope(auth.Scopes, "channel:manage:clips") || !auth.ExpiresAt.After(time.Now()) {
		return nil, ErrClipTranscriptionNotAuthorized
	}

	mediaURL, err := s.downloads.GetDownloadURL(ctx, broadcasterID, clip.TwitchClipID, auth.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("getting authorized Twitch clip media: %w", err)
	}
	wavPath, err := s.extractor.Extract(ctx, mediaURL)
	if err != nil {
		return nil, fmt.Errorf("streaming clip audio: %w", err)
	}
	defer os.Remove(wavPath)

	result, err := s.transcriber.TranscribeAudio(ctx, wavPath)
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
