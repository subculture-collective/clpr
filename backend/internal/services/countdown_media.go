package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

// ErrCountdownMediaUnavailable marks a failure that is not about one clip:
// a missing tool, or an authorization store that cannot be read. The worker
// fails the edition instead of skipping clip after clip.
var ErrCountdownMediaUnavailable = errors.New("countdown media fetching is unavailable")

// countdownMaxMediaBytes bounds one clip file. A 60 s 1080p60 clip is far
// below this.
const countdownMaxMediaBytes = 300 << 20

// CountdownMediaStore is the storage the media worker needs.
type CountdownMediaStore interface {
	ClaimEditionForFetch(ctx context.Context, staleAfter time.Duration) (*models.CountdownEdition, *models.CountdownSeries, error)
	RecordMedia(ctx context.Context, m *models.CountdownEditionMedia) error
	BackfillClip(ctx context.Context, edition *models.CountdownEdition, series *models.CountdownSeries, clipID uuid.UUID, reason string) (*models.CountdownEditionClip, error)
	FinishFetch(ctx context.Context, editionID uuid.UUID, status string, reason *string) error
	ListMediaToDelete(ctx context.Context, cutoff time.Time) ([]models.CountdownEditionMedia, error)
	MarkMediaDeleted(ctx context.Context, editionID, clipID uuid.UUID) error
}

// ClipMediaFetcher writes one clip's video to dest and reports the source.
type ClipMediaFetcher interface {
	Fetch(ctx context.Context, clip *models.CountdownEditionClip, dest string) (source string, err error)
}

// MediaInfo is what ffprobe reports about a fetched file.
type MediaInfo struct {
	DurationSeconds float64
	Width, Height   int
}

// MediaProber inspects a fetched file.
type MediaProber interface {
	Probe(ctx context.Context, path string) (*MediaInfo, error)
}

// TwitchClipMediaFetcher fetches through Helix GET /clips/downloads when the
// broadcaster granted channel:manage:clips, and otherwise downloads the
// playback files Twitch serves to its own player with yt-dlp (ADR-002).
// yt-dlp uses its own client; clpr's Twitch credentials are never sent to the
// player API.
type TwitchClipMediaFetcher struct {
	Auth       TwitchAuthLookup
	Downloads  ClipDownloadURLProvider
	HTTPClient *http.Client
	YTDLPPath  string
	// HelixOnly refuses the player fallback: weekly clip packs (#86) use only
	// the broadcaster's own authorized download.
	HelixOnly bool
	Now       func() time.Time
}

func (f *TwitchClipMediaFetcher) now() time.Time {
	if f.Now != nil {
		return f.Now()
	}
	return time.Now()
}

// Fetch implements ClipMediaFetcher.
func (f *TwitchClipMediaFetcher) Fetch(ctx context.Context, clip *models.CountdownEditionClip, dest string) (string, error) {
	if f.Auth != nil && f.Downloads != nil && clip.BroadcasterID != nil && strings.TrimSpace(*clip.BroadcasterID) != "" {
		broadcasterID := strings.TrimSpace(*clip.BroadcasterID)
		auth, err := f.Auth.GetTwitchAuthByTwitchUserID(ctx, broadcasterID)
		if err != nil {
			return "", fmt.Errorf("%w: loading broadcaster Twitch authorization: %v", ErrCountdownMediaUnavailable, err)
		}
		if auth != nil && hasScope(auth.Scopes, "channel:manage:clips") && auth.ExpiresAt.After(f.now()) {
			mediaURL, err := f.Downloads.GetDownloadURL(ctx, broadcasterID, clip.TwitchClipID, auth.AccessToken)
			if err == nil {
				if err = f.download(ctx, mediaURL, dest); err == nil {
					return models.CountdownMediaSourceHelix, nil
				}
			}
			if f.HelixOnly {
				return "", fmt.Errorf("%w: Helix clip download failed: %v", ErrCountdownMediaUnavailable, err)
			}
			utils.Warn("Helix clip download failed; using the player files", map[string]interface{}{
				"clip_id": clip.ClipID.String(), "error": err.Error(),
			})
		}
	}
	if f.HelixOnly {
		return "", fmt.Errorf("%w: the broadcaster has not authorized clip downloads", ErrCountdownMediaUnavailable)
	}
	if err := f.ytdlp(ctx, clip.TwitchClipURL, dest); err != nil {
		return "", err
	}
	return models.CountdownMediaSourcePlayer, nil
}

