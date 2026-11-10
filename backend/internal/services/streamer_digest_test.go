package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

func sampleDigest() *models.StreamerDigest {
	return &models.StreamerDigest{Name: "Moon <Moon>", BroadcasterID: "1234", WeekStart: "2026-11-02",
		TopClips: []models.StreamerDigestClip{{ClipID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			Title: `<script>alert("x")</script> ace`, ViewCount: 1234567}},
		Placements: []models.CountdownFeature{{EditionDate: "2026-11-04", SeriesSlug: "just-chatting", SeriesName: "Just Chatting",
			Position: 3, Title: "chat moment"}},
		PackStatus: models.ClipPackReady, PackClipCount: 7}
}

func TestRenderStreamerDigest(t *testing.T) {
	subject, html, text, err := renderStreamerDigest("https://clpr.tv/", sampleDigest(), "https://clpr.tv/api/v1/notifications/unsubscribe?token=abc")
	if err != nil {
		t.Fatal(err)
	}
	if subject != "Your week on clpr: Week of November 2, 2026" {
		t.Fatalf("subject = %q", subject)
	}
	for _, want := range []string{
		"Your week on clpr, Moon &lt;Moon&gt;",
		"&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt; ace",
		"1,234,567 views",
		`href="https://clpr.tv/clip/11111111-1111-4111-8111-111111111111"`,
		`href="https://clpr.tv/countdowns/2026-11-04/just-chatting"`,
		"#3 in Just Chatting, Nov 4",
		"Your 7 top clips of the week are ready to post",
		`href="https://clpr.tv/creator/1234/dashboard"`,
		`href="https://clpr.tv/api/v1/notifications/unsubscribe?token=abc"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML lacks %q", want)
		}
	}
	if strings.Contains(html, "<script>") {
		t.Fatal("the HTML carries an unescaped script tag")
	}
	for _, want := range []string{"Moon <Moon>", "1,234,567 views): https://clpr.tv/clip/", "Stop the weekly digest: https://clpr.tv/api/v1/notifications/unsubscribe?token=abc"} {
		if !strings.Contains(text, want) {
			t.Errorf("text lacks %q:\n%s", want, text)
		}
	}

	d := sampleDigest()
	d.PackStatus = models.ClipPackBuilding
	if _, html, _, _ := renderStreamerDigest("https://clpr.tv", d, ""); !strings.Contains(html, "is being made") || strings.Contains(html, "Stop the weekly digest") {
		t.Fatal("pending pack or missing unsubscribe link rendered wrong")
	}
}

type fakeDigestStore struct {
	recipients []models.StreamerDigestRecipient
	claimed    map[uuid.UUID]bool
	content    map[string]*models.StreamerDigest
	finished   map[uuid.UUID]string
	weeks      []time.Time
}

func (f *fakeDigestStore) ListDigestRecipients(_ context.Context, week time.Time) ([]models.StreamerDigestRecipient, error) {
	f.weeks = append(f.weeks, week)
	return f.recipients, nil
}
func (f *fakeDigestStore) ClaimDigest(_ context.Context, id uuid.UUID, _ time.Time) (bool, error) {
	if f.claimed[id] {
		return false, nil
	}
	f.claimed[id] = true
	return true, nil
}
func (f *fakeDigestStore) FinishDigest(_ context.Context, id uuid.UUID, _ time.Time, status string, _ *string) error {
	f.finished[id] = status
	return nil
}
func (f *fakeDigestStore) StreamerDigestContent(_ context.Context, broadcasterID string, _ time.Time) (*models.StreamerDigest, error) {
	if d, ok := f.content[broadcasterID]; ok {
		return d, nil
	}
	return &models.StreamerDigest{BroadcasterID: broadcasterID}, nil
}

type fakeDigestSender struct {
	sent []string
	fail string
}

func (f *fakeDigestSender) SendStreamerDigest(_ context.Context, rcpt models.StreamerDigestRecipient, d *models.StreamerDigest) error {
	if rcpt.BroadcasterID == f.fail {
		return errors.New("sendgrid said no")
	}
	if d.Name != rcpt.Name {
		return errors.New("digest has the wrong name")
	}
	f.sent = append(f.sent, rcpt.BroadcasterID)
	return nil
}

func TestStreamerDigestServiceSendsOncePerWeek(t *testing.T) {
	ok, empty, failing, taken := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store := &fakeDigestStore{
		recipients: []models.StreamerDigestRecipient{
			{UserID: ok, BroadcasterID: "ok", Name: "OK"}, {UserID: empty, BroadcasterID: "empty"},
			{UserID: failing, BroadcasterID: "failing"}, {UserID: taken, BroadcasterID: "taken"},
		},
		claimed:  map[uuid.UUID]bool{taken: true},
		content:  map[string]*models.StreamerDigest{"ok": sampleDigest(), "failing": sampleDigest()},
		finished: map[uuid.UUID]string{},
	}
	sender := &fakeDigestSender{fail: "failing"}
	svc := &StreamerDigestService{Store: store, Sender: sender, Now: func() time.Time { return time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC) }}
	sent, err := svc.RunOnce(context.Background())
	if err != nil || sent != 1 || len(sender.sent) != 1 || sender.sent[0] != "ok" {
		t.Fatalf("sent %d (%v), %v", sent, sender.sent, err)
	}
	if store.finished[ok] != models.StreamerDigestSent || store.finished[empty] != models.StreamerDigestSkipped ||
		store.finished[failing] != models.StreamerDigestFailed || store.finished[taken] != "" {
		t.Fatalf("finished = %v", store.finished)
	}
	if store.weeks[0] != time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("week = %v", store.weeks[0])
	}
}

func TestStreamerDigestWeek(t *testing.T) {
	svc := &StreamerDigestService{}
	for _, tc := range []struct{ now, want string }{
		{"2026-11-09T08:59:00Z", "2026-10-26"},
		{"2026-11-09T09:00:00Z", "2026-11-02"},
		{"2026-11-15T23:59:00Z", "2026-11-02"},
	} {
		now, _ := time.Parse(time.RFC3339, tc.now)
		if got := svc.DigestWeek(now).Format("2006-01-02"); got != tc.want {
			t.Errorf("DigestWeek(%s) = %s, want %s", tc.now, got, tc.want)
		}
	}
}

func TestSendStreamerDigestNeedsEmail(t *testing.T) {
	s := NewEmailService(&EmailConfig{Enabled: false}, nil, nil)
	if err := s.SendStreamerDigest(context.Background(), models.StreamerDigestRecipient{}, sampleDigest()); !errors.Is(err, ErrEmailDisabled) {
		t.Fatalf("disabled email = %v", err)
	}
}

func TestStreamerDigestEmptyIgnoresFailedPacks(t *testing.T) {
	for status, empty := range map[string]bool{"": true, models.ClipPackFailed: true, models.ClipPackExpired: true,
		models.ClipPackReady: false, models.ClipPackPending: false} {
		if got := (&models.StreamerDigest{PackStatus: status}).Empty(); got != empty {
			t.Errorf("pack %q: Empty() = %v, want %v", status, got, empty)
		}
	}
}
