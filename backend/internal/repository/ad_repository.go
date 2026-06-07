package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// AdRepository handles database operations for ads
type AdRepository struct {
	pool *pgxpool.Pool
}

// NewAdRepository creates a new AdRepository
func NewAdRepository(pool *pgxpool.Pool) *AdRepository {
	return &AdRepository{
		pool: pool,
	}
}

// GetActiveAds retrieves all active ads that are within their date range and budget
func (r *AdRepository) GetActiveAds(ctx context.Context, adType *string, width, height *int) ([]models.Ad, error) {
	whereClauses := []string{
		"is_active = true",
		"(start_date IS NULL OR start_date <= NOW())",
		"(end_date IS NULL OR end_date > NOW())",
		"(daily_budget_cents IS NULL OR spent_today_cents < daily_budget_cents)",
		"(total_budget_cents IS NULL OR spent_total_cents < total_budget_cents)",
	}
	args := []interface{}{}
	argIndex := 1

	if adType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("ad_type = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *adType)
		argIndex++
	}

	if width != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(width IS NULL OR width = %s)", utils.SQLPlaceholder(argIndex)))
		args = append(args, *width)
		argIndex++
	}

	if height != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(height IS NULL OR height = %s)", utils.SQLPlaceholder(argIndex)))
		args = append(args, *height)
		argIndex++
	}

	whereClause := whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}

	query := fmt.Sprintf(`
		SELECT id, name, advertiser_name, ad_type, content_url, click_url, alt_text,
			width, height, priority, weight, daily_budget_cents, total_budget_cents,
			spent_today_cents, spent_total_cents, cpm_cents, is_active, start_date,
			end_date, targeting_criteria, created_at, updated_at
		FROM ads
		WHERE %s
		ORDER BY priority DESC, weight DESC
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get active ads: %w", err)
	}
	defer rows.Close()

	var ads []models.Ad
	for rows.Next() {
		var ad models.Ad
		var targetingJSON []byte
		err := rows.Scan(
			&ad.ID, &ad.Name, &ad.AdvertiserName, &ad.AdType, &ad.ContentURL, &ad.ClickURL,
			&ad.AltText, &ad.Width, &ad.Height, &ad.Priority, &ad.Weight, &ad.DailyBudgetCents,
			&ad.TotalBudgetCents, &ad.SpentTodayCents, &ad.SpentTotalCents, &ad.CPMCents,
			&ad.IsActive, &ad.StartDate, &ad.EndDate, &targetingJSON, &ad.CreatedAt, &ad.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ad: %w", err)
		}

		if targetingJSON != nil {
			if err := json.Unmarshal(targetingJSON, &ad.TargetingCriteria); err != nil {
				// Log warning but continue - targeting will be empty
				ad.TargetingCriteria = nil
			}
		}
		ads = append(ads, ad)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ads: %w", err)
	}

	return ads, nil
}

// GetAdByID retrieves an ad by ID
func (r *AdRepository) GetAdByID(ctx context.Context, id uuid.UUID) (*models.Ad, error) {
	query := `
		SELECT id, name, advertiser_name, ad_type, content_url, click_url, alt_text,
			width, height, priority, weight, daily_budget_cents, total_budget_cents,
			spent_today_cents, spent_total_cents, cpm_cents, is_active, start_date,
			end_date, targeting_criteria, created_at, updated_at
		FROM ads
		WHERE id = $1
	`

	var ad models.Ad
	var targetingJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ad.ID, &ad.Name, &ad.AdvertiserName, &ad.AdType, &ad.ContentURL, &ad.ClickURL,
		&ad.AltText, &ad.Width, &ad.Height, &ad.Priority, &ad.Weight, &ad.DailyBudgetCents,
		&ad.TotalBudgetCents, &ad.SpentTodayCents, &ad.SpentTotalCents, &ad.CPMCents,
		&ad.IsActive, &ad.StartDate, &ad.EndDate, &targetingJSON, &ad.CreatedAt, &ad.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ad by ID: %w", err)
	}

	if targetingJSON != nil {
		if err := json.Unmarshal(targetingJSON, &ad.TargetingCriteria); err != nil {
			// Log warning but continue - targeting will be empty
			ad.TargetingCriteria = nil
		}
	}

	return &ad, nil
}

// CreateImpression creates a new ad impression record
func (r *AdRepository) CreateImpression(ctx context.Context, impression *models.AdImpression) error {
	query := `
		INSERT INTO ad_impressions (id, ad_id, user_id, session_id, platform, ip_address,
			user_agent, page_url, viewability_time_ms, is_viewable, is_clicked, cost_cents)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		impression.ID, impression.AdID, impression.UserID, impression.SessionID,
		impression.Platform, impression.IPAddress, impression.UserAgent, impression.PageURL,
		impression.ViewabilityTimeMs, impression.IsViewable, impression.IsClicked, impression.CostCents,
	)
	if err != nil {
		return fmt.Errorf("failed to create impression: %w", err)
	}

	return nil
}

