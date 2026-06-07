package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	internalutils "git.subcult.tv/subculture-collective/clpr/internal/utils"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	justChattingGameID       = "509658"
	justChattingPerGameLimit = 50
	defaultPerGameLimit      = 5
	maxTrendingGames         = 10
	defaultTrendingMaxPages  = 3
)

// DefaultTrendingPageWindow exposes the default page rotation window for schedulers
const DefaultTrendingPageWindow = defaultTrendingMaxPages

var defaultTrendingGameIDs = []string{
	justChattingGameID, // Just Chatting
	"32982",            // Grand Theft Auto V
	"33214",            // Fortnite
	"516575",           // Valorant
	"21779",            // League of Legends
	"27471",            // Minecraft
	"512710",           // Call of Duty: Warzone
	"511224",           // Apex Legends
	"29595",            // Dota 2
	"488552",           // Overwatch 2
}

// SyncClipsByGameOptions controls pagination behaviour for game syncs
type SyncClipsByGameOptions struct {
	InitialCursor  string
	SinglePage     bool
	PageIndex      int
	LanguageFilter string
}

// TrendingGameConfig pairs a game with its per-run fetch limit
type TrendingGameConfig struct {
	GameID string
	Limit  int
}

// TrendingSyncOptions configures a trending sync run
type TrendingSyncOptions struct {
	Games                []TrendingGameConfig
	StateStore           TrendingStateStore
	MaxPages             int
	ForceResetPagination bool
	LanguageFilter       string
}

// ClipSyncService handles fetching and syncing clips from Twitch
type ClipSyncService struct {
	twitchClient *twitch.Client
	clipRepo     *repository.ClipRepository
	tagRepo      *repository.TagRepository
	userRepo     *repository.UserRepository
	stateStore   TrendingStateStore
	maxPages     int
	defaultLang  string
}

// NewClipSyncService creates a new ClipSyncService
func NewClipSyncService(twitchClient *twitch.Client, clipRepo *repository.ClipRepository, tagRepo *repository.TagRepository, userRepo *repository.UserRepository, redisClient *redispkg.Client) *ClipSyncService {
	var stateStore TrendingStateStore
	if redisClient != nil {
		stateStore = NewRedisTrendingStateStore(redisClient)
	}

	return &ClipSyncService{
		twitchClient: twitchClient,
		clipRepo:     clipRepo,
		tagRepo:      tagRepo,
		userRepo:     userRepo,
		stateStore:   stateStore,
		maxPages:     defaultTrendingMaxPages,
		defaultLang:  normalizeLanguageFilter("en"),
	}
}

// SetDefaultLanguage overrides the service-level language filter (use "all" or "" to disable)
func (s *ClipSyncService) SetDefaultLanguage(lang string) {
	s.defaultLang = normalizeLanguageFilter(lang)
}

// SyncStats contains statistics about a sync operation
type SyncStats struct {
	ClipsFetched int
	ClipsCreated int
	ClipsUpdated int
	ClipsSkipped int
	Errors       []string
	StartTime    time.Time
	EndTime      time.Time
}

