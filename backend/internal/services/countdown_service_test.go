package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
)

type fakeCountdownStore struct {
	CountdownStore
	series   []models.CountdownSeries
	existing map[string]bool
	selected []string
	failFor  map[string]bool
	created  *models.CountdownExclusionInput
}

func (f *fakeCountdownStore) ListSeries(context.Context) ([]models.CountdownSeries, error) {
	return f.series, nil
}

func (f *fakeCountdownStore) GetSeries(_ context.Context, id uuid.UUID) (*models.CountdownSeries, error) {
	for i := range f.series {
		if f.series[i].ID == id {
			return &f.series[i], nil
		}
	}
	return nil, repository.ErrCountdownSeriesNotFound
}

func (f *fakeCountdownStore) EditionExists(_ context.Context, seriesID uuid.UUID, date string) (bool, error) {
	return f.existing[seriesID.String()+date], nil
}

func (f *fakeCountdownStore) SelectEdition(_ context.Context, s *models.CountdownSeries, date string, start, end time.Time, _ bool) (*models.CountdownEdition, error) {
	f.selected = append(f.selected, s.Slug+" "+date+" "+start.Format(time.RFC3339)+" "+end.Format(time.RFC3339))
	if f.failFor[s.Slug] {
		reason := "only 3 of 10 clips were eligible"
		return &models.CountdownEdition{Status: models.CountdownEditionFailed, FailureReason: &reason}, nil
	}
	return &models.CountdownEdition{Status: models.CountdownEditionSelected, Clips: make([]models.CountdownEditionClip, s.Size)}, nil
}

func (f *fakeCountdownStore) CreateExclusion(_ context.Context, in *models.CountdownExclusionInput, source string, _ *uuid.UUID) (*models.CountdownExclusion, error) {
	f.created = in
	return &models.CountdownExclusion{Reason: in.Reason, Source: source}, nil
}

func countdownSeriesFixture(slug string, hour int, enabled bool) models.CountdownSeries {
	return models.CountdownSeries{ID: uuid.New(), Slug: slug, Size: 10, WindowHours: 24, PerBroadcasterCap: 1,
		SelectionHourUTC: hour, Enabled: enabled}
}

func TestDueCountdownDate(t *testing.T) {
	series := countdownSeriesFixture("overall", 12, true)
	cases := []struct {
		now  string
		want string
	}{
		{"2026-11-07T11:59:59Z", "2026-11-06"},
		{"2026-11-07T12:00:00Z", "2026-11-07"},
		{"2026-11-07T23:30:00Z", "2026-11-07"},
		// 06:00 on Nov 8 in UTC+9 is 21:00 UTC on Nov 7, after that day's noon.
		{"2026-11-08T06:00:00+09:00", "2026-11-07"},
	}
	for _, tc := range cases {
		now, err := time.Parse(time.RFC3339, tc.now)
		if err != nil {
			t.Fatal(err)
		}
		if got := dueCountdownDate(&series, now).Format(countdownDateLayout); got != tc.want {
			t.Errorf("dueCountdownDate(%s) = %s, want %s", tc.now, got, tc.want)
		}
	}
}

func TestCountdownWindowEndsAtSelectionHour(t *testing.T) {
	series := countdownSeriesFixture("overall", 15, true)
	series.WindowHours = 48
	start, end := countdownWindow(&series, time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC))
	if want := time.Date(2026, 11, 7, 15, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("end = %s, want %s", end, want)
	}
	if want := time.Date(2026, 11, 5, 15, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start = %s, want %s", start, want)
	}
}

func TestGenerateDueEditionsSkipsDisabledAndExisting(t *testing.T) {
	overall := countdownSeriesFixture("overall", 12, true)
	later := countdownSeriesFixture("late", 20, true)
	disabled := countdownSeriesFixture("off", 0, false)
	done := countdownSeriesFixture("done", 0, true)
	short := countdownSeriesFixture("short", 6, true)
	store := &fakeCountdownStore{
		series:   []models.CountdownSeries{overall, later, disabled, done, short},
		existing: map[string]bool{done.ID.String() + "2026-11-07": true},
		failFor:  map[string]bool{"short": true},
	}
	svc := NewCountdownService(store)
	svc.now = func() time.Time { return time.Date(2026, 11, 7, 13, 0, 0, 0, time.UTC) }

	selected, failed, err := svc.GenerateDueEditions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if selected != 2 || failed != 1 {
		t.Fatalf("selected=%d failed=%d, want 2 and 1", selected, failed)
	}
	want := []string{
		"overall 2026-11-07 2026-11-06T12:00:00Z 2026-11-07T12:00:00Z",
		// 13:00 is before the 20:00 selection hour, so the late series is due for the day before.
		"late 2026-11-06 2026-11-05T20:00:00Z 2026-11-06T20:00:00Z",
		"short 2026-11-07 2026-11-06T06:00:00Z 2026-11-07T06:00:00Z",
	}
	if len(store.selected) != len(want) {
		t.Fatalf("selected %v, want %v", store.selected, want)
	}
	for i := range want {
		if store.selected[i] != want[i] {
			t.Errorf("selection %d = %q, want %q", i, store.selected[i], want[i])
		}
	}
}

