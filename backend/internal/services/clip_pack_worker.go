package services

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

// ClipPackStore is the storage the clip pack worker needs.
type ClipPackStore interface {
	EnsureWeeklyPacks(ctx context.Context, weekStart time.Time, perPack int) (int, error)
	ClaimPackForBuild(ctx context.Context, staleAfter time.Duration) (*models.ClipPack, error)
	CompletePack(ctx context.Context, id uuid.UUID, fileName, sha256 string, bytes int64, clipCount int, expiresAt time.Time) error
	FailPack(ctx context.Context, id uuid.UUID, reason string) error
	ListPacksToDelete(ctx context.Context, now time.Time) ([]models.ClipPack, error)
	MarkPackDeleted(ctx context.Context, id uuid.UUID) error
}

// ClipPackRenderer draws a pack's videos; countdownrender.Renderer
// implements it.
type ClipPackRenderer interface {
	RenderCountdown(ctx context.Context, ed countdownrender.Edition, out string) error
	RenderShort(ctx context.Context, c countdownrender.Clip, out string) error
}

// ClipPackMetrics counts built packs and deleted zips.
type ClipPackMetrics struct {
	Packs   *prometheus.CounterVec
	Deleted prometheus.Counter
}

// NewClipPackMetrics creates the metrics and registers them with reg when it
// is not nil.
func NewClipPackMetrics(reg prometheus.Registerer) *ClipPackMetrics {
	m := &ClipPackMetrics{
		Packs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clpr_clip_packs_total",
			Help: "Weekly clip packs that finished building, by outcome (ready, failed).",
		}, []string{"outcome"}),
		Deleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "clpr_clip_pack_files_deleted_total",
			Help: "Expired clip pack zips deleted from the render host.",
		}),
	}
	for _, outcome := range []string{"ready", "failed"} {
		m.Packs.WithLabelValues(outcome)
	}
	if reg != nil {
		reg.MustRegister(m.Packs, m.Deleted)
	}
	return m
}

// ClipPackWorker builds weekly clip packs (#86): for each broadcaster who
// features their clips, their most-viewed clips of the past week as
// captioned 9:16 files and a 16:9 highlight reel, zipped on the render host
// and deleted Retention after they are ready. Clips are fetched only through
// the broadcaster's own Helix authorization.
type ClipPackWorker struct {
	Store    ClipPackStore
	Fetcher  ClipMediaFetcher
	Prober   MediaProber
	Renderer ClipPackRenderer
	// Captioner, when set, captions the 9:16 files.
	Captioner CountdownCaptioner
	// WorkDir holds the downloaded clips and renders while a pack builds. It
	// is under the media directory, so its 24-hour sweep catches leftovers.
	WorkDir string
	// PackDir holds the finished zips.
	PackDir string
	// Retention is how long a ready zip is kept; default 7 days.
	Retention time.Duration
	// PerPack is the most clips in a pack; default 10.
	PerPack int
	// ReadyHourUTC is the hour on Monday from which last week's packs are
	// made; default 6.
	ReadyHourUTC int
	Metrics      *ClipPackMetrics
	Now          func() time.Time

	lastEnsure time.Time
}

func (w *ClipPackWorker) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

func (w *ClipPackWorker) retention() time.Duration {
	if w.Retention > 0 {
		return w.Retention
	}
	return 7 * 24 * time.Hour
}

// lastWeek returns the Monday (UTC) starting the newest week whose packs are
// due by now.
func (w *ClipPackWorker) lastWeek(now time.Time) time.Time {
	now = now.UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	monday := day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
	if now.Before(monday.Add(time.Duration(w.ReadyHourUTC) * time.Hour)) {
		monday = monday.AddDate(0, 0, -7)
	}
	return monday.AddDate(0, 0, -7)
}

// RunOnce deletes expired zips, makes last week's packs (at most once an
// hour, so a broadcaster who opts in later in the week still gets one), and
// builds every pending pack.
func (w *ClipPackWorker) RunOnce(ctx context.Context) error {
	errs := []error{w.Sweep(ctx)}
	if now := w.now(); !now.Before(w.lastEnsure.Add(time.Hour)) {
		perPack := w.PerPack
		if perPack <= 0 {
			perPack = 10
		}
		week := w.lastWeek(now)
		created, err := w.Store.EnsureWeeklyPacks(ctx, week, perPack)
		if err != nil {
			errs = append(errs, err)
		} else {
			w.lastEnsure = now
			if created > 0 {
				utils.Info("Weekly clip packs queued", map[string]interface{}{"week_start": week.Format("2006-01-02"), "packs": created})
			}
		}
	}
	for {
		built, err := w.BuildNext(ctx)
		if err != nil || !built {
			return errors.Join(append(errs, err)...)
		}
	}
}

