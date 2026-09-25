package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type engagementFake struct {
	batches   int
	published int
	failed    []uuid.UUID
	observed  map[uuid.UUID]int
	clips     []repository.EngagementPollClip
	repeat    bool
}

func (f *engagementFake) Publish(context.Context) error { f.published++; return nil }
func (f *engagementFake) ClaimDue(context.Context, int) ([]repository.EngagementPollClip, error) {
	f.batches++
	if f.batches > 1 && !f.repeat {
		return nil, nil
	}
	return f.clips, nil
}
func (f *engagementFake) Observe(ctx context.Context, id uuid.UUID, count int, at time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.observed[id] = count
	return nil
}
func (f *engagementFake) PollFailed(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.failed = append(f.failed, id)
	return nil
}

func TestEngagementDeadlineStillRecordsMissedObservation(t *testing.T) {
	for _, phase := range []string{"provider", "persistence"} {
		t.Run(phase, func(t *testing.T) {
			id := uuid.New()
			store := &engagementFake{clips: []repository.EngagementPollClip{{ID: id, TwitchID: "slow-provider"}}, observed: map[uuid.UUID]int{}}
			provider := engagementProviderFunc(func(ctx context.Context, _ *twitch.ClipParams) (*twitch.ClipsResponse, error) {
				<-ctx.Done()
				if phase == "persistence" {
					return &twitch.ClipsResponse{Data: []twitch.Clip{{ID: "slow-provider", ViewCount: 25}}}, nil
				}
				return nil, ctx.Err()
			})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			defer cancel()
			err := NewEngagementScheduler(store, provider).poll(ctx)
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.Equal(t, []uuid.UUID{id}, store.failed, "the exhausted network budget must not cancel failure bookkeeping")
			require.Empty(t, store.observed)
		})
	}
}

type engagementProviderFunc func(context.Context, *twitch.ClipParams) (*twitch.ClipsResponse, error)

func (f engagementProviderFunc) GetClips(c context.Context, p *twitch.ClipParams) (*twitch.ClipsResponse, error) {
	return f(c, p)
}

func TestEngagementPollingBudgetAndProviderFailure(t *testing.T) {
	for _, scenario := range []string{"missing clip", "provider backoff", "bounded sweep", "no provider"} {
		t.Run(scenario, func(t *testing.T) {
			id := uuid.New()
			f := &engagementFake{clips: []repository.EngagementPollClip{{ID: id, TwitchID: "clip"}}, observed: map[uuid.UUID]int{}, repeat: scenario == "bounded sweep"}
			var provider EngagementProvider = engagementProviderFunc(func(ctx context.Context, p *twitch.ClipParams) (*twitch.ClipsResponse, error) {
				_, hasDeadline := ctx.Deadline()
				require.True(t, hasDeadline)
				require.Equal(t, []string{"clip"}, p.ClipIDs)
				if scenario == "provider backoff" {
					return nil, errors.New("rate limited")
				}
				if scenario == "missing clip" {
					return &twitch.ClipsResponse{}, nil
				}
				return &twitch.ClipsResponse{Data: []twitch.Clip{{ID: "clip", ViewCount: 25}}}, nil
			})
			if scenario == "no provider" {
				provider = nil
			}
			scheduler := NewEngagementScheduler(f, provider)
			err := scheduler.RunOnce(context.Background())
			require.Equal(t, 1, f.published, "publish local engagement even when provider fails")
			switch scenario {
			case "provider backoff":
				require.Error(t, err)
				require.Equal(t, 1, f.batches)
				require.Equal(t, []uuid.UUID{id}, f.failed)
			case "missing clip":
				require.NoError(t, err)
				require.Equal(t, []uuid.UUID{id}, f.failed)
				require.Empty(t, f.observed)
			case "bounded sweep":
				require.NoError(t, err)
				require.Equal(t, 10, f.batches)
				require.Equal(t, 25, f.observed[id])
			case "no provider":
				require.NoError(t, err)
				require.Zero(t, f.batches)
			}
			scheduler.Stop()
			scheduler.Stop()
		})
	}
}

// Rankings are republished at most once per publish interval (default five
// minutes) while polling continues every tick.
func TestEngagementTicksPollEveryMinuteButPublishOnInterval(t *testing.T) {
	id := uuid.New()
	f := &engagementFake{clips: []repository.EngagementPollClip{{ID: id, TwitchID: "clip"}}, observed: map[uuid.UUID]int{}, repeat: true}
	polls := 0
	provider := engagementProviderFunc(func(context.Context, *twitch.ClipParams) (*twitch.ClipsResponse, error) {
		polls++
		return &twitch.ClipsResponse{Data: []twitch.Clip{{ID: "clip", ViewCount: 25}}}, nil
	})
	s := NewEngagementScheduler(f, provider)
	s.SetPublishInterval(0) // ignored: keeps the default
	require.Equal(t, DefaultEngagementPublishInterval, s.publishEvery)

	start := time.Date(2026, 9, 25, 5, 0, 0, 0, time.UTC)
	for minute := 0; minute < 11; minute++ {
		require.NoError(t, s.runTick(context.Background(), start.Add(time.Duration(minute)*time.Minute)))
	}
	require.Equal(t, 3, f.published, "publishes at minutes 0, 5 and 10")
	require.Equal(t, 11*10, polls, "every tick polls its full bounded sweep")

	s.SetPublishInterval(time.Minute)
	require.NoError(t, s.runTick(context.Background(), start.Add(11*time.Minute)))
	require.Equal(t, 4, f.published)
}
