package scheduler

import (
	"context"
	"os"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
)

const autoTagSchedulerName = "auto_tag"

// AutoTagScheduler periodically finds untagged clips and applies
// structural + AI-derived tags to them.  It runs on a short ticker
// so new clips are tagged near-real-time.
type AutoTagScheduler struct {
	autoTag   *services.AutoTagService
	whisper   *services.WhisperService
	thumbnail *services.ThumbnailService
	clipRepo  *repository.ClipRepository
	tagRepo   *repository.TagRepository

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
	tagRepo *repository.TagRepository,
	intervalSeconds int,
) *AutoTagScheduler {
	if intervalSeconds <= 0 {
		intervalSeconds = 30
	}
	return &AutoTagScheduler{
		autoTag:   autoTag,
		whisper:   whisper,
		thumbnail: thumbnail,
		clipRepo:  clipRepo,
		tagRepo:   tagRepo,
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

		// Full pipeline: download → audio extraction → whisper,
		// thumbnail extraction → vision AI classification.
		// These are best-effort — failures are logged and we move on.
		if s.whisper != nil || (s.thumbnail != nil && s.thumbnail.Operational()) {
			s.processClipMedia(ctx, clip)
		}
	}

	duration := time.Since(startTime)
	utils.Info("Auto-tag batch completed", map[string]interface{}{
		"scheduler": autoTagSchedulerName,
		"tagged":    tagged,
		"total":     len(untagged),
		"duration":  duration.String(),
	})
}

// processClipMedia downloads the clip video, extracts audio for Whisper
// transcription, extracts thumbnails for vision AI classification, and
// stores the resulting content tags.  All steps are best-effort — individual
// failures are logged and skipped.
func (s *AutoTagScheduler) processClipMedia(ctx context.Context, clip *models.Clip) {
	// Derive the video download URL from the thumbnail URL.
	// Twitch thumbnail URLs follow the pattern:
	//   https://clips-media-assets2.twitch.tv/AT-cm-<id>-preview-480x272.jpg
	// The actual mp4 is at the same path with .mp4 suffix.
	if clip.ThumbnailURL == nil || *clip.ThumbnailURL == "" {
		utils.Warn("No thumbnail URL available for media processing", map[string]interface{}{
			"clip_id": clip.ID.String(),
		})
		return
	}

	videoPath, dlErr := services.DownloadClipVideo(ctx, *clip.ThumbnailURL, "/tmp/clpr-tags")
	if dlErr != nil {
		utils.Warn("Failed to download clip video, skipping media analysis", map[string]interface{}{
			"clip_id": clip.ID.String(),
		})
		return
	}
	defer os.Remove(videoPath)

	// Whisper transcription
	if s.whisper != nil {
		wavPath, extErr := services.ExtractAudio(ctx, "ffmpeg", videoPath, "/tmp/clpr-tags")
		if extErr == nil {
			defer os.Remove(wavPath)
			result, whErr := s.whisper.TranscribeAudio(ctx, wavPath)
			if whErr == nil && result != nil {
				utils.Info("Clip transcribed", map[string]interface{}{
					"clip_id":  clip.ID.String(),
					"language": result.Language,
					"text_len": len(result.FullText),
				})
			} else if whErr != nil {
				utils.Warn("Whisper transcription failed", map[string]interface{}{
					"clip_id": clip.ID.String(),
				})
			}
		} else {
			utils.Warn("Audio extraction failed", map[string]interface{}{
				"clip_id": clip.ID.String(),
			})
		}
	}

	// Thumbnail extraction + vision AI classification
	if s.thumbnail != nil && s.thumbnail.Operational() {
		duration := 60.0 // default if clip.Duration is nil
		if clip.Duration != nil {
			duration = *clip.Duration
		}
		thumbs, thErr := s.thumbnail.ExtractThumbnails(ctx, videoPath, duration)
		if thErr == nil {
			defer func() {
				for _, t := range thumbs {
					os.Remove(t)
				}
			}()
			tags, visErr := s.thumbnail.ClassifyThumbnails(ctx, thumbs, gameName(clip))
			if visErr == nil {
				// Ensure content tags exist and attach them to the clip.
				s.ensureAndAttachContentTags(ctx, clip.ID, tags)
			} else {
				utils.Warn("Vision classification failed", map[string]interface{}{
					"clip_id": clip.ID.String(),
				})
			}
		} else {
			utils.Warn("Thumbnail extraction failed", map[string]interface{}{
				"clip_id": clip.ID.String(),
			})
		}
	}
}

// ensureAndAttachContentTags ensures that content/ prefixed tags exist in
// the tags table and then attaches them to the clip via clip_tags.
func (s *AutoTagScheduler) ensureAndAttachContentTags(ctx context.Context, clipID uuid.UUID, tagSlugs []string) {
	if s.tagRepo == nil {
		return
	}

	// Ensure the content parent tag exists.
	_, _ = s.tagRepo.GetOrCreateTag(ctx, "Content", "content", nil)

	for _, slug := range tagSlugs {
		fullSlug := "content/" + slug

		// Ensure the tag row exists in the tags table.
		_, _ = s.tagRepo.GetOrCreateTag(ctx, slug, fullSlug, nil)

		// Attach the tag to the clip (idempotent).
		if addErr := s.clipRepo.AddTagBySlug(ctx, clipID, fullSlug); addErr != nil {
			utils.Warn("Failed to add content tag", map[string]interface{}{
				"clip_id": clipID.String(),
				"tag":     fullSlug,
			})
		}
	}
}

// gameName returns the game name for a clip, or "Unknown Game" if not set.
func gameName(clip *models.Clip) string {
	if clip != nil && clip.GameName != nil {
		return *clip.GameName
	}
	return "Unknown Game"
}