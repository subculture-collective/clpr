//go:build integration

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

type transcriptionAuthStub struct{}

func (transcriptionAuthStub) GetTwitchAuthByTwitchUserID(_ context.Context, id string) (*models.TwitchAuth, error) {
	return &models.TwitchAuth{TwitchUserID: id, AccessToken: "token", Scopes: "channel:manage:clips", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type transcriptionDownloadStub struct {
	err   error
	calls int
}

func (s *transcriptionDownloadStub) GetDownloadURL(context.Context, string, string, string) (string, error) {
	s.calls++
	return "", s.err
}

type unusedExtractor struct{}

func (unusedExtractor) Extract(context.Context, string) (string, error) { return "", errors.New("unused") }

type unusedTranscriber struct{}

func (unusedTranscriber) TranscribeAudio(context.Context, string) (*services.WhisperResult, error) {
	return nil, errors.New("unused")
}

// A Twitch outage charged every queued clip a failed transcription attempt,
// so three outages used up a clip's whole budget.
func TestTranscriptionProviderOutageKeepsClipAttempts(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "clips", "users")
	ctx := context.Background()
	clipRepo := repository.NewClipRepository(pool)

	broadcasterID := "bc-" + uuid.NewString()[:8]
	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, username, display_name) VALUES ($1, $2, $2)`, userID, broadcasterID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO twitch_auth (user_id, twitch_user_id, twitch_username, access_token, refresh_token, expires_at, scopes)
		VALUES ($1, $2, $2, 'a', 'r', NOW() + INTERVAL '1 hour', 'channel:manage:clips')`, userID, broadcasterID); err != nil {
		t.Fatalf("create twitch_auth: %v", err)
	}
	var clipIDs []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		clip := models.Clip{
			ID: id, TwitchClipID: id.String(), TwitchClipURL: "https://clips.twitch.tv/" + id.String(),
			EmbedURL: "https://clips.twitch.tv/embed?clip=" + id.String(), Title: fmt.Sprintf("clip %d", i),
			CreatorName: broadcasterID, BroadcasterName: broadcasterID, BroadcasterID: &broadcasterID,
			CreatedAt: time.Now(), ImportedAt: time.Now(),
		}
		if err := clipRepo.Create(ctx, &clip); err != nil {
			t.Fatalf("create clip: %v", err)
		}
		clipIDs = append(clipIDs, id)
	}
	attempts := func() []int {
		t.Helper()
		out := make([]int, len(clipIDs))
		for i, id := range clipIDs {
			if err := pool.QueryRow(ctx, `SELECT transcription_attempt_count FROM clips WHERE id = $1`, id).Scan(&out[i]); err != nil {
				t.Fatalf("read attempts: %v", err)
			}
		}
		return out
	}
	run := func(downloadErr error) *transcriptionDownloadStub {
		downloads := &transcriptionDownloadStub{err: downloadErr}
		s := &AutoTagScheduler{
			clipRepo:      clipRepo,
			transcription: services.NewClipTranscriptionService(transcriptionAuthStub{}, downloads, unusedExtractor{}, unusedTranscriber{}),
		}
		s.processTranscriptionClips(ctx)
		return downloads
	}

	outage := run(fmt.Errorf("%w: status 503", services.ErrTwitchClipDownloadUnavailable))
	if got := attempts(); got[0] != 0 || got[1] != 0 {
		t.Fatalf("attempt counts after a Twitch outage = %v, want [0 0]", got)
	}
	if outage.calls != 1 {
		t.Fatalf("Twitch was called %d times during the outage, want 1", outage.calls)
	}

	run(errors.New("Twitch clip download API returned status 404"))
	if got := attempts(); got[0] != 1 || got[1] != 1 {
		t.Fatalf("attempt counts after clip-specific failures = %v, want [1 1]", got)
	}
}