func TestReselectEditionRejectsFutureSelectionHour(t *testing.T) {
	series := countdownSeriesFixture("overall", 12, true)
	store := &fakeCountdownStore{series: []models.CountdownSeries{series}}
	svc := NewCountdownService(store)
	svc.now = func() time.Time { return time.Date(2026, 11, 7, 11, 0, 0, 0, time.UTC) }

	if _, err := svc.ReselectEdition(context.Background(), series.ID, "2026-11-07"); !errors.Is(err, ErrCountdownValidation) {
		t.Fatalf("future date err = %v, want validation error", err)
	}
	if _, err := svc.ReselectEdition(context.Background(), series.ID, "11/07/2026"); !errors.Is(err, ErrCountdownValidation) {
		t.Fatalf("bad date err = %v, want validation error", err)
	}
	if _, err := svc.ReselectEdition(context.Background(), series.ID, "2026-11-06"); err != nil {
		t.Fatalf("past date err = %v", err)
	}
	if _, err := svc.ReselectEdition(context.Background(), uuid.New(), "2026-11-06"); !errors.Is(err, repository.ErrCountdownSeriesNotFound) {
		t.Fatalf("unknown series err = %v", err)
	}
}

func intPtr(v int) *int { return &v }

func TestNormalizeCountdownSeries(t *testing.T) {
	lang := " EN "
	empty := "  "
	in := models.CountdownSeriesInput{Slug: " just-chatting ", Name: " Just Chatting ", Language: &lang, GameID: &empty}
	if err := normalizeCountdownSeries(&in); err != nil {
		t.Fatal(err)
	}
	if in.Slug != "just-chatting" || in.Name != "Just Chatting" || *in.Language != "en" || in.GameID != nil {
		t.Fatalf("normalized = %+v", in)
	}
	if in.Size != 10 || in.WindowHours != 24 || in.PerBroadcasterCap != 1 || in.Source != models.CountdownSourceVelocity {
		t.Fatalf("defaults = %+v", in)
	}
	room := uuid.New()
	roomSeries := models.CountdownSeriesInput{Slug: "room", Name: "Room", Source: " clip_room ", ClipRoomID: &room}
	if err := normalizeCountdownSeries(&roomSeries); err != nil || roomSeries.Source != models.CountdownSourceClipRoom {
		t.Fatalf("clip room series = %+v, %v", roomSeries, err)
	}

	bad := []models.CountdownSeriesInput{
		{Slug: "Has Space", Name: "x"},
		{Slug: "double--hyphen", Name: "x"},
		{Slug: "ok", Name: ""},
		{Slug: "ok", Name: "x", Size: 51},
		{Slug: "ok", Name: "x", WindowHours: 169},
		{Slug: "ok", Name: "x", PerBroadcasterCap: 11},
		{Slug: "ok", Name: "x", SelectionHourUTC: 24},
		{Slug: "ok", Name: "x", SelectionHourUTC: -1},
		{Slug: "ok", Name: "x", Size: 5, MinSize: intPtr(0)},
		{Slug: "ok", Name: "x", Size: 5, MinSize: intPtr(6)},
		{Slug: "ok", Name: "x", Size: 5, MinSize: intPtr(2), ShortsCount: 3},
		{Slug: "ok", Name: "x", Source: "trending"},
		{Slug: "ok", Name: "x", Source: models.CountdownSourceClipRoom},
		{Slug: "ok", Name: "x", ClipRoomID: &room},
	}
	for _, b := range bad {
		b := b
		if err := normalizeCountdownSeries(&b); !errors.Is(err, ErrCountdownValidation) {
			t.Errorf("normalize(%+v) err = %v, want validation error", b, err)
		}
	}
}

func TestCreateExclusionValidation(t *testing.T) {
	store := &fakeCountdownStore{}
	svc := NewCountdownService(store)
	admin := uuid.New()
	broadcaster := " 12345 "
	clipID := uuid.New()

	bad := []models.CountdownExclusionInput{
		{Reason: "request"},
		{BroadcasterID: &broadcaster, ClipID: &clipID, Reason: "request"},
		{BroadcasterID: &broadcaster, Reason: "because"},
		{ClipID: &uuid.Nil, Reason: "dmca"},
	}
	for _, b := range bad {
		b := b
		if _, err := svc.CreateExclusion(context.Background(), &b, admin); !errors.Is(err, ErrCountdownValidation) {
			t.Errorf("CreateExclusion(%+v) err = %v, want validation error", b, err)
		}
	}

	x, err := svc.CreateExclusion(context.Background(), &models.CountdownExclusionInput{BroadcasterID: &broadcaster, Reason: "request"}, admin)
	if err != nil {
		t.Fatal(err)
	}
	if x.Source != models.CountdownExclusionSourceAdmin || *store.created.BroadcasterID != "12345" {
		t.Fatalf("exclusion = %+v, stored broadcaster %q", x, *store.created.BroadcasterID)
	}
}
