package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

// SearchTranscriptionStore selects clips to transcribe for search and
// stores the results.
type SearchTranscriptionStore interface {
	NextClipsToTranscribe(ctx context.Context, q repository.SearchTranscriptionQuery) ([]repository.SearchTranscriptionClip, error)
	SaveSearchTranscript(ctx context.Context, clipID uuid.UUID, language, fullText string, segments json.RawMessage) (bool, error)
	RecordSearchTranscriptionAttempt(ctx context.Context, clipID uuid.UUID, outcome string) error
	SearchClipsAttemptedOn(ctx context.Context, t time.Time) (int, error)
	ClearExcludedTranscripts(ctx context.Context) (int64, error)
	RecordTranscriptionUsage(ctx context.Context, u repository.TranscriptionUsage) error
}

// SearchTranscriptionWorker transcribes public clips above a view threshold
// for transcript search (ADR-003). Each clip file is downloaded to the
// render host, transcribed and deleted before the next one; only the
// transcript is kept.
type SearchTranscriptionWorker struct {
	Store     SearchTranscriptionStore
	Fetcher   ClipMediaFetcher
	Prober    MediaProber
	Captioner CountdownCaptioner
	// Dir is the media directory; files go in its .search subdirectory,
	// under the directory's 24-hour sweep.
	Dir string
	// MinViews, MaxAge and DailyLimit bound which clips and how many a day.
	MinViews   int
	MaxAge     time.Duration
	DailyLimit int
	// BatchSize is the most clips one cycle transcribes, so countdown work
	// in the same loop is not held up.
	BatchSize int
	Metrics   *prometheus.CounterVec
	Now       func() time.Time
}

const (
	searchTranscriptionRetryAfter  = 6 * time.Hour
	searchTranscriptionMaxAttempts = 3
)

// NewSearchTranscriptionMetrics counts clips by outcome.
func NewSearchTranscriptionMetrics(reg prometheus.Registerer) *prometheus.CounterVec {
	m := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clpr_search_transcription_clips_total",
		Help: "Clips transcribed for transcript search on the render host, by outcome (transcribed, no_speech, failed).",
	}, []string{"outcome"})
	for _, outcome := range []string{repository.SearchTranscriptionTranscribed, repository.SearchTranscriptionNoSpeech, repository.SearchTranscriptionFailed} {
		m.WithLabelValues(outcome)
	}
	if reg != nil {
		reg.MustRegister(m)
	}
	return m
}

func (w *SearchTranscriptionWorker) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

// RunOnce empties the transcripts of excluded broadcasters, then
// transcribes up to BatchSize clips within the day's limit.
func (w *SearchTranscriptionWorker) RunOnce(ctx context.Context) error {
	if w.Metrics == nil {
		w.Metrics = NewSearchTranscriptionMetrics(nil)
	}
	if cleared, err := w.Store.ClearExcludedTranscripts(ctx); err != nil {
		return fmt.Errorf("clearing excluded transcripts: %w", err)
	} else if cleared > 0 {
		utils.Info("Transcripts of excluded clips emptied", map[string]interface{}{"transcripts": cleared})
	}
	now := w.now()
	done, err := w.Store.SearchClipsAttemptedOn(ctx, now)
	if err != nil {
		return err
	}
	limit := min(w.BatchSize, w.DailyLimit-done)
	if limit <= 0 {
		return nil
	}
	clips, err := w.Store.NextClipsToTranscribe(ctx, repository.SearchTranscriptionQuery{
		MinViews:     w.MinViews,
		CreatedAfter: now.Add(-w.MaxAge),
		RetryBefore:  now.Add(-searchTranscriptionRetryAfter),
		MaxAttempts:  searchTranscriptionMaxAttempts,
		Limit:        limit,
	})
	if err != nil {
		return err
	}
	dir := filepath.Join(w.Dir, ".search")
	if len(clips) > 0 {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	for _, clip := range clips {
		if err := w.transcribe(ctx, dir, clip); err != nil {
			return err
		}
	}
	return nil
}

// transcribe handles one clip. It returns an error only when the pass
// should stop: the tools or the database are unavailable, or shutdown.
func (w *SearchTranscriptionWorker) transcribe(ctx context.Context, dir string, clip repository.SearchTranscriptionClip) error {
	start := w.now()
	path := filepath.Join(dir, clip.ClipID.String()+".mp4")
	defer os.Remove(path)

	usage := repository.TranscriptionUsage{Day: start, Source: repository.TranscriptSourceSearch}
	outcome, reason := repository.SearchTranscriptionFailed, ""
	_, err := w.Fetcher.Fetch(ctx, &models.CountdownEditionClip{ClipID: clip.ClipID, TwitchClipID: clip.TwitchClipID,
		TwitchClipURL: clip.TwitchClipURL, BroadcasterID: clip.BroadcasterID}, path)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, ErrCountdownMediaUnavailable) {
		return err
	}
	if err != nil {
		reason = err.Error()
	} else if info, err := w.Prober.Probe(ctx, path); err != nil {
		reason = err.Error()
	} else {
		usage.AudioSeconds = info.DurationSeconds
		words, err := w.Captioner.Words(ctx, path)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		switch text, segments := transcriptFromWords(words); {
		case err != nil:
			reason = err.Error()
		case text == "":
			outcome = repository.SearchTranscriptionNoSpeech
		default:
			outcome = repository.SearchTranscriptionTranscribed
			if usage.Transcribed, err = w.Store.SaveSearchTranscript(ctx, clip.ClipID, "en", text, segments); err != nil {
				return fmt.Errorf("storing the transcript of clip %s: %w", clip.ClipID, err)
			}
		}
	}
	_ = os.Remove(path)
	usage.ComputeSeconds = w.now().Sub(start).Seconds()
	w.Metrics.WithLabelValues(outcome).Inc()
	if reason != "" {
		utils.Warn("Search transcription failed", map[string]interface{}{"clip_id": clip.ClipID.String(), "reason": reason})
	}
	if err := w.Store.RecordSearchTranscriptionAttempt(ctx, clip.ClipID, outcome); err != nil {
		return err
	}
	return w.Store.RecordTranscriptionUsage(ctx, usage)
}
