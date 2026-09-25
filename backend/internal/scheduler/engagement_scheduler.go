package scheduler

import (
	"context"
	"fmt"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
	"sync"
	"time"
)

type EngagementStore interface {
	Publish(context.Context) error
	ClaimDue(context.Context, int) ([]repository.EngagementPollClip, error)
	Observe(context.Context, uuid.UUID, int, time.Time) error
	PollFailed(context.Context, uuid.UUID) error
}

type EngagementProvider interface {
	GetClips(context.Context, *twitch.ClipParams) (*twitch.ClipsResponse, error)
}
type EngagementScheduler struct {
	store    EngagementStore
	provider EngagementProvider
	stop     chan struct{}
	once     sync.Once

	// publishEvery spaces out ranking publishes. Each publish rewrites every
	// period's ranking (~140k rows in production), so publishing on every
	// one-minute poll dominated database writes and mostly timed out.
	publishEvery  time.Duration
	lastPublishAt time.Time
}

// DefaultEngagementPublishInterval is how often rankings are republished.
// Rankings stay resolvable for an hour after publication.
const DefaultEngagementPublishInterval = 5 * time.Minute

func NewEngagementScheduler(store EngagementStore, provider EngagementProvider) *EngagementScheduler {
	return &EngagementScheduler{store: store, provider: provider, stop: make(chan struct{}), publishEvery: DefaultEngagementPublishInterval}
}

// SetPublishInterval changes how often Start republishes rankings; values
// below one minute publish on every poll.
func (s *EngagementScheduler) SetPublishInterval(d time.Duration) {
	if d > 0 {
		s.publishEvery = d
	}
}

// publishDue reports whether a scheduled tick should publish. Failed
// attempts count too, so a timing-out publish is not retried every minute.
func (s *EngagementScheduler) publishDue(now time.Time) bool {
	return s.lastPublishAt.IsZero() || now.Sub(s.lastPublishAt) >= s.publishEvery
}

// runTick polls every tick and publishes when due.
func (s *EngagementScheduler) runTick(ctx context.Context, now time.Time) error {
	if s.publishDue(now) {
		s.lastPublishAt = now
		return s.RunOnce(ctx)
	}
	pollCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	return s.poll(pollCtx)
}

func (s *EngagementScheduler) Stop() { s.once.Do(func() { close(s.stop) }) }
func (s *EngagementScheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-s.stop:
			cancel()
		case <-ctx.Done():
		}
	}()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		if err := s.runTick(ctx, time.Now()); err != nil && ctx.Err() == nil {
			utils.Error("Recent engagement refresh failed", err, nil)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *EngagementScheduler) RunOnce(ctx context.Context) error {
	// Provider failure must not prevent publishing local engagement or its stale-data notice.
	pollCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	pollErr := s.poll(pollCtx)
	cancel()
	// Matches the publish transaction's statement budget in EngagementRepository.Publish.
	publishCtx, publishCancel := context.WithTimeout(ctx, 150*time.Second)
	defer publishCancel()
	if err := s.store.Publish(publishCtx); err != nil {
		return err
	}
	return pollErr
}
func (s *EngagementScheduler) poll(ctx context.Context) error {
	if s.provider == nil {
		return nil
	}
	for batch := 0; batch < 10; batch++ {
		clips, err := s.store.ClaimDue(ctx, 100)
		if err != nil {
			return err
		}
		if len(clips) == 0 {
			return nil
		}
		ids := make([]string, len(clips))
		for i, c := range clips {
			ids[i] = c.TwitchID
		}
		response, err := s.provider.GetClips(ctx, &twitch.ClipParams{ClipIDs: ids})
		observed := time.Now().UTC()
		if err != nil {
			return s.recordFailedBatch(ctx, clips, fmt.Errorf("engagement provider request: %w", err))
		}
		counts := map[string]int{}
		for _, c := range response.Data {
			counts[c.ID] = c.ViewCount
		}
		for i, c := range clips {
			count, ok := counts[c.TwitchID]
			if !ok {
				if err = s.store.PollFailed(ctx, c.ID); err != nil {
					return s.recordFailedBatch(ctx, clips[i:], err)
				}
				continue
			}
			if err = s.store.Observe(ctx, c.ID, count, observed); err != nil {
				return s.recordFailedBatch(ctx, clips[i:], err)
			}
		}
	}
	return nil
}

func (s *EngagementScheduler) recordFailedBatch(ctx context.Context, clips []repository.EngagementPollClip, cause error) error {
	// The polling budget may expire during network I/O or persistence. Record
	// only unfinished observations under a separate, bounded cleanup budget.
	failureCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	for _, clip := range clips {
		if err := s.store.PollFailed(failureCtx, clip.ID); err != nil {
			return fmt.Errorf("%w; record missed observation: %w", cause, err)
		}
	}
	return cause
}
