package services

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

// CountdownRenderStore is the storage the render worker needs.
type CountdownRenderStore interface {
	ClaimEditionForRender(ctx context.Context, staleAfter time.Duration) (*models.CountdownEdition, *models.CountdownSeries, []models.CountdownEditionMedia, error)
	RecordRender(ctx context.Context, x *models.CountdownRender) error
	FinishRender(ctx context.Context, editionID uuid.UUID, status string, reason *string) error
	MarkMediaDeleted(ctx context.Context, editionID, clipID uuid.UUID) error
	ListRendersToDelete(ctx context.Context, cutoff time.Time) ([]models.CountdownRender, error)
	MarkRenderDeleted(ctx context.Context, id uuid.UUID) error
}

// CountdownRenderer draws the videos; countdownrender.Renderer implements it.
type CountdownRenderer interface {
	RenderCountdown(ctx context.Context, ed countdownrender.Edition, out string) error
	RenderShort(ctx context.Context, c countdownrender.Clip, out string) error
	RenderThumbnail(ctx context.Context, ed countdownrender.Edition, out string) error
}

// CountdownRenderMetrics counts render outcomes and time.
type CountdownRenderMetrics struct {
	Seconds  *prometheus.HistogramVec
	Editions *prometheus.CounterVec
	Deleted  prometheus.Counter
	Captions *prometheus.CounterVec
}

// NewCountdownRenderMetrics creates the metrics and registers them with reg
// when it is not nil.
func NewCountdownRenderMetrics(reg prometheus.Registerer) *CountdownRenderMetrics {
	m := &CountdownRenderMetrics{
		Seconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "clpr_countdown_render_seconds",
			Help:    "Time to render one countdown output, by kind (countdown, short, thumbnail).",
			Buckets: []float64{1, 5, 15, 30, 60, 120, 300, 600, 1200},
		}, []string{"kind"}),
		Editions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clpr_countdown_render_editions_total",
			Help: "Countdown editions that finished rendering, by outcome (rendered, failed).",
		}, []string{"outcome"}),
		Deleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "clpr_countdown_render_files_deleted_total",
			Help: "Rendered countdown files deleted from the render host.",
		}),
		Captions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clpr_countdown_short_captions_total",
			Help: "Shorts by caption outcome (captioned, none for no English speech, failed).",
		}, []string{"outcome"}),
	}
	// Outcomes start at zero so an alert on increase() sees the first one.
	for _, outcome := range []string{"rendered", "failed"} {
		m.Editions.WithLabelValues(outcome)
	}
	if reg != nil {
		reg.MustRegister(m.Seconds, m.Editions, m.Deleted, m.Captions)
	}
	return m
}

// CountdownRenderWorker renders fetched editions into a countdown video, a
// thumbnail and the series' Shorts, then deletes the source clips.
type CountdownRenderWorker struct {
	Store     CountdownRenderStore
	Renderer  CountdownRenderer
	Prober    MediaProber
	MediaDir  string
	RenderDir string
	// Retention is how long rendered files stay before deletion; they are
	// published (#79) well within it.
	Retention  time.Duration
	StaleAfter time.Duration
	Metrics    *CountdownRenderMetrics
	Now        func() time.Time
	// Notifier, when set, announces each rendered edition (#83).
	Notifier interface {
		EditionPublished(ctx context.Context, edition *models.CountdownEdition) error
	}
	// Captioner, when set, transcribes the clips of Shorts for captions
	// (#81). A clip it cannot transcribe gets a Short without captions.
	Captioner CountdownCaptioner
}

// CountdownCaptioner transcribes a clip file for captions.
type CountdownCaptioner interface {
	Words(ctx context.Context, mediaPath string) ([]countdownrender.Word, error)
}

func (w *CountdownRenderWorker) ensure() {
	if w.Metrics == nil {
		w.Metrics = NewCountdownRenderMetrics(nil)
	}
}

func (w *CountdownRenderWorker) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

// RunOnce deletes expired renders, then renders every fetched edition.
func (w *CountdownRenderWorker) RunOnce(ctx context.Context) error {
	sweepErr := w.Sweep(ctx)
	for {
		claimed, err := w.RenderNext(ctx)
		if err != nil {
			return errors.Join(sweepErr, err)
		}
		if !claimed {
			return sweepErr
		}
	}
}