// BuildNext claims one pending pack and builds it. It reports whether a pack
// was claimed.
func (w *ClipPackWorker) BuildNext(ctx context.Context) (bool, error) {
	if w.Metrics == nil {
		w.Metrics = NewClipPackMetrics(nil)
	}
	pack, err := w.Store.ClaimPackForBuild(ctx, 2*time.Hour)
	if err != nil || pack == nil {
		return false, err
	}
	fields := map[string]interface{}{"pack_id": pack.ID.String(), "broadcaster_id": pack.BroadcasterID, "week_start": pack.WeekStart}
	name, digest, size, count, err := w.build(ctx, pack)
	if ctx.Err() != nil {
		return true, ctx.Err() // the claim goes stale and is taken again
	}
	var finishErr error
	if err != nil {
		fields["reason"] = err.Error()
		utils.Warn("Clip pack failed", fields)
		w.Metrics.Packs.WithLabelValues("failed").Inc()
		finishErr = w.Store.FailPack(ctx, pack.ID, err.Error())
	} else {
		fields["clips"] = count
		utils.Info("Clip pack ready", fields)
		w.Metrics.Packs.WithLabelValues("ready").Inc()
		finishErr = w.Store.CompletePack(ctx, pack.ID, name, digest, size, count, w.now().Add(w.retention()))
	}
	if errors.Is(finishErr, repository.ErrClipPackNotBuilding) {
		finishErr = nil
	}
	return true, finishErr
}

var packNameUnsafe = regexp.MustCompile(`[^a-z0-9]+`)

// packFileName is the zip's name inside its directory, for example
// clpr-pack-moonmoon-2026-11-02.zip.
func packFileName(pack *models.ClipPack) string {
	slug := strings.Trim(packNameUnsafe.ReplaceAllString(strings.ToLower(pack.BroadcasterName), "-"), "-")
	if slug == "" || len(slug) > 40 {
		slug = pack.BroadcasterID
	}
	return fmt.Sprintf("clpr-pack-%s-%s.zip", slug, pack.WeekStart)
}

func (w *ClipPackWorker) build(ctx context.Context, pack *models.ClipPack) (name, digest string, size int64, count int, err error) {
	work := filepath.Join(w.WorkDir, pack.ID.String())
	if err := os.MkdirAll(work, 0o700); err != nil {
		return "", "", 0, 0, fmt.Errorf("creating the work directory: %w", err)
	}
	// ADR-002: the downloaded clips and intermediate renders go when the
	// pack is built, whatever the outcome.
	defer os.RemoveAll(work)

	var clips []countdownrender.Clip
	var credits strings.Builder
	fmt.Fprintf(&credits, "clpr weekly clip pack for %s, week of %s\n\n", pack.BroadcasterName, pack.WeekStart)
	for _, c := range pack.Clips {
		path := filepath.Join(work, fmt.Sprintf("clip-%02d.mp4", c.Position))
		if err := w.fetch(ctx, pack, c, path); err != nil {
			if ctx.Err() != nil {
				return "", "", 0, 0, ctx.Err()
			}
			utils.Warn("Clip pack left out a clip", map[string]interface{}{
				"pack_id": pack.ID.String(), "clip_id": c.ClipID.String(), "reason": err.Error(),
			})
			continue
		}
		rank := len(clips) + 1
		clips = append(clips, countdownrender.Clip{Rank: rank, Path: path, Title: c.Title, Broadcaster: c.BroadcasterName})
		fmt.Fprintf(&credits, "%02d. %s (%d views)\n    %s\n", rank, c.Title, c.ViewCount, c.TwitchClipURL)
	}
	if len(clips) == 0 {
		return "", "", 0, 0, errors.New("no clip of the week could be downloaded")
	}

	out := filepath.Join(work, "out")
	if err := os.MkdirAll(out, 0o700); err != nil {
		return "", "", 0, 0, err
	}
	files := []string{}
	reel := filepath.Join(out, "highlights-16x9.mp4")
	ed := countdownrender.Edition{SeriesName: pack.BroadcasterName + "'s week", Date: pack.WeekStart, Clips: clips}
	if err := w.Renderer.RenderCountdown(ctx, ed, reel); err != nil {
		return "", "", 0, 0, fmt.Errorf("render the highlight reel: %w", err)
	}
	files = append(files, reel)
	for _, c := range clips {
		if w.Captioner != nil {
			words, err := w.Captioner.Words(ctx, c.Path)
			if err != nil {
				utils.Warn("Clip pack captions failed; rendering without them", map[string]interface{}{
					"pack_id": pack.ID.String(), "rank": c.Rank, "reason": err.Error(),
				})
			}
			c.Words = words
		}
		short := filepath.Join(out, fmt.Sprintf("%02d-9x16.mp4", c.Rank))
		if err := w.Renderer.RenderShort(ctx, c, short); err != nil {
			return "", "", 0, 0, fmt.Errorf("render clip %d at 9:16: %w", c.Rank, err)
		}
		files = append(files, short)
	}
	creditsPath := filepath.Join(out, "credits.txt")
	if err := os.WriteFile(creditsPath, []byte(credits.String()), 0o600); err != nil {
		return "", "", 0, 0, err
	}
	files = append(files, creditsPath)

	name = filepath.Join(pack.ID.String(), packFileName(pack))
	final := filepath.Join(w.PackDir, name)
	if err := os.MkdirAll(filepath.Dir(final), 0o700); err != nil {
		return "", "", 0, 0, err
	}
	temp := filepath.Join(filepath.Dir(final), "partial.zip")
	if err := writeZip(temp, files); err != nil {
		_ = os.Remove(temp)
		return "", "", 0, 0, fmt.Errorf("zip the pack: %w", err)
	}
	digest, size, err = fileSHA256(temp)
	if err == nil {
		err = os.Rename(temp, final)
	}
	if err != nil {
		_ = os.Remove(temp)
		return "", "", 0, 0, fmt.Errorf("store the pack: %w", err)
	}
	return name, digest, size, len(clips), nil
}

