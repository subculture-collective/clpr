package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

// executeStrategy runs the appropriate curation strategy and returns matching clips.
func (s *PlaylistScriptService) executeStrategy(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	// Twitch-powered strategies need the clip sync service
	if strings.HasPrefix(script.Strategy, "twitch_") {
		return s.executeTwitchStrategy(ctx, script)
	}

	// Database-only strategies need the curation repository
	if s.curationRepo == nil {
		return nil, fmt.Errorf("curation repository not configured")
	}

	switch script.Strategy {
	case "sleeper_hits":
		return s.curationRepo.SleeperHits(ctx, script)

	case "viral_velocity":
		return s.curationRepo.ViralVelocity(ctx, script)

	case "community_favorites":
		return s.curationRepo.CommunityFavorites(ctx, script)

	case "deep_cuts":
		return s.curationRepo.DeepCuts(ctx, script)

	case "fresh_faces":
		return s.curationRepo.FreshFaces(ctx, script)

	case "one_per_creator":
		return s.curationRepo.OnePerCreator(ctx, script)

	case "similar_vibes":
		if script.SeedClipID == nil {
			return nil, fmt.Errorf("similar_vibes strategy requires seed_clip_id")
		}
		return s.curationRepo.SimilarVibes(ctx, script, *script.SeedClipID)

	case "cross_game_hits":
		if len(script.GameIDs) == 0 {
			return nil, fmt.Errorf("cross_game_hits strategy requires game_ids")
		}
		return s.curationRepo.CrossGameHits(ctx, script, script.GameIDs)

	case "controversial":
		return s.curationRepo.Controversial(ctx, script)

	case "binge_worthy":
		return s.curationRepo.BingeWorthy(ctx, script)

	case "rising_stars":
		return s.curationRepo.RisingStars(ctx, script)

	default:
		return nil, fmt.Errorf("unknown strategy: %s", script.Strategy)
	}
}

// executeTwitchStrategy runs a Twitch-powered curation strategy.
// These strategies fetch fresh clips from the Twitch API, import them into the local DB,
// and return the results — making playlist scripts the driver of content discovery.
func (s *PlaylistScriptService) executeTwitchStrategy(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	if s.clipSyncService == nil {
		return nil, fmt.Errorf("twitch strategies require clip sync service (Twitch client not configured)")
	}

	langFilter := ""
	if script.Language != nil {
		langFilter = *script.Language
	}

	switch script.Strategy {
	case "twitch_top_game":
		return s.twitchTopGame(ctx, script, langFilter)

	case "twitch_top_broadcaster":
		return s.twitchTopBroadcaster(ctx, script, langFilter)

	case "twitch_trending":
		return s.twitchTrending(ctx, script, langFilter)

	case "twitch_discovery":
		return s.twitchDiscovery(ctx, script, langFilter)

	default:
		return nil, fmt.Errorf("unknown twitch strategy: %s", script.Strategy)
	}
}

// twitchTopGame fetches top clips for a specific game from Twitch.
// Example: "Top League Moments", "Best Valorant Clips This Week"
func (s *PlaylistScriptService) twitchTopGame(ctx context.Context, script *models.PlaylistScript, langFilter string) ([]models.Clip, error) {
	if script.GameID == nil || *script.GameID == "" {
		return nil, fmt.Errorf("twitch_top_game strategy requires game_id")
	}

	params := &twitch.ClipParams{
		GameID: *script.GameID,
		First:  minInt(script.ClipLimit, 100),
	}

	hours := TimeframeToHours(script.Timeframe)
	if hours > 0 {
		params.EndedAt = time.Now()
		params.StartedAt = params.EndedAt.Add(-time.Duration(hours) * time.Hour)
	}
	// If hours == 0 (all-time), leave StartedAt/EndedAt as zero — Twitch returns all-time top clips

	return s.clipSyncService.FetchAndImportClips(ctx, params, script.ClipLimit, langFilter, BotUserID)
}

