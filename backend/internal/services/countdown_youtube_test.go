package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func youTubeClips(n int, title string) []models.CountdownEditionClip {
	clips := make([]models.CountdownEditionClip, n)
	for i := range clips {
		clips[i] = models.CountdownEditionClip{Position: i + 1, ClipID: uuid.New(), Title: fmt.Sprintf("%s %d", title, i+1),
			BroadcasterName: fmt.Sprintf("Streamer%d", i+1), TwitchClipURL: fmt.Sprintf("https://clips.twitch.tv/clip%d", i+1)}
	}
	return clips
}

func TestCountdownYouTubeVideoMetadata(t *testing.T) {
	series := &models.CountdownSeries{Slug: "overall", Name: "Overall"}
	clips := youTubeClips(3, "Clutch <ace>")
	durations := map[uuid.UUID]float64{clips[0].ClipID: 30, clips[1].ClipID: 20.5, clips[2].ClipID: 12}
	v := CountdownYouTubeVideo(series, "2026-11-07", clips, durations, 2.5, "https://clpr.tv/")
	if v.Title != "Top 3 Twitch Clips of November 7, 2026: Overall" || v.CategoryID != "20" {
		t.Fatalf("title %q category %q", v.Title, v.CategoryID)
	}
	// The video plays #3, #2, #1, then the credits; each chapter is a card
	// (2.5 s) plus its clip.
	for _, line := range []string{
		"0:00 #3 Clutch ‹ace› 3 (Streamer3)",
		"0:14 #2 Clutch ‹ace› 2 (Streamer2)",
		"0:37 #1 Clutch ‹ace› 1 (Streamer1)",
		"1:10 Credits",
		"#1 Clutch ‹ace› 1 by Streamer1: https://www.twitch.tv/streamer1 (clip: https://clips.twitch.tv/clip1)",
		"Full list with every clip: https://clpr.tv/countdowns/2026-11-07/overall",
		"https://clpr.tv/contact",
	} {
		if !strings.Contains(v.Description, line+"\n") && !strings.Contains(v.Description, line) {
			t.Errorf("description lacks %q:\n%s", line, v.Description)
		}
	}
	if strings.ContainsAny(v.Description+v.Title, "<>") {
		t.Error("angle brackets reach YouTube")
	}

	long := youTubeClips(50, strings.Repeat("a very long clip title ", 10))
	v = CountdownYouTubeVideo(series, "2026-11-07", long, nil, 2.5, "https://clpr.tv")
	if len(v.Description) > youTubeDescriptionBytes || !utf8.ValidString(v.Description) {
		t.Fatalf("description is %d bytes", len(v.Description))
	}
	tags := 0
	for _, tag := range v.Tags {
		tags += utf8.RuneCountInString(tag) + 1
	}
	if tags > youTubeTagsLimit {
		t.Fatalf("tags use %d characters", tags)
	}

	short := ShortYouTubeVideo(series, "2026-11-07", models.CountdownEditionClip{Position: 2, Title: strings.Repeat("x", 200),
		BroadcasterName: "Bravo", TwitchClipURL: "https://clips.twitch.tv/b"}, "https://clpr.tv")
	if !strings.HasSuffix(short.Title, " #Shorts") || utf8.RuneCountInString(short.Title) > youTubeTitleLimit {
		t.Fatalf("short title %q", short.Title)
	}
	if !strings.Contains(short.Description, "https://www.twitch.tv/bravo") || !strings.Contains(short.Description, "#2 in the Overall countdown") {
		t.Fatalf("short description %q", short.Description)
	}
}

type fakePublishStore struct {
	edition  *models.CountdownEdition
	series   *models.CountdownSeries
	renders  []models.CountdownRender
	uploads  map[uuid.UUID]models.CountdownUpload
	playlist string
	status   string
	reason   string
	claims   int
}

func (s *fakePublishStore) ClaimEditionForPublish(context.Context, time.Duration) (*models.CountdownEdition, *models.CountdownSeries, error) {
	if s.status != models.CountdownEditionRendered {
		return nil, nil, nil
	}
	s.claims++
	s.status = models.CountdownEditionPublishing
	return s.edition, s.series, nil
}
func (s *fakePublishStore) ListEditionRenders(context.Context, uuid.UUID) ([]models.CountdownRender, error) {
	return s.renders, nil
}
func (s *fakePublishStore) ClipDurations(context.Context, uuid.UUID) (map[uuid.UUID]float64, error) {
	return map[uuid.UUID]float64{}, nil
}
func (s *fakePublishStore) ListUploads(context.Context, uuid.UUID) (map[uuid.UUID]models.CountdownUpload, error) {
	out := map[uuid.UUID]models.CountdownUpload{}
	for k, v := range s.uploads {
		out[k] = v
	}
	return out, nil
}
func (s *fakePublishStore) RecordUpload(_ context.Context, u *models.CountdownUpload) error {
	s.uploads[u.RenderID] = *u
	return nil
}
func (s *fakePublishStore) FinishPublish(_ context.Context, _ uuid.UUID, status string, reason *string) error {
	s.status, s.reason = status, ""
	if reason != nil {
		s.reason = *reason
	}
	return nil
}
func (s *fakePublishStore) SeriesPlaylist(context.Context, uuid.UUID) (string, error) {
	return s.playlist, nil
}
func (s *fakePublishStore) SaveSeriesPlaylist(_ context.Context, _ uuid.UUID, id string) error {
	s.playlist = id
	return nil
}