// SyncClipsByGame fetches and syncs clips for a specific game
func (s *ClipSyncService) SyncClipsByGame(ctx context.Context, gameID string, hours int, limit int, opts *SyncClipsByGameOptions) (*SyncStats, string, error) {
	stats := &SyncStats{StartTime: time.Now()}
	languageFilter := s.defaultLang

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	params := &twitch.ClipParams{
		GameID:    gameID,
		StartedAt: startTime,
		EndedAt:   endTime,
		First:     internalutils.Min(limit, 100), // Twitch API max is 100
	}

	pageIndex := 1
	if opts != nil {
		if opts.InitialCursor != "" {
			params.After = opts.InitialCursor
		}
		if opts.PageIndex > 0 {
			pageIndex = opts.PageIndex
		}
		if opts.LanguageFilter != "" {
			languageFilter = opts.LanguageFilter
		}
	}

	languageFilter = normalizeLanguageFilter(languageFilter)

	utils.Info("Syncing clips for game", map[string]interface{}{
		"game_id":    gameID,
		"start_time": startTime,
		"end_time":   endTime,
		"page":       pageIndex,
		"cursor":     params.After,
		"limit":      params.First,
	})

	totalFetched := 0
	var nextCursor string
	var fetchErr error

	for totalFetched < limit {
		clipsResp, err := s.twitchClient.GetClips(ctx, params)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Game category %s not found (404) - may have been removed or merged", gameID))
				utils.Warn("Game category returned 404 - skipping", map[string]interface{}{"game_id": gameID})
			} else {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to fetch clips: %v", err))
			}
			fetchErr = err
			break
		}

		if len(clipsResp.Data) == 0 {
			break
		}

		var channelTags map[string][]string
		if s.tagRepo != nil {
			ids := make([]string, 0, len(clipsResp.Data))
			seenIDs := map[string]bool{}
			for _, clip := range clipsResp.Data {
				if clip.BroadcasterID != "" && !seenIDs[clip.BroadcasterID] {
					seenIDs[clip.BroadcasterID] = true
					ids = append(ids, clip.BroadcasterID)
				}
			}
			channelTags = s.fetchChannelTags(ctx, ids)
		}

		for _, twitchClip := range clipsResp.Data {
			if !languageMatches(twitchClip.Language, languageFilter) {
				stats.ClipsSkipped++
				continue
			}

			if err := s.processClip(ctx, &twitchClip, stats, channelTags[twitchClip.BroadcasterID]); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to process clip %s: %v", twitchClip.ID, err))
			}
			totalFetched++
			if totalFetched >= limit {
				break
			}
		}

		nextCursor = clipsResp.Pagination.Cursor
		if opts != nil && opts.SinglePage {
			break
		}
		if nextCursor == "" || totalFetched >= limit {
			break
		}

		params.After = nextCursor
	}

	stats.ClipsFetched = totalFetched
	stats.EndTime = time.Now()

	utils.Info("Sync completed", map[string]interface{}{
		"fetched":     stats.ClipsFetched,
		"created":     stats.ClipsCreated,
		"updated":     stats.ClipsUpdated,
		"skipped":     stats.ClipsSkipped,
		"errors":      len(stats.Errors),
		"duration":    stats.EndTime.Sub(stats.StartTime),
		"next_cursor": nextCursor,
	})

	return stats, nextCursor, fetchErr
}

// SyncClipsByBroadcaster fetches and syncs clips for a specific broadcaster
type SyncClipsByBroadcasterOptions struct {
	LanguageFilter string
}

func (s *ClipSyncService) SyncClipsByBroadcaster(ctx context.Context, broadcasterID string, hours int, limit int, opts *SyncClipsByBroadcasterOptions) (*SyncStats, error) {
	stats := &SyncStats{
		StartTime: time.Now(),
	}
	languageFilter := normalizeLanguageFilter(s.defaultLang)
	if opts != nil && opts.LanguageFilter != "" {
		languageFilter = normalizeLanguageFilter(opts.LanguageFilter)
	}

	// Calculate time range
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	params := &twitch.ClipParams{
		BroadcasterID: broadcasterID,
		StartedAt:     startTime,
		EndedAt:       endTime,
		First:         internalutils.Min(limit, 100),
	}

	utils.Info("Syncing clips for broadcaster", map[string]interface{}{
		"broadcaster_id": broadcasterID,
		"start_time":     startTime,
		"end_time":       endTime,
	})

	// Fetch clips with pagination
	totalFetched := 0
	for totalFetched < limit {
		clipsResp, err := s.twitchClient.GetClips(ctx, params)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to fetch clips: %v", err))
			break
		}

		if len(clipsResp.Data) == 0 {
			break
		}

		var channelTags map[string][]string
		if s.tagRepo != nil {
			channelTags = s.fetchChannelTags(ctx, []string{broadcasterID})
		}

		// Process each clip
		for _, twitchClip := range clipsResp.Data {
			if !languageMatches(twitchClip.Language, languageFilter) {
				stats.ClipsSkipped++
				continue
			}

			if err := s.processClip(ctx, &twitchClip, stats, channelTags[twitchClip.BroadcasterID]); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to process clip %s: %v", twitchClip.ID, err))
			}
			totalFetched++
			if totalFetched >= limit {
				break
			}
		}

		// Check if there are more pages
		if clipsResp.Pagination.Cursor == "" || totalFetched >= limit {
			break
		}

		params.After = clipsResp.Pagination.Cursor
	}

	stats.ClipsFetched = totalFetched
	stats.EndTime = time.Now()

	utils.Info("Sync completed", map[string]interface{}{
		"fetched":  stats.ClipsFetched,
		"created":  stats.ClipsCreated,
		"updated":  stats.ClipsUpdated,
		"skipped":  stats.ClipsSkipped,
		"errors":   len(stats.Errors),
		"duration": stats.EndTime.Sub(stats.StartTime),
	})

	return stats, nil
}

