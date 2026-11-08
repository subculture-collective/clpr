package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

// CountdownPublishStore is the storage the YouTube publisher needs.
type CountdownPublishStore interface {
	ClaimEditionForPublish(ctx context.Context, staleAfter time.Duration) (*models.CountdownEdition, *models.CountdownSeries, error)
	ListEditionRenders(ctx context.Context, editionID uuid.UUID) ([]models.CountdownRender, error)
	ClipDurations(ctx context.Context, editionID uuid.UUID) (map[uuid.UUID]float64, error)
	ListUploads(ctx context.Context, editionID uuid.UUID) (map[uuid.UUID]models.CountdownUpload, error)
	RecordUpload(ctx context.Context, u *models.CountdownUpload) error
	FinishPublish(ctx context.Context, editionID uuid.UUID, status string, reason *string) error
	SeriesPlaylist(ctx context.Context, seriesID uuid.UUID) (string, error)
	SaveSeriesPlaylist(ctx context.Context, seriesID uuid.UUID, playlistID string) error
}

// YouTubeUploader is the part of the YouTube Data API the publisher uses.
type YouTubeUploader interface {
	UploadVideo(ctx context.Context, path string, v YouTubeVideo) (string, error)
	SetThumbnail(ctx context.Context, videoID, path string) error
	CreatePlaylist(ctx context.Context, title, description, privacy string) (string, error)
	AddToPlaylist(ctx context.Context, playlistID, videoID string) error
}

// CountdownPublishMetrics counts uploads and published editions.
type CountdownPublishMetrics struct {
	Uploads  *prometheus.CounterVec
	Editions *prometheus.CounterVec
}

// NewCountdownPublishMetrics creates the metrics and registers them with reg
// when it is not nil.
func NewCountdownPublishMetrics(reg prometheus.Registerer) *CountdownPublishMetrics {
	m := &CountdownPublishMetrics{
		Uploads: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clpr_countdown_youtube_uploads_total",
			Help: "YouTube uploads of countdown renders, by kind (countdown, short, thumbnail, playlist) and outcome (ok, quota, failed).",
		}, []string{"kind", "outcome"}),
		Editions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clpr_countdown_publish_editions_total",
			Help: "Countdown editions leaving the publish stage, by outcome (published, retry, quota, auth, failed).",
		}, []string{"outcome"}),
	}
	// Outcomes start at zero so an alert on increase() sees the first one.
	for _, outcome := range []string{"published", "retry", "quota", "auth", "failed"} {
		m.Editions.WithLabelValues(outcome)
	}
	for _, kind := range []string{"countdown", "short"} {
		for _, outcome := range []string{"ok", "quota", "failed"} {
			m.Uploads.WithLabelValues(kind, outcome)
		}
	}
	if reg != nil {
		reg.MustRegister(m.Uploads, m.Editions)
	}
	return m
}

// YouTube privacy settings for the publisher. PrivacyScheduled uploads
// private with a publish time.
const (
	YouTubePrivacyPrivate   = "private"
	YouTubePrivacyUnlisted  = "unlisted"
	YouTubePrivacyPublic    = "public"
	YouTubePrivacyScheduled = "scheduled"
)

// CountdownPublisher uploads rendered editions to YouTube (#79): the
// countdown video with its thumbnail into the series playlist, then the
// Shorts. Until the API project passes YouTube's audit, YouTube keeps every
// upload private, so Privacy defaults to private.
type CountdownPublisher struct {
	Store     CountdownPublishStore
	YouTube   YouTubeUploader
	RenderDir string
	BaseURL   string
	// Privacy is private, unlisted, public or scheduled.
	Privacy string
	// PublishHourUTC is when scheduled uploads go public on the edition date.
	PublishHourUTC int
	// RankCardSeconds is the render template's card length, for chapters.
	RankCardSeconds float64
	Metrics         *CountdownPublishMetrics
	Now             func() time.Time

	pausedUntil time.Time
}

func (p *CountdownPublisher) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