type fakeUploader struct {
	uploads    []YouTubeVideo
	thumbnails []string
	items      []string
	failOn     int // fail the nth upload (1-based) with err
	err        error
	thumbErr   error
}

func (u *fakeUploader) UploadVideo(_ context.Context, path string, v YouTubeVideo) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	if u.failOn == len(u.uploads)+1 && u.err != nil {
		return "", u.err
	}
	u.uploads = append(u.uploads, v)
	return fmt.Sprintf("video%07d", len(u.uploads)), nil
}
func (u *fakeUploader) SetThumbnail(_ context.Context, id, _ string) error {
	u.thumbnails = append(u.thumbnails, id)
	return u.thumbErr
}
func (u *fakeUploader) CreatePlaylist(context.Context, string, string, string) (string, error) {
	return "PLnew00000000", nil
}
func (u *fakeUploader) AddToPlaylist(_ context.Context, playlist, id string) error {
	u.items = append(u.items, playlist+"/"+id)
	return nil
}

func publishFixture(t *testing.T, shorts int) (*fakePublishStore, *fakeUploader, *CountdownPublisher) {
	t.Helper()
	dir := t.TempDir()
	edition := &models.CountdownEdition{ID: uuid.New(), SeriesSlug: "overall", EditionDate: "2026-11-07", Clips: youTubeClips(3, "clip")}
	store := &fakePublishStore{edition: edition, series: &models.CountdownSeries{ID: uuid.New(), Slug: "overall", Name: "Overall"},
		uploads: map[uuid.UUID]models.CountdownUpload{}, status: models.CountdownEditionRendered}
	add := func(kind, name string, clip *uuid.UUID) {
		rel := filepath.Join(edition.ID.String(), name)
		_ = os.MkdirAll(filepath.Join(dir, edition.ID.String()), 0o700)
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(kind), 0o600); err != nil {
			t.Fatal(err)
		}
		store.renders = append(store.renders, models.CountdownRender{ID: uuid.New(), EditionID: edition.ID, Kind: kind, ClipID: clip, FileName: rel})
	}
	add(models.CountdownRenderCountdown, "countdown.mp4", nil)
	add(models.CountdownRenderThumbnail, "thumbnail.jpg", nil)
	for i := 0; i < shorts; i++ {
		add(models.CountdownRenderShort, fmt.Sprintf("short-%02d.mp4", i+1), &edition.Clips[i].ClipID)
	}
	uploader := &fakeUploader{}
	now := time.Date(2026, 11, 7, 13, 0, 0, 0, time.UTC)
	p := &CountdownPublisher{Store: store, YouTube: uploader, RenderDir: dir, BaseURL: "https://clpr.tv", RankCardSeconds: 2.5,
		Metrics: NewCountdownPublishMetrics(nil), Now: func() time.Time { return now }}
	return store, uploader, p
}

func TestCountdownPublisherUploadsEdition(t *testing.T) {
	store, uploader, p := publishFixture(t, 2)
	if err := p.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionPublished {
		t.Fatalf("status %q (%s)", store.status, store.reason)
	}
	if len(uploader.uploads) != 3 || !strings.HasPrefix(uploader.uploads[0].Title, "Top 3") ||
		!strings.HasSuffix(uploader.uploads[1].Title, "#Shorts") || !strings.HasPrefix(uploader.uploads[1].Description, "#1 in") {
		t.Fatalf("uploads = %+v", uploader.uploads)
	}
	for _, v := range uploader.uploads {
		if v.PrivacyStatus != "private" || v.PublishAt != nil {
			t.Fatalf("privacy %q publishAt %v; uploads must stay private by default", v.PrivacyStatus, v.PublishAt)
		}
	}
	if len(uploader.thumbnails) != 1 || uploader.thumbnails[0] != "video0000001" ||
		len(uploader.items) != 1 || uploader.items[0] != "PLnew00000000/video0000001" || store.playlist != "PLnew00000000" {
		t.Fatalf("thumbnails %v, playlist items %v", uploader.thumbnails, uploader.items)
	}
	if len(store.uploads) != 3 {
		t.Fatalf("recorded %d uploads", len(store.uploads))
	}
}