// UpdateImpression updates an existing impression with viewability/click data
func (r *AdRepository) UpdateImpression(ctx context.Context, impressionID uuid.UUID, viewabilityTimeMs int, isViewable, isClicked bool) error {
	query := `
		UPDATE ad_impressions
		SET viewability_time_ms = $2, is_viewable = $3, is_clicked = $4,
			clicked_at = CASE WHEN $4 = true AND clicked_at IS NULL THEN NOW() ELSE clicked_at END
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, impressionID, viewabilityTimeMs, isViewable, isClicked)
	if err != nil {
		return fmt.Errorf("failed to update impression: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("impression not found")
	}

	return nil
}

// GetImpressionByID retrieves an impression by ID
func (r *AdRepository) GetImpressionByID(ctx context.Context, id uuid.UUID) (*models.AdImpression, error) {
	query := `
		SELECT id, ad_id, user_id, session_id, platform, ip_address, user_agent, page_url,
			viewability_time_ms, is_viewable, is_clicked, clicked_at, cost_cents, created_at
		FROM ad_impressions
		WHERE id = $1
	`

	var imp models.AdImpression
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&imp.ID, &imp.AdID, &imp.UserID, &imp.SessionID, &imp.Platform, &imp.IPAddress,
		&imp.UserAgent, &imp.PageURL, &imp.ViewabilityTimeMs, &imp.IsViewable, &imp.IsClicked,
		&imp.ClickedAt, &imp.CostCents, &imp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get impression: %w", err)
	}

	return &imp, nil
}

// IncrementAdSpend increments the spend counters for an ad
func (r *AdRepository) IncrementAdSpend(ctx context.Context, adID uuid.UUID, costCents int) error {
	query := `
		UPDATE ads
		SET spent_today_cents = spent_today_cents + $2,
			spent_total_cents = spent_total_cents + $2
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, adID, costCents)
	if err != nil {
		return fmt.Errorf("failed to increment ad spend: %w", err)
	}

	return nil
}

// ResetDailySpend resets the daily spend for all ads (should be called at midnight)
func (r *AdRepository) ResetDailySpend(ctx context.Context) error {
	query := `UPDATE ads SET spent_today_cents = 0`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to reset daily spend: %w", err)
	}

	return nil
}