// nextQuotaReset is the next 08:00 UTC. YouTube resets quotas at midnight
// Pacific time, which is 08:00 UTC in winter and 07:00 UTC in summer.
func nextQuotaReset(now time.Time) time.Time {
	reset := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.UTC)
	if !reset.After(now) {
		reset = reset.Add(24 * time.Hour)
	}
	return reset
}

// RunOnce publishes every rendered edition, unless a quota error or a
// transient failure paused publishing.
func (p *CountdownPublisher) RunOnce(ctx context.Context) error {
	if p.Metrics == nil {
		p.Metrics = NewCountdownPublishMetrics(nil)
	}
	for p.now().After(p.pausedUntil) {
		claimed, err := p.PublishNext(ctx)
		if err != nil || !claimed {
			return err
		}
	}
	return nil
}

// PublishNext claims one rendered edition and uploads it. It reports whether
// an edition was claimed.
func (p *CountdownPublisher) PublishNext(ctx context.Context) (bool, error) {
	edition, series, err := p.Store.ClaimEditionForPublish(ctx, 3*time.Hour)
	if err != nil || edition == nil {
		return false, err
	}
	fields := map[string]interface{}{"edition_id": edition.ID.String(), "series": edition.SeriesSlug, "edition_date": edition.EditionDate}
	err = p.publish(ctx, edition, series)
	if ctx.Err() != nil {
		return true, ctx.Err() // keep the claim; it is taken again when stale
	}
	var outcome string
	var yt *YouTubeError
	switch {
	case err == nil:
		outcome = models.CountdownEditionPublished
		p.Metrics.Editions.WithLabelValues("published").Inc()
		utils.Info("Countdown edition published", fields)
	case IsYouTubeQuotaError(err):
		outcome = models.CountdownEditionRendered
		p.pausedUntil = nextQuotaReset(p.now())
		p.Metrics.Editions.WithLabelValues("quota").Inc()
		fields["reason"], fields["retry_at"] = err.Error(), p.pausedUntil.Format(time.RFC3339)
		utils.Warn("YouTube quota reached; publishing paused", fields)
	case errors.As(err, &yt) && yt.AuthFailed():
		// Failing the edition would delete its renders; wait for the channel
		// to be authorized again.
		outcome = models.CountdownEditionRendered
		p.pausedUntil = p.now().Add(time.Hour)
		p.Metrics.Editions.WithLabelValues("auth").Inc()
		fields["reason"] = err.Error()
		utils.Error("YouTube authorization failed; publishing paused for an hour", err, fields)
	case errors.As(err, &yt) && yt.Transient(), !errors.As(err, &yt) && !errors.Is(err, errCountdownPublishPermanent):
		// Network errors and YouTube's 5xx and rate limits: retry later.
		outcome = models.CountdownEditionRendered
		p.pausedUntil = p.now().Add(15 * time.Minute)
		p.Metrics.Editions.WithLabelValues("retry").Inc()
		fields["reason"] = err.Error()
		utils.Warn("Countdown publish failed; retrying in 15 minutes", fields)
	default:
		outcome = models.CountdownEditionFailed
		p.Metrics.Editions.WithLabelValues("failed").Inc()
		fields["reason"] = err.Error()
		utils.Warn("Countdown publish failed", fields)
	}
	var reason *string
	if err != nil {
		message := err.Error()
		reason = &message
	}
	finishErr := p.Store.FinishPublish(ctx, edition.ID, outcome, reason)
	if errors.Is(finishErr, repository.ErrCountdownEditionNotPublishing) {
		finishErr = nil
	}
	if outcome == models.CountdownEditionFailed {
		return true, errors.Join(err, finishErr)
	}
	return true, finishErr
}

var errCountdownPublishPermanent = errors.New("countdown cannot be published")