// SyncTrendingClips fetches trending clips from multiple top games with pagination rotation
func (s *ClipSyncService) SyncTrendingClips(ctx context.Context, hours int, opts *TrendingSyncOptions) (*SyncStats, error) {
	stats := &SyncStats{StartTime: time.Now()}
	resolved := s.applyTrendingDefaults(opts)

	games := append([]TrendingGameConfig(nil), resolved.Games...)
	if len(games) == 0 {
		resolvedGames, resolveErr := s.resolveTrendingGames(ctx, resolved.StateStore)
		if resolveErr != nil {
			stats.Errors = append(stats.Errors, resolveErr.Error())
		}
		games = resolvedGames
	}

	if len(games) == 0 {
		games = buildTrendingGameConfigs(defaultTrendingGameIDs)
	}

	utils.Info("Syncing trending clips", map[string]interface{}{
		"games":       len(games),
		"max_pages":   resolved.MaxPages,
		"force_reset": resolved.ForceResetPagination,
	})

	for _, game := range games {
		var cursorState *TrendingCursorState
		if resolved.StateStore != nil && !resolved.ForceResetPagination {
			state, err := resolved.StateStore.GetCursor(ctx, game.GameID)
			if err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to load cursor for game %s: %v", game.GameID, err))
			} else {
				cursorState = state
			}
		}

		pageIndex := 1
		initialCursor := ""
		if cursorState != nil {
			initialCursor = cursorState.Cursor
			if cursorState.Page > 0 {
				pageIndex = cursorState.Page
			}
		}

		gameStats, nextCursor, err := s.SyncClipsByGame(ctx, game.GameID, hours, game.Limit, &SyncClipsByGameOptions{
			InitialCursor:  initialCursor,
			SinglePage:     true,
			PageIndex:      pageIndex,
			LanguageFilter: resolved.LanguageFilter,
		})
		if err != nil {
			utils.Warn("Failed to sync game", map[string]interface{}{"game_id": game.GameID, "error": err})
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to sync game %s: %v", game.GameID, err))
		}

		stats.ClipsFetched += gameStats.ClipsFetched
		stats.ClipsCreated += gameStats.ClipsCreated
		stats.ClipsUpdated += gameStats.ClipsUpdated
		stats.ClipsSkipped += gameStats.ClipsSkipped
		stats.Errors = append(stats.Errors, gameStats.Errors...)

		// For admin-triggered runs we force page 1 and leave stored cursors untouched
		if resolved.ForceResetPagination || resolved.StateStore == nil {
			utils.Info("Trending game sync", map[string]interface{}{
				"game_id":    game.GameID,
				"page":       pageIndex,
				"cursor_in":  initialCursor,
				"cursor_out": nextCursor,
				"reset":      true,
				"fetched":    gameStats.ClipsFetched,
			})
			continue
		}

		nextPage := pageIndex + 1
		shouldReset := nextCursor == "" || (resolved.MaxPages > 0 && nextPage > resolved.MaxPages)
		if shouldReset {
			if err := resolved.StateStore.ClearCursor(ctx, game.GameID); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to clear cursor for game %s: %v", game.GameID, err))
			}
		} else {
			if err := resolved.StateStore.SaveCursor(ctx, game.GameID, &TrendingCursorState{Cursor: nextCursor, Page: nextPage}); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to persist cursor for game %s: %v", game.GameID, err))
			}
		}

		utils.Info("Trending game sync", map[string]interface{}{
			"game_id":    game.GameID,
			"page":       pageIndex,
			"next_page":  nextPage,
			"cursor_in":  initialCursor,
			"cursor_out": nextCursor,
			"reset":      shouldReset,
			"fetched":    gameStats.ClipsFetched,
		})
	}

	stats.EndTime = time.Now()

	utils.Info("Trending sync completed", map[string]interface{}{
		"fetched":  stats.ClipsFetched,
		"created":  stats.ClipsCreated,
		"updated":  stats.ClipsUpdated,
		"skipped":  stats.ClipsSkipped,
		"errors":   len(stats.Errors),
		"duration": stats.EndTime.Sub(stats.StartTime),
	})

	return stats, nil
}

