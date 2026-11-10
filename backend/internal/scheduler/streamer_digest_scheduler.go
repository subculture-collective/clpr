package scheduler

import (
	"context"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// StreamerDigestRunner sends the digests that are due.
type StreamerDigestRunner interface {
	RunOnce(ctx context.Context) (int, error)
}

// StreamerDigestScheduler checks hourly for weekly streamer digests (#85).
type StreamerDigestScheduler struct {
	runner   StreamerDigestRunner
	interval time.Duration
}

// NewStreamerDigestScheduler creates the scheduler; interval defaults to an
// hour.
func NewStreamerDigestScheduler(runner StreamerDigestRunner, interval time.Duration) *StreamerDigestScheduler {
	if interval <= 0 {
		interval = time.Hour
	}
	return &StreamerDigestScheduler{runner: runner, interval: interval}
}

// Start runs until ctx is done.
func (s *StreamerDigestScheduler) Start(ctx context.Context) {
	s.run(ctx)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.run(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *StreamerDigestScheduler) run(ctx context.Context) {
	sent, err := s.runner.RunOnce(ctx)
	if err != nil && ctx.Err() == nil {
		utils.Error("Streamer digests failed", err, map[string]interface{}{"sent": sent})
		return
	}
	if sent > 0 {
		utils.Info("Streamer digests sent", map[string]interface{}{"sent": sent})
	}
}