// GetFrequencyCap gets the current frequency cap for a user/session and ad
func (r *AdRepository) GetFrequencyCap(ctx context.Context, adID uuid.UUID, userID *uuid.UUID, sessionID *string, windowType string) (*models.AdFrequencyCap, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, ad_id, user_id, session_id, impression_count, window_start, window_type, created_at, updated_at
			FROM ad_frequency_caps
			WHERE ad_id = $1 AND user_id = $2 AND window_type = $3
		`
		args = []interface{}{adID, *userID, windowType}
	} else if sessionID != nil {
		query = `
			SELECT id, ad_id, user_id, session_id, impression_count, window_start, window_type, created_at, updated_at
			FROM ad_frequency_caps
			WHERE ad_id = $1 AND session_id = $2 AND window_type = $3
		`
		args = []interface{}{adID, *sessionID, windowType}
	} else {
		return nil, fmt.Errorf("either userID or sessionID must be provided")
	}

	var cap models.AdFrequencyCap
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&cap.ID, &cap.AdID, &cap.UserID, &cap.SessionID, &cap.ImpressionCount,
		&cap.WindowStart, &cap.WindowType, &cap.CreatedAt, &cap.UpdatedAt,
	)
	if err != nil {
		return nil, err // Let caller handle "no rows" error
	}

	return &cap, nil
}

// UpsertFrequencyCap creates or updates a frequency cap record
func (r *AdRepository) UpsertFrequencyCap(ctx context.Context, adID uuid.UUID, userID *uuid.UUID, sessionID *string, windowType string, windowStart time.Time) error {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			INSERT INTO ad_frequency_caps (id, ad_id, user_id, window_type, window_start, impression_count)
			VALUES ($1, $2, $3, $4, $5, 1)
			ON CONFLICT (ad_id, user_id, window_type)
			DO UPDATE SET 
				impression_count = CASE 
					WHEN ad_frequency_caps.window_start < $5 THEN 1 
					ELSE ad_frequency_caps.impression_count + 1 
				END,
				window_start = CASE 
					WHEN ad_frequency_caps.window_start < $5 THEN $5 
					ELSE ad_frequency_caps.window_start 
				END,
				updated_at = NOW()
		`
		args = []interface{}{uuid.New(), adID, *userID, windowType, windowStart}
	} else if sessionID != nil {
		query = `
			INSERT INTO ad_frequency_caps (id, ad_id, session_id, window_type, window_start, impression_count)
			VALUES ($1, $2, $3, $4, $5, 1)
			ON CONFLICT (ad_id, session_id, window_type)
			DO UPDATE SET 
				impression_count = CASE 
					WHEN ad_frequency_caps.window_start < $5 THEN 1 
					ELSE ad_frequency_caps.impression_count + 1 
				END,
				window_start = CASE 
					WHEN ad_frequency_caps.window_start < $5 THEN $5 
					ELSE ad_frequency_caps.window_start 
				END,
				updated_at = NOW()
		`
		args = []interface{}{uuid.New(), adID, *sessionID, windowType, windowStart}
	} else {
		return fmt.Errorf("either userID or sessionID must be provided")
	}

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to upsert frequency cap: %w", err)
	}

	return nil
}

// GetFrequencyLimits gets all frequency limits for an ad
func (r *AdRepository) GetFrequencyLimits(ctx context.Context, adID uuid.UUID) ([]models.AdFrequencyLimit, error) {
	query := `
		SELECT id, ad_id, window_type, max_impressions, created_at
		FROM ad_frequency_limits
		WHERE ad_id = $1
	`

	rows, err := r.pool.Query(ctx, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get frequency limits: %w", err)
	}
	defer rows.Close()

	var limits []models.AdFrequencyLimit
	for rows.Next() {
		var limit models.AdFrequencyLimit
		err := rows.Scan(&limit.ID, &limit.AdID, &limit.WindowType, &limit.MaxImpressions, &limit.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan frequency limit: %w", err)
		}
		limits = append(limits, limit)
	}

	return limits, nil
}

// GetUserImpressionCount gets the count of impressions for a user/session within a time window
func (r *AdRepository) GetUserImpressionCount(ctx context.Context, adID uuid.UUID, userID *uuid.UUID, sessionID *string, windowType string) (int, error) {
	windowStart := r.calculateWindowStart(windowType)

	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT COALESCE(impression_count, 0)
			FROM ad_frequency_caps
			WHERE ad_id = $1 AND user_id = $2 AND window_type = $3 AND window_start >= $4
		`
		args = []interface{}{adID, *userID, windowType, windowStart}
	} else if sessionID != nil {
		query = `
			SELECT COALESCE(impression_count, 0)
			FROM ad_frequency_caps
			WHERE ad_id = $1 AND session_id = $2 AND window_type = $3 AND window_start >= $4
		`
		args = []interface{}{adID, *sessionID, windowType, windowStart}
	} else {
		return 0, nil // No tracking for completely anonymous
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		// No rows is not an error - return 0
		return 0, nil
	}

	return count, nil
}

// calculateWindowStart returns the start time for a given window type
func (r *AdRepository) calculateWindowStart(windowType string) time.Time {
	now := time.Now().UTC()
	switch windowType {
	case models.FrequencyWindowHourly:
		return now.Truncate(time.Hour)
	case models.FrequencyWindowDaily:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case models.FrequencyWindowWeekly:
		// Start of the week (Sunday)
		daysSinceSunday := int(now.Weekday())
		return time.Date(now.Year(), now.Month(), now.Day()-daysSinceSunday, 0, 0, 0, 0, time.UTC)
	case models.FrequencyWindowLifetime:
		// Return epoch for lifetime
		return time.Time{}
	default:
		return now.Truncate(time.Hour)
	}
}

// CountRecentImpressions counts impressions from a specific IP in the last minute (fraud prevention)
func (r *AdRepository) CountRecentImpressions(ctx context.Context, adID uuid.UUID, ipAddress string, minutes int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM ad_impressions
		WHERE ad_id = $1 AND ip_address = $2 AND created_at > NOW() - INTERVAL '1 minute' * $3
	`

	var count int
	err := r.pool.QueryRow(ctx, query, adID, ipAddress, minutes).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent impressions: %w", err)
	}

	return count, nil
}

