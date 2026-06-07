package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

type siteFreshnessPreset struct {
	Name          string
	Description   string
	Sort          string
	Timeframe     *string
	ClipLimit     int
	Visibility    string
	Schedule      string
	Strategy      string
	RetentionDays int
	TitleTemplate string
	RequiresTwitch bool
}

type seedSummary struct {
	Created        []string
	Skipped        []string
	Generated      []string
	GenerateErrors []string
}

func main() {
	generateNow := flag.Bool("generate-now", false, "Generate playlists immediately after ensuring the presets exist")
	dryRun := flag.Bool("dry-run", false, "Show what would be created or generated without changing data")
	ownerFlag := flag.String("owner", services.BotUserID.String(), "Owner user UUID for the default smart playlists")
	flag.Parse()

	log.Println("Bootstrapping site freshness automation...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ownerID, err := uuid.Parse(*ownerFlag)
	if err != nil {
		log.Fatalf("Invalid owner UUID %q: %v", *ownerFlag, err)
	}

	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db.Pool)
	if _, err := userRepo.GetByID(ctx, ownerID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Fatalf("Owner user %s was not found. Run migrations first or choose a different owner with -owner.", ownerID)
		}
		log.Fatalf("Failed to verify owner user %s: %v", ownerID, err)
	}

	scriptRepo := repository.NewPlaylistScriptRepository(db.Pool)
	playlistRepo := repository.NewPlaylistRepository(db.Pool)
	clipRepo := repository.NewClipRepository(db.Pool)
	curationRepo := repository.NewPlaylistCurationRepository(db.Pool)
	playlistScriptService := services.NewPlaylistScriptService(scriptRepo, playlistRepo, clipRepo, curationRepo, nil)

	includeTwitchPresets := hasTwitchCredentials(cfg)
	presets := defaultSiteFreshnessPresets(includeTwitchPresets)
	if len(presets) == 0 {
		log.Println("No presets selected. Nothing to do.")
		return
	}

	var clipSyncService *services.ClipSyncService
	if *generateNow && includeTwitchPresets {
		clipSyncService = initClipSyncService(cfg, db)
		if clipSyncService != nil {
			playlistScriptService.SetClipSyncService(clipSyncService)
		}
	}

	summary, err := seedPresets(ctx, ownerID, presets, scriptRepo, playlistScriptService, *generateNow, *dryRun, clipSyncService != nil)
	if err != nil {
		log.Fatalf("Failed to seed site freshness presets: %v", err)
	}

	printSummary(summary, *dryRun, cfg.Server.BaseURL)

	if len(summary.GenerateErrors) > 0 {
		os.Exit(1)
	}
}