func (f *TwitchClipMediaFetcher) download(ctx context.Context, mediaURL, dest string) error {
	parsed, err := url.Parse(mediaURL)
	if err != nil || parsed.Scheme != "https" {
		return fmt.Errorf("Twitch returned an unusable download URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return err
	}
	client := f.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading clip media: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clip media download returned status %d", resp.StatusCode)
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("%w: creating media file: %v", ErrCountdownMediaUnavailable, err)
	}
	n, copyErr := io.Copy(out, io.LimitReader(resp.Body, countdownMaxMediaBytes+1))
	closeErr := out.Close()
	if copyErr == nil && n > countdownMaxMediaBytes {
		copyErr = fmt.Errorf("clip media is larger than %d bytes", countdownMaxMediaBytes)
	}
	if copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		_ = os.Remove(dest)
		return fmt.Errorf("writing clip media: %w", copyErr)
	}
	return nil
}

// countdownClipURL accepts only Twitch clip pages, so a stored URL can never
// point yt-dlp somewhere else.
func countdownClipURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil {
		return "", fmt.Errorf("clip URL is not a Twitch clip")
	}
	host := strings.ToLower(parsed.Hostname())
	switch {
	case host == "clips.twitch.tv" && strings.Trim(parsed.Path, "/") != "":
	case (host == "www.twitch.tv" || host == "twitch.tv" || host == "m.twitch.tv") && strings.Contains(parsed.Path, "/clip/"):
	default:
		return "", fmt.Errorf("clip URL is not a Twitch clip")
	}
	return parsed.String(), nil
}

func (f *TwitchClipMediaFetcher) ytdlp(ctx context.Context, rawURL, dest string) error {
	clipURL, err := countdownClipURL(rawURL)
	if err != nil {
		return err
	}
	path := f.YTDLPPath
	if path == "" {
		path = "yt-dlp"
	}
	cmd := exec.CommandContext(ctx, path,
		"--no-config", "--no-playlist", "--no-progress", "--quiet", "--no-warnings", "--no-part",
		"--max-filesize", strconv.Itoa(countdownMaxMediaBytes),
		"-f", "best[ext=mp4][vcodec!=none][acodec!=none]/best[ext=mp4]/best",
		"-o", dest, "--", clipURL)
	output, err := cmd.CombinedOutput()
	if toolMissing(err) {
		return fmt.Errorf("%w: %v", ErrCountdownMediaUnavailable, err)
	}
	if err != nil {
		_ = os.Remove(dest)
		message := strings.TrimSpace(string(output))
		if len(message) > 300 {
			message = message[len(message)-300:]
		}
		return fmt.Errorf("yt-dlp failed: %v: %s", err, message)
	}
	if _, err := os.Stat(dest); err != nil {
		return fmt.Errorf("yt-dlp wrote no file")
	}
	return nil
}

// toolMissing reports whether a command could not start because the binary
// is absent, by name on PATH or at a configured path.
func toolMissing(err error) bool {
	var exitErr *exec.ExitError
	return err != nil && !errors.As(err, &exitErr) && (errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist))
}

// FFprobeMediaProber reads duration and frame size with ffprobe.
type FFprobeMediaProber struct {
	FFprobePath string
}

// Probe implements MediaProber.
func (p *FFprobeMediaProber) Probe(ctx context.Context, path string) (*MediaInfo, error) {
	bin := p.FFprobePath
	if bin == "" {
		bin = "ffprobe"
	}
	output, err := exec.CommandContext(ctx, bin, "-v", "error", "-print_format", "json",
		"-show_entries", "format=duration:stream=codec_type,width,height", "--", path).Output()
	if toolMissing(err) {
		return nil, fmt.Errorf("%w: %v", ErrCountdownMediaUnavailable, err)
	}
	if err != nil {
		return nil, fmt.Errorf("ffprobe could not read the file: %v", err)
	}
	return parseFFprobe(output)
}

