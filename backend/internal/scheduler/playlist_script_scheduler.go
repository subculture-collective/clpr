package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	playlistScriptSchedulerName = "playlist_script"
	playlistScriptJobGenerate   = "playlist_script_generate"
	playlistScriptJobCleanup    = "playlist_script_cleanup"
)

// PlaylistScriptServiceInterface defines the methods the scheduler needs from the service.
type PlaylistScriptServiceInterface interface {
	ListDueForExecution(ctx context.Context) ([]*models.PlaylistScript, error)
	GeneratePlaylist(ctx context.Context, scriptID uuid.UUID) (*models.Playlist, error)
	DeleteStaleGeneratedPlaylists(ctx context.Context) (int64, error)
}

// PlaylistScriptScheduler manages periodic automated playlist generation.
type PlaylistScriptScheduler struct {
	service  PlaylistScriptServiceInterface
	interval time.Duration
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewPlaylistScriptScheduler creates a new scheduler that checks every intervalMinutes for due scripts.
func NewPlaylistScriptScheduler(service PlaylistScriptServiceInterface, intervalMinutes int) *PlaylistScriptScheduler {
	return &PlaylistScriptScheduler{
		service:  service,
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan struct{}),
	}
}

// Start begins the periodic playlist script execution loop.
func (s *PlaylistScriptScheduler) Start(ctx context.Context) {
	utils.Info("Starting playlist script scheduler", map[string]interface{}{
		"scheduler": playlistScriptSchedulerName,
		"interval":  s.interval.String(),
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial check
	s.runDueScripts(ctx)
	s.cleanupStale(ctx)

	for {
		select {
		case <-ticker.C:
			s.runDueScripts(ctx)
			s.cleanupStale(ctx)
		case <-s.stopChan:
			utils.Info("Playlist script scheduler stopped", map[string]interface{}{
				"scheduler": playlistScriptSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Playlist script scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": playlistScriptSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner.
func (s *PlaylistScriptScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// runDueScripts fetches and executes all scripts that are due for their scheduled run.
func (s *PlaylistScriptScheduler) runDueScripts(ctx context.Context) {
	startTime := time.Now()

	scripts, err := s.service.ListDueForExecution(ctx)
	if err != nil {
		utils.Error("Failed to list due playlist scripts", err, map[string]interface{}{
			"scheduler": playlistScriptSchedulerName,
			"job":       playlistScriptJobGenerate,
		})
		metrics.JobExecutionTotal.WithLabelValues(playlistScriptJobGenerate, "failed").Inc()
		return
	}

	if len(scripts) == 0 {
		return
	}

	utils.Info("Found playlist scripts due for execution", map[string]interface{}{
		"scheduler": playlistScriptSchedulerName,
		"count":     len(scripts),
	})

	successCount := 0
	failCount := 0

	for _, script := range scripts {
		// Run each script sequentially to avoid overwhelming the database
		playlist, genErr := s.service.GeneratePlaylist(ctx, script.ID)
		if genErr != nil {
			utils.Error("Failed to generate playlist from script", genErr, map[string]interface{}{
				"scheduler": playlistScriptSchedulerName,
				"script_id": script.ID.String(),
				"script":    script.Name,
				"strategy":  script.Strategy,
			})
			failCount++
			continue
		}

		utils.Info("Generated playlist from script", map[string]interface{}{
			"scheduler":   playlistScriptSchedulerName,
			"script_id":   script.ID.String(),
			"script":      script.Name,
			"playlist_id": playlist.ID.String(),
			"strategy":    script.Strategy,
		})
		successCount++
	}

	duration := time.Since(startTime)
	metrics.JobExecutionDuration.WithLabelValues(playlistScriptJobGenerate).Observe(duration.Seconds())

	if failCount > 0 {
		metrics.JobExecutionTotal.WithLabelValues(playlistScriptJobGenerate, "partial").Inc()
	} else {
		metrics.JobExecutionTotal.WithLabelValues(playlistScriptJobGenerate, "success").Inc()
	}
	metrics.JobLastSuccessTimestamp.WithLabelValues(playlistScriptJobGenerate).Set(float64(time.Now().Unix()))

	utils.Info("Playlist script generation cycle completed", map[string]interface{}{
		"scheduler": playlistScriptSchedulerName,
		"success":   successCount,
		"failed":    failCount,
		"duration":  duration.String(),
	})
}

// cleanupStale removes generated playlists past their retention period.
func (s *PlaylistScriptScheduler) cleanupStale(ctx context.Context) {
	startTime := time.Now()

	deleted, err := s.service.DeleteStaleGeneratedPlaylists(ctx)
	duration := time.Since(startTime)

	metrics.JobExecutionDuration.WithLabelValues(playlistScriptJobCleanup).Observe(duration.Seconds())

	if err != nil {
		utils.Error("Stale playlist cleanup failed", err, map[string]interface{}{
			"scheduler": playlistScriptSchedulerName,
			"job":       playlistScriptJobCleanup,
		})
		metrics.JobExecutionTotal.WithLabelValues(playlistScriptJobCleanup, "failed").Inc()
		return
	}

	metrics.JobExecutionTotal.WithLabelValues(playlistScriptJobCleanup, "success").Inc()
	metrics.JobLastSuccessTimestamp.WithLabelValues(playlistScriptJobCleanup).Set(float64(time.Now().Unix()))

	if deleted > 0 {
		utils.Info("Stale playlist cleanup completed", map[string]interface{}{
			"scheduler": playlistScriptSchedulerName,
			"job":       playlistScriptJobCleanup,
			"deleted":   deleted,
			"duration":  duration.String(),
		})
	}
}
