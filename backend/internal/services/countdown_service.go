package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
)

// ErrCountdownValidation wraps every rejected series or exclusion input.
var ErrCountdownValidation = errors.New("invalid countdown input")

var countdownSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

const countdownDateLayout = "2006-01-02"

// CountdownStore is the storage the countdown service needs.
type CountdownStore interface {
	ListSeries(ctx context.Context) ([]models.CountdownSeries, error)
	GetSeries(ctx context.Context, id uuid.UUID) (*models.CountdownSeries, error)
	CreateSeries(ctx context.Context, in *models.CountdownSeriesInput) (*models.CountdownSeries, error)
	UpdateSeries(ctx context.Context, id uuid.UUID, in *models.CountdownSeriesInput) (*models.CountdownSeries, error)
	EditionExists(ctx context.Context, seriesID uuid.UUID, date string) (bool, error)
	SelectEdition(ctx context.Context, series *models.CountdownSeries, date string, windowStart, windowEnd time.Time, replace bool) (*models.CountdownEdition, error)
	ListEditions(ctx context.Context, date *string, seriesID *uuid.UUID, limit int) ([]models.CountdownEdition, error)
	ListExclusions(ctx context.Context) ([]models.CountdownExclusion, error)
	CreateExclusion(ctx context.Context, in *models.CountdownExclusionInput, source string, createdBy *uuid.UUID) (*models.CountdownExclusion, error)
	DeleteExclusion(ctx context.Context, id uuid.UUID) error
}

// CountdownService selects the daily countdown editions (#76).
type CountdownService struct {
	store CountdownStore
	now   func() time.Time
}

// NewCountdownService creates a CountdownService.
func NewCountdownService(store CountdownStore) *CountdownService {
	return &CountdownService{store: store, now: time.Now}
}

func countdownInvalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrCountdownValidation, fmt.Sprintf(format, args...))
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// normalizeCountdownSeries fills defaults and checks the input against the
// limits the database enforces, so admins get a 400 instead of a 500.
func normalizeCountdownSeries(in *models.CountdownSeriesInput) error {
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	in.GameID = trimOptional(in.GameID)
	in.Language = trimOptional(in.Language)
	if in.Language != nil {
		lower := strings.ToLower(*in.Language)
		in.Language = &lower
	}
	if in.Size == 0 {
		in.Size = 10
	}
	if in.WindowHours == 0 {
		in.WindowHours = 24
	}
	if in.PerBroadcasterCap == 0 {
		in.PerBroadcasterCap = 1
	}
	switch {
	case len(in.Slug) > 64 || !countdownSlugPattern.MatchString(in.Slug):
		return countdownInvalid("slug must be lowercase letters, digits and single hyphens, at most 64 characters")
	case in.Name == "" || len([]rune(in.Name)) > 120:
		return countdownInvalid("name must be 1 to 120 characters")
	case in.GameID != nil && len(*in.GameID) > 50:
		return countdownInvalid("game_id must be at most 50 characters")
	case in.Language != nil && (len(*in.Language) < 2 || len(*in.Language) > 10):
		return countdownInvalid("language must be 2 to 10 characters")
	case in.Size < 1 || in.Size > 50:
		return countdownInvalid("size must be between 1 and 50")
	case in.WindowHours < 1 || in.WindowHours > 168:
		return countdownInvalid("window_hours must be between 1 and 168")
	case in.PerBroadcasterCap < 1 || in.PerBroadcasterCap > 10:
		return countdownInvalid("per_broadcaster_cap must be between 1 and 10")
	case in.SelectionHourUTC < 0 || in.SelectionHourUTC > 23:
		return countdownInvalid("selection_hour_utc must be between 0 and 23")
	case in.ShortsCount < 0 || in.ShortsCount > 10 || in.ShortsCount > in.Size:
		return countdownInvalid("shorts_count must be between 0 and 10 and at most size")
	case in.MinSize != nil && (*in.MinSize < 1 || *in.MinSize > in.Size):
		return countdownInvalid("min_size must be between 1 and size")
	case in.MinSize != nil && in.ShortsCount > *in.MinSize:
		return countdownInvalid("shorts_count must be at most min_size")
	}
	return nil
}

// countdownWindow returns the window of the series edition on date: it ends
// at the series selection hour (UTC) that day.
func countdownWindow(series *models.CountdownSeries, date time.Time) (start, end time.Time) {
	end = time.Date(date.Year(), date.Month(), date.Day(), series.SelectionHourUTC, 0, 0, 0, time.UTC)
	return end.Add(-time.Duration(series.WindowHours) * time.Hour), end
}

// dueCountdownDate returns the most recent edition date whose selection hour
// has passed at now. A missed day is not backfilled beyond that one date.
func dueCountdownDate(series *models.CountdownSeries, now time.Time) time.Time {
	now = now.UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if _, end := countdownWindow(series, today); now.Before(end) {
		return today.AddDate(0, 0, -1)
	}
	return today
}

// ListSeries returns every series in selection order.
func (s *CountdownService) ListSeries(ctx context.Context) ([]models.CountdownSeries, error) {
	return s.store.ListSeries(ctx)
}

