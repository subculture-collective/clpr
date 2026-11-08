package services

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCountdownOverdue(t *testing.T) {
	created := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	series := models.CountdownSeriesProgress{Slug: "overall", SelectionHourUTC: 12, CreatedAt: created}
	day := func(s string) *string { return &s }
	at := func(h, m int) time.Time { return time.Date(2026, 11, 7, h, m, 0, 0, time.UTC) }
	cases := []struct {
		name   string
		latest *string
		now    time.Time
		want   float64
	}{
		{"today's edition rendered", day("2026-11-07"), at(14, 0), 0},
		{"today's edition not rendered 2 h after selection", day("2026-11-06"), at(14, 0), 7200},
		{"before today's selection, yesterday's is due", day("2026-11-06"), at(9, 0), 0},
		{"yesterday's missing too", day("2026-11-05"), at(9, 0), 21 * 3600},
		{"never rendered", nil, at(12, 30), 1800},
	}
	for _, tc := range cases {
		if got := countdownOverdue(series, tc.latest, tc.now); got != tc.want {
			t.Errorf("%s: overdue %v, want %v", tc.name, got, tc.want)
		}
	}
	fresh := series
	fresh.CreatedAt = at(13, 0)
	if got := countdownOverdue(fresh, nil, at(14, 0)); got != 0 {
		t.Errorf("a series created after today's selection is overdue by %v", got)
	}
}

type fakeProgressSource []models.CountdownSeriesProgress

func (f fakeProgressSource) ListSeriesProgress(context.Context) ([]models.CountdownSeriesProgress, error) {
	return f, nil
}

func TestCountdownProgressUpdate(t *testing.T) {
	rendered := "2026-11-07"
	source := fakeProgressSource{{Slug: "overall", SelectionHourUTC: 12, CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), LatestRendered: &rendered}}
	p := NewCountdownProgress(source, true, nil)
	if err := p.Update(context.Background(), time.Date(2026, 11, 7, 13, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if got := testutil.ToFloat64(p.Overdue.WithLabelValues("overall", "rendered")); got != 0 {
		t.Fatalf("rendered overdue = %v", got)
	}
	if got := testutil.ToFloat64(p.Overdue.WithLabelValues("overall", "published")); got != 3600 {
		t.Fatalf("published overdue = %v", got)
	}
	// A series that is disabled disappears from the gauge.
	p.Source = fakeProgressSource{}
	_ = p.Update(context.Background(), time.Now())
	if n := testutil.CollectAndCount(p.Overdue); n != 0 {
		t.Fatalf("%d series still exported", n)
	}
}