func (s *ClipSyncService) applyTrendingDefaults(opts *TrendingSyncOptions) *TrendingSyncOptions {
	resolved := &TrendingSyncOptions{
		MaxPages:       defaultTrendingMaxPages,
		StateStore:     s.stateStore,
		LanguageFilter: normalizeLanguageFilter(s.defaultLang),
	}

	if s.maxPages > 0 {
		resolved.MaxPages = s.maxPages
	}

	if opts != nil {
		resolved.ForceResetPagination = opts.ForceResetPagination
		if opts.MaxPages > 0 {
			resolved.MaxPages = opts.MaxPages
		}
		if opts.StateStore != nil {
			resolved.StateStore = opts.StateStore
		}
		resolved.Games = append(resolved.Games, opts.Games...)
		if opts.LanguageFilter != "" {
			resolved.LanguageFilter = opts.LanguageFilter
		}
	}

	if resolved.MaxPages == 0 {
		resolved.MaxPages = defaultTrendingMaxPages
	}

	return resolved
}

func (s *ClipSyncService) resolveTrendingGames(ctx context.Context, store TrendingStateStore) ([]TrendingGameConfig, error) {
	if s.twitchClient != nil {
		topGamesResp, err := s.twitchClient.GetTopGames(ctx, maxTrendingGames, "")
		if err == nil && topGamesResp != nil && len(topGamesResp.Data) > 0 {
			ids := ensureTrendingGameIDs(topGamesResp.Data)
			configs := buildTrendingGameConfigs(ids)
			if store != nil {
				if err := store.SaveGameIDs(ctx, ids); err != nil {
					utils.Warn("Failed to cache trending game IDs", map[string]interface{}{"error": err})
				}
			}
			return configs, nil
		}
		if err != nil {
			utils.Warn("Failed to fetch top games", map[string]interface{}{"error": err})
		}
	}

	if store != nil {
		cached, err := store.LoadGameIDs(ctx)
		if err != nil {
			utils.Warn("Failed to load cached top games", map[string]interface{}{"error": err})
		} else if len(cached) > 0 {
			return buildTrendingGameConfigs(ensureJustChattingGameIDs(cached)), nil
		}
	}

	return buildTrendingGameConfigs(defaultTrendingGameIDs), fmt.Errorf("using fallback trending game list")
}

func ensureTrendingGameIDs(games []twitch.Game) []string {
	ids := make([]string, 0, len(games))
	for _, g := range games {
		if g.ID != "" {
			ids = append(ids, g.ID)
		}
	}
	return ensureJustChattingGameIDs(ids)
}

func ensureJustChattingGameIDs(ids []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, maxTrendingGames)

	// Ensure Just Chatting is always present and first
	seen[justChattingGameID] = true
	result = append(result, justChattingGameID)

	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
		if len(result) >= maxTrendingGames {
			break
		}
	}

	if len(result) > maxTrendingGames {
		return result[:maxTrendingGames]
	}

	return result
}