func (p *CountdownPublisher) privacy(edition *models.CountdownEdition) (string, *time.Time) {
	switch p.Privacy {
	case YouTubePrivacyUnlisted, YouTubePrivacyPublic:
		return p.Privacy, nil
	case YouTubePrivacyScheduled:
		day, err := time.Parse("2006-01-02", edition.EditionDate)
		if err != nil {
			return YouTubePrivacyPrivate, nil
		}
		at := day.Add(time.Duration(p.PublishHourUTC) * time.Hour)
		if earliest := p.now().Add(15 * time.Minute); at.Before(earliest) {
			at = earliest
		}
		at = at.Truncate(time.Minute)
		return YouTubePrivacyPrivate, &at
	}
	return YouTubePrivacyPrivate, nil
}

func (p *CountdownPublisher) publish(ctx context.Context, edition *models.CountdownEdition, series *models.CountdownSeries) error {
	renders, err := p.Store.ListEditionRenders(ctx, edition.ID)
	if err != nil {
		return err
	}
	var countdown, thumbnail *models.CountdownRender
	shorts := map[uuid.UUID]models.CountdownRender{}
	for i := range renders {
		switch x := renders[i]; x.Kind {
		case models.CountdownRenderCountdown:
			countdown = &renders[i]
		case models.CountdownRenderThumbnail:
			thumbnail = &renders[i]
		case models.CountdownRenderShort:
			if x.ClipID != nil {
				shorts[*x.ClipID] = x
			}
		}
	}
	uploads, err := p.Store.ListUploads(ctx, edition.ID)
	if err != nil {
		return err
	}
	privacy, publishAt := p.privacy(edition)
	clips := append([]models.CountdownEditionClip(nil), edition.Clips...)
	sort.Slice(clips, func(i, j int) bool { return clips[i].Position < clips[j].Position })

	if countdown == nil {
		if !uploadedKind(uploads, models.CountdownRenderCountdown) {
			return fmt.Errorf("%w: the countdown video is no longer on the render host", errCountdownPublishPermanent)
		}
	} else if _, done := uploads[countdown.ID]; !done {
		durations, err := p.Store.ClipDurations(ctx, edition.ID)
		if err != nil {
			return err
		}
		video := CountdownYouTubeVideo(series, edition.EditionDate, clips, durations, p.RankCardSeconds, p.BaseURL)
		video.PrivacyStatus, video.PublishAt = privacy, publishAt
		id, err := p.upload(ctx, countdown, video)
		if err != nil {
			return err
		}
		if thumbnail != nil {
			p.bestEffort(ctx, "thumbnail", edition, func() error {
				return p.YouTube.SetThumbnail(ctx, id, filepath.Join(p.RenderDir, thumbnail.FileName))
			})
		}
		p.bestEffort(ctx, "playlist", edition, func() error { return p.addToPlaylist(ctx, series, id, privacy) })
	}

	for _, clip := range clips {
		x, ok := shorts[clip.ClipID]
		if !ok {
			continue
		}
		if _, done := uploads[x.ID]; done {
			continue
		}
		video := ShortYouTubeVideo(series, edition.EditionDate, clip, p.BaseURL)
		video.PrivacyStatus, video.PublishAt = privacy, publishAt
		if _, err := p.upload(ctx, &x, video); err != nil {
			return err
		}
	}
	return nil
}

func uploadedKind(uploads map[uuid.UUID]models.CountdownUpload, kind string) bool {
	for _, u := range uploads {
		if u.Kind == kind {
			return true
		}
	}
	return false
}