func (w *ClipPackWorker) fetch(ctx context.Context, pack *models.ClipPack, c models.ClipPackClip, path string) error {
	broadcasterID := c.BroadcasterID
	if broadcasterID == nil {
		broadcasterID = &pack.BroadcasterID
	}
	if _, err := w.Fetcher.Fetch(ctx, &models.CountdownEditionClip{Position: c.Position, ClipID: c.ClipID, Title: c.Title,
		BroadcasterName: c.BroadcasterName, BroadcasterID: broadcasterID, TwitchClipID: c.TwitchClipID,
		TwitchClipURL: c.TwitchClipURL, Duration: c.Duration}, path); err != nil {
		_ = os.Remove(path)
		return err
	}
	info, err := w.Prober.Probe(ctx, path)
	if err == nil && (info.DurationSeconds <= 0 || info.Width <= 0 || info.Height <= 0) {
		err = errors.New("the file has no video")
	}
	if err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}

// writeZip stores files (already compressed video) under their base names.
func writeZip(path string, files []string) error {
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	for _, file := range files {
		if err := addZipFile(zw, file); err != nil {
			_ = zw.Close()
			_ = out.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func addZipFile(zw *zip.Writer, path string) error {
	in, err := os.Open(path) // #nosec G304 -- the worker's own render output
	if err != nil {
		return err
	}
	defer in.Close()
	method := zip.Store
	if strings.HasSuffix(path, ".txt") {
		method = zip.Deflate
	}
	w, err := zw.CreateHeader(&zip.FileHeader{Name: filepath.Base(path), Method: method})
	if err != nil {
		return err
	}
	_, err = io.Copy(w, in)
	return err
}

// Sweep deletes the zips that expired.
func (w *ClipPackWorker) Sweep(ctx context.Context) error {
	if w.Metrics == nil {
		w.Metrics = NewClipPackMetrics(nil)
	}
	packs, err := w.Store.ListPacksToDelete(ctx, w.now())
	if err != nil || len(packs) == 0 {
		return err
	}
	root, err := os.OpenRoot(w.PackDir)
	if errors.Is(err, fs.ErrNotExist) {
		root = nil
	} else if err != nil {
		return err
	}
	if root != nil {
		defer root.Close()
	}
	var errs []error
	for _, p := range packs {
		if root != nil && p.FileName != nil {
			name := filepath.ToSlash(filepath.Clean(*p.FileName))
			if !fs.ValidPath(name) || name == "." {
				errs = append(errs, fmt.Errorf("refusing to delete %q outside the pack directory", *p.FileName))
				continue
			}
			if err := root.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, err)
				continue
			}
			_ = root.Remove(filepath.ToSlash(filepath.Dir(name))) // the pack's directory, once empty
		}
		if err := w.Store.MarkPackDeleted(ctx, p.ID); err != nil {
			errs = append(errs, err)
			continue
		}
		w.Metrics.Deleted.Inc()
	}
	return errors.Join(errs...)
}