func parseFFprobe(output []byte) (*MediaInfo, error) {
	var report struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("ffprobe output is not JSON")
	}
	info := &MediaInfo{}
	for _, s := range report.Streams {
		if s.CodecType == "video" && s.Width > 0 && s.Height > 0 {
			info.Width, info.Height = s.Width, s.Height
			break
		}
	}
	if info.Width == 0 {
		return nil, fmt.Errorf("file has no video stream")
	}
	duration, err := strconv.ParseFloat(report.Format.Duration, 64)
	if err != nil || duration <= 0 || math.IsInf(duration, 0) || math.IsNaN(duration) {
		return nil, fmt.Errorf("file has no duration")
	}
	info.DurationSeconds = duration
	return info, nil
}

// CountdownMediaMetrics counts the worker's outcomes.
type CountdownMediaMetrics struct {
	Fetches  *prometheus.CounterVec
	Skips    prometheus.Counter
	Editions *prometheus.CounterVec
	Deleted  prometheus.Counter
}

// NewCountdownMediaMetrics creates the metrics and registers them with reg
// when it is not nil.
func NewCountdownMediaMetrics(reg prometheus.Registerer) *CountdownMediaMetrics {
	m := &CountdownMediaMetrics{
		Fetches: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clpr_countdown_media_fetches_total",
			Help: "Clip media fetches for countdown editions, by source (helix, player, none) and outcome.",
		}, []string{"source", "outcome"}),
		Skips: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "clpr_countdown_media_skips_total",
			Help: "Clips skipped and backfilled because their media could not be fetched or verified.",
		}),
		Editions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clpr_countdown_media_editions_total",
			Help: "Countdown editions that finished fetching, by outcome (fetched, failed).",
		}, []string{"outcome"}),
		Deleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "clpr_countdown_media_files_deleted_total",
			Help: "Countdown media files deleted from the render host.",
		}),
	}
	// Outcomes start at zero so an alert on increase() sees the first one.
	for _, outcome := range []string{"fetched", "failed"} {
		m.Editions.WithLabelValues(outcome)
	}
	if reg != nil {
		reg.MustRegister(m.Fetches, m.Skips, m.Editions, m.Deleted)
	}
	return m
}

// CountdownMediaWorker fetches the media of selected countdown editions onto
// the render host and deletes it again within the retention window.
type CountdownMediaWorker struct {
	Store     CountdownMediaStore
	Fetcher   ClipMediaFetcher
	Prober    MediaProber
	Dir       string
	Retention time.Duration // files are deleted this long after fetching; ADR-002 caps it at 24 h
	// MaxSkips is how many clips one edition may skip before the worker gives
	// up on it. Many failures in a row usually mean the source changed.
	MaxSkips   int
	StaleAfter time.Duration
	Metrics    *CountdownMediaMetrics
	Now        func() time.Time
}

func (w *CountdownMediaWorker) ensureMetrics() {
	if w.Metrics == nil {
		w.Metrics = NewCountdownMediaMetrics(nil)
	}
}

func (w *CountdownMediaWorker) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

// RunOnce deletes expired media, then fetches every waiting edition.
func (w *CountdownMediaWorker) RunOnce(ctx context.Context) error {
	sweepErr := w.Sweep(ctx)
	for {
		claimed, err := w.FetchNext(ctx)
		if err != nil {
			return errors.Join(sweepErr, err)
		}
		if !claimed {
			return sweepErr
		}
	}
}