// RenderNext claims one fetched edition and renders it. It reports whether
// an edition was claimed.
func (w *CountdownRenderWorker) RenderNext(ctx context.Context) (bool, error) {
	w.ensure()
	stale := w.StaleAfter
	if stale <= 0 {
		stale = time.Hour
	}
	edition, series, media, err := w.Store.ClaimEditionForRender(ctx, stale)
	if err != nil || edition == nil {
		return false, err
	}
	err = w.render(ctx, edition, series, media)
	if ctx.Err() != nil {
		// Shutting down: keep the claim and the source clips for a retry.
		return true, ctx.Err()
	}
	outcome, reason := models.CountdownEditionRendered, (*string)(nil)
	if err != nil {
		outcome = models.CountdownEditionFailed
		message := err.Error()
		reason = &message
		_ = os.RemoveAll(filepath.Join(w.RenderDir, edition.ID.String()))
	}
	finishErr := w.Store.FinishRender(ctx, edition.ID, outcome, reason)
	if errors.Is(finishErr, repository.ErrCountdownEditionNotRendering) {
		finishErr = nil
	}
	// ADR-002: source clips go once the render finishes, whatever the outcome.
	deleteErr := w.deleteSources(ctx, media)
	var notifyErr error
	if finishErr == nil {
		w.Metrics.Editions.WithLabelValues(outcome).Inc()
		if outcome == models.CountdownEditionRendered && w.Notifier != nil {
			if notifyErr = w.Notifier.EditionPublished(ctx, edition); notifyErr != nil {
				notifyErr = fmt.Errorf("announce edition %s: %w", edition.ID, notifyErr)
			}
		}
	}
	fields := map[string]interface{}{"edition_id": edition.ID.String(), "series": edition.SeriesSlug, "edition_date": edition.EditionDate}
	if err != nil {
		fields["reason"] = err.Error()
		utils.Warn("Countdown edition render failed", fields)
	} else {
		utils.Info("Countdown edition rendered", fields)
	}
	return true, errors.Join(finishErr, deleteErr, notifyErr)
}

func (w *CountdownRenderWorker) render(ctx context.Context, edition *models.CountdownEdition, series *models.CountdownSeries, media []models.CountdownEditionMedia) error {
	files := make(map[uuid.UUID]string, len(media))
	for _, m := range media {
		path, ok := pathInside(w.MediaDir, m.FileName)
		if !ok {
			return fmt.Errorf("media path %q is outside the media directory", m.FileName)
		}
		files[m.ClipID] = path
	}
	clips := append([]models.CountdownEditionClip(nil), edition.Clips...)
	sort.Slice(clips, func(i, j int) bool { return clips[i].Position < clips[j].Position })
	ed := countdownrender.Edition{SeriesName: series.Name, Date: edition.EditionDate}
	for _, c := range clips {
		path, ok := files[c.ClipID]
		if !ok {
			return fmt.Errorf("media for clip #%d is missing", c.Position)
		}
		ed.Clips = append(ed.Clips, countdownrender.Clip{Rank: c.Position, Path: path, Title: c.Title, Broadcaster: c.BroadcasterName})
	}
	if len(ed.Clips) == 0 {
		return fmt.Errorf("edition has no clips")
	}

	dir := filepath.Join(w.RenderDir, edition.ID.String())
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating the render directory: %w", err)
	}
	if err := w.output(ctx, edition.ID, models.CountdownRenderCountdown, nil, "countdown.mp4", func(out string) error {
		return w.Renderer.RenderCountdown(ctx, ed, out)
	}); err != nil {
		return err
	}
	if err := w.output(ctx, edition.ID, models.CountdownRenderThumbnail, nil, "thumbnail.jpg", func(out string) error {
		return w.Renderer.RenderThumbnail(ctx, ed, out)
	}); err != nil {
		return err
	}
	for i := 0; i < series.ShortsCount && i < len(ed.Clips); i++ {
		c := ed.Clips[i]
		clipID := clips[i].ClipID
		c.Words = w.captionWords(ctx, edition, c)
		if err := w.output(ctx, edition.ID, models.CountdownRenderShort, &clipID, fmt.Sprintf("short-%02d.mp4", c.Rank), func(out string) error {
			return w.Renderer.RenderShort(ctx, c, out)
		}); err != nil {
			return err
		}
	}
	return nil
}

