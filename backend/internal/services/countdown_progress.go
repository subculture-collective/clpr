package services

import (
	"context"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

// CountdownProgressSource lists the progress of every enabled series.
type CountdownProgressSource interface {
	ListSeriesProgress(ctx context.Context) ([]models.CountdownSeriesProgress, error)
}

// CountdownProgress exports how overdue each series' latest edition is, so
// a missed daily edition raises an alert (#80).
type CountdownProgress struct {
	Source  CountdownProgressSource
	Overdue *prometheus.GaugeVec
	// Published adds the "published" stage, when YouTube publishing is on.
	Published bool
}

// NewCountdownProgress creates the gauge and registers it with reg when it
// is not nil.
func NewCountdownProgress(source CountdownProgressSource, published bool, reg prometheus.Registerer) *CountdownProgress {
	p := &CountdownProgress{Source: source, Published: published, Overdue: prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "clpr_countdown_edition_overdue_seconds",
		Help: "Seconds since the due edition of an enabled series should have been selected, while it is not yet rendered (stage=rendered) or published (stage=published); 0 when it is.",
	}, []string{"series", "stage"})}
	if reg != nil {
		reg.MustRegister(p.Overdue)
	}
	return p
}

// overdue returns how long the edition due at now has been missing at
// stage, given the newest date that reached it.
func countdownOverdue(series models.CountdownSeriesProgress, latest *string, now time.Time) float64 {
	now = now.UTC()
	due := time.Date(now.Year(), now.Month(), now.Day(), series.SelectionHourUTC, 0, 0, 0, time.UTC)
	if now.Before(due) {
		due = due.AddDate(0, 0, -1)
	}
	if series.CreatedAt.After(due) {
		return 0 // the series did not exist when this edition was due
	}
	if latest != nil && *latest >= due.Format("2006-01-02") {
		return 0
	}
	return now.Sub(due).Seconds()
}

// Update sets the gauge for every enabled series.
func (p *CountdownProgress) Update(ctx context.Context, now time.Time) error {
	series, err := p.Source.ListSeriesProgress(ctx)
	if err != nil {
		return err
	}
	p.Overdue.Reset()
	for _, s := range series {
		p.Overdue.WithLabelValues(s.Slug, "rendered").Set(countdownOverdue(s, s.LatestRendered, now))
		if p.Published {
			p.Overdue.WithLabelValues(s.Slug, "published").Set(countdownOverdue(s, s.LatestPublished, now))
		}
	}
	return nil
}