// FetchNext claims one edition and fetches its media. It reports whether an
// edition was claimed.
func (w *CountdownMediaWorker) FetchNext(ctx context.Context) (bool, error) {
	w.ensureMetrics()
	stale := w.StaleAfter
	if stale <= 0 {
		stale = time.Hour
	}
	edition, series, err := w.Store.ClaimEditionForFetch(ctx, stale)
	if err != nil || edition == nil {
		return false, err
	}
	editionDir := filepath.Join(w.Dir, edition.ID.String())
	if err := os.MkdirAll(editionDir, 0o700); err != nil {
		return true, w.fail(ctx, edition, fmt.Sprintf("cannot create the media directory: %v", err))
	}

	maxSkips := w.MaxSkips
	if maxSkips <= 0 {
		maxSkips = 3
	}
	skips := 0
	queue := append([]models.CountdownEditionClip(nil), edition.Clips...)
	for len(queue) > 0 {
		clip := queue[0]
		queue = queue[1:]
		source, fetchErr := w.fetchClip(ctx, edition, &clip)
		if fetchErr == nil {
			w.Metrics.Fetches.WithLabelValues(source, "ok").Inc()
			continue
		}
		if errors.Is(fetchErr, repository.ErrCountdownEditionNotFetching) {
			// Reselected while we worked; the new selection gets a fresh claim.
			_ = os.RemoveAll(editionDir)
			return true, nil
		}
		if source == "" {
			source = "none"
		}
		w.Metrics.Fetches.WithLabelValues(source, "failed").Inc()
		if ctx.Err() != nil {
			// Shutting down: leave the edition claimed; it is reclaimed once stale.
			return true, ctx.Err()
		}
		if errors.Is(fetchErr, ErrCountdownMediaUnavailable) {
			return true, errors.Join(fetchErr, w.fail(ctx, edition, fmt.Sprintf("media fetching is unavailable: %v", fetchErr)))
		}
		skips++
		w.Metrics.Skips.Inc()
		utils.Warn("Countdown clip media could not be fetched; backfilling", map[string]interface{}{
			"edition_id": edition.ID.String(), "clip_id": clip.ClipID.String(), "error": fetchErr.Error(),
		})
		if skips > maxSkips {
			return true, w.fail(ctx, edition, fmt.Sprintf("%d clips could not be fetched; last error: %v", skips, fetchErr))
		}
		next, err := w.Store.BackfillClip(ctx, edition, series, clip.ClipID, fetchErr.Error())
		if errors.Is(err, repository.ErrCountdownEditionNotFetching) {
			_ = os.RemoveAll(editionDir)
			return true, nil
		}
		if err != nil {
			return true, errors.Join(err, w.fail(ctx, edition, fmt.Sprintf("backfill failed: %v", err)))
		}
		if next == nil {
			return true, w.fail(ctx, edition, fmt.Sprintf("not enough eligible clips after %d skips", skips))
		}
		queue = append(queue, *next)
	}

	if err := w.Store.FinishFetch(ctx, edition.ID, models.CountdownEditionFetched, nil); err != nil {
		if errors.Is(err, repository.ErrCountdownEditionNotFetching) {
			return true, nil
		}
		return true, err
	}
	w.Metrics.Editions.WithLabelValues(models.CountdownEditionFetched).Inc()
	utils.Info("Countdown edition media fetched", map[string]interface{}{
		"edition_id": edition.ID.String(), "series": edition.SeriesSlug, "edition_date": edition.EditionDate, "skips": skips,
	})
	return true, nil
}

