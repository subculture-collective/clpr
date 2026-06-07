package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// SearchRepository handles search operations
type SearchRepository struct {
	db *pgxpool.Pool
}

// NewSearchRepository creates a new SearchRepository
func NewSearchRepository(db *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{db: db}
}

// Search performs a universal search across all types
func (r *SearchRepository) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	response := &models.SearchResponse{
		Query:   req.Query,
		Results: models.SearchResultsByType{},
		Counts:  models.SearchCounts{},
		Meta: models.SearchMeta{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}

	// Parse query into tsquery format
	tsQuery := parseQueryToTSQuery(req.Query)

	// Determine what types to search
	searchAll := req.Type == "" || req.Type == "all"
	searchClips := searchAll || req.Type == "clips"
	searchCreators := searchAll || req.Type == "creators"
	searchGames := searchAll || req.Type == "games"
	searchTags := searchAll || req.Type == "tags"

	// Search clips
	if searchClips {
		clips, count, err := r.searchClips(ctx, tsQuery, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search clips: %w", err)
		}
		response.Results.Clips = clips
		response.Counts.Clips = count
	}

	// Search creators
	if searchCreators {
		creators, count, err := r.searchCreators(ctx, tsQuery, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search creators: %w", err)
		}
		response.Results.Creators = creators
		response.Counts.Creators = count
	}

	// Search games
	if searchGames {
		games, count, err := r.searchGames(ctx, tsQuery, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search games: %w", err)
		}
		response.Results.Games = games
		response.Counts.Games = count
	}

	// Search tags
	if searchTags {
		tags, count, err := r.searchTags(ctx, tsQuery, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search tags: %w", err)
		}
		response.Results.Tags = tags
		response.Counts.Tags = count
	}

	// Calculate total items and pages
	totalItems := response.Counts.Clips + response.Counts.Creators + response.Counts.Games + response.Counts.Tags
	response.Meta.TotalItems = totalItems
	if req.Limit > 0 {
		response.Meta.TotalPages = (totalItems + req.Limit - 1) / req.Limit
	}

	return response, nil
}

// searchClips searches for clips
func (r *SearchRepository) searchClips(ctx context.Context, tsQuery string, req *models.SearchRequest) ([]models.Clip, int, error) {
	// Build WHERE clause with filters
	// Belt-and-suspenders: only search user-submitted clips (scraped clips are in discovery_clips)
	whereClause := "c.is_removed = false AND c.submitted_by_user_id IS NOT NULL"
	args := []interface{}{}
	argPos := 1

	// Full-text search condition
	if tsQuery != "" {
		whereClause += fmt.Sprintf(" AND c.search_vector @@ to_tsquery('english', %s)", utils.SQLPlaceholder(argPos))
		args = append(args, tsQuery)
		argPos++
	}

	// Apply filters
	if req.GameID != nil && *req.GameID != "" {
		whereClause += fmt.Sprintf(" AND c.game_id = %s", utils.SQLPlaceholder(argPos))
		args = append(args, *req.GameID)
		argPos++
	}

	if req.CreatorID != nil && *req.CreatorID != "" {
		whereClause += fmt.Sprintf(" AND c.creator_id = %s", utils.SQLPlaceholder(argPos))
		args = append(args, *req.CreatorID)
		argPos++
	}

	if req.Language != nil && *req.Language != "" {
		whereClause += fmt.Sprintf(" AND c.language = %s", utils.SQLPlaceholder(argPos))
		args = append(args, *req.Language)
		argPos++
	}

	if req.MinVotes != nil {
		whereClause += fmt.Sprintf(" AND c.vote_score >= %s", utils.SQLPlaceholder(argPos))
		args = append(args, *req.MinVotes)
		argPos++
	}

	if req.DateFrom != nil && *req.DateFrom != "" {
		whereClause += fmt.Sprintf(" AND c.created_at >= %s::timestamp", utils.SQLPlaceholder(argPos))
		args = append(args, *req.DateFrom)
		argPos++
	}

	if req.DateTo != nil && *req.DateTo != "" {
		whereClause += fmt.Sprintf(" AND c.created_at <= %s::timestamp", utils.SQLPlaceholder(argPos))
		args = append(args, *req.DateTo)
		argPos++
	}

	// Handle tag filters
	if len(req.Tags) > 0 {
		whereClause += fmt.Sprintf(` AND c.id IN (
			SELECT ct.clip_id
			FROM clip_tags ct
			JOIN tags t ON ct.tag_id = t.id
			WHERE t.slug = ANY(%s)
		)`, utils.SQLPlaceholder(argPos))
		args = append(args, req.Tags)
		argPos++
	}

	// Build ORDER BY clause
	orderBy := "c.created_at DESC" // Default to recent
	if req.Sort == "relevance" && tsQuery != "" {
		orderBy = "ts_rank(c.search_vector, to_tsquery('english', $1)) DESC, c.vote_score DESC, c.created_at DESC"
	} else if req.Sort == "popular" {
		orderBy = "c.vote_score DESC, c.created_at DESC"
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM clips c WHERE %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (req.Page - 1) * req.Limit
	query := fmt.Sprintf(`
		SELECT
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url,
			c.title, c.creator_name, c.creator_id, c.broadcaster_name,
			c.broadcaster_id, c.game_id, c.game_name, c.language,
			c.thumbnail_url, c.duration, c.view_count, c.created_at,
			c.imported_at, c.vote_score, c.comment_count, c.favorite_count,
			c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason
		FROM clips c
		WHERE %s
		ORDER BY %s
		LIMIT %s OFFSET %s
	`, whereClause, orderBy, utils.SQLPlaceholder(argPos), utils.SQLPlaceholder(argPos+1))
	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
		)
		if err != nil {
			return nil, 0, err
		}
		clips = append(clips, clip)
	}

	return clips, totalCount, nil
}