// CreateSeries validates and stores a series.
func (s *CountdownService) CreateSeries(ctx context.Context, in *models.CountdownSeriesInput) (*models.CountdownSeries, error) {
	if err := normalizeCountdownSeries(in); err != nil {
		return nil, err
	}
	return s.store.CreateSeries(ctx, in)
}

// UpdateSeries validates and replaces a series configuration.
func (s *CountdownService) UpdateSeries(ctx context.Context, id uuid.UUID, in *models.CountdownSeriesInput) (*models.CountdownSeries, error) {
	if err := normalizeCountdownSeries(in); err != nil {
		return nil, err
	}
	return s.store.UpdateSeries(ctx, id, in)
}

// ListEditions returns editions, newest first.
func (s *CountdownService) ListEditions(ctx context.Context, date *string, seriesID *uuid.UUID, limit int) ([]models.CountdownEdition, error) {
	if date != nil {
		if _, err := time.Parse(countdownDateLayout, *date); err != nil {
			return nil, countdownInvalid("date must be YYYY-MM-DD")
		}
	}
	return s.store.ListEditions(ctx, date, seriesID, limit)
}

// ReselectEdition selects the edition of a series for date again, replacing
// one that has not moved past selection. The date's selection hour must have
// passed.
func (s *CountdownService) ReselectEdition(ctx context.Context, seriesID uuid.UUID, date string) (*models.CountdownEdition, error) {
	day, err := time.Parse(countdownDateLayout, date)
	if err != nil {
		return nil, countdownInvalid("date must be YYYY-MM-DD")
	}
	series, err := s.store.GetSeries(ctx, seriesID)
	if err != nil {
		return nil, err
	}
	start, end := countdownWindow(series, day)
	if end.After(s.now()) {
		return nil, countdownInvalid("the selection hour of %s has not passed yet", date)
	}
	return s.store.SelectEdition(ctx, series, date, start, end, true)
}

// GenerateDueEditions selects today's edition for every enabled series whose
// selection hour has passed, in priority order. It returns how many editions
// were selected and how many were stored as failed.
func (s *CountdownService) GenerateDueEditions(ctx context.Context) (selected, failed int, err error) {
	series, err := s.store.ListSeries(ctx)
	if err != nil {
		return 0, 0, err
	}
	now := s.now()
	var errs []error
	for i := range series {
		sr := &series[i]
		if !sr.Enabled {
			continue
		}
		day := dueCountdownDate(sr, now)
		date := day.Format(countdownDateLayout)
		exists, err := s.store.EditionExists(ctx, sr.ID, date)
		if err != nil {
			errs = append(errs, fmt.Errorf("series %s: %w", sr.Slug, err))
			continue
		}
		if exists {
			continue
		}
		start, end := countdownWindow(sr, day)
		edition, err := s.store.SelectEdition(ctx, sr, date, start, end, false)
		if errors.Is(err, repository.ErrCountdownEditionExists) {
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("series %s: %w", sr.Slug, err))
			continue
		}
		fields := map[string]interface{}{"series": sr.Slug, "edition_date": date, "status": edition.Status, "clips": len(edition.Clips)}
		if edition.Status == models.CountdownEditionSkipped {
			if edition.FailureReason != nil {
				fields["reason"] = *edition.FailureReason
			}
			utils.Info("Countdown edition skipped: too few clips", fields)
			continue
		}
		if edition.Status == models.CountdownEditionFailed {
			failed++
			if edition.FailureReason != nil {
				fields["reason"] = *edition.FailureReason
			}
			utils.Warn("Countdown edition could not be filled", fields)
			continue
		}
		selected++
		utils.Info("Countdown edition selected", fields)
	}
	return selected, failed, errors.Join(errs...)
}

// ListExclusions returns every exclusion.
func (s *CountdownService) ListExclusions(ctx context.Context) ([]models.CountdownExclusion, error) {
	return s.store.ListExclusions(ctx)
}

// CreateExclusion validates and stores an admin exclusion.
func (s *CountdownService) CreateExclusion(ctx context.Context, in *models.CountdownExclusionInput, adminID uuid.UUID) (*models.CountdownExclusion, error) {
	in.BroadcasterID = trimOptional(in.BroadcasterID)
	in.Note = trimOptional(in.Note)
	in.Reason = strings.TrimSpace(in.Reason)
	switch {
	case (in.BroadcasterID == nil) == (in.ClipID == nil):
		return nil, countdownInvalid("set exactly one of broadcaster_id and clip_id")
	case in.BroadcasterID != nil && len(*in.BroadcasterID) > 50:
		return nil, countdownInvalid("broadcaster_id must be at most 50 characters")
	case in.ClipID != nil && *in.ClipID == uuid.Nil:
		return nil, countdownInvalid("clip_id is not a clip")
	case in.Reason != models.CountdownExclusionRequest && in.Reason != models.CountdownExclusionDMCA &&
		in.Reason != models.CountdownExclusionTakedown:
		return nil, countdownInvalid("reason must be request, dmca or takedown")
	case in.Note != nil && len([]rune(*in.Note)) > 500:
		return nil, countdownInvalid("note must be at most 500 characters")
	}
	return s.store.CreateExclusion(ctx, in, models.CountdownExclusionSourceAdmin, &adminID)
}

// DeleteExclusion removes an exclusion.
func (s *CountdownService) DeleteExclusion(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteExclusion(ctx, id)
}