func defaultSiteFreshnessPresets(includeTwitch bool) []siteFreshnessPreset {
	day := stringPtr("day")
	week := stringPtr("week")
	month := stringPtr("month")

	presets := []siteFreshnessPreset{
		{
			Name:           "Viral Velocity",
			Description:    "Automatically refreshed playlist highlighting clips with the fastest recent momentum.",
			Sort:           "trending",
			Timeframe:      day,
			ClipLimit:      20,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "viral_velocity",
			RetentionDays:  7,
			TitleTemplate:  "Viral Velocity • {date}",
			RequiresTwitch: false,
		},
		{
			Name:           "Fresh Faces",
			Description:    "Daily playlist surfacing standout clips from newer creators so discovery does not get stuck on the usual suspects.",
			Sort:           "top",
			Timeframe:      week,
			ClipLimit:      20,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "fresh_faces",
			RetentionDays:  10,
			TitleTemplate:  "Fresh Faces • {date}",
			RequiresTwitch: false,
		},
		{
			Name:           "Creator Roulette",
			Description:    "Daily variety rail with one standout clip per creator so the homepage stays diverse and surprising.",
			Sort:           "top",
			Timeframe:      week,
			ClipLimit:      20,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "one_per_creator",
			RetentionDays:  7,
			TitleTemplate:  "Creator Roulette • {date}",
			RequiresTwitch: false,
		},
		{
			Name:           "Hidden Gems",
			Description:    "Daily playlist of sleeper-hit clips with strong retention that deserve a much bigger audience.",
			Sort:           "top",
			Timeframe:      week,
			ClipLimit:      20,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "sleeper_hits",
			RetentionDays:  10,
			TitleTemplate:  "Hidden Gems • {date}",
			RequiresTwitch: false,
		},
		{
			Name:           "Community Favorites",
			Description:    "Automatically refreshed playlist of the clips people save and come back to most often.",
			Sort:           "top",
			Timeframe:      month,
			ClipLimit:      20,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "community_favorites",
			RetentionDays:  14,
			TitleTemplate:  "Community Favorites • {date}",
			RequiresTwitch: false,
		},
		{
			Name:           "Breakout Board",
			Description:    "Daily spotlight on creators whose recent clips are outperforming their usual baseline.",
			Sort:           "top",
			Timeframe:      month,
			ClipLimit:      20,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "rising_stars",
			RetentionDays:  10,
			TitleTemplate:  "Breakout Board • {date}",
			RequiresTwitch: false,
		},
		{
			Name:           "Deep Cuts Weekly",
			Description:    "Weekly playlist of underrated gems with strong watch-through and engagement.",
			Sort:           "top",
			Timeframe:      week,
			ClipLimit:      25,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "weekly",
			Strategy:       "deep_cuts",
			RetentionDays:  21,
			TitleTemplate:  "Deep Cuts • Week of {week_start}",
			RequiresTwitch: false,
		},
		{
			Name:           "Binge Loop",
			Description:    "Daily playlist of clips that tend to keep people watching through multi-clip sessions.",
			Sort:           "top",
			Timeframe:      week,
			ClipLimit:      24,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "binge_worthy",
			RetentionDays:  7,
			TitleTemplate:  "Binge Loop • {date}",
			RequiresTwitch: false,
		},
		{
			Name:           "Hot Takes",
			Description:    "Daily playlist of the clips that sparked the most debate, reactions, and comment-fueled chaos.",
			Sort:           "trending",
			Timeframe:      week,
			ClipLimit:      18,
			Visibility:     models.PlaylistVisibilityPublic,
			Schedule:       "daily",
			Strategy:       "controversial",
			RetentionDays:  7,
			TitleTemplate:  "Hot Takes • {date}",
			RequiresTwitch: false,
		},
	}

	if includeTwitch {
		presets = append(presets,
			siteFreshnessPreset{
				Name:           "Trending Now",
				Description:    "Automatically imports and publishes a fresh mix of top clips from Twitch's current hottest games.",
				Sort:           "trending",
				Timeframe:      day,
				ClipLimit:      25,
				Visibility:     models.PlaylistVisibilityPublic,
				Schedule:       "daily",
				Strategy:       "twitch_trending",
				RetentionDays:  7,
				TitleTemplate:  "Trending Now • {date}",
				RequiresTwitch: true,
			},
			siteFreshnessPreset{
				Name:           "Discovery Mix",
				Description:    "Daily discovery playlist pulling clips from beyond Twitch's main categories to keep the catalog surprising.",
				Sort:           "top",
				Timeframe:      week,
				ClipLimit:      25,
				Visibility:     models.PlaylistVisibilityPublic,
				Schedule:       "daily",
				Strategy:       "twitch_discovery",
				RetentionDays:  7,
				TitleTemplate:  "Discovery Mix • {date}",
				RequiresTwitch: true,
			},
		)
	}

	return presets
}

func seedPresets(
	ctx context.Context,
	ownerID uuid.UUID,
	presets []siteFreshnessPreset,
	scriptRepo *repository.PlaylistScriptRepository,
	service *services.PlaylistScriptService,
	generateNow bool,
	dryRun bool,
	canGenerateTwitch bool,
) (*seedSummary, error) {
	existingScripts, err := scriptRepo.ListByUser(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing scripts: %w", err)
	}

	existingByName := make(map[string]*models.PlaylistScript, len(existingScripts))
	for _, script := range existingScripts {
		existingByName[script.Name] = script
	}

	summary := &seedSummary{}

	for _, preset := range presets {
		script, exists := existingByName[preset.Name]
		if exists {
			summary.Skipped = append(summary.Skipped, preset.Name)
		} else {
			if dryRun {
				summary.Created = append(summary.Created, fmt.Sprintf("%s (dry-run)", preset.Name))
			} else {
				created, err := service.CreateScript(ctx, ownerID, preset.toCreateRequest())
				if err != nil {
					return nil, fmt.Errorf("failed to create preset %q: %w", preset.Name, err)
				}
				script = created
				existingByName[preset.Name] = created
				summary.Created = append(summary.Created, preset.Name)
			}
		}

		if !generateNow {
			continue
		}

		if preset.RequiresTwitch && !canGenerateTwitch {
			summary.GenerateErrors = append(summary.GenerateErrors, fmt.Sprintf("%s: skipped immediate generation because Twitch/Redis initialization was unavailable", preset.Name))
			continue
		}

		if dryRun {
			summary.Generated = append(summary.Generated, fmt.Sprintf("%s (dry-run)", preset.Name))
			continue
		}

		playlist, err := service.GeneratePlaylist(ctx, script.ID)
		if err != nil {
			summary.GenerateErrors = append(summary.GenerateErrors, fmt.Sprintf("%s: %v", preset.Name, err))
			continue
		}

		summary.Generated = append(summary.Generated, fmt.Sprintf("%s -> %s", preset.Name, playlist.ID.String()))
	}

	sort.Strings(summary.Created)
	sort.Strings(summary.Skipped)
	sort.Strings(summary.Generated)
	sort.Strings(summary.GenerateErrors)

	return summary, nil
}