func buildTrendingGameConfigs(gameIDs []string) []TrendingGameConfig {
	ids := ensureJustChattingGameIDs(gameIDs)
	if len(ids) > maxTrendingGames {
		ids = ids[:maxTrendingGames]
	}

	configs := make([]TrendingGameConfig, 0, len(ids))
	for _, id := range ids {
		configs = append(configs, TrendingGameConfig{
			GameID: id,
			Limit:  perGameLimit(id),
		})
	}

	return configs
}

func perGameLimit(gameID string) int {
	if gameID == justChattingGameID {
		return justChattingPerGameLimit
	}
	return defaultPerGameLimit
}

// FetchClipByURL fetches a single clip by its Twitch URL or ID
func (s *ClipSyncService) FetchClipByURL(ctx context.Context, clipURLOrID string) (*models.Clip, error) {
	// Extract clip ID from URL if needed
	clipID := ExtractClipID(clipURLOrID)
	if clipID == "" {
		return nil, fmt.Errorf("invalid clip URL or ID: %s", clipURLOrID)
	}

	// Fetch from Twitch
	params := &twitch.ClipParams{
		ClipIDs: []string{clipID},
	}

	clipsResp, err := s.twitchClient.GetClips(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clip from Twitch: %w", err)
	}

	if len(clipsResp.Data) == 0 {
		return nil, fmt.Errorf("clip not found: %s", clipID)
	}

	twitchClip := clipsResp.Data[0]

	// Check if already exists
	exists, err := s.clipRepo.ExistsByTwitchClipID(ctx, twitchClip.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check clip existence: %w", err)
	}

	if exists {
		// Return existing clip
		return s.clipRepo.GetByTwitchClipID(ctx, twitchClip.ID)
	}

	// Transform and save
	clip := transformTwitchClip(&twitchClip)
	if err := s.clipRepo.Create(ctx, clip); err != nil {
		return nil, fmt.Errorf("failed to save clip: %w", err)
	}

	if s.tagRepo != nil && twitchClip.BroadcasterID != "" {
		if tags := s.fetchChannelTags(ctx, []string{twitchClip.BroadcasterID}); len(tags) > 0 {
			_ = s.applyStreamerTags(ctx, clip, tags[twitchClip.BroadcasterID])
		}
	}

	return clip, nil
}

// processClip processes a single clip from Twitch (create or update)
func (s *ClipSyncService) processClip(ctx context.Context, twitchClip *twitch.Clip, stats *SyncStats, streamerTags []string) error {
	// Check if clip already exists
	exists, err := s.clipRepo.ExistsByTwitchClipID(ctx, twitchClip.ID)
	if err != nil {
		return fmt.Errorf("failed to check clip existence: %w", err)
	}

	if exists {
		// Update view count for existing clip
		if err := s.clipRepo.UpdateViewCount(ctx, twitchClip.ID, twitchClip.ViewCount); err != nil {
			return fmt.Errorf("failed to update view count: %w", err)
		}
		stats.ClipsUpdated++
		return nil
	}

	// Transform Twitch clip to our model
	clip := transformTwitchClip(twitchClip)

	// Ensure unclaimed users exist for creator and broadcaster
	if err := s.ensureUnclaimedUser(ctx, twitchClip.CreatorID, twitchClip.CreatorName); err != nil {
		utils.Warn("Failed to ensure unclaimed user for creator", map[string]interface{}{"creator": twitchClip.CreatorName, "error": err})
	}
	if err := s.ensureUnclaimedUser(ctx, twitchClip.BroadcasterID, twitchClip.BroadcasterName); err != nil {
		utils.Warn("Failed to ensure unclaimed user for broadcaster", map[string]interface{}{"broadcaster": twitchClip.BroadcasterName, "error": err})
	}

	// Save to database
	if err := s.clipRepo.Create(ctx, clip); err != nil {
		return fmt.Errorf("failed to create clip: %w", err)
	}

	if len(streamerTags) > 0 && s.tagRepo != nil {
		if err := s.applyStreamerTags(ctx, clip, streamerTags); err != nil {
			stats.Errors = append(stats.Errors, err.Error())
		}
	}

	stats.ClipsCreated++
	return nil
}