// CountViewableImpressions counts viewable impressions for an ad
func (r *AdRepository) CountViewableImpressions(ctx context.Context, adID uuid.UUID, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM ad_impressions
		WHERE ad_id = $1 AND is_viewable = true AND created_at >= $2
	`

	var count int
	err := r.pool.QueryRow(ctx, query, adID, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count viewable impressions: %w", err)
	}

	return count, nil
}

// GetTargetingRules retrieves all targeting rules for an ad
func (r *AdRepository) GetTargetingRules(ctx context.Context, adID uuid.UUID) ([]models.AdTargetingRule, error) {
	query := `
		SELECT id, ad_id, rule_type, operator, values, created_at
		FROM ad_targeting_rules
		WHERE ad_id = $1
		ORDER BY rule_type
	`

	rows, err := r.pool.Query(ctx, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get targeting rules: %w", err)
	}
	defer rows.Close()

	var rules []models.AdTargetingRule
	for rows.Next() {
		var rule models.AdTargetingRule
		err := rows.Scan(&rule.ID, &rule.AdID, &rule.RuleType, &rule.Operator, &rule.Values, &rule.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan targeting rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// CreateTargetingRule creates a new targeting rule for an ad
func (r *AdRepository) CreateTargetingRule(ctx context.Context, rule *models.AdTargetingRule) error {
	query := `
		INSERT INTO ad_targeting_rules (id, ad_id, rule_type, operator, values)
		VALUES ($1, $2, $3, $4, $5)
	`

	rule.ID = uuid.New()
	_, err := r.pool.Exec(ctx, query, rule.ID, rule.AdID, rule.RuleType, rule.Operator, rule.Values)
	if err != nil {
		return fmt.Errorf("failed to create targeting rule: %w", err)
	}

	return nil
}

// DeleteTargetingRules deletes all targeting rules for an ad
func (r *AdRepository) DeleteTargetingRules(ctx context.Context, adID uuid.UUID) error {
	query := `DELETE FROM ad_targeting_rules WHERE ad_id = $1`
	_, err := r.pool.Exec(ctx, query, adID)
	if err != nil {
		return fmt.Errorf("failed to delete targeting rules: %w", err)
	}
	return nil
}

// GetExperiment retrieves an experiment by ID
func (r *AdRepository) GetExperiment(ctx context.Context, experimentID uuid.UUID) (*models.AdExperiment, error) {
	query := `
		SELECT id, name, description, status, start_date, end_date, traffic_percent, winning_variant, created_at, updated_at
		FROM ad_experiments
		WHERE id = $1
	`

	var exp models.AdExperiment
	err := r.pool.QueryRow(ctx, query, experimentID).Scan(
		&exp.ID, &exp.Name, &exp.Description, &exp.Status, &exp.StartDate,
		&exp.EndDate, &exp.TrafficPercent, &exp.WinningVariant, &exp.CreatedAt, &exp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	return &exp, nil
}

// GetRunningExperiments retrieves all currently running experiments
func (r *AdRepository) GetRunningExperiments(ctx context.Context) ([]models.AdExperiment, error) {
	query := `
		SELECT id, name, description, status, start_date, end_date, traffic_percent, winning_variant, created_at, updated_at
		FROM ad_experiments
		WHERE status = 'running'
		  AND (start_date IS NULL OR start_date <= NOW())
		  AND (end_date IS NULL OR end_date > NOW())
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get running experiments: %w", err)
	}
	defer rows.Close()

	var experiments []models.AdExperiment
	for rows.Next() {
		var exp models.AdExperiment
		err := rows.Scan(
			&exp.ID, &exp.Name, &exp.Description, &exp.Status, &exp.StartDate,
			&exp.EndDate, &exp.TrafficPercent, &exp.WinningVariant, &exp.CreatedAt, &exp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan experiment: %w", err)
		}
		experiments = append(experiments, exp)
	}

	return experiments, rows.Err()
}

// GetAdsByExperiment retrieves all ads for a given experiment
func (r *AdRepository) GetAdsByExperiment(ctx context.Context, experimentID uuid.UUID) ([]models.Ad, error) {
	query := `
		SELECT id, name, advertiser_name, ad_type, content_url, click_url, alt_text,
			width, height, priority, weight, daily_budget_cents, total_budget_cents,
			spent_today_cents, spent_total_cents, cpm_cents, is_active, start_date,
			end_date, targeting_criteria, slot_id, experiment_id, experiment_variant, created_at, updated_at
		FROM ads
		WHERE experiment_id = $1 AND is_active = true
	`

	rows, err := r.pool.Query(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ads by experiment: %w", err)
	}
	defer rows.Close()

	var ads []models.Ad
	for rows.Next() {
		var ad models.Ad
		var targetingJSON []byte
		err := rows.Scan(
			&ad.ID, &ad.Name, &ad.AdvertiserName, &ad.AdType, &ad.ContentURL, &ad.ClickURL,
			&ad.AltText, &ad.Width, &ad.Height, &ad.Priority, &ad.Weight, &ad.DailyBudgetCents,
			&ad.TotalBudgetCents, &ad.SpentTodayCents, &ad.SpentTotalCents, &ad.CPMCents,
			&ad.IsActive, &ad.StartDate, &ad.EndDate, &targetingJSON, &ad.SlotID,
			&ad.ExperimentID, &ad.ExperimentVariant, &ad.CreatedAt, &ad.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ad: %w", err)
		}

		if targetingJSON != nil {
			if err := json.Unmarshal(targetingJSON, &ad.TargetingCriteria); err != nil {
				ad.TargetingCriteria = nil
			}
		}
		ads = append(ads, ad)
	}

	return ads, rows.Err()
}

// GetCTRReportByCampaign retrieves CTR report grouped by campaign (ad)
func (r *AdRepository) GetCTRReportByCampaign(ctx context.Context, since time.Time) ([]models.AdCTRReport, error) {
	query := `
		SELECT
			a.id as ad_id,
			a.name as ad_name,
			COUNT(i.id) as impressions,
			COUNT(CASE WHEN i.is_viewable THEN 1 END) as viewable_impressions,
			COUNT(CASE WHEN i.is_clicked THEN 1 END) as clicks,
			COALESCE(SUM(i.cost_cents), 0) as spend_cents
		FROM ads a
		LEFT JOIN ad_impressions i ON a.id = i.ad_id AND i.created_at >= $1
		GROUP BY a.id, a.name
		ORDER BY impressions DESC
	`

	rows, err := r.pool.Query(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get CTR report by campaign: %w", err)
	}
	defer rows.Close()

	var reports []models.AdCTRReport
	for rows.Next() {
		var report models.AdCTRReport
		err := rows.Scan(
			&report.AdID, &report.AdName, &report.Impressions,
			&report.ViewableImpressions, &report.Clicks, &report.SpendCents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CTR report: %w", err)
		}

		// Calculate CTR and viewability rate
		if report.ViewableImpressions > 0 {
			report.CTR = float64(report.Clicks) / float64(report.ViewableImpressions) * 100
		}
		if report.Impressions > 0 {
			report.ViewabilityRate = float64(report.ViewableImpressions) / float64(report.Impressions) * 100
		}

		reports = append(reports, report)
	}

	return reports, rows.Err()
}

// GetCTRReportBySlot retrieves CTR report grouped by ad slot
func (r *AdRepository) GetCTRReportBySlot(ctx context.Context, since time.Time) ([]models.AdSlotReport, error) {
	query := `
		SELECT
			COALESCE(i.slot_id, 'unassigned') as slot_id,
			COUNT(i.id) as impressions,
			COUNT(CASE WHEN i.is_viewable THEN 1 END) as viewable_impressions,
			COUNT(CASE WHEN i.is_clicked THEN 1 END) as clicks,
			COALESCE(SUM(i.cost_cents), 0) as spend_cents,
			COUNT(DISTINCT i.ad_id) as unique_ads
		FROM ad_impressions i
		WHERE i.created_at >= $1
		GROUP BY i.slot_id
		ORDER BY impressions DESC
	`

	rows, err := r.pool.Query(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get CTR report by slot: %w", err)
	}
	defer rows.Close()

	var reports []models.AdSlotReport
	for rows.Next() {
		var report models.AdSlotReport
		err := rows.Scan(
			&report.SlotID, &report.Impressions, &report.ViewableImpressions,
			&report.Clicks, &report.SpendCents, &report.UniqueAds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slot report: %w", err)
		}

		// Calculate CTR and viewability rate
		if report.ViewableImpressions > 0 {
			report.CTR = float64(report.Clicks) / float64(report.ViewableImpressions) * 100
		}
		if report.Impressions > 0 {
			report.ViewabilityRate = float64(report.ViewableImpressions) / float64(report.Impressions) * 100
		}

		reports = append(reports, report)
	}

	return reports, rows.Err()
}

// GetExperimentReport retrieves analytics for an experiment with variant comparison
func (r *AdRepository) GetExperimentReport(ctx context.Context, experimentID uuid.UUID, since time.Time) (*models.AdExperimentReport, error) {
	// Get experiment info
	exp, err := r.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	// Get variant metrics
	query := `
		SELECT
			experiment_variant as variant,
			COUNT(id) as impressions,
			COUNT(CASE WHEN is_viewable THEN 1 END) as viewable_impressions,
			COUNT(CASE WHEN is_clicked THEN 1 END) as clicks
		FROM ad_impressions
		WHERE experiment_id = $1 AND created_at >= $2
		GROUP BY experiment_variant
		ORDER BY experiment_variant
	`

	rows, err := r.pool.Query(ctx, query, experimentID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment variant metrics: %w", err)
	}
	defer rows.Close()

	var variants []models.AdExperimentVariantReport
	for rows.Next() {
		var variant models.AdExperimentVariantReport
		err := rows.Scan(
			&variant.Variant, &variant.Impressions,
			&variant.ViewableImpressions, &variant.Clicks,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant report: %w", err)
		}

		// Calculate CTR
		if variant.ViewableImpressions > 0 {
			variant.CTR = float64(variant.Clicks) / float64(variant.ViewableImpressions) * 100
		}
		// Note: ConversionRate calculation requires conversion tracking implementation
		// Conversions are tracked separately in ad_experiment_analytics table

		variants = append(variants, variant)
	}

	return &models.AdExperimentReport{
		ExperimentID:   exp.ID,
		ExperimentName: exp.Name,
		Status:         exp.Status,
		Variants:       variants,
	}, nil
}

// UpsertCampaignAnalytics updates or inserts daily campaign analytics
func (r *AdRepository) UpsertCampaignAnalytics(ctx context.Context, adID uuid.UUID, date time.Time, slotID *string, impressions, viewableImpressions, clicks int, spendCents int64, uniqueUsers int) error {
	query := `
		INSERT INTO ad_campaign_analytics (id, ad_id, date, slot_id, impressions, viewable_impressions, clicks, spend_cents, unique_users)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (ad_id, date, slot_id)
		DO UPDATE SET
			impressions = ad_campaign_analytics.impressions + EXCLUDED.impressions,
			viewable_impressions = ad_campaign_analytics.viewable_impressions + EXCLUDED.viewable_impressions,
			clicks = ad_campaign_analytics.clicks + EXCLUDED.clicks,
			spend_cents = ad_campaign_analytics.spend_cents + EXCLUDED.spend_cents,
			unique_users = ad_campaign_analytics.unique_users + EXCLUDED.unique_users,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query, adID, date, slotID, impressions, viewableImpressions, clicks, spendCents, uniqueUsers)
	if err != nil {
		return fmt.Errorf("failed to upsert campaign analytics: %w", err)
	}

	return nil
}

// ListCampaigns retrieves all campaigns with optional filtering
func (r *AdRepository) ListCampaigns(ctx context.Context, page, limit int, status *string) ([]models.Ad, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if status != nil {
		switch *status {
		case "active":
			whereClauses = append(whereClauses, "is_active = true AND (end_date IS NULL OR end_date > NOW())")
		case "inactive":
			whereClauses = append(whereClauses, "is_active = false")
		case "ended":
			whereClauses = append(whereClauses, "end_date IS NOT NULL AND end_date <= NOW()")
		case "scheduled":
			whereClauses = append(whereClauses, "is_active = true AND start_date IS NOT NULL AND start_date > NOW()")
		}
	}

	whereClause := whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM ads WHERE %s`, whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Get campaigns with pagination
	offset := (page - 1) * limit
	args = append(args, limit, offset)
	// Use len(args)-1 and len(args) for placeholders to correctly track argument positions
	limitPlaceholder := utils.SQLPlaceholder(len(args) - 1)
	offsetPlaceholder := utils.SQLPlaceholder(len(args))
	query := fmt.Sprintf(`
		SELECT id, name, advertiser_name, ad_type, content_url, click_url, alt_text,
			width, height, priority, weight, daily_budget_cents, total_budget_cents,
			spent_today_cents, spent_total_cents, cpm_cents, is_active, start_date,
			end_date, targeting_criteria, created_at, updated_at
		FROM ads
		WHERE %s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s
	`, whereClause, limitPlaceholder, offsetPlaceholder)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []models.Ad
	for rows.Next() {
		var ad models.Ad
		var targetingJSON []byte
		err := rows.Scan(
			&ad.ID, &ad.Name, &ad.AdvertiserName, &ad.AdType, &ad.ContentURL, &ad.ClickURL,
			&ad.AltText, &ad.Width, &ad.Height, &ad.Priority, &ad.Weight, &ad.DailyBudgetCents,
			&ad.TotalBudgetCents, &ad.SpentTodayCents, &ad.SpentTotalCents, &ad.CPMCents,
			&ad.IsActive, &ad.StartDate, &ad.EndDate, &targetingJSON, &ad.CreatedAt, &ad.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan campaign: %w", err)
		}

		if targetingJSON != nil {
			if err := json.Unmarshal(targetingJSON, &ad.TargetingCriteria); err != nil {
				ad.TargetingCriteria = nil
			}
		}
		campaigns = append(campaigns, ad)
	}

	return campaigns, total, rows.Err()
}

// CreateCampaign creates a new ad campaign
func (r *AdRepository) CreateCampaign(ctx context.Context, ad *models.Ad) error {
	var targetingJSON []byte
	var err error
	if ad.TargetingCriteria != nil {
		targetingJSON, err = json.Marshal(ad.TargetingCriteria)
		if err != nil {
			return fmt.Errorf("failed to marshal targeting criteria: %w", err)
		}
	}

	query := `
		INSERT INTO ads (id, name, advertiser_name, ad_type, content_url, click_url, alt_text,
			width, height, priority, weight, daily_budget_cents, total_budget_cents,
			cpm_cents, is_active, start_date, end_date, targeting_criteria)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	if ad.ID == uuid.Nil {
		ad.ID = uuid.New()
	}

	_, err = r.pool.Exec(ctx, query,
		ad.ID, ad.Name, ad.AdvertiserName, ad.AdType, ad.ContentURL, ad.ClickURL, ad.AltText,
		ad.Width, ad.Height, ad.Priority, ad.Weight, ad.DailyBudgetCents, ad.TotalBudgetCents,
		ad.CPMCents, ad.IsActive, ad.StartDate, ad.EndDate, targetingJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create campaign: %w", err)
	}

	return nil
}

// UpdateCampaign updates an existing ad campaign
func (r *AdRepository) UpdateCampaign(ctx context.Context, ad *models.Ad) error {
	var targetingJSON []byte
	var err error
	if ad.TargetingCriteria != nil {
		targetingJSON, err = json.Marshal(ad.TargetingCriteria)
		if err != nil {
			return fmt.Errorf("failed to marshal targeting criteria: %w", err)
		}
	}

	query := `
		UPDATE ads SET
			name = $2,
			advertiser_name = $3,
			ad_type = $4,
			content_url = $5,
			click_url = $6,
			alt_text = $7,
			width = $8,
			height = $9,
			priority = $10,
			weight = $11,
			daily_budget_cents = $12,
			total_budget_cents = $13,
			cpm_cents = $14,
			is_active = $15,
			start_date = $16,
			end_date = $17,
			targeting_criteria = $18,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		ad.ID, ad.Name, ad.AdvertiserName, ad.AdType, ad.ContentURL, ad.ClickURL, ad.AltText,
		ad.Width, ad.Height, ad.Priority, ad.Weight, ad.DailyBudgetCents, ad.TotalBudgetCents,
		ad.CPMCents, ad.IsActive, ad.StartDate, ad.EndDate, targetingJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("campaign not found")
	}

	return nil
}