func (p siteFreshnessPreset) toCreateRequest() *models.CreatePlaylistScriptRequest {
	visibility := p.Visibility
	schedule := p.Schedule
	strategy := p.Strategy
	retention := p.RetentionDays
	isActive := true
	description := p.Description
	titleTemplate := p.TitleTemplate

	return &models.CreatePlaylistScriptRequest{
		Name:          p.Name,
		Description:   &description,
		Sort:          p.Sort,
		Timeframe:     p.Timeframe,
		ClipLimit:     p.ClipLimit,
		Visibility:    &visibility,
		IsActive:      &isActive,
		Schedule:      &schedule,
		Strategy:      &strategy,
		ExcludeNSFW:   boolPtr(true),
		RetentionDays: &retention,
		TitleTemplate: &titleTemplate,
	}
}

func initClipSyncService(cfg *config.Config, db *database.DB) *services.ClipSyncService {
	redisClient, err := redispkg.NewClient(&cfg.Redis)
	if err != nil {
		log.Printf("WARNING: Failed to initialize Redis client for immediate Twitch generation: %v", err)
		return nil
	}

	twitchClient, err := twitch.NewClient(&cfg.Twitch, redisClient)
	if err != nil {
		log.Printf("WARNING: Failed to initialize Twitch client for immediate generation: %v", err)
		return nil
	}

	return services.NewClipSyncService(
		twitchClient,
		repository.NewClipRepository(db.Pool),
		repository.NewTagRepository(db.Pool),
		repository.NewUserRepository(db.Pool),
		redisClient,
	)
}

func printSummary(summary *seedSummary, dryRun bool, baseURL string) {
	modeLabel := "completed"
	if dryRun {
		modeLabel = "dry-run completed"
	}

	log.Printf("Site freshness bootstrap %s", modeLabel)
	log.Printf("Created presets: %d", len(summary.Created))
	for _, item := range summary.Created {
		log.Printf("  + %s", item)
	}

	log.Printf("Skipped existing presets: %d", len(summary.Skipped))
	for _, item := range summary.Skipped {
		log.Printf("  = %s", item)
	}

	if len(summary.Generated) > 0 {
		log.Printf("Generated playlists: %d", len(summary.Generated))
		for _, item := range summary.Generated {
			if strings.Contains(item, "->") {
				parts := strings.SplitN(item, "->", 2)
				playlistID := strings.TrimSpace(parts[1])
				log.Printf("  * %s (%s/playlists/%s)", strings.TrimSpace(parts[0]), strings.TrimRight(baseURL, "/"), playlistID)
				continue
			}
			log.Printf("  * %s", item)
		}
	}

	if len(summary.GenerateErrors) > 0 {
		log.Printf("Immediate generation warnings/errors: %d", len(summary.GenerateErrors))
		for _, item := range summary.GenerateErrors {
			log.Printf("  ! %s", item)
		}
	}

	if len(summary.GenerateErrors) == 0 {
		log.Println("Tip: the API scheduler checks for due playlist scripts every 5 minutes, so these rules will keep producing fresh public playlists automatically.")
	}
}

func hasTwitchCredentials(cfg *config.Config) bool {
	return strings.TrimSpace(cfg.Twitch.ClientID) != "" && strings.TrimSpace(cfg.Twitch.ClientSecret) != ""
}

func stringPtr(v string) *string {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}
