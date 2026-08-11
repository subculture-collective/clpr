package scheduler

import (
	"context"
	"encoding/json"
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
	autoTag       *services.AutoTagService
	thumbnail     *services.ThumbnailService
	transcription *services.ClipTranscriptionService
	clipRepo      *repository.ClipRepository
	tagRepo       *repository.TagRepository

	interval time.Duration
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewAutoTagScheduler creates a new AutoTagScheduler.
//
// thumbnail may be nil when vision enrichment is disabled.
func NewAutoTagScheduler(
	autoTag *services.AutoTagService,
	thumbnail *services.ThumbnailService,
	transcription *services.ClipTranscriptionService,
	clipRepo *repository.ClipRepository,
	tagRepo *repository.TagRepository,
	intervalSeconds int,
) *AutoTagScheduler {
	if intervalSeconds <= 0 {
		intervalSeconds = 30
	}
	return &AutoTagScheduler{
		autoTag:       autoTag,
		thumbnail:     thumbnail,
		transcription: transcription,
		clipRepo:      clipRepo,
		tagRepo:       tagRepo,
		interval:      time.Duration(intervalSeconds) * time.Second,
		stopChan:      make(chan struct{}),
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
		if s.transcription != nil {
			s.processTranscriptionClips(ctx)
		}
		if s.thumbnail != nil && s.thumbnail.Operational() {
			s.processVisionClips(ctx)
		}
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

		if err := s.clipRepo.MarkAutoTagged(ctx, clip.ID); err != nil {
			utils.Error("Failed to record auto-tag completion", err, map[string]interface{}{
				"scheduler": autoTagSchedulerName,
				"clip_id":   clip.ID.String(),
			})
		}
	}

	duration := time.Since(startTime)
	utils.Info("Auto-tag batch completed", map[string]interface{}{
		"scheduler": autoTagSchedulerName,
		"tagged":    tagged,
		"total":     len(untagged),
		"duration":  duration.String(),
	})

	if s.transcription != nil {
		s.processTranscriptionClips(ctx)
	}
	if s.thumbnail != nil && s.thumbnail.Operational() {
		s.processVisionClips(ctx)
	}
}

func (s *AutoTagScheduler) processTranscriptionClips(ctx context.Context) {
	clips, err := s.clipRepo.GetClipsNeedingTranscription(ctx, 2)
	if err != nil {
		utils.Error("Failed to fetch clips needing transcription", err, map[string]interface{}{
			"scheduler": autoTagSchedulerName,
		})
		return
	}
	for i := range clips {
		clip := &clips[i]
		clipCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		result, transcribeErr := s.transcription.TranscribeClip(clipCtx, clip)
		cancel()
		if transcribeErr != nil {
			utils.Warn("Authorized clip transcription failed", map[string]interface{}{
				"clip_id": clip.ID.String(),
			})
			if recordErr := s.clipRepo.RecordTranscriptionFailure(ctx, clip.ID, transcribeErr); recordErr != nil {
				utils.Error("Failed to record transcription failure", recordErr, map[string]interface{}{
					"clip_id": clip.ID.String(),
				})
			}
			continue
		}
		segments, marshalErr := json.Marshal(result.Segments)
		if marshalErr != nil {
			_ = s.clipRepo.RecordTranscriptionFailure(ctx, clip.ID, marshalErr)
			continue
		}
		if recordErr := s.clipRepo.RecordClipTranscript(ctx, &models.ClipTranscript{
			ClipID: clip.ID, Language: result.Language, FullText: result.FullText,
			Segments: segments, Source: "twitch_authorized_whisper",
		}); recordErr != nil {
			utils.Error("Failed to store clip transcript", recordErr, map[string]interface{}{
				"clip_id": clip.ID.String(),
			})
		}
	}
}

// processVisionClips analyzes Twitch's public thumbnail URL directly. It does
// not download clip video or require broadcaster/editor permissions.
func (s *AutoTagScheduler) processVisionClips(ctx context.Context) {
	clips, err := s.clipRepo.GetClipsNeedingVision(ctx, 10)
	if err != nil {
		utils.Error("Failed to fetch clips needing thumbnail enrichment", err, map[string]interface{}{
			"scheduler": autoTagSchedulerName,
		})
		return
	}
	for i := range clips {
		clip := &clips[i]
		var result *services.ClipThumbnailEnrichment
		transcript, transcriptErr := s.clipRepo.GetClipTranscript(ctx, clip.ID)
		if transcriptErr != nil {
			utils.Warn("Failed to load clip transcript for enrichment", map[string]interface{}{
				"clip_id": clip.ID.String(),
			})
		}
		var analyzeErr error
		if transcript != nil && transcript.FullText != "" {
			result, analyzeErr = s.thumbnail.AnalyzeClipWithTranscript(ctx, clip, transcript.FullText)
		} else {
			result, analyzeErr = s.thumbnail.AnalyzeClipThumbnail(ctx, clip)
		}
		if analyzeErr != nil {
			utils.Warn("Twitch thumbnail enrichment failed", map[string]interface{}{
				"clip_id": clip.ID.String(),
			})
			if recordErr := s.clipRepo.RecordVisionFailure(ctx, clip.ID, analyzeErr); recordErr != nil {
				utils.Error("Failed to record thumbnail enrichment failure", recordErr, map[string]interface{}{
					"clip_id": clip.ID.String(),
				})
			}
			continue
		}

		s.ensureAndAttachContentTags(ctx, clip.ID, result.Tags)
		enrichment := &models.ClipEnrichment{
			ClipID:         clip.ID,
			SourceTitle:    clip.Title,
			SuggestedTitle: result.SuggestedTitle,
			Confidence:     result.Confidence,
			Basis:          result.Basis,
			Evidence:       result.Evidence,
			Tags:           result.Tags,
			TitleAccepted:  services.ShouldApplySuggestedTitle(clip, result),
		}
		if recordErr := s.clipRepo.RecordThumbnailEnrichment(ctx, enrichment); recordErr != nil {
			utils.Error("Failed to store thumbnail enrichment", recordErr, map[string]interface{}{
				"clip_id": clip.ID.String(),
			})
			if failureErr := s.clipRepo.RecordVisionFailure(ctx, clip.ID, recordErr); failureErr != nil {
				utils.Error("Failed to record thumbnail persistence failure", failureErr, map[string]interface{}{
					"clip_id": clip.ID.String(),
				})
			}
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