// twitchTopBroadcaster fetches top clips for a specific broadcaster from Twitch.
// Example: "Hasanabi Most Viewed", "xQc Best of the Week"
func (s *PlaylistScriptService) twitchTopBroadcaster(ctx context.Context, script *models.PlaylistScript, langFilter string) ([]models.Clip, error) {
	if script.BroadcasterID == nil || *script.BroadcasterID == "" {
		return nil, fmt.Errorf("twitch_top_broadcaster strategy requires broadcaster_id")
	}

	params := &twitch.ClipParams{
		BroadcasterID: *script.BroadcasterID,
		First:         minInt(script.ClipLimit, 100),
	}

	hours := TimeframeToHours(script.Timeframe)
	if hours > 0 {
		params.EndedAt = time.Now()
		params.StartedAt = params.EndedAt.Add(-time.Duration(hours) * time.Hour)
	}

	return s.clipSyncService.FetchAndImportClips(ctx, params, script.ClipLimit, langFilter, BotUserID)
}

// twitchTrending fetches the hottest clips across Twitch's currently popular games.
// Dynamically discovers what's trending via GetTopGames, then fetches top clips from each.
// Example: "Twitch Top 10 Today", "What's Popping on Twitch"
func (s *PlaylistScriptService) twitchTrending(ctx context.Context, script *models.PlaylistScript, langFilter string) ([]models.Clip, error) {
	hours := TimeframeToHours(script.Timeframe)
	if hours == 0 {
		hours = 24 // default to last 24 hours for trending
	}

	// Use SyncTrendingClips which already handles top game resolution,
	// then query our local DB for the freshest imports
	_, err := s.clipSyncService.SyncTrendingClips(ctx, hours, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sync trending clips: %w", err)
	}

	// After syncing, query local DB for recently imported clips sorted by view count
	// This leverages the existing standard filter path with a recent import window
	filters := buildFiltersFromScript(script)
	filters.Sort = "top"
	clips, _, err := s.clipRepo.ListWithFilters(ctx, filters, script.ClipLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to query trending clips: %w", err)
	}

	return clips, nil
}

// twitchDiscovery fetches clips from a diverse set of games for content discovery.
// Goes beyond the usual top 10 to surface clips from lesser-known categories.
// If game_ids are provided, uses those; otherwise picks from top 20 games on Twitch.
// Example: "Discover Something New", "Beyond the Mainstream"
func (s *PlaylistScriptService) twitchDiscovery(ctx context.Context, script *models.PlaylistScript, langFilter string) ([]models.Clip, error) {
	hours := TimeframeToHours(script.Timeframe)
	if hours == 0 {
		hours = 168 // default to last week for discovery
	}

	var gameIDs []string

	if len(script.GameIDs) > 0 {
		gameIDs = script.GameIDs
	} else {
		// Fetch top 20 games from Twitch, then use the bottom half (ranks 11-20)
		// to discover content beyond the usual top categories
		topGamesResp, err := s.clipSyncService.GetTopGames(ctx, 20)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch top games for discovery: %w", err)
		}
		// Take games ranked 11-20 for true discovery
		startIdx := 10
		if len(topGamesResp) <= 10 {
			startIdx = 0
		}
		for i := startIdx; i < len(topGamesResp); i++ {
			gameIDs = append(gameIDs, topGamesResp[i])
		}
	}

	if len(gameIDs) == 0 {
		return nil, fmt.Errorf("no games found for discovery")
	}

	// Fetch a few clips from each game to build a diverse playlist
	clipsPerGame := maxInt(script.ClipLimit/len(gameIDs), 1)
	var allClips []models.Clip

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	for _, gid := range gameIDs {
		if len(allClips) >= script.ClipLimit {
			break
		}

		params := &twitch.ClipParams{
			GameID:    gid,
			StartedAt: startTime,
			EndedAt:   endTime,
			First:     minInt(clipsPerGame, 100),
		}

		clips, err := s.clipSyncService.FetchAndImportClips(ctx, params, clipsPerGame, langFilter, BotUserID)
		if err != nil {
			continue // skip this game on error, try the next
		}
		allClips = append(allClips, clips...)
	}

	// Trim to limit
	if len(allClips) > script.ClipLimit {
		allClips = allClips[:script.ClipLimit]
	}

	return allClips, nil
}

// minInt returns the smaller of two ints (avoids redeclaring builtin)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the larger of two ints (avoids redeclaring builtin)
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