// processClipAsPosted imports a Twitch clip and marks it as "posted" by the given submitter.
// If the clip already exists and is unclaimed, it claims it for the submitter.
// If the clip already exists and is already posted, it just updates the view count.
func (s *ClipSyncService) processClipAsPosted(ctx context.Context, twitchClip *twitch.Clip, stats *SyncStats, streamerTags []string, submitterID uuid.UUID) error {
	// Check if clip already exists
	existing, err := s.clipRepo.GetByTwitchClipID(ctx, twitchClip.ID)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return fmt.Errorf("failed to check clip existence: %w", err)
	}

	if existing != nil {
		// Update view count
		if err := s.clipRepo.UpdateViewCount(ctx, twitchClip.ID, twitchClip.ViewCount); err != nil {
			return fmt.Errorf("failed to update view count: %w", err)
		}

		// If unclaimed, claim it for the bot
		if existing.SubmittedByUserID == nil {
			if err := s.clipRepo.ClaimScrapedClip(ctx, existing.ID, submitterID, nil, false, nil, time.Now()); err != nil {
				utils.Warn("Failed to claim scraped clip for bot", map[string]interface{}{
					"clip_id":    existing.ID.String(),
					"twitch_id":  twitchClip.ID,
					"error":      err,
				})
			}
		}

		stats.ClipsUpdated++
		return nil
	}

	// New clip — create with submitter attribution
	clip := transformTwitchClip(twitchClip)
	clip.SubmittedByUserID = &submitterID
	now := time.Now()
	clip.SubmittedAt = &now

	// Ensure unclaimed users exist for creator and broadcaster
	if err := s.ensureUnclaimedUser(ctx, twitchClip.CreatorID, twitchClip.CreatorName); err != nil {
		utils.Warn("Failed to ensure unclaimed user for creator", map[string]interface{}{"creator": twitchClip.CreatorName, "error": err})
	}
	if err := s.ensureUnclaimedUser(ctx, twitchClip.BroadcasterID, twitchClip.BroadcasterName); err != nil {
		utils.Warn("Failed to ensure unclaimed user for broadcaster", map[string]interface{}{"broadcaster": twitchClip.BroadcasterName, "error": err})
	}

	if err := s.clipRepo.Create(ctx, clip); err != nil {
		return fmt.Errorf("failed to create clip: %w", err)
	}

	if len(streamerTags) > 0 && s.tagRepo != nil {
		if err := s.applyStreamerTags(ctx, clip, streamerTags); err != nil {
			stats.Errors = append(stats.Errors, err.Error())
		}
	}

	stats.ClipsCreated++
	return nil
}

// transformTwitchClip converts a Twitch API clip to our database model
func transformTwitchClip(twitchClip *twitch.Clip) *models.Clip {
	return &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    twitchClip.ID,
		TwitchClipURL:   twitchClip.URL,
		EmbedURL:        twitchClip.EmbedURL,
		Title:           twitchClip.Title,
		CreatorName:     twitchClip.CreatorName,
		CreatorID:       internalutils.StringPtr(twitchClip.CreatorID),
		BroadcasterName: twitchClip.BroadcasterName,
		BroadcasterID:   internalutils.StringPtr(twitchClip.BroadcasterID),
		GameID:          internalutils.StringPtr(twitchClip.GameID),
		GameName:        nil, // Will be enriched separately if needed
		Language:        internalutils.StringPtr(twitchClip.Language),
		ThumbnailURL:    internalutils.StringPtr(twitchClip.ThumbnailURL),
		Duration:        internalutils.Float64Ptr(twitchClip.Duration),
		ViewCount:       twitchClip.ViewCount,
		CreatedAt:       twitchClip.CreatedAt,
		ImportedAt:      time.Now(),
		VoteScore:       0,
		CommentCount:    0,
		FavoriteCount:   0,
		IsFeatured:      false,
		IsNSFW:          false,
		IsRemoved:       false,
		RemovedReason:   nil,
	}
}

