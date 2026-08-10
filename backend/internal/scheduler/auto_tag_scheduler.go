package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const autoTagSchedulerName = "auto_tag"

// AutoTagScheduler periodically finds untagged clips and applies
// structural + AI-derived tags to them.  It runs on a short ticker
// so new clips are tagged near-real-time.
//
// TODO: The full pipeline (Whisper transcription, thumbnail extraction,
// vision AI classification) requires downloading the video file first.
// For now only structural tags are applied.  Once a video download worker
// is available, wire the whisper and thumbnail services in processClips.
type AutoTagScheduler struct {
	autoTag  *services.AutoTagService
	whisper  *services.WhisperService
	thumbnail *services.ThumbnailService
	clipRepo *repository.ClipRepository

	interval time.Duration
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewAutoTagScheduler creates a new AutoTagScheduler.
//
// whisper and thumbnail may be nil until their respective video-download
// prerequisite is implemented.  The scheduler skips those steps when
// the service pointer is nil.
func NewAutoTagScheduler(
	autoTag *services.AutoTagService,
	whisper *services.WhisperService,
	thumbnail *services.ThumbnailService,
	clipRepo *repository.ClipRepository,
	intervalSeconds int,
) *AutoTagScheduler {
	return &AutoTagScheduler{
		autoTag:   autoTag,
		whisper:   whisper,
		thumbnail: thumbnail,
		clipRepo:  clipRepo,
		interval:  time.Duration(intervalSeconds) * time.Second,
		stopChan:  make(chan struct{}),
	}
}

// Start begins the periodic auto-tagging loop.  It runs immediately on
// start and then every tick thereafter.
func (s *AutoTagScheduler) Start(ctx context.Context) {
	utils.Info("Starting auto-tag scheduler", map[string]interface{}{
		"scheduler": autoTagSchedulerName,
		"interval":  s.interval.String(),
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Process immediately on start
	s.processClips(ctx)

	for {
		select {
		case <-ticker.C:
			s.processClips(ctx)
		case <-s.stopChan:
			utils.Info("Auto-tag scheduler stopped", map[string]interface{}{
				"scheduler": autoTagSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Auto-tag scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": autoTagSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner.
func (s *AutoTagScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// processClips fetches a batch of untagged clips and applies tags to them.
func (s *AutoTagScheduler) processClips(ctx context.Context) {
	startTime := time.Now()

	untagged, err := s.clipRepo.GetUntaggedClips(ctx, 100)
	if err != nil {
		utils.Error("Failed to fetch untagged clips", err, map[string]interface{}{
			"scheduler": autoTagSchedulerName,
		})
		return
	}

	if len(untagged) == 0 {
		return
	}

	utils.Info("Processing untagged clips", map[string]interface{}{
		"scheduler": autoTagSchedulerName,
		"count":     len(untagged),
	})

	tagged := 0
	for i := range untagged {
		clip := &untagged[i]

		// Apply structural tags (duration, language, game, broadcaster,
		// pattern-matching from title).  Tagging is best-effort — individual
		// failures are logged but do not stop the batch.
		if err := s.autoTag.ApplyAutoTags(ctx, clip); err != nil {
			utils.Error("Failed to apply auto-tags", err, map[string]interface{}{
				"scheduler": autoTagSchedulerName,
				"clip_id":   clip.ID.String(),
			})
			continue
		}
		tagged++

		// TODO: Full pipeline — download video, extract audio, run Whisper
		// transcription, extract thumbnails, classify with vision AI.
		// Example skeleton:
		//
		//   if s.whisper != nil {
		//       wavPath := downloadAndExtractAudio(clip.TwitchClipURL)
		//       result, err := s.whisper.TranscribeAudio(ctx, wavPath)
		//       …
		//   }
		//   if s.thumbnail != nil {
		//       thumbs, err := s.thumbnail.ExtractThumbnails(ctx, videoPath, *clip.Duration)
		//       contentTags, err := s.thumbnail.ClassifyThumbnails(ctx, thumbs, clip.GameName)
		//       …
		//   }
		_ = s.whisper      // placeholder — not yet wired
		_ = s.thumbnail    // placeholder — not yet wired
		_ = fmt.Sprintf("") // keep import clean
	}

	duration := time.Since(startTime)
	utils.Info("Auto-tag batch completed", map[string]interface{}{
		"scheduler": autoTagSchedulerName,
		"tagged":    tagged,
		"total":     len(untagged),
		"duration":  duration.String(),
	})
}