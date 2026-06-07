package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// AnalyticsRepository handles analytics data access
type AnalyticsRepository struct {
	db *pgxpool.Pool
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// TrackEvent records an analytics event
func (r *AnalyticsRepository) TrackEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	query := `
		INSERT INTO analytics_events (event_type, user_id, clip_id, metadata, ip_address, user_agent, referrer)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	return r.db.QueryRow(ctx, query,
		event.EventType,
		event.UserID,
		event.ClipID,
		event.Metadata,
		event.IPAddress,
		event.UserAgent,
		event.Referrer,
	).Scan(&event.ID, &event.CreatedAt)
}

// GetCreatorAnalytics retrieves analytics for a specific creator
func (r *AnalyticsRepository) GetCreatorAnalytics(ctx context.Context, creatorName string) (*models.CreatorAnalytics, error) {
	query := `
		SELECT creator_name, creator_id, total_clips, total_views, total_upvotes,
		       total_downvotes, total_comments, total_favorites, avg_engagement_rate,
		       follower_count, updated_at
		FROM creator_analytics
		WHERE creator_name = $1
	`

	var analytics models.CreatorAnalytics
	err := r.db.QueryRow(ctx, query, creatorName).Scan(
		&analytics.CreatorName,
		&analytics.CreatorID,
		&analytics.TotalClips,
		&analytics.TotalViews,
		&analytics.TotalUpvotes,
		&analytics.TotalDownvotes,
		&analytics.TotalComments,
		&analytics.TotalFavorites,
		&analytics.AvgEngagementRate,
		&analytics.FollowerCount,
		&analytics.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &analytics, nil
}

// GetCreatorTopClips returns top clips for a creator sorted by a specific metric
func (r *AnalyticsRepository) GetCreatorTopClips(ctx context.Context, creatorName, sortBy string, limit int) ([]models.CreatorTopClip, error) {
	// Determine sort column
	sortColumn := "c.vote_score"
	switch sortBy {
	case "views":
		sortColumn = "COALESCE(ca.total_views, 0)"
	case "comments":
		sortColumn = "c.comment_count"
	case "votes":
		sortColumn = "c.vote_score"
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
		       c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
		       c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
		       c.view_count, c.created_at, c.imported_at, c.vote_score,
		       c.comment_count, c.favorite_count, c.is_featured, c.is_nsfw,
		       c.is_removed, c.removed_reason,
		       COALESCE(ca.total_views, 0) as views,
		       CASE
		           WHEN COALESCE(ca.total_views, 0) > 0
		           THEN (c.vote_score::float + c.comment_count::float) / ca.total_views::float
		           ELSE 0
		       END as engagement_rate
		FROM clips c
		LEFT JOIN clip_analytics ca ON c.id = ca.clip_id
		WHERE c.creator_name = $1 AND c.is_removed = false
		ORDER BY %s DESC
		LIMIT $2
	`, sortColumn)

	rows, err := r.db.Query(ctx, query, creatorName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clips []models.CreatorTopClip
	for rows.Next() {
		var clip models.CreatorTopClip
		err := rows.Scan(
			&clip.ID,
			&clip.TwitchClipID,
			&clip.TwitchClipURL,
			&clip.EmbedURL,
			&clip.Title,
			&clip.CreatorName,
			&clip.CreatorID,
			&clip.BroadcasterName,
			&clip.BroadcasterID,
			&clip.GameID,
			&clip.GameName,
			&clip.Language,
			&clip.ThumbnailURL,
			&clip.Duration,
			&clip.ViewCount,
			&clip.CreatedAt,
			&clip.ImportedAt,
			&clip.VoteScore,
			&clip.CommentCount,
			&clip.FavoriteCount,
			&clip.IsFeatured,
			&clip.IsNSFW,
			&clip.IsRemoved,
			&clip.RemovedReason,
			&clip.Views,
			&clip.EngagementRate,
		)
		if err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}

	return clips, rows.Err()
}

// GetCreatorTrends returns time-series data for a creator's performance metrics
func (r *AnalyticsRepository) GetCreatorTrends(ctx context.Context, creatorName, metricType string, days int) ([]models.TrendDataPoint, error) {
	query := `
		SELECT da.date, COALESCE(SUM(da.value), 0) as value
		FROM daily_analytics da
		WHERE da.entity_type = 'creator'
		  AND da.entity_id = $1
		  AND da.metric_type = $2
		  AND da.date >= CURRENT_DATE - $3 * INTERVAL '1 day'
		GROUP BY da.date
		ORDER BY da.date ASC
	`

	rows, err := r.db.Query(ctx, query, creatorName, metricType, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []models.TrendDataPoint
	for rows.Next() {
		var point models.TrendDataPoint
		err := rows.Scan(&point.Date, &point.Value)
		if err != nil {
			return nil, err
		}
		trends = append(trends, point)
	}

	return trends, rows.Err()
}

// GetClipAnalytics retrieves analytics for a specific clip
func (r *AnalyticsRepository) GetClipAnalytics(ctx context.Context, clipID uuid.UUID) (*models.ClipAnalytics, error) {
	query := `
		SELECT clip_id, total_views, unique_viewers, avg_view_duration,
		       total_shares, peak_concurrent_viewers, retention_rate,
		       first_viewed_at, last_viewed_at, updated_at
		FROM clip_analytics
		WHERE clip_id = $1
	`

	var analytics models.ClipAnalytics
	err := r.db.QueryRow(ctx, query, clipID).Scan(
		&analytics.ClipID,
		&analytics.TotalViews,
		&analytics.UniqueViewers,
		&analytics.AvgViewDuration,
		&analytics.TotalShares,
		&analytics.PeakConcurrentViews,
		&analytics.RetentionRate,
		&analytics.FirstViewedAt,
		&analytics.LastViewedAt,
		&analytics.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &analytics, nil
}

// GetUserAnalytics retrieves personal statistics for a user
func (r *AnalyticsRepository) GetUserAnalytics(ctx context.Context, userID uuid.UUID) (*models.UserAnalytics, error) {
	query := `
		SELECT user_id, clips_upvoted, clips_downvoted, comments_posted,
		       clips_favorited, searches_performed, days_active,
		       total_karma_earned, last_active_at, updated_at
		FROM user_analytics
		WHERE user_id = $1
	`

	var analytics models.UserAnalytics
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&analytics.UserID,
		&analytics.ClipsUpvoted,
		&analytics.ClipsDownvoted,
		&analytics.CommentsPosted,
		&analytics.ClipsFavorited,
		&analytics.SearchesPerformed,
		&analytics.DaysActive,
		&analytics.TotalKarmaEarned,
		&analytics.LastActiveAt,
		&analytics.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &analytics, nil
}

// GetPlatformAnalytics retrieves platform-wide statistics
func (r *AnalyticsRepository) GetPlatformAnalytics(ctx context.Context, date time.Time) (*models.PlatformAnalytics, error) {
	query := `
		SELECT id, date, total_users, active_users_daily, active_users_weekly,
		       active_users_monthly, new_users_today, total_clips, new_clips_today,
		       total_votes, votes_today, total_comments, comments_today,
		       total_views, views_today, avg_session_duration, metadata, created_at
		FROM platform_analytics
		WHERE date = $1
	`

	var analytics models.PlatformAnalytics
	err := r.db.QueryRow(ctx, query, date).Scan(
		&analytics.ID,
		&analytics.Date,
		&analytics.TotalUsers,
		&analytics.ActiveUsersDaily,
		&analytics.ActiveUsersWeekly,
		&analytics.ActiveUsersMonthly,
		&analytics.NewUsersToday,
		&analytics.TotalClips,
		&analytics.NewClipsToday,
		&analytics.TotalVotes,
		&analytics.VotesToday,
		&analytics.TotalComments,
		&analytics.CommentsToday,
		&analytics.TotalViews,
		&analytics.ViewsToday,
		&analytics.AvgSessionDuration,
		&analytics.Metadata,
		&analytics.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &analytics, nil
}

// GetPlatformOverviewMetrics retrieves current platform KPIs
func (r *AnalyticsRepository) GetPlatformOverviewMetrics(ctx context.Context) (*models.PlatformOverviewMetrics, error) {
	query := `
		WITH latest_analytics AS (
			SELECT * FROM platform_analytics
			ORDER BY date DESC
			LIMIT 1
		)
		SELECT
			COALESCE(total_users, 0),
			COALESCE(active_users_daily, 0),
			COALESCE(active_users_monthly, 0),
			COALESCE(total_clips, 0),
			COALESCE(new_clips_today, 0),
			COALESCE(total_votes, 0),
			COALESCE(total_comments, 0),
			COALESCE(avg_session_duration, 0)
		FROM latest_analytics
	`

	var metrics models.PlatformOverviewMetrics
	err := r.db.QueryRow(ctx, query).Scan(
		&metrics.TotalUsers,
		&metrics.ActiveUsersDaily,
		&metrics.ActiveUsersMonthly,
		&metrics.TotalClips,
		&metrics.ClipsAddedToday,
		&metrics.TotalVotes,
		&metrics.TotalComments,
		&metrics.AvgSessionDuration,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Return zero-value metrics when no analytics data exists yet
			return &models.PlatformOverviewMetrics{}, nil
		}
		return nil, err
	}

	return &metrics, nil
}

// GetMostPopularGames retrieves top games by clip count and views
func (r *AnalyticsRepository) GetMostPopularGames(ctx context.Context, limit int) ([]models.GameMetric, error) {
	query := `
		SELECT
			c.game_id,
			c.game_name,
			COUNT(*) as clip_count,
			COALESCE(SUM(ca.total_views), 0) as view_count
		FROM clips c
		LEFT JOIN clip_analytics ca ON c.id = ca.clip_id
		WHERE c.is_removed = false AND c.game_name IS NOT NULL
		GROUP BY c.game_id, c.game_name
		ORDER BY clip_count DESC, view_count DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.GameMetric
	for rows.Next() {
		var game models.GameMetric
		err := rows.Scan(&game.GameID, &game.GameName, &game.ClipCount, &game.ViewCount)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, rows.Err()
}

// GetMostPopularCreators retrieves top creators by metrics
func (r *AnalyticsRepository) GetMostPopularCreators(ctx context.Context, limit int) ([]models.CreatorMetric, error) {
	query := `
		SELECT
			creator_id,
			creator_name,
			total_clips,
			total_views,
			total_upvotes
		FROM creator_analytics
		ORDER BY total_views DESC, total_upvotes DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creators []models.CreatorMetric
	for rows.Next() {
		var creator models.CreatorMetric
		err := rows.Scan(
			&creator.CreatorID,
			&creator.CreatorName,
			&creator.ClipCount,
			&creator.ViewCount,
			&creator.VoteScore,
		)
		if err != nil {
			return nil, err
		}
		creators = append(creators, creator)
	}

	return creators, rows.Err()
}

// GetTrendingTags returns tags that are trending based on recent clip activity
func (r *AnalyticsRepository) GetTrendingTags(ctx context.Context, days, limit int) ([]models.TagMetric, error) {
	query := `
		SELECT t.id, t.name, COUNT(*) as usage_count
		FROM tags t
		JOIN clip_tags ct ON t.id = ct.tag_id
		JOIN clips c ON ct.clip_id = c.id
		WHERE ct.created_at >= CURRENT_DATE - $1 * INTERVAL '1 day'
		  AND c.is_removed = false
		GROUP BY t.id, t.name
		ORDER BY usage_count DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, days, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.TagMetric
	for rows.Next() {
		var tag models.TagMetric
		err := rows.Scan(&tag.TagID, &tag.TagName, &tag.UsageCount)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// GetPlatformTrends retrieves time-series data for platform metrics
func (r *AnalyticsRepository) GetPlatformTrends(ctx context.Context, metricType string, days int) ([]models.TrendDataPoint, error) {
	// Map metric types to column names
	var column string
	switch metricType {
	case "users":
		column = "new_users_today"
	case "clips":
		column = "new_clips_today"
	case "views":
		column = "views_today"
	case "votes":
		column = "votes_today"
	case "comments":
		column = "comments_today"
	default:
		return nil, fmt.Errorf("invalid metric type: %s", metricType)
	}

	query := fmt.Sprintf(`
		SELECT date, %s as value
		FROM platform_analytics
		WHERE date >= CURRENT_DATE - $1 * INTERVAL '1 day'
		ORDER BY date ASC
	`, column)

	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []models.TrendDataPoint
	for rows.Next() {
		var point models.TrendDataPoint
		var value int
		err := rows.Scan(&point.Date, &value)
		if err != nil {
			return nil, err
		}
		point.Value = int64(value)
		trends = append(trends, point)
	}

	return trends, rows.Err()
}

// GetAverageClipVoteScore retrieves the average vote score across all clips
func (r *AnalyticsRepository) GetAverageClipVoteScore(ctx context.Context) (float64, error) {
	query := `
SELECT COALESCE(AVG(vote_score), 0) as avg_vote_score
FROM clips
WHERE is_removed = false
`

	var avgScore float64
	err := r.db.QueryRow(ctx, query).Scan(&avgScore)
	if err != nil {
		return 0, err
	}

	return avgScore, nil
}

// parseDeviceType categorizes user agents into device types for analytics purposes.
//
// Returns one of:
//   - "mobile": for smartphones (e.g., Android, iPhone, etc.), explicitly excluding iPads.
//   - "tablet": for tablets and iPads (e.g., "ipad", "tablet", "kindle", "playbook").
//   - "desktop": for computers and known desktop browsers (e.g., "windows", "macintosh", "linux", "chrome", "firefox", "safari", "edge").
//   - "unknown": for user agents that do not match any known keywords.
//
// Keywords are checked in order: mobile, then tablet, then desktop.
// If a user agent matches multiple categories, the first matching category in this order is returned.
// For example, "ipad" is excluded from "mobile" and included in "tablet".
// The function is case-insensitive and returns "unknown" for empty or unrecognized user agents.
func parseDeviceType(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	ua := strings.ToLower(userAgent)

	// Check for mobile devices
	mobileKeywords := []string{"mobile", "android", "iphone", "ipod", "blackberry", "windows phone"}
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) && !strings.Contains(ua, "ipad") {
			return "mobile"
		}
	}

	// Check for tablets
	tabletKeywords := []string{"ipad", "tablet", "kindle", "playbook"}
	for _, keyword := range tabletKeywords {
		if strings.Contains(ua, keyword) {
			return "tablet"
		}
	}

	// Check for desktop/known browsers
	desktopKeywords := []string{"windows", "macintosh", "linux", "chrome", "firefox", "safari", "edge"}
	for _, keyword := range desktopKeywords {
		if strings.Contains(ua, keyword) {
			return "desktop"
		}
	}

	return "unknown"
}

// extractCountryFromIP extracts a country code from an IP address
// This is a simplified implementation that returns "XX" (unknown) for all IPs
// In production, this would use a GeoIP database like MaxMind GeoLite2
func extractCountryFromIP(ipAddress string) string {
	// Use empty string to represent invalid or missing IP addresses
	if ipAddress == "" {
		return "XX" // Unknown country code
	}

	// For now, return XX (unknown) for all IPs
	// In production, you would use a GeoIP library:
	// - github.com/oschwald/geoip2-golang with MaxMind GeoLite2 database
	// - or use a GeoIP service API
	return "XX"
}

// GetCreatorAudienceInsights retrieves audience insights (geography and devices) for a creator
func (r *AnalyticsRepository) GetCreatorAudienceInsights(ctx context.Context, creatorName string, limit int) (*models.CreatorAudienceInsights, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Get all clip IDs for this creator
	clipsQuery := `
		SELECT id FROM clips
		WHERE creator_name = $1 AND is_removed = false
	`

	rows, err := r.db.Query(ctx, clipsQuery, creatorName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clipIDs []uuid.UUID
	for rows.Next() {
		var clipID uuid.UUID
		if err := rows.Scan(&clipID); err != nil {
			return nil, err
		}
		clipIDs = append(clipIDs, clipID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(clipIDs) == 0 {
		// No clips for this creator
		return &models.CreatorAudienceInsights{
			TopCountries: []models.GeographyMetric{},
			DeviceTypes:  []models.DeviceMetric{},
			TotalViews:   0,
		}, nil
	}

	// Query analytics events for these clips
	eventsQuery := `
		SELECT user_agent, ip_address
		FROM analytics_events
		WHERE event_type = 'clip_view'
		  AND clip_id = ANY($1)
		  AND created_at >= CURRENT_DATE - INTERVAL '90 days'
	`

	eventsRows, err := r.db.Query(ctx, eventsQuery, clipIDs)
	if err != nil {
		return nil, err
	}
	defer eventsRows.Close()

	// Count views by device type and country
	deviceCounts := make(map[string]int64)
	countryCounts := make(map[string]int64)
	totalViews := int64(0)

	for eventsRows.Next() {
		var userAgent, ipAddress *string
		if err := eventsRows.Scan(&userAgent, &ipAddress); err != nil {
			return nil, err
		}

		totalViews++

		// Parse device type
		ua := ""
		if userAgent != nil {
			ua = *userAgent
		}
		deviceType := parseDeviceType(ua)
		deviceCounts[deviceType]++

		// Extract country
		ip := ""
		if ipAddress != nil {
			ip = *ipAddress
		}
		country := extractCountryFromIP(ip)
		countryCounts[country]++
	}

	if err := eventsRows.Err(); err != nil {
		return nil, err
	}

	// Convert device counts to sorted slice
	deviceMetrics := make([]models.DeviceMetric, 0, len(deviceCounts))
	for deviceType, count := range deviceCounts {
		percentage := 0.0
		if totalViews > 0 {
			percentage = float64(count) / float64(totalViews) * 100
		}
		deviceMetrics = append(deviceMetrics, models.DeviceMetric{
			DeviceType: deviceType,
			ViewCount:  count,
			Percentage: percentage,
		})
	}

	// Sort device metrics by view count (descending)
	sort.Slice(deviceMetrics, func(i, j int) bool {
		return deviceMetrics[i].ViewCount > deviceMetrics[j].ViewCount
	})

	// Convert country counts to sorted slice (top N countries)
	type countryCount struct {
		country string
		count   int64
	}
	countryCountsSlice := make([]countryCount, 0, len(countryCounts))
	for country, count := range countryCounts {
		countryCountsSlice = append(countryCountsSlice, countryCount{country, count})
	}

	// Sort by count (descending)
	sort.Slice(countryCountsSlice, func(i, j int) bool {
		return countryCountsSlice[i].count > countryCountsSlice[j].count
	})

	// Take top N countries
	topN := limit
	if topN > len(countryCountsSlice) {
		topN = len(countryCountsSlice)
	}

	geographyMetrics := make([]models.GeographyMetric, topN)
	for i := 0; i < topN; i++ {
		percentage := 0.0
		if totalViews > 0 {
			percentage = float64(countryCountsSlice[i].count) / float64(totalViews) * 100
		}
		geographyMetrics[i] = models.GeographyMetric{
			Country:    countryCountsSlice[i].country,
			ViewCount:  countryCountsSlice[i].count,
			Percentage: percentage,
		}
	}

	return &models.CreatorAudienceInsights{
		TopCountries: geographyMetrics,
		DeviceTypes:  deviceMetrics,
		TotalViews:   totalViews,
	}, nil
}

// GetUserPostsCount returns the number of posts/submissions by a user in the last N days
func (r *AnalyticsRepository) GetUserPostsCount(ctx context.Context, userID uuid.UUID, days int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM clip_submissions
		WHERE user_id = $1
		  AND created_at >= CURRENT_DATE - $2 * INTERVAL '1 day'
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID, days).Scan(&count)
	return count, err
}

// GetUserCommentsCount returns the number of comments posted by a user in the last N days
func (r *AnalyticsRepository) GetUserCommentsCount(ctx context.Context, userID uuid.UUID, days int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM comments
		WHERE user_id = $1
		  AND created_at >= CURRENT_DATE - $2 * INTERVAL '1 day'
		  AND is_deleted = false
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID, days).Scan(&count)
	return count, err
}

// GetUserVotesCount returns the number of votes cast by a user in the last N days
func (r *AnalyticsRepository) GetUserVotesCount(ctx context.Context, userID uuid.UUID, days int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM votes
		WHERE user_id = $1
		  AND created_at >= CURRENT_DATE - $2 * INTERVAL '1 day'
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID, days).Scan(&count)
	return count, err
}

// GetUserLoginDays returns the number of unique days a user logged in over the last N days
func (r *AnalyticsRepository) GetUserLoginDays(ctx context.Context, userID uuid.UUID, days int) (int, error) {
	query := `
		SELECT COUNT(DISTINCT DATE(created_at))
		FROM analytics_events
		WHERE user_id = $1
		  AND event_type = 'login'
		  AND created_at >= CURRENT_DATE - $2 * INTERVAL '1 day'
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID, days).Scan(&count)
	return count, err
}

// GetUserAvgDailyMinutes returns average daily minutes spent by a user over the last N days
func (r *AnalyticsRepository) GetUserAvgDailyMinutes(ctx context.Context, userID uuid.UUID, days int) (float64, error) {
	query := `
		SELECT COALESCE(AVG(daily_minutes), 0)
		FROM (
			SELECT DATE(created_at) as day,
			       COUNT(*) * 2 as daily_minutes
			FROM analytics_events
			WHERE user_id = $1
			  AND created_at >= CURRENT_DATE - $2 * INTERVAL '1 day'
			  AND event_type IN ('clip_view', 'page_view')
			GROUP BY DATE(created_at)
		) as daily_activity
	`

	var avgMinutes float64
	err := r.db.QueryRow(ctx, query, userID, days).Scan(&avgMinutes)
	return avgMinutes, err
}

// GetDAU returns Daily Active Users count
func (r *AnalyticsRepository) GetDAU(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE created_at >= CURRENT_DATE
		  AND user_id IS NOT NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetWAU returns Weekly Active Users count
func (r *AnalyticsRepository) GetWAU(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
		  AND user_id IS NOT NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetMAU returns Monthly Active Users count
func (r *AnalyticsRepository) GetMAU(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
		  AND user_id IS NOT NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetRetentionRate returns the retention rate for users after N days
func (r *AnalyticsRepository) GetRetentionRate(ctx context.Context, days int) (float64, error) {
	query := `
		WITH cohort AS (
			SELECT DISTINCT user_id,
			       DATE(created_at) as signup_date
			FROM users
			WHERE created_at >= CURRENT_DATE - INTERVAL '60 days'
		),
		retained AS (
			SELECT c.user_id
			FROM cohort c
			INNER JOIN analytics_events ae ON c.user_id = ae.user_id
			WHERE DATE(ae.created_at) = c.signup_date + $1 * INTERVAL '1 day'
		)
		SELECT CASE
			WHEN COUNT(DISTINCT c.user_id) > 0
			THEN (COUNT(DISTINCT r.user_id)::float / COUNT(DISTINCT c.user_id)::float) * 100
			ELSE 0
		END as retention_rate
		FROM cohort c
		LEFT JOIN retained r ON c.user_id = r.user_id
	`

	var retentionRate float64
	err := r.db.QueryRow(ctx, query, days).Scan(&retentionRate)
	return retentionRate, err
}

// GetMonthlyChurnRate returns the monthly churn rate
func (r *AnalyticsRepository) GetMonthlyChurnRate(ctx context.Context) (float64, error) {
	query := `
		WITH active_last_month AS (
			SELECT DISTINCT user_id
			FROM analytics_events
			WHERE created_at >= CURRENT_DATE - INTERVAL '60 days'
			  AND created_at < CURRENT_DATE - INTERVAL '30 days'
			  AND user_id IS NOT NULL
		),
		churned_users AS (
			SELECT alm.user_id
			FROM active_last_month alm
			WHERE NOT EXISTS (
				SELECT 1
				FROM analytics_events ae
				WHERE ae.user_id = alm.user_id
				  AND ae.created_at >= CURRENT_DATE - INTERVAL '30 days'
			)
		)
		SELECT CASE
			WHEN COUNT(DISTINCT alm.user_id) > 0
			THEN (COUNT(DISTINCT cu.user_id)::float / COUNT(DISTINCT alm.user_id)::float) * 100
			ELSE 0
		END as churn_rate
		FROM active_last_month alm
		LEFT JOIN churned_users cu ON alm.user_id = cu.user_id
	`

	var churnRate float64
	err := r.db.QueryRow(ctx, query).Scan(&churnRate)
	return churnRate, err
}

// GetDAUChangeWoW returns the week-over-week percentage change in DAU
func (r *AnalyticsRepository) GetDAUChangeWoW(ctx context.Context) (float64, error) {
	query := `
		WITH last_week AS (
			SELECT COUNT(DISTINCT user_id) as count
			FROM analytics_events
			WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
			  AND user_id IS NOT NULL
		),
		prev_week AS (
			SELECT COUNT(DISTINCT user_id) as count
			FROM analytics_events
			WHERE created_at >= CURRENT_DATE - INTERVAL '14 days'
			  AND created_at < CURRENT_DATE - INTERVAL '7 days'
			  AND user_id IS NOT NULL
		)
		SELECT CASE
			WHEN pw.count > 0
			THEN ((lw.count - pw.count)::float / pw.count::float) * 100
			ELSE 0
		END as change
		FROM last_week lw, prev_week pw
	`

	var change float64
	err := r.db.QueryRow(ctx, query).Scan(&change)
	return change, err
}

// GetMAUChangeMoM returns the month-over-month percentage change in MAU
func (r *AnalyticsRepository) GetMAUChangeMoM(ctx context.Context) (float64, error) {
	query := `
		WITH this_month AS (
			SELECT COUNT(DISTINCT user_id) as count
			FROM analytics_events
			WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
			  AND user_id IS NOT NULL
		),
		prev_month AS (
			SELECT COUNT(DISTINCT user_id) as count
			FROM analytics_events
			WHERE created_at >= CURRENT_DATE - INTERVAL '60 days'
			  AND created_at < CURRENT_DATE - INTERVAL '30 days'
			  AND user_id IS NOT NULL
		)
		SELECT CASE
			WHEN pm.count > 0
			THEN ((tm.count - pm.count)::float / pm.count::float) * 100
			ELSE 0
		END as change
		FROM this_month tm, prev_month pm
	`

	var change float64
	err := r.db.QueryRow(ctx, query).Scan(&change)
	return change, err
}

// GetTrendingData returns trend data points for a specific metric
func (r *AnalyticsRepository) GetTrendingData(ctx context.Context, metric string, days int) ([]models.TrendDataPoint, error) {
	var query string

	switch metric {
	case "dau":
		query = `
			SELECT DATE(created_at) as date,
			       COUNT(DISTINCT user_id) as value
			FROM analytics_events
			WHERE created_at >= CURRENT_DATE - $1 * INTERVAL '1 day'
			  AND user_id IS NOT NULL
			GROUP BY DATE(created_at)
			ORDER BY date ASC
		`
	case "clips":
		query = `
			SELECT DATE(created_at) as date,
			       COUNT(*) as value
			FROM clips
			WHERE created_at >= CURRENT_DATE - $1 * INTERVAL '1 day'
			GROUP BY DATE(created_at)
			ORDER BY date ASC
		`
	case "votes":
		query = `
			SELECT DATE(created_at) as date,
			       COUNT(*) as value
			FROM votes
			WHERE created_at >= CURRENT_DATE - $1 * INTERVAL '1 day'
			GROUP BY DATE(created_at)
			ORDER BY date ASC
		`
	case "comments":
		query = `
			SELECT DATE(created_at) as date,
			       COUNT(*) as value
			FROM comments
			WHERE created_at >= CURRENT_DATE - $1 * INTERVAL '1 day'
			  AND is_deleted = false
			GROUP BY DATE(created_at)
			ORDER BY date ASC
		`
	default:
		return nil, fmt.Errorf("unsupported metric: %s", metric)
	}

	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataPoints []models.TrendDataPoint
	for rows.Next() {
		var point models.TrendDataPoint
		err := rows.Scan(&point.Date, &point.Value)
		if err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, point)
	}

	return dataPoints, rows.Err()
}

// GetClipViewCount returns the total view count for a clip
func (r *AnalyticsRepository) GetClipViewCount(ctx context.Context, clipID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM analytics_events
		WHERE clip_id = $1
		  AND event_type = 'clip_view'
	`

	var count int64
	err := r.db.QueryRow(ctx, query, clipID).Scan(&count)
	return count, err
}

// GetClipShareCount returns the total share count for a clip
func (r *AnalyticsRepository) GetClipShareCount(ctx context.Context, clipID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM analytics_events
		WHERE clip_id = $1
		  AND event_type = 'clip_share'
	`

	var count int64
	err := r.db.QueryRow(ctx, query, clipID).Scan(&count)
	return count, err
}

// GetClipVoteCounts returns the separate upvote and downvote counts for a clip
func (r *AnalyticsRepository) GetClipVoteCounts(ctx context.Context, clipID uuid.UUID) (upvotes int64, downvotes int64, err error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE vote_type = 1) as upvotes,
			COUNT(*) FILTER (WHERE vote_type = -1) as downvotes
		FROM votes
		WHERE clip_id = $1
	`

	err = r.db.QueryRow(ctx, query, clipID).Scan(&upvotes, &downvotes)
	return upvotes, downvotes, err
}

// WatchPartyAnalytics represents aggregated analytics for a watch party
type WatchPartyAnalytics struct {
	PartyID               uuid.UUID `json:"party_id" db:"party_id"`
	UniqueViewers         int       `json:"unique_viewers" db:"unique_viewers"`
	PeakConcurrentViewers int       `json:"peak_concurrent_viewers" db:"peak_concurrent_viewers"`
	AvgDurationSeconds    int       `json:"avg_duration_seconds" db:"avg_watch_duration_seconds"`
	ChatMessages          int       `json:"chat_messages" db:"chat_messages"`
	Reactions             int       `json:"reactions" db:"reactions"`
}

// HostStats represents statistics for a host user
type HostStats struct {
	TotalPartiesHosted int     `json:"total_parties_hosted" db:"total_parties_hosted"`
	TotalViewers       int     `json:"total_viewers" db:"total_viewers"`
	AvgViewersPerParty float64 `json:"avg_viewers_per_party" db:"avg_viewers_per_party"`
	TotalChatMessages  int     `json:"total_chat_messages" db:"total_chat_messages"`
	TotalReactions     int     `json:"total_reactions" db:"total_reactions"`
}

// TrackWatchPartyEvent records a watch party analytics event
func (r *AnalyticsRepository) TrackWatchPartyEvent(ctx context.Context, partyID uuid.UUID, userID *uuid.UUID, eventType string, metadata interface{}) error {
	query := `
		INSERT INTO watch_party_events (party_id, user_id, event_type, metadata, time)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := r.db.Exec(ctx, query, partyID, userID, eventType, metadata)
	if err != nil {
		return fmt.Errorf("failed to track watch party event: %w", err)
	}

	return nil
}

// GetWatchPartyAnalytics retrieves analytics for a specific watch party
func (r *AnalyticsRepository) GetWatchPartyAnalytics(ctx context.Context, partyID uuid.UUID) (*WatchPartyAnalytics, error) {
	// Note: In production, the materialized view should be refreshed periodically
	// via a background job rather than synchronously in the request path.
	// For now, we skip the refresh and use the existing materialized view data.

	query := `
		SELECT
			party_id,
			unique_viewers,
			peak_concurrent_viewers,
			avg_watch_duration_seconds,
			chat_messages,
			reactions
		FROM watch_party_analytics
		WHERE party_id = $1
	`

	var analytics WatchPartyAnalytics
	err := r.db.QueryRow(ctx, query, partyID).Scan(
		&analytics.PartyID,
		&analytics.UniqueViewers,
		&analytics.PeakConcurrentViewers,
		&analytics.AvgDurationSeconds,
		&analytics.ChatMessages,
		&analytics.Reactions,
	)

	if err != nil {
		// If no analytics found yet, return zero values
		if err == pgx.ErrNoRows {
			return &WatchPartyAnalytics{
				PartyID:               partyID,
				UniqueViewers:         0,
				PeakConcurrentViewers: 0,
				AvgDurationSeconds:    0,
				ChatMessages:          0,
				Reactions:             0,
			}, nil
		}
		// Return actual database errors
		return nil, fmt.Errorf("failed to get watch party analytics: %w", err)
	}

	return &analytics, nil
}

// GetHostStats retrieves statistics for a host user
func (r *AnalyticsRepository) GetHostStats(ctx context.Context, userID uuid.UUID) (*HostStats, error) {
	// Note: In production, the materialized view should be refreshed periodically
	// via a background job rather than synchronously in the request path.

	query := `
		SELECT
			COUNT(DISTINCT wpa.party_id) as total_parties_hosted,
			COALESCE(SUM(wpa.unique_viewers), 0) as total_viewers,
			COALESCE(AVG(wpa.unique_viewers), 0) as avg_viewers_per_party,
			COALESCE(SUM(wpa.chat_messages), 0) as total_chat_messages,
			COALESCE(SUM(wpa.reactions), 0) as total_reactions
		FROM watch_party_analytics wpa
		WHERE wpa.host_user_id = $1
	`

	var stats HostStats
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&stats.TotalPartiesHosted,
		&stats.TotalViewers,
		&stats.AvgViewersPerParty,
		&stats.TotalChatMessages,
		&stats.TotalReactions,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get host stats: %w", err)
	}

	return &stats, nil
}

// GetRealtimeViewerCount gets the current viewer count for a party
func (r *AnalyticsRepository) GetRealtimeViewerCount(ctx context.Context, partyID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM watch_party_participants
		WHERE party_id = $1 AND left_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, partyID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get realtime viewer count: %w", err)
	}

	return count, nil
}