// ensureUnclaimedUser creates an unclaimed user account if one doesn't exist for the given Twitch ID
func (s *ClipSyncService) ensureUnclaimedUser(ctx context.Context, twitchID, displayName string) error {
	if s.userRepo == nil || twitchID == "" {
		return nil
	}

	// Check if user already exists with this Twitch ID
	_, err := s.userRepo.GetByTwitchID(ctx, twitchID)
	if err == nil {
		// User already exists, nothing to do
		return nil
	}

	// If error is not "not found", return it
	if err != repository.ErrUserNotFound {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	// Create unclaimed user account
	username := generateUnclaimedUsername(displayName, twitchID)
	unclaimedUser := &models.User{
		ID:            uuid.New(),
		TwitchID:      &twitchID,
		Username:      username,
		DisplayName:   displayName,
		Role:          "user",
		AccountType:   models.AccountTypeMember,
		AccountStatus: "unclaimed",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, unclaimedUser); err != nil {
		// Ignore duplicate errors (race condition)
		if err == repository.ErrUserAlreadyExists {
			return nil
		}
		return fmt.Errorf("failed to create unclaimed user: %w", err)
	}

	utils.Info("Created unclaimed user account", map[string]interface{}{"display_name": displayName, "twitch_id": twitchID})
	return nil
}

// generateUnclaimedUsername creates a username for unclaimed accounts
func generateUnclaimedUsername(displayName, twitchID string) string {
	// Normalize the display name to create a username
	username := strings.ToLower(displayName)
	username = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(username, "_")
	username = regexp.MustCompile(`_+`).ReplaceAllString(username, "_")
	username = strings.Trim(username, "_")

	// If username is empty or too short, use a prefix
	if len(username) < 3 {
		username = "user_" + twitchID[:8]
	}

	// Truncate if too long (max 50 chars from schema)
	if len(username) > 50 {
		username = username[:50]
	}

	return username
}

func (s *ClipSyncService) fetchChannelTags(ctx context.Context, broadcasterIDs []string) map[string][]string {
	result := make(map[string][]string, len(broadcasterIDs))
	if s.twitchClient == nil || len(broadcasterIDs) == 0 {
		return result
	}

	resp, err := s.twitchClient.GetChannels(ctx, broadcasterIDs)
	if err != nil || resp == nil {
		return result
	}

	for _, ch := range resp.Data {
		if len(ch.Tags) > 0 {
			result[ch.BroadcasterID] = append([]string{}, ch.Tags...)
		}
	}

	return result
}

func (s *ClipSyncService) applyStreamerTags(ctx context.Context, clip *models.Clip, tags []string) error {
	if s.tagRepo == nil || len(tags) == 0 || clip == nil {
		return nil
	}

	seen := make(map[string]bool, len(tags))
	var lastErr error

	for _, raw := range tags {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		slug := utils.Slugify(name)
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true

		tag, err := s.tagRepo.GetOrCreateTag(ctx, name, slug, nil)
		if err != nil {
			lastErr = err
			continue
		}

		if err := s.tagRepo.AddTagToClip(ctx, clip.ID, tag.ID); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("failed to apply streamer tags: %w", lastErr)
	}

	return nil
}

func normalizeLanguageFilter(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" || lang == "*" || lang == "all" {
		return ""
	}
	return lang
}

func languageMatches(clipLang, filter string) bool {
	filter = normalizeLanguageFilter(filter)
	if filter == "" {
		return true
	}
	clipLang = strings.ToLower(strings.TrimSpace(clipLang))
	if clipLang == "" {
		return false
	}
	return clipLang == filter || strings.HasPrefix(clipLang, filter+"-")
}


// ExtractClipID extracts the clip ID from a Twitch clip URL or returns the ID if already provided
func ExtractClipID(clipURLOrID string) string {
	// If it's already just an ID (alphanumeric), return it
	if !strings.Contains(clipURLOrID, "/") && !strings.Contains(clipURLOrID, "twitch.tv") {
		return clipURLOrID
	}

	// Handle full URLs: https://www.twitch.tv/username/clip/ClipIDHere
	// or https://clips.twitch.tv/ClipIDHere
	parts := strings.Split(clipURLOrID, "/")
	if len(parts) > 0 {
		// Get the last non-empty part
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" && parts[i] != "clip" {
				// Remove query parameters if present
				clipID := strings.Split(parts[i], "?")[0]
				return clipID
			}
		}
	}

	return ""
}