// fetchClip downloads, verifies and records one clip.
func (w *CountdownMediaWorker) fetchClip(ctx context.Context, edition *models.CountdownEdition, clip *models.CountdownEditionClip) (string, error) {
	name := filepath.Join(edition.ID.String(), clip.ClipID.String()+".mp4")
	dest := filepath.Join(w.Dir, name)
	temp := dest + ".partial"
	_ = os.Remove(temp)
	source, err := w.Fetcher.Fetch(ctx, clip, temp)
	if err != nil {
		return source, err
	}
	info, err := w.Prober.Probe(ctx, temp)
	if err == nil {
		err = verifyCountdownMedia(clip, info)
	}
	if err != nil {
		_ = os.Remove(temp)
		return source, err
	}
	digest, size, err := fileSHA256(temp)
	if err != nil {
		_ = os.Remove(temp)
		return source, fmt.Errorf("%w: hashing media: %v", ErrCountdownMediaUnavailable, err)
	}
	if err := os.Rename(temp, dest); err != nil {
		_ = os.Remove(temp)
		return source, fmt.Errorf("%w: storing media: %v", ErrCountdownMediaUnavailable, err)
	}
	err = w.Store.RecordMedia(ctx, &models.CountdownEditionMedia{
		EditionID: edition.ID, ClipID: clip.ClipID, Source: source, FileName: name, SHA256: digest, Bytes: size,
		DurationSeconds: info.DurationSeconds, Width: info.Width, Height: info.Height,
	})
	if err != nil {
		_ = os.Remove(dest)
		if errors.Is(err, repository.ErrCountdownEditionNotFetching) {
			return source, err
		}
		return source, fmt.Errorf("%w: recording media: %v", ErrCountdownMediaUnavailable, err)
	}
	return source, nil
}

// verifyCountdownMedia checks a probed file against the clip it should be.
func verifyCountdownMedia(clip *models.CountdownEditionClip, info *MediaInfo) error {
	if info.DurationSeconds > repository.CountdownMaxClipSeconds+5 {
		return fmt.Errorf("file lasts %.1f s, longer than any eligible clip", info.DurationSeconds)
	}
	if clip.Duration != nil && math.Abs(info.DurationSeconds-*clip.Duration) > 2 {
		return fmt.Errorf("file lasts %.1f s but the clip lasts %.1f s", info.DurationSeconds, *clip.Duration)
	}
	return nil
}

func fileSHA256(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func (w *CountdownMediaWorker) fail(ctx context.Context, edition *models.CountdownEdition, reason string) error {
	utils.Warn("Countdown edition media failed", map[string]interface{}{
		"edition_id": edition.ID.String(), "series": edition.SeriesSlug, "edition_date": edition.EditionDate, "reason": reason,
	})
	err := w.Store.FinishFetch(context.WithoutCancel(ctx), edition.ID, models.CountdownEditionFailed, &reason)
	if errors.Is(err, repository.ErrCountdownEditionNotFetching) {
		err = nil
	}
	if err == nil {
		w.Metrics.Editions.WithLabelValues(models.CountdownEditionFailed).Inc()
	}
	// Files of a failed edition are deleted by the next sweep.
	return err
}

// Sweep deletes the files of failed editions and files older than the
// retention window, then any stray file older than the window (from a crash).
// Every lookup and removal goes through os.Root, so a symlink cannot lead it
// outside the media directory.
func (w *CountdownMediaWorker) Sweep(ctx context.Context) error {
	w.ensureMetrics()
	retention := w.Retention
	if retention <= 0 || retention > 24*time.Hour {
		retention = 24 * time.Hour
	}
	cutoff := w.now().Add(-retention)
	root, err := os.OpenRoot(w.Dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer root.Close()

	var errs []error
	media, err := w.Store.ListMediaToDelete(ctx, cutoff)
	if err != nil {
		errs = append(errs, err)
	}
	for _, m := range media {
		name := filepath.ToSlash(filepath.Clean(m.FileName))
		if !fs.ValidPath(name) || name == "." {
			errs = append(errs, fmt.Errorf("refusing to delete %q outside the media directory", m.FileName))
			continue
		}
		if err := root.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
			continue
		}
		if err := w.Store.MarkMediaDeleted(ctx, m.EditionID, m.ClipID); err != nil {
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

// removeStaleFiles deletes regular files modified before cutoff under root,
// then the directories left empty.
func removeStaleFiles(root *os.Root, cutoff time.Time) (int, error) {
	removed := 0
	var dirs []string
	err := fs.WalkDir(root.FS(), ".", func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() {
			if name != "." {
				dirs = append(dirs, name)
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, err := d.Info()
		if err != nil || !info.ModTime().Before(cutoff) {
			return nil
		}
		if err := root.Remove(name); err == nil {
			removed++
		}
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- {
		_ = root.Remove(dirs[i]) // only succeeds when empty
	}
	return removed, err
}