// upload sends one render and records it at once, so a retry never uploads
// it twice.
func (p *CountdownPublisher) upload(ctx context.Context, x *models.CountdownRender, video YouTubeVideo) (string, error) {
	path, ok := pathInside(p.RenderDir, x.FileName)
	if !ok {
		return "", fmt.Errorf("%w: render path %q is outside the render directory", errCountdownPublishPermanent, x.FileName)
	}
	id, err := p.YouTube.UploadVideo(ctx, path, video)
	if err != nil {
		p.Metrics.Uploads.WithLabelValues(x.Kind, uploadOutcome(err)).Inc()
		return "", fmt.Errorf("upload %s: %w", x.Kind, err)
	}
	p.Metrics.Uploads.WithLabelValues(x.Kind, "ok").Inc()
	err = p.Store.RecordUpload(context.WithoutCancel(ctx), &models.CountdownUpload{RenderID: x.ID, EditionID: x.EditionID, Kind: x.Kind,
		ClipID: x.ClipID, YouTubeVideoID: id, PrivacyStatus: video.PrivacyStatus, PublishAt: video.PublishAt})
	if err != nil {
		// The video exists; failing permanently keeps the edition from
		// uploading it again.
		return "", fmt.Errorf("%w: record upload %s of %s: %v", errCountdownPublishPermanent, id, x.Kind, err)
	}
	return id, nil
}

func uploadOutcome(err error) string {
	if IsYouTubeQuotaError(err) {
		return "quota"
	}
	return "failed"
}

// bestEffort runs a step whose failure leaves the video published.
func (p *CountdownPublisher) bestEffort(ctx context.Context, kind string, edition *models.CountdownEdition, step func() error) {
	if err := step(); err != nil {
		p.Metrics.Uploads.WithLabelValues(kind, uploadOutcome(err)).Inc()
		utils.Warn("YouTube "+kind+" step failed; the video is uploaded", map[string]interface{}{
			"edition_id": edition.ID.String(), "reason": err.Error(),
		})
		return
	}
	p.Metrics.Uploads.WithLabelValues(kind, "ok").Inc()
}

func (p *CountdownPublisher) addToPlaylist(ctx context.Context, series *models.CountdownSeries, videoID, privacy string) error {
	playlist, err := p.Store.SeriesPlaylist(ctx, series.ID)
	if err != nil {
		return err
	}
	if playlist == "" {
		playlist, err = p.YouTube.CreatePlaylist(ctx, youTubeText(fmt.Sprintf("Daily Twitch clip countdown: %s", series.Name), 150),
			"The Twitch clips that drew views fastest each day, counted down. Every clip belongs to its broadcaster.", privacy)
		if err != nil {
			return err
		}
		if err := p.Store.SaveSeriesPlaylist(ctx, series.ID, playlist); err != nil {
			return err
		}
	}
	return p.YouTube.AddToPlaylist(ctx, playlist, videoID)
}

// YouTube rejects angle brackets in titles and descriptions.
var youTubeUnsafe = strings.NewReplacer("<", "‹", ">", "›")

// youTubeText cleans text for a title and cuts it to limit characters.
func youTubeText(s string, limit int) string {
	s = youTubeUnsafe.Replace(strings.Join(strings.Fields(s), " "))
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:limit-1])) + "…"
}

func chapterTime(seconds float64) string {
	total := int(seconds)
	if total >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", total/3600, total/60%60, total%60)
	}
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

func twitchChannelURL(broadcaster string) string {
	return "https://www.twitch.tv/" + strings.ToLower(strings.TrimSpace(broadcaster))
}

func countdownPageURL(baseURL, date, slug string) string {
	return fmt.Sprintf("%s/countdowns/%s/%s", strings.TrimRight(baseURL, "/"), date, slug)
}

func longDate(date string) string {
	if t, err := time.Parse("2006-01-02", date); err == nil {
		return t.Format("January 2, 2006")
	}
	return date
}

// YouTube limits: 100 characters of title, 5,000 bytes of description and
// 500 characters of tags.
const (
	youTubeTitleLimit       = 100
	youTubeDescriptionBytes = 5000
	youTubeTagsLimit        = 500
	youTubeGamingCategory   = "20"
)