// searchCreators searches for creators (users)
func (r *SearchRepository) searchCreators(ctx context.Context, tsQuery string, req *models.SearchRequest) ([]models.User, int, error) {
	whereClause := "u.is_banned = false"
	args := []interface{}{}
	argPos := 1

	if tsQuery != "" {
		whereClause += fmt.Sprintf(" AND u.search_vector @@ to_tsquery('english', %s)", utils.SQLPlaceholder(argPos))
		args = append(args, tsQuery)
		argPos++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u WHERE %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY
	orderBy := "u.karma_points DESC, u.created_at DESC"
	if req.Sort == "relevance" && tsQuery != "" {
		orderBy = "ts_rank(u.search_vector, to_tsquery('english', $1)) DESC, u.karma_points DESC"
	}

	// Get paginated results
	offset := (req.Page - 1) * req.Limit
	query := fmt.Sprintf(`
		SELECT
			u.id, u.twitch_id, u.username, u.display_name,
			u.email, u.avatar_url, u.bio, u.karma_points,
			u.role, u.is_banned, u.created_at, u.updated_at,
			u.last_login_at
		FROM users u
		WHERE %s
		ORDER BY %s
		LIMIT %s OFFSET %s
	`, whereClause, orderBy, utils.SQLPlaceholder(argPos), utils.SQLPlaceholder(argPos+1))
	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.TwitchID, &user.Username, &user.DisplayName,
			&user.Email, &user.AvatarURL, &user.Bio, &user.KarmaPoints,
			&user.Role, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, totalCount, nil
}

// searchGames searches for games (aggregated from clips)
func (r *SearchRepository) searchGames(ctx context.Context, tsQuery string, req *models.SearchRequest) ([]models.GameSearchResult, int, error) {
	whereClause := "c.game_id IS NOT NULL AND c.game_name IS NOT NULL AND c.is_removed = false AND c.submitted_by_user_id IS NOT NULL"
	args := []interface{}{}
	argPos := 1

	if tsQuery != "" {
		whereClause += fmt.Sprintf(" AND to_tsvector('english', c.game_name) @@ to_tsquery('english', %s)", utils.SQLPlaceholder(argPos))
		args = append(args, tsQuery)
		argPos++
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT c.game_id)
		FROM clips c
		WHERE %s
	`, whereClause)
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (req.Page - 1) * req.Limit
	orderBy := "clip_count DESC"
	if req.Sort == "relevance" && tsQuery != "" {
		orderBy = "clip_count DESC" // Still sort by popularity for games
	}

	query := fmt.Sprintf(`
		SELECT c.game_id, c.game_name, COUNT(*) as clip_count
		FROM clips c
		WHERE %s
		GROUP BY c.game_id, c.game_name
		ORDER BY %s
		LIMIT %s OFFSET %s
	`, whereClause, orderBy, utils.SQLPlaceholder(argPos), utils.SQLPlaceholder(argPos+1))
	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var games []models.GameSearchResult
	for rows.Next() {
		var game models.GameSearchResult
		err := rows.Scan(&game.ID, &game.Name, &game.ClipCount)
		if err != nil {
			return nil, 0, err
		}
		games = append(games, game)
	}

	return games, totalCount, nil
}

// searchTags searches for tags
func (r *SearchRepository) searchTags(ctx context.Context, tsQuery string, req *models.SearchRequest) ([]models.Tag, int, error) {
	whereClause := "1=1"
	args := []interface{}{}
	argPos := 1

	if tsQuery != "" {
		whereClause += fmt.Sprintf(" AND t.search_vector @@ to_tsquery('english', %s)", utils.SQLPlaceholder(argPos))
		args = append(args, tsQuery)
		argPos++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tags t WHERE %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY
	orderBy := "t.usage_count DESC, t.created_at DESC"
	if req.Sort == "relevance" && tsQuery != "" {
		orderBy = "ts_rank(t.search_vector, to_tsquery('english', $1)) DESC, t.usage_count DESC"
	}

	// Get paginated results
	offset := (req.Page - 1) * req.Limit
	query := fmt.Sprintf(`
		SELECT t.id, t.name, t.slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t
		WHERE %s
		ORDER BY %s
		LIMIT %s OFFSET %s
	`, whereClause, orderBy, utils.SQLPlaceholder(argPos), utils.SQLPlaceholder(argPos+1))
	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
			&tag.Color, &tag.UsageCount, &tag.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tags = append(tags, tag)
	}

	return tags, totalCount, nil
}

// GetSuggestions returns search suggestions for autocomplete
func (r *SearchRepository) GetSuggestions(ctx context.Context, query string, limit int) ([]models.SearchSuggestion, error) {
	if query == "" || len(query) < 2 {
		return []models.SearchSuggestion{}, nil
	}

	suggestions := []models.SearchSuggestion{}

	// Search for matching games
	gameQuery := `
		SELECT DISTINCT game_name
		FROM clips
		WHERE game_name ILIKE $1 AND game_name IS NOT NULL AND is_removed = false AND submitted_by_user_id IS NOT NULL
		ORDER BY COUNT(*) OVER (PARTITION BY game_name) DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, gameQuery, "%"+query+"%", limit/2)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var gameName string
			if err := rows.Scan(&gameName); err == nil {
				suggestions = append(suggestions, models.SearchSuggestion{
					Text: gameName,
					Type: "game",
				})
			}
		}
	}

	// Search for matching creators
	creatorQuery := `
		SELECT DISTINCT creator_name
		FROM clips
		WHERE creator_name ILIKE $1 AND is_removed = false AND submitted_by_user_id IS NOT NULL
		ORDER BY COUNT(*) OVER (PARTITION BY creator_name) DESC
		LIMIT $2
	`
	rows2, err := r.db.Query(ctx, creatorQuery, "%"+query+"%", limit/2)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var creatorName string
			if err := rows2.Scan(&creatorName); err == nil {
				suggestions = append(suggestions, models.SearchSuggestion{
					Text: creatorName,
					Type: "creator",
				})
			}
		}
	}

	// Search for matching tags
	tagQuery := `
		SELECT name
		FROM tags
		WHERE name ILIKE $1
		ORDER BY usage_count DESC
		LIMIT $2
	`
	rows3, err := r.db.Query(ctx, tagQuery, "%"+query+"%", limit/3)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var tagName string
			if err := rows3.Scan(&tagName); err == nil {
				suggestions = append(suggestions, models.SearchSuggestion{
					Text: tagName,
					Type: "tag",
				})
			}
		}
	}

	return suggestions, nil
}