// Helper functions
func float64Ptr(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FetchAndImportClips fetches clips from Twitch using the given parameters,
// imports any new ones into the local database as "posted" by the submitter,
// and returns the local clip models ordered by Twitch view count descending.
// New clips get submitted_by_user_id set; unclaimed existing clips get claimed.
// This is the primary entry point for Twitch-powered playlist curation strategies.
func (s *ClipSyncService) FetchAndImportClips(ctx context.Context, params *twitch.ClipParams, limit int, langFilter string, submitterID uuid.UUID) ([]models.Clip, error) {
	if langFilter == "" {
		langFilter = s.defaultLang
	}
	langFilter = normalizeLanguageFilter(langFilter)

	// Cap per-request to Twitch API limit
	if params.First == 0 || params.First > 100 {
		params.First = internalutils.Min(limit, 100)
	}

	var twitchClipIDs []string
	totalFetched := 0

	for totalFetched < limit {
		clipsResp, err := s.twitchClient.GetClips(ctx, params)
		if err != nil {
			if len(twitchClipIDs) > 0 {
				// Partial success — return what we have
				utils.Warn("Twitch API error during fetch, returning partial results", map[string]interface{}{
					"error":   err,
					"fetched": totalFetched,
				})
				break
			}
			return nil, fmt.Errorf("failed to fetch clips from Twitch: %w", err)
		}

		if len(clipsResp.Data) == 0 {
			break
		}

		stats := &SyncStats{StartTime: time.Now()}

		// Batch fetch channel tags for this page
		var channelTags map[string][]string
		if s.tagRepo != nil {
			ids := make([]string, 0, len(clipsResp.Data))
			seenIDs := map[string]bool{}
			for _, clip := range clipsResp.Data {
				if clip.BroadcasterID != "" && !seenIDs[clip.BroadcasterID] {
					seenIDs[clip.BroadcasterID] = true
					ids = append(ids, clip.BroadcasterID)
				}
			}
			channelTags = s.fetchChannelTags(ctx, ids)
		}

		for _, twitchClip := range clipsResp.Data {
			if !languageMatches(twitchClip.Language, langFilter) {
				continue
			}

			if err := s.processClipAsPosted(ctx, &twitchClip, stats, channelTags[twitchClip.BroadcasterID], submitterID); err != nil {
				utils.Warn("Failed to process clip during fetch-and-import", map[string]interface{}{
					"clip_id": twitchClip.ID,
					"error":   err,
				})
				continue
			}

			twitchClipIDs = append(twitchClipIDs, twitchClip.ID)
			totalFetched++
			if totalFetched >= limit {
				break
			}
		}

		if clipsResp.Pagination.Cursor == "" || totalFetched >= limit {
			break
		}
		params.After = clipsResp.Pagination.Cursor
	}

	if len(twitchClipIDs) == 0 {
		return nil, nil
	}

	// Resolve local clip models for the fetched Twitch clip IDs
	clips, err := s.clipRepo.GetByTwitchClipIDs(ctx, twitchClipIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local clips: %w", err)
	}

	return clips, nil
}

// GetTopGames returns the IDs of the top games currently on Twitch.
func (s *ClipSyncService) GetTopGames(ctx context.Context, count int) ([]string, error) {
	resp, err := s.twitchClient.GetTopGames(ctx, count, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top games: %w", err)
	}

	gameIDs := make([]string, 0, len(resp.Data))
	for _, g := range resp.Data {
		gameIDs = append(gameIDs, g.ID)
	}
	return gameIDs, nil
}

// TimeframeToHours converts a timeframe string to hours. Returns 0 for all-time.
func TimeframeToHours(timeframe *string) int {
	if timeframe == nil {
		return 0
	}
	switch *timeframe {
	case "hour":
		return 1
	case "day":
		return 24
	case "week":
		return 168
	case "month":
		return 720
	case "year":
		return 8760
	default:
		return 0
	}
}