// captionWords transcribes one Short's clip. Failures leave the Short
// without captions rather than failing the edition.
func (w *CountdownRenderWorker) captionWords(ctx context.Context, edition *models.CountdownEdition, c countdownrender.Clip) []countdownrender.Word {
	if w.Captioner == nil {
		return nil
	}
	words, err := w.Captioner.Words(ctx, c.Path)
	switch {
	case err != nil:
		w.Metrics.Captions.WithLabelValues("failed").Inc()
		utils.Warn("Short captions failed; rendering without them", map[string]interface{}{
			"edition_id": edition.ID.String(), "rank": c.Rank, "reason": err.Error(),
		})
		return nil
	case len(words) == 0:
		w.Metrics.Captions.WithLabelValues("none").Inc()
	default:
		w.Metrics.Captions.WithLabelValues("captioned").Inc()
	}
	return words
}

// output renders one file under a temporary name, checks it, and records it.
func (w *CountdownRenderWorker) output(ctx context.Context, editionID uuid.UUID, kind string, clipID *uuid.UUID, name string, render func(string) error) error {
	rel := filepath.Join(editionID.String(), name)
	final := filepath.Join(w.RenderDir, rel)
	temp := filepath.Join(filepath.Dir(final), "partial-"+name)
	start := w.now()
	if err := render(temp); err != nil {
		_ = os.Remove(temp)
		return fmt.Errorf("render %s: %w", kind, err)
	}
	w.Metrics.Seconds.WithLabelValues(kind).Observe(w.now().Sub(start).Seconds())
	var duration *float64
	if kind != models.CountdownRenderThumbnail {
		info, err := w.Prober.Probe(ctx, temp)
		if err != nil {
			_ = os.Remove(temp)
			return fmt.Errorf("check %s: %w", kind, err)
		}
		duration = &info.DurationSeconds
	}
	digest, size, err := fileSHA256(temp)
	if err == nil {
		err = os.Rename(temp, final)
	}
	if err != nil {
		_ = os.Remove(temp)
		return fmt.Errorf("store %s: %w", kind, err)
	}
	return w.Store.RecordRender(ctx, &models.CountdownRender{EditionID: editionID, Kind: kind, ClipID: clipID,
		FileName: rel, SHA256: digest, Bytes: size, DurationSeconds: duration})
}

func (w *CountdownRenderWorker) deleteSources(ctx context.Context, media []models.CountdownEditionMedia) error {
	var errs []error
	for _, m := range media {
		path, ok := pathInside(w.MediaDir, m.FileName)
		if !ok {
			continue
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
			continue
		}
		errs = append(errs, w.Store.MarkMediaDeleted(context.WithoutCancel(ctx), m.EditionID, m.ClipID))
	}
	return errors.Join(errs...)
}

// Sweep deletes renders older than the retention window, the renders of
// failed editions, and leftover files older than the window.
func (w *CountdownRenderWorker) Sweep(ctx context.Context) error {
	w.ensure()
	retention := w.Retention
	if retention <= 0 {
		retention = 7 * 24 * time.Hour
	}
	cutoff := w.now().Add(-retention)
	root, err := os.OpenRoot(w.RenderDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer root.Close()
	var errs []error
	renders, err := w.Store.ListRendersToDelete(ctx, cutoff)
	if err != nil {
		errs = append(errs, err)
	}
	for _, x := range renders {
		name := filepath.ToSlash(filepath.Clean(x.FileName))
		if !fs.ValidPath(name) || name == "." {
			errs = append(errs, fmt.Errorf("refusing to delete %q outside the render directory", x.FileName))
			continue
		}
		if err := root.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
			continue
		}
		if err := w.Store.MarkRenderDeleted(ctx, x.ID); err != nil {
			errs = append(errs, err)
			continue
		}
		w.Metrics.Deleted.Inc()
	}
	removed, err := removeStaleFiles(root, cutoff)
	w.Metrics.Deleted.Add(float64(removed))
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// pathInside joins name to dir and refuses results outside dir.
func pathInside(dir, name string) (string, bool) {
	clean := filepath.ToSlash(filepath.Clean(name))
	if !fs.ValidPath(clean) || clean == "." {
		return "", false
	}
	return filepath.Join(dir, filepath.FromSlash(clean)), true
}