// TrackSearch records a search query for analytics
func (r *SearchRepository) TrackSearch(ctx context.Context, userID *uuid.UUID, query string, resultCount int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO search_queries (user_id, query, result_count)
		VALUES ($1, $2, $3)
	`, userID, query, resultCount)
	return err
}

// GetTrendingSearches returns the most popular search queries in a given time period
func (r *SearchRepository) GetTrendingSearches(ctx context.Context, days int, limit int) ([]models.TrendingSearch, error) {
	if days <= 0 {
		days = 7
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT
			query,
			COUNT(*) as search_count,
			COUNT(DISTINCT user_id) as unique_users,
			AVG(result_count)::int as avg_results
		FROM search_queries
		WHERE created_at >= NOW() - $1 * INTERVAL '1 day'
			AND query != ''
		GROUP BY query
		ORDER BY search_count DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, days, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending searches: %w", err)
	}
	defer rows.Close()

	var searches []models.TrendingSearch
	for rows.Next() {
		var search models.TrendingSearch
		if err := rows.Scan(&search.Query, &search.SearchCount, &search.UniqueUsers, &search.AvgResults); err != nil {
			return nil, fmt.Errorf("failed to scan trending search: %w", err)
		}
		searches = append(searches, search)
	}

	return searches, nil
}

// GetFailedSearches returns searches that returned no results
func (r *SearchRepository) GetFailedSearches(ctx context.Context, days int, limit int) ([]models.FailedSearch, error) {
	if days <= 0 {
		days = 7
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT
			query,
			COUNT(*) as search_count,
			MAX(created_at) as last_searched
		FROM search_queries
		WHERE created_at >= NOW() - $1 * INTERVAL '1 day'
			AND result_count = 0
			AND query != ''
		GROUP BY query
		ORDER BY search_count DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, days, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed searches: %w", err)
	}
	defer rows.Close()

	var searches []models.FailedSearch
	for rows.Next() {
		var search models.FailedSearch
		if err := rows.Scan(&search.Query, &search.SearchCount, &search.LastSearched); err != nil {
			return nil, fmt.Errorf("failed to scan failed search: %w", err)
		}
		searches = append(searches, search)
	}

	return searches, nil
}

// GetUserSearchHistory returns a user's recent search queries
func (r *SearchRepository) GetUserSearchHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.SearchHistoryItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT
			query,
			result_count,
			created_at
		FROM search_queries
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user search history: %w", err)
	}
	defer rows.Close()

	var history []models.SearchHistoryItem
	for rows.Next() {
		var item models.SearchHistoryItem
		if err := rows.Scan(&item.Query, &item.ResultCount, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan search history item: %w", err)
		}
		history = append(history, item)
	}

	return history, nil
}

// GetSearchAnalyticsSummary returns overall search analytics
func (r *SearchRepository) GetSearchAnalyticsSummary(ctx context.Context, days int) (*models.SearchAnalyticsSummary, error) {
	if days <= 0 {
		days = 7
	}

	query := `
		SELECT
			COUNT(*) as total_searches,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(CASE WHEN result_count = 0 THEN 1 END) as failed_searches,
			AVG(result_count)::int as avg_results_per_search
		FROM search_queries
		WHERE created_at >= NOW() - $1 * INTERVAL '1 day'
	`

	var summary models.SearchAnalyticsSummary
	err := r.db.QueryRow(ctx, query, days).Scan(
		&summary.TotalSearches,
		&summary.UniqueUsers,
		&summary.FailedSearches,
		&summary.AvgResultsPerSearch,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get search analytics summary: %w", err)
	}

	// Calculate success rate
	if summary.TotalSearches > 0 {
		summary.SuccessRate = float64(summary.TotalSearches-summary.FailedSearches) / float64(summary.TotalSearches) * 100
	}

	return &summary, nil
}

// parseQueryToTSQuery converts a search query to PostgreSQL tsquery format
func parseQueryToTSQuery(query string) string {
	if query == "" {
		return ""
	}

	// Simple parsing: split by spaces and join with & (AND)
	words := strings.Fields(query)

	// Remove special characters and empty strings
	var cleanWords []string
	for _, word := range words {
		// Basic sanitization
		cleaned := strings.TrimSpace(word)
		if cleaned != "" && len(cleaned) > 0 {
			// Add prefix matching for partial words
			cleanWords = append(cleanWords, cleaned+":*")
		}
	}

	if len(cleanWords) == 0 {
		return ""
	}

	return strings.Join(cleanWords, " & ")
}
