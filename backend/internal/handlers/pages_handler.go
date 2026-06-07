package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ClipRepositoryForPages defines clip repo methods needed by PagesHandler.
type ClipRepositoryForPages interface {
	ListClipsByBroadcaster(ctx context.Context, broadcasterID, sort string, limit, offset int) ([]models.Clip, int, error)
	ListClipsByGame(ctx context.Context, gameID string, limit, offset int) ([]models.Clip, int, error)
	ListClipsForBestOf(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]models.Clip, int, error)
	ListClipsForStreamerGame(ctx context.Context, broadcasterID, gameID string, limit, offset int) ([]models.Clip, int, error)
}

// BroadcasterRepositoryForPages defines broadcaster repo methods needed by PagesHandler.
type BroadcasterRepositoryForPages interface {
	GetBroadcasterByName(ctx context.Context, broadcasterName string) (string, error)
	GetBroadcasterStats(ctx context.Context, broadcasterID string) (int, int64, float64, error)
	GetFollowerCount(ctx context.Context, broadcasterID string) (int, error)
	ListBroadcasterGames(ctx context.Context, broadcasterID string) ([]models.GameWithClipCount, error)
}

// GameRepositoryForPages defines game repo methods needed by PagesHandler.
type GameRepositoryForPages interface {
	GetBySlug(ctx context.Context, slug string) (*models.GameEntity, error)
	GetByTwitchGameID(ctx context.Context, twitchGameID string) (*models.GameEntity, error)
	ListTopBroadcastersForGame(ctx context.Context, gameID string, limit int) ([]models.BroadcasterWithClipCount, error)
}

// PagesHandler renders server-side HTML pages for SEO.
type PagesHandler struct {
	clipRepo        ClipRepositoryForPages
	broadcasterRepo BroadcasterRepositoryForPages
	gameRepo        GameRepositoryForPages
}

// NewPagesHandler creates a new PagesHandler.
func NewPagesHandler(clipRepo ClipRepositoryForPages, broadcasterRepo BroadcasterRepositoryForPages, gameRepo GameRepositoryForPages) *PagesHandler {
	return &PagesHandler{
		clipRepo:        clipRepo,
		broadcasterRepo: broadcasterRepo,
		gameRepo:        gameRepo,
	}
}

const pagesClipLimit = 50

// GetStreamerPage renders the streamer profile pSEO page.
func (h *PagesHandler) GetStreamerPage(c *gin.Context) {
	broadcasterName := c.Param("broadcasterName")
	if broadcasterName == "" {
		h.render404(c)
		return
	}

	ctx := c.Request.Context()
	baseURL := getBaseURL(c)

	broadcasterID, err := h.broadcasterRepo.GetBroadcasterByName(ctx, broadcasterName)
	if err != nil {
		h.render404(c)
		return
	}

	totalClips, totalViews, avgScore, _ := h.broadcasterRepo.GetBroadcasterStats(ctx, broadcasterID)
	followerCount, _ := h.broadcasterRepo.GetFollowerCount(ctx, broadcasterID)
	games, _ := h.broadcasterRepo.ListBroadcasterGames(ctx, broadcasterID)

	clips, _, err := h.clipRepo.ListClipsByBroadcaster(ctx, broadcasterID, "popular", pagesClipLimit, 0)
	if err != nil {
		log.Printf("Error listing broadcaster clips: %v", err)
		clips = nil
	}

	ogImage := ""
	if len(clips) > 0 && clips[0].ThumbnailURL != nil {
		ogImage = *clips[0].ThumbnailURL
	}

	schema := buildProfileSchema(broadcasterName, baseURL, totalClips, totalViews)

	data := models.StreamerPageData{
		PageData: models.PageData{
			Title:        fmt.Sprintf("%s Clips", broadcasterName),
			Description:  fmt.Sprintf("Watch the best %s Twitch clips on clpr.tv. %d clips, %d total views.", broadcasterName, totalClips, totalViews),
			CanonicalURL: fmt.Sprintf("%s/clips/streamer/%s", baseURL, broadcasterName),
			OGImage:      ogImage,
			SchemaJSON:   schema,
			BaseURL:      baseURL,
		},
		BroadcasterName: broadcasterName,
		BroadcasterID:   broadcasterID,
		TotalClips:      totalClips,
		TotalViews:      totalViews,
		AvgVoteScore:    avgScore,
		FollowerCount:   followerCount,
		TopClips:        clips,
		GamesPlayed:     games,
	}

	c.HTML(http.StatusOK, "streamer.html", data)
}

// GetGamePage renders the game pSEO page.
func (h *PagesHandler) GetGamePage(c *gin.Context) {
	gameSlug := c.Param("gameSlug")
	if gameSlug == "" {
		h.render404(c)
		return
	}

	ctx := c.Request.Context()
	baseURL := getBaseURL(c)

	game, err := h.gameRepo.GetBySlug(ctx, gameSlug)
	if err != nil {
		h.render404(c)
		return
	}

	topBroadcasters, _ := h.gameRepo.ListTopBroadcastersForGame(ctx, game.TwitchGameID, 20)

	clips, total, err := h.clipRepo.ListClipsByGame(ctx, game.TwitchGameID, pagesClipLimit, 0)
	if err != nil {
		log.Printf("Error listing game clips: %v", err)
		clips = nil
		total = 0
	}

	ogImage := ""
	if game.BoxArtURL != nil {
		ogImage = *game.BoxArtURL
	}

	schema := buildCollectionSchema(game.Name+" Clips", baseURL, total)

	data := models.GamePageData{
		PageData: models.PageData{
			Title:        fmt.Sprintf("%s Clips & Highlights", game.Name),
			Description:  fmt.Sprintf("Watch the best %s Twitch clips and highlights. %d clips from top streamers.", game.Name, total),
			CanonicalURL: fmt.Sprintf("%s/clips/game/%s", baseURL, gameSlug),
			OGImage:      ogImage,
			SchemaJSON:   schema,
			BaseURL:      baseURL,
		},
		Game:            *game,
		TopClips:        clips,
		TopBroadcasters: topBroadcasters,
		TotalClips:      total,
	}

	c.HTML(http.StatusOK, "game.html", data)
}

// GetBestOfPage renders the temporal best-of pSEO page (this-week, this-month).
func (h *PagesHandler) GetBestOfPage(c *gin.Context) {
	period := c.Param("period")
	start, end, label, ok := parsePeriod(period)
	if !ok {
		h.render404(c)
		return
	}

	ctx := c.Request.Context()
	baseURL := getBaseURL(c)

	clips, total, err := h.clipRepo.ListClipsForBestOf(ctx, start, end, pagesClipLimit, 0)
	if err != nil {
		log.Printf("Error listing best-of clips: %v", err)
		clips = nil
	}

	ogImage := ""
	if len(clips) > 0 && clips[0].ThumbnailURL != nil {
		ogImage = *clips[0].ThumbnailURL
	}

	schema := buildItemListSchema(clips, baseURL)

	data := models.BestOfPageData{
		PageData: models.PageData{
			Title:        fmt.Sprintf("Best Twitch Clips: %s", label),
			Description:  fmt.Sprintf("The top Twitch clips from %s. %d clips ranked by community votes.", label, total),
			CanonicalURL: fmt.Sprintf("%s/clips/best/%s", baseURL, period),
			OGImage:      ogImage,
			SchemaJSON:   schema,
			BaseURL:      baseURL,
		},
		Period:    label,
		PeriodKey: period,
		StartDate: start,
		EndDate:   end,
		TopClips:  clips,
		Total:     total,
	}

	c.HTML(http.StatusOK, "best.html", data)
}

