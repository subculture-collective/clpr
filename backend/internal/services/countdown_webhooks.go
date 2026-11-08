package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

// WebhookEventTrigger queues a webhook event for every subscriber.
type WebhookEventTrigger interface {
	TriggerEvent(ctx context.Context, eventType string, eventID uuid.UUID, data map[string]interface{}) error
}

// PublicCountdownReader reads an edition as its public page shows it.
type PublicCountdownReader interface {
	GetPublicEdition(ctx context.Context, date, seriesSlug string) (*models.PublicCountdownEdition, error)
}

// CountdownPublishNotifier queues countdown.published when an edition's page
// goes public (#83). The event ID is the edition ID, so rendering an edition
// again does not post it twice.
type CountdownPublishNotifier struct {
	Editions PublicCountdownReader
	Webhooks WebhookEventTrigger
	BaseURL  string
}

// EditionPublished queues the event for one rendered edition.
func (n *CountdownPublishNotifier) EditionPublished(ctx context.Context, edition *models.CountdownEdition) error {
	public, err := n.Editions.GetPublicEdition(ctx, edition.EditionDate, edition.SeriesSlug)
	if err != nil || public == nil {
		return err
	}
	base := strings.TrimRight(n.BaseURL, "/")
	data := CountdownPublishedData{
		EditionDate: public.EditionDate,
		SeriesSlug:  public.SeriesSlug,
		SeriesName:  public.SeriesName,
		Size:        public.Size,
		URL:         fmt.Sprintf("%s/countdowns/%s/%s", base, public.EditionDate, public.SeriesSlug),
		Clips:       make([]WebhookClip, 0, len(public.Clips)),
	}
	for _, c := range public.Clips {
		clip := WebhookClip{
			Position:        c.Position,
			Title:           c.Title,
			BroadcasterName: c.BroadcasterName,
			URL:             base + "/clip/" + c.ClipID.String(),
			TwitchURL:       c.TwitchClipURL,
		}
		if c.ThumbnailURL != nil {
			clip.ThumbnailURL = *c.ThumbnailURL
		}
		data.Clips = append(data.Clips, clip)
	}
	payload, err := webhookEventData(data)
	if err != nil {
		return err
	}
	return n.Webhooks.TriggerEvent(ctx, models.WebhookEventCountdownPublished, edition.ID, payload)
}

// DailyTopClipsSource lists the clips of the daily top-clips post.
type DailyTopClipsSource interface {
	ListDailyTopClips(ctx context.Context, start, end time.Time, limit int) ([]models.DailyTopClip, error)
}

type dailyTopClips struct {
	source  DailyTopClipsSource
	baseURL string
	hour    int
	mu      sync.Mutex
	done    string
}

// SetDailyTopClips enables the daily.top_clips event: once a day after
// hourUTC, the top clips of the 24 hours ending then.
func (s *OutboundWebhookService) SetDailyTopClips(source DailyTopClipsSource, baseURL string, hourUTC int) {
	if hourUTC < 0 || hourUTC > 23 {
		hourUTC = 16
	}
	s.daily = &dailyTopClips{source: source, baseURL: strings.TrimRight(baseURL, "/"), hour: hourUTC}
}

// dailyTopClipsEventID is the same for every run on one date, so instances
// and restarts cannot queue the post twice.
func dailyTopClipsEventID(date string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("clpr:daily.top_clips:"+date))
}

// QueueDailyTopClips queues today's daily.top_clips event when it is due and
// someone subscribes to it.
func (s *OutboundWebhookService) QueueDailyTopClips(ctx context.Context, now time.Time) error {
	d := s.daily
	if d == nil {
		return nil
	}
	now = now.UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), d.hour, 0, 0, 0, time.UTC)
	if now.Before(end) {
		return nil
	}
	date := end.Format("2006-01-02")
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.done == date {
		return nil
	}
	subscriptions, err := s.webhookRepo.GetActiveSubscriptionsByEvent(ctx, models.WebhookEventDailyTopClips)
	if err != nil || len(subscriptions) == 0 {
		// A subscription added later today still gets today's post.
		return err
	}
	clips, err := d.source.ListDailyTopClips(ctx, end.Add(-24*time.Hour), end, 10)
	if err != nil {
		return err
	}
	if len(clips) > 0 {
		data := DailyTopClipsData{Date: date, Clips: make([]WebhookClip, 0, len(clips))}
		for i, c := range clips {
			clip := WebhookClip{
				Position:        i + 1,
				Title:           c.Title,
				BroadcasterName: c.BroadcasterName,
				URL:             d.baseURL + "/clip/" + c.ClipID.String(),
				TwitchURL:       c.TwitchClipURL,
				ViewCount:       c.ViewCount,
			}
			if c.ThumbnailURL != nil {
				clip.ThumbnailURL = *c.ThumbnailURL
			}
			data.Clips = append(data.Clips, clip)
		}
		payload, err := webhookEventData(data)
		if err != nil {
			return err
		}
		if err := s.TriggerEvent(ctx, models.WebhookEventDailyTopClips, dailyTopClipsEventID(date), payload); err != nil {
			return err
		}
	}
	d.done = date
	return nil
}
