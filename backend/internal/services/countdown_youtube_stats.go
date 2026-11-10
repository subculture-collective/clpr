package services

import (
	"context"
	"errors"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// CountdownStatsStore is the storage the view-count refresher needs.
type CountdownStatsStore interface {
	ListUploadsForStats(ctx context.Context, uploadedAfter, staleBefore time.Time, limit int) ([]string, error)
	SaveUploadStats(ctx context.Context, stats []models.CountdownVideoStats, missing []string, at time.Time) error
}

// YouTubeStatsReader reads view counts from the YouTube Data API.
type YouTubeStatsReader interface {
	VideoStats(ctx context.Context, ids []string) (map[string]YouTubeVideoStats, error)
}

// CountdownStatsRefresher keeps the view counts of recent countdown uploads
// current (#84), for the creator dashboard. Each video is read at most once
// per Every, for Window after its upload, 50 videos per videos.list call.
type CountdownStatsRefresher struct {
	Store   CountdownStatsStore
	YouTube YouTubeStatsReader
	// Every is how often one video is read again; default 6 hours.
	Every time.Duration
	// Window is how long after the upload views are still read; default 90 days.
	Window time.Duration
	// PerRun caps the videos read in one run; default 500 (10 quota units).
	PerRun int
	// Interval is the least time between runs, so the videos that came due
	// meanwhile share videos.list calls; default 1 hour.
	Interval time.Duration
	Now      func() time.Time

	pausedUntil time.Time
	lastRun     time.Time
}

func (r *CountdownStatsRefresher) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

func durationOr(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}

// RunOnce reads the view counts that are due. A quota, authorization or
// temporary YouTube error pauses refreshing and is logged rather than
// returned, since stale view counts harm nothing.
func (r *CountdownStatsRefresher) RunOnce(ctx context.Context) error {
	now := r.now()
	if now.Before(r.pausedUntil) || now.Before(r.lastRun.Add(durationOr(r.Interval, time.Hour))) {
		return nil
	}
	r.lastRun = now
	perRun := r.PerRun
	if perRun <= 0 {
		perRun = 500
	}
	ids, err := r.Store.ListUploadsForStats(ctx, now.Add(-durationOr(r.Window, 90*24*time.Hour)), now.Add(-durationOr(r.Every, 6*time.Hour)), perRun)
	if err != nil {
		return err
	}
	for start := 0; start < len(ids); start += 50 {
		batch := ids[start:min(start+50, len(ids))]
		found, err := r.YouTube.VideoStats(ctx, batch)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return r.pause(err)
		}
		stats := make([]models.CountdownVideoStats, 0, len(found))
		var missing []string
		for _, id := range batch {
			// A video YouTube rejected, failed to process or deleted cannot
			// be watched, so it counts as missing.
			if s, ok := found[id]; ok && s.UploadStatus != "rejected" && s.UploadStatus != "failed" && s.UploadStatus != "deleted" {
				stats = append(stats, models.CountdownVideoStats{YouTubeVideoID: id, ViewCount: s.ViewCount, PrivacyStatus: s.PrivacyStatus})
			} else {
				missing = append(missing, id)
			}
		}
		if err := r.Store.SaveUploadStats(ctx, stats, missing, r.now()); err != nil {
			return err
		}
	}
	return nil
}

func (r *CountdownStatsRefresher) pause(err error) error {
	var yt *YouTubeError
	switch {
	case IsYouTubeQuotaError(err):
		r.pausedUntil = nextQuotaReset(r.now())
	case errors.As(err, &yt) && yt.AuthFailed():
		r.pausedUntil = r.now().Add(time.Hour)
	case errors.As(err, &yt) && !yt.Transient():
		// A request YouTube rejects outright will not succeed on retry; wait
		// for the next refresh period instead of repeating it every cycle.
		r.pausedUntil = r.now().Add(durationOr(r.Every, 6*time.Hour))
	default:
		r.pausedUntil = r.now().Add(15 * time.Minute)
	}
	utils.Warn("Countdown view counts not refreshed", map[string]interface{}{
		"reason": err.Error(), "retry_at": r.pausedUntil.Format(time.RFC3339),
	})
	return nil
}