// DeleteCampaign deletes an ad campaign by ID
func (r *AdRepository) DeleteCampaign(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ads WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("campaign not found")
	}

	return nil
}

// GetCampaignReportByDate retrieves campaign performance report by date range
func (r *AdRepository) GetCampaignReportByDate(ctx context.Context, adID *uuid.UUID, startDate, endDate time.Time) ([]models.AdCampaignAnalytics, error) {
	whereClauses := []string{"date >= $1", "date <= $2"}
	args := []interface{}{startDate, endDate}
	argIndex := 3

	if adID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("ad_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *adID)
	}

	whereClause := whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}

	query := fmt.Sprintf(`
		SELECT id, ad_id, date, slot_id, impressions, viewable_impressions, clicks, spend_cents, unique_users, created_at, updated_at
		FROM ad_campaign_analytics
		WHERE %s
		ORDER BY date DESC
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign report by date: %w", err)
	}
	defer rows.Close()

	var reports []models.AdCampaignAnalytics
	for rows.Next() {
		var report models.AdCampaignAnalytics
		err := rows.Scan(
			&report.ID, &report.AdID, &report.Date, &report.SlotID, &report.Impressions,
			&report.ViewableImpressions, &report.Clicks, &report.SpendCents, &report.UniqueUsers,
			&report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign report: %w", err)
		}
		reports = append(reports, report)
	}

	return reports, rows.Err()
}

// GetCampaignReportByPlacement retrieves campaign performance report grouped by placement/slot
func (r *AdRepository) GetCampaignReportByPlacement(ctx context.Context, adID *uuid.UUID, since time.Time) ([]models.AdSlotReport, error) {
	args := []interface{}{since}
	whereClause := "i.created_at >= $1"

	if adID != nil {
		whereClause += " AND i.ad_id = $2"
		args = append(args, *adID)
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(i.slot_id, 'unassigned') as slot_id,
			COUNT(i.id) as impressions,
			COUNT(CASE WHEN i.is_viewable THEN 1 END) as viewable_impressions,
			COUNT(CASE WHEN i.is_clicked THEN 1 END) as clicks,
			COALESCE(SUM(i.cost_cents), 0) as spend_cents,
			COUNT(DISTINCT i.ad_id) as unique_ads
		FROM ad_impressions i
		WHERE %s
		GROUP BY i.slot_id
		ORDER BY impressions DESC
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign report by placement: %w", err)
	}
	defer rows.Close()

	var reports []models.AdSlotReport
	for rows.Next() {
		var report models.AdSlotReport
		err := rows.Scan(
			&report.SlotID, &report.Impressions, &report.ViewableImpressions,
			&report.Clicks, &report.SpendCents, &report.UniqueAds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan placement report: %w", err)
		}

		// Calculate CTR and viewability rate
		if report.ViewableImpressions > 0 {
			report.CTR = float64(report.Clicks) / float64(report.ViewableImpressions) * 100
		}
		if report.Impressions > 0 {
			report.ViewabilityRate = float64(report.ViewableImpressions) / float64(report.Impressions) * 100
		}

		reports = append(reports, report)
	}

	return reports, rows.Err()
}