func TestCountdownPublisherQuotaPausesAndResumesWithoutDuplicates(t *testing.T) {
	store, uploader, p := publishFixture(t, 2)
	uploader.failOn, uploader.err = 2, &YouTubeError{Status: 403, Reason: "uploadLimitExceeded"}
	if err := p.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionRendered || !strings.Contains(store.reason, "uploadLimitExceeded") {
		t.Fatalf("status %q reason %q", store.status, store.reason)
	}
	if got := testutil.ToFloat64(p.Metrics.Editions.WithLabelValues("quota")); got != 1 {
		t.Fatalf("quota editions = %v", got)
	}
	// Paused until 08:00 UTC: nothing is claimed in the meantime.
	if err := p.RunOnce(context.Background()); err != nil || store.claims != 1 {
		t.Fatalf("claimed during the pause: %d, %v", store.claims, err)
	}
	p.Now = func() time.Time { return time.Date(2026, 11, 8, 8, 1, 0, 0, time.UTC) }
	uploader.err = nil
	if err := p.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionPublished || len(uploader.uploads) != 3 {
		t.Fatalf("status %q after %d uploads (the countdown must not upload twice)", store.status, len(uploader.uploads))
	}
}

func TestCountdownPublisherFailureModes(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status string
		paused time.Duration
	}{
		{"rejected metadata", &YouTubeError{Status: 400, Reason: "invalidTitle"}, models.CountdownEditionFailed, 0},
		{"revoked authorization", &YouTubeError{Status: 400, Reason: "invalid_grant"}, models.CountdownEditionRendered, time.Hour},
		{"YouTube outage", &YouTubeError{Status: 503}, models.CountdownEditionRendered, 15 * time.Minute},
		{"network", errors.New("connection reset"), models.CountdownEditionRendered, 15 * time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, uploader, p := publishFixture(t, 0)
			uploader.failOn, uploader.err = 1, tc.err
			_ = p.RunOnce(context.Background())
			if store.status != tc.status {
				t.Fatalf("status %q, want %q", store.status, tc.status)
			}
			if got := p.pausedUntil.Sub(p.now()); got != tc.paused && !(tc.paused == 0 && got < 0) {
				t.Fatalf("paused for %s, want %s", got, tc.paused)
			}
		})
	}
}

func TestCountdownPublisherThumbnailFailureKeepsPublishing(t *testing.T) {
	store, uploader, p := publishFixture(t, 0)
	uploader.thumbErr = &YouTubeError{Status: 403, Reason: "forbidden", Message: "custom thumbnails need a verified channel"}
	if err := p.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionPublished {
		t.Fatalf("status %q", store.status)
	}
	if got := testutil.ToFloat64(p.Metrics.Uploads.WithLabelValues("thumbnail", "failed")); got != 1 {
		t.Fatalf("thumbnail failures = %v", got)
	}
}

func TestCountdownPublisherScheduledPrivacy(t *testing.T) {
	store, uploader, p := publishFixture(t, 1)
	p.Privacy, p.PublishHourUTC = YouTubePrivacyScheduled, 15
	if err := p.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 11, 7, 15, 0, 0, 0, time.UTC)
	for _, v := range uploader.uploads {
		if v.PrivacyStatus != "private" || v.PublishAt == nil || !v.PublishAt.Equal(want) {
			t.Fatalf("privacy %q publishAt %v", v.PrivacyStatus, v.PublishAt)
		}
	}
	for _, u := range store.uploads {
		if u.PublishAt == nil || !u.PublishAt.Equal(want) {
			t.Fatalf("recorded publish time %v", u.PublishAt)
		}
	}
	// Past the publish hour, YouTube needs a time in the future.
	late := time.Date(2026, 11, 7, 16, 0, 0, 0, time.UTC)
	p.Now = func() time.Time { return late }
	if _, at := p.privacy(store.edition); at == nil || !at.Equal(late.Add(15*time.Minute)) {
		t.Fatalf("late publish time %v", at)
	}
}

func TestCountdownPublisherFailsWhenRendersAreGone(t *testing.T) {
	store, _, p := publishFixture(t, 0)
	store.renders = nil
	_ = p.RunOnce(context.Background())
	if store.status != models.CountdownEditionFailed || !strings.Contains(store.reason, "no longer on the render host") {
		t.Fatalf("status %q reason %q", store.status, store.reason)
	}
}

func TestNextQuotaReset(t *testing.T) {
	if got := nextQuotaReset(time.Date(2026, 11, 7, 7, 59, 0, 0, time.UTC)); !got.Equal(time.Date(2026, 11, 7, 8, 0, 0, 0, time.UTC)) {
		t.Fatal(got)
	}
	if got := nextQuotaReset(time.Date(2026, 11, 7, 8, 0, 0, 0, time.UTC)); !got.Equal(time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)) {
		t.Fatal(got)
	}
}