// CountdownYouTubeVideo builds the title, description (with chapters and
// credits) and tags of a countdown video. clips are in rank order, #1 first;
// the video plays them from the lowest rank up.
func CountdownYouTubeVideo(series *models.CountdownSeries, date string, clips []models.CountdownEditionClip, durations map[uuid.UUID]float64, cardSeconds float64, baseURL string) YouTubeVideo {
	title := youTubeText(fmt.Sprintf("Top %d Twitch Clips of %s: %s", len(clips), longDate(date), series.Name), youTubeTitleLimit)
	var chapters, credits strings.Builder
	offset := 0.0
	for i := len(clips) - 1; i >= 0; i-- {
		c := clips[i]
		fmt.Fprintf(&chapters, "%s #%d %s (%s)\n", chapterTime(offset), c.Position, youTubeText(c.Title, 60), c.BroadcasterName)
		duration := durations[c.ClipID]
		if duration == 0 && c.Duration != nil {
			duration = *c.Duration
		}
		offset += cardSeconds + duration
	}
	fmt.Fprintf(&chapters, "%s Credits\n", chapterTime(offset))
	for _, c := range clips {
		fmt.Fprintf(&credits, "#%d %s by %s: %s (clip: %s)\n", c.Position, youTubeText(c.Title, 80), c.BroadcasterName,
			twitchChannelURL(c.BroadcasterName), c.TwitchClipURL)
	}
	base := strings.TrimRight(baseURL, "/")
	description := fmt.Sprintf("The %d Twitch clips that drew views fastest on %s, counted down to #1.\n\n"+
		"Full list with every clip: %s\n\n%s\nClips and channels\n%s\n"+
		"Every clip belongs to its broadcaster; follow them on Twitch. To have a clip removed, or your channel left out of future countdowns, write to us at %s/contact.",
		len(clips), longDate(date), countdownPageURL(base, date, series.Slug), chapters.String(), credits.String(), base)
	return YouTubeVideo{
		Title:       title,
		Description: cutBytes(youTubeUnsafe.Replace(description), youTubeDescriptionBytes),
		Tags:        youTubeTags(append([]string{"twitch", "twitch clips", "top clips", "daily countdown", series.Name}, broadcasterNames(clips)...)),
		CategoryID:  youTubeGamingCategory,
	}
}

// ShortYouTubeVideo builds the title, description and tags of one Short.
func ShortYouTubeVideo(series *models.CountdownSeries, date string, clip models.CountdownEditionClip, baseURL string) YouTubeVideo {
	const suffix = " #Shorts"
	title := youTubeText(clip.Title, youTubeTitleLimit-utf8.RuneCountInString(suffix)) + suffix
	description := fmt.Sprintf("#%d in the %s countdown of %s.\n\nClip by %s: %s\nOriginal clip: %s\nFull countdown: %s\n\n"+
		"Every clip belongs to its broadcaster; follow them on Twitch.\n#Shorts",
		clip.Position, series.Name, longDate(date), clip.BroadcasterName, twitchChannelURL(clip.BroadcasterName), clip.TwitchClipURL,
		countdownPageURL(baseURL, date, series.Slug))
	return YouTubeVideo{
		Title:       title,
		Description: cutBytes(youTubeUnsafe.Replace(description), youTubeDescriptionBytes),
		Tags:        youTubeTags([]string{"twitch", "twitch clips", "shorts", clip.BroadcasterName, series.Name}),
		CategoryID:  youTubeGamingCategory,
	}
}

func broadcasterNames(clips []models.CountdownEditionClip) []string {
	out := make([]string, 0, len(clips))
	for _, c := range clips {
		out = append(out, c.BroadcasterName)
	}
	return out
}

// youTubeTags drops duplicates and stops before the 500-character limit,
// counting the separators and quotes YouTube adds.
func youTubeTags(tags []string) []string {
	seen := map[string]bool{}
	out := []string{}
	total := 0
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.NewReplacer("<", "", ">", "", ",", "").Replace(tag))
		key := strings.ToLower(tag)
		if tag == "" || seen[key] {
			continue
		}
		cost := utf8.RuneCountInString(tag) + 1
		if strings.Contains(tag, " ") {
			cost += 2
		}
		if total+cost > youTubeTagsLimit {
			break
		}
		seen[key] = true
		total += cost
		out = append(out, tag)
	}
	return out
}

func cutBytes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