// GetBestOfMonthPage renders the monthly archive pSEO page.
func (h *PagesHandler) GetBestOfMonthPage(c *gin.Context) {
	yearStr := c.Param("year")
	monthStr := c.Param("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > time.Now().Year()+1 {
		h.render404(c)
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		h.render404(c)
		return
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	label := start.Format("January 2006")

	ctx := c.Request.Context()
	baseURL := getBaseURL(c)

	clips, total, err := h.clipRepo.ListClipsForBestOf(ctx, start, end, pagesClipLimit, 0)
	if err != nil {
		log.Printf("Error listing monthly best-of clips: %v", err)
		clips = nil
	}

	ogImage := ""
	if len(clips) > 0 && clips[0].ThumbnailURL != nil {
		ogImage = *clips[0].ThumbnailURL
	}

	schema := buildItemListSchema(clips, baseURL)
	periodKey := fmt.Sprintf("%d/%02d", year, month)

	data := models.BestOfPageData{
		PageData: models.PageData{
			Title:        fmt.Sprintf("Best Twitch Clips: %s", label),
			Description:  fmt.Sprintf("The top Twitch clips from %s. %d clips ranked by community votes.", label, total),
			CanonicalURL: fmt.Sprintf("%s/clips/best/%s", baseURL, periodKey),
			OGImage:      ogImage,
			SchemaJSON:   schema,
			BaseURL:      baseURL,
		},
		Period:    label,
		PeriodKey: periodKey,
		StartDate: start,
		EndDate:   end,
		TopClips:  clips,
		Total:     total,
	}

	c.HTML(http.StatusOK, "best.html", data)
}

// GetStreamerGamePage renders the streamer+game combo pSEO page.
func (h *PagesHandler) GetStreamerGamePage(c *gin.Context) {
	broadcasterName := c.Param("broadcasterName")
	gameSlug := c.Param("gameSlug")
	if broadcasterName == "" || gameSlug == "" {
		h.render404(c)
		return
	}

	ctx := c.Request.Context()
	baseURL := getBaseURL(c)

	broadcasterID, err := h.broadcasterRepo.GetBroadcasterByName(ctx, broadcasterName)
	if err != nil {
		h.render404(c)
		return
	}

	game, err := h.gameRepo.GetBySlug(ctx, gameSlug)
	if err != nil {
		h.render404(c)
		return
	}

	clips, total, err := h.clipRepo.ListClipsForStreamerGame(ctx, broadcasterID, game.TwitchGameID, pagesClipLimit, 0)
	if err != nil {
		log.Printf("Error listing streamer+game clips: %v", err)
		clips = nil
	}

	ogImage := ""
	if len(clips) > 0 && clips[0].ThumbnailURL != nil {
		ogImage = *clips[0].ThumbnailURL
	}

	schema := buildCollectionSchema(fmt.Sprintf("%s playing %s", broadcasterName, game.Name), baseURL, total)

	data := models.StreamerGamePageData{
		PageData: models.PageData{
			Title:        fmt.Sprintf("%s playing %s - Clips", broadcasterName, game.Name),
			Description:  fmt.Sprintf("Watch %s play %s. %d clips on clpr.tv.", broadcasterName, game.Name, total),
			CanonicalURL: fmt.Sprintf("%s/clips/streamer/%s/%s", baseURL, broadcasterName, gameSlug),
			OGImage:      ogImage,
			SchemaJSON:   schema,
			BaseURL:      baseURL,
		},
		BroadcasterName: broadcasterName,
		BroadcasterID:   broadcasterID,
		Game:            *game,
		Clips:           clips,
		Total:           total,
	}

	c.HTML(http.StatusOK, "streamer_game.html", data)
}

func (h *PagesHandler) render404(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", models.PageData{
		Title:       "Page Not Found",
		Description: "The page you're looking for doesn't exist.",
	})
}

// getBaseURL extracts the base URL from gin context.
func getBaseURL(c *gin.Context) string {
	baseURL := c.GetString("base_url")
	if baseURL == "" {
		baseURL = "https://clpr.tv"
	}
	return strings.TrimRight(baseURL, "/")
}

// parsePeriod converts "this-week" or "this-month" to date ranges.
func parsePeriod(period string) (start, end time.Time, label string, ok bool) {
	now := time.Now().UTC()
	switch period {
	case "this-week":
		// Start from last Monday
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 7)
		label = "This Week"
		ok = true
	case "this-month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
		label = "This Month"
		ok = true
	default:
		ok = false
	}
	return
}

// Schema.org JSON-LD builders

func buildProfileSchema(name, baseURL string, totalClips int, totalViews int64) string {
	schema := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "ProfilePage",
		"name":     name + " Clips",
		"url":      fmt.Sprintf("%s/clips/streamer/%s", baseURL, name),
		"mainEntity": map[string]interface{}{
			"@type":               "Person",
			"name":                name,
			"url":                 fmt.Sprintf("https://twitch.tv/%s", name),
			"interactionStatistic": []map[string]interface{}{
				{"@type": "InteractionCounter", "interactionType": "https://schema.org/WatchAction", "userInteractionCount": totalViews},
			},
		},
	}
	b, _ := json.Marshal(schema)
	return string(b)
}

func buildCollectionSchema(name, baseURL string, totalItems int) string {
	schema := map[string]interface{}{
		"@context":         "https://schema.org",
		"@type":            "CollectionPage",
		"name":             name,
		"numberOfItems":    totalItems,
	}
	b, _ := json.Marshal(schema)
	return string(b)
}

func buildItemListSchema(clips []models.Clip, baseURL string) string {
	items := make([]map[string]interface{}, 0, len(clips))
	for i, clip := range clips {
		if i >= 10 {
			break
		}
		item := map[string]interface{}{
			"@type":    "ListItem",
			"position": i + 1,
			"url":      fmt.Sprintf("%s/clip/%s", baseURL, clip.ID),
			"name":     clip.Title,
		}
		items = append(items, item)
	}
	schema := map[string]interface{}{
		"@context":        "https://schema.org",
		"@type":           "ItemList",
		"itemListElement": items,
	}
	b, _ := json.Marshal(schema)
	return string(b)
}

