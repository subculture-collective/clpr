package services

import (
	"context"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
)

// StreamerDigestStore is the storage the weekly digest needs.
type StreamerDigestStore interface {
	ListDigestRecipients(ctx context.Context, weekStart time.Time) ([]models.StreamerDigestRecipient, error)
	ClaimDigest(ctx context.Context, userID uuid.UUID, weekStart time.Time) (bool, error)
	FinishDigest(ctx context.Context, userID uuid.UUID, weekStart time.Time, status string, reason *string) error
	StreamerDigestContent(ctx context.Context, broadcasterID string, weekStart time.Time) (*models.StreamerDigest, error)
}

// StreamerDigestSender sends one digest email.
type StreamerDigestSender interface {
	SendStreamerDigest(ctx context.Context, rcpt models.StreamerDigestRecipient, d *models.StreamerDigest) error
}

// StreamerDigestService sends the weekly streamer digest (#85): from
// SendHourUTC on Monday, one email per broadcaster who features their clips,
// with last week's top clips, countdown placements and clip pack.
type StreamerDigestService struct {
	Store  StreamerDigestStore
	Sender StreamerDigestSender
	// SendHourUTC is the hour on Monday from which last week's digests go
	// out, after the clip packs; default 9.
	SendHourUTC int
	Now         func() time.Time
}

func (s *StreamerDigestService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// DigestWeek returns the Monday (UTC) starting the newest week whose digests
// are due by now.
func (s *StreamerDigestService) DigestWeek(now time.Time) time.Time {
	hour := s.SendHourUTC
	if hour <= 0 || hour > 23 {
		hour = 9
	}
	now = now.UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	monday := day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
	if now.Before(monday.Add(time.Duration(hour) * time.Hour)) {
		monday = monday.AddDate(0, 0, -7)
	}
	return monday.AddDate(0, 0, -7)
}

// RunOnce sends every digest due for the latest week and returns how many it
// sent. Each digest is claimed before sending, so a digest goes out at most
// once even when several API replicas run this.
func (s *StreamerDigestService) RunOnce(ctx context.Context) (int, error) {
	week := s.DigestWeek(s.now())
	recipients, err := s.Store.ListDigestRecipients(ctx, week)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, rcpt := range recipients {
		if ctx.Err() != nil {
			return sent, ctx.Err()
		}
		claimed, err := s.Store.ClaimDigest(ctx, rcpt.UserID, week)
		if err != nil {
			return sent, err
		}
		if !claimed {
			continue
		}
		status, reason := s.sendOne(ctx, rcpt, week)
		if status == models.StreamerDigestSent {
			sent++
		}
		if err := s.Store.FinishDigest(context.WithoutCancel(ctx), rcpt.UserID, week, status, reason); err != nil {
			return sent, err
		}
	}
	return sent, nil
}

func (s *StreamerDigestService) sendOne(ctx context.Context, rcpt models.StreamerDigestRecipient, week time.Time) (string, *string) {
	fields := map[string]interface{}{"user_id": rcpt.UserID.String(), "week_start": week.Format("2006-01-02")}
	digest, err := s.Store.StreamerDigestContent(ctx, rcpt.BroadcasterID, week)
	if err != nil {
		reason := "loading the digest: " + err.Error()
		utils.Warn("Streamer digest failed", fields)
		return models.StreamerDigestFailed, &reason
	}
	if digest.Empty() {
		reason := "no clips, placements or clip pack that week"
		return models.StreamerDigestSkipped, &reason
	}
	digest.Name = rcpt.Name
	if err := s.Sender.SendStreamerDigest(ctx, rcpt, digest); err != nil {
		reason := err.Error()
		fields["reason"] = reason
		utils.Warn("Streamer digest failed", fields)
		return models.StreamerDigestFailed, &reason
	}
	return models.StreamerDigestSent, nil
}
