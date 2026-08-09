package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrPlaylistScriptNotFound   = repository.ErrPlaylistScriptNotFound
	ErrPlaylistScriptInactive   = errors.New("playlist script is inactive")
	ErrPlaylistScriptValidation = errors.New("invalid playlist script")
	ErrPlaylistGenerationEmpty  = errors.New("playlist generation returned no clips")
)

// BotUserID is the well-known UUID of the system bot user that posts clips
// on behalf of automated playlist scripts. Created in migration 000107.
var BotUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type generatedPlaylistPresentation struct {
	isCurated    bool
	isFeatured   bool
	displayOrder int
}

type playlistGenerationWriter interface {
	Persist(context.Context, *models.PlaylistScript, *models.Playlist, []models.Clip) error
}

var siteFreshnessDisplayOrder = map[string]int{
	"Viral Velocity":      1,
	"Trending Now":        2,
	"Hidden Gems":         3,
	"Fresh Faces":         4,
	"Creator Roulette":    5,
	"Breakout Board":      6,
	"Community Favorites": 7,
	"Discovery Mix":       8,
	"Binge Loop":          9,
	"Deep Cuts Weekly":    10,
	"Hot Takes":           11,
}

// PlaylistScriptService handles script-based playlist automation
type PlaylistScriptService struct {
	scriptRepo       *repository.PlaylistScriptRepository
	playlistRepo     *repository.PlaylistRepository
	clipRepo         *repository.ClipRepository
	curationRepo     *repository.PlaylistCurationRepository
	clipSyncService  *ClipSyncService // nil when Twitch client is not configured
	generationWriter playlistGenerationWriter
}

// NewPlaylistScriptService creates a new PlaylistScriptService
func NewPlaylistScriptService(scriptRepo *repository.PlaylistScriptRepository, playlistRepo *repository.PlaylistRepository, clipRepo *repository.ClipRepository, curationRepo *repository.PlaylistCurationRepository, clipSyncService *ClipSyncService) *PlaylistScriptService {
	return &PlaylistScriptService{
		scriptRepo:       scriptRepo,
		playlistRepo:     playlistRepo,
		clipRepo:         clipRepo,
		curationRepo:     curationRepo,
		clipSyncService:  clipSyncService,
		generationWriter: repository.NewPlaylistGenerationWriter(scriptRepo),
	}
}

// SetClipSyncService sets the clip sync service for Twitch-powered strategies.
// Called after initialization since ClipSyncService may be nil when Twitch is not configured.
func (s *PlaylistScriptService) SetClipSyncService(clipSyncService *ClipSyncService) {
	s.clipSyncService = clipSyncService
}

// ListScripts returns all playlist scripts
func (s *PlaylistScriptService) ListScripts(ctx context.Context, limit int) ([]*models.PlaylistScript, error) {
	return s.scriptRepo.List(ctx, limit)
}

// ListUserScripts returns playlist scripts owned by a specific user
func (s *PlaylistScriptService) ListUserScripts(ctx context.Context, userID uuid.UUID) ([]*models.PlaylistScript, error) {
	return s.scriptRepo.ListByUser(ctx, userID)
}

// allowedUserSchedules defines the schedules available to non-admin users.
var allowedUserSchedules = map[string]bool{
	"manual": true,
	"daily":  true,
	"weekly": true,
}

// CreateUserScript creates a playlist script scoped to a regular user.
// Strategy is forced to "standard" and schedule is restricted.
func (s *PlaylistScriptService) CreateUserScript(ctx context.Context, userID uuid.UUID, req *models.CreatePlaylistScriptRequest) (*models.PlaylistScript, error) {
	if req.Strategy != nil && *req.Strategy != "standard" {
		return nil, fmt.Errorf("%w: user scripts only support the standard strategy", ErrPlaylistScriptValidation)
	}
	// Force strategy to standard for user-created scripts
	standard := "standard"
	req.Strategy = &standard

	// Validate schedule
	schedule := "daily"
	if req.Schedule != nil {
		schedule = *req.Schedule
	}
	if !allowedUserSchedules[schedule] {
		return nil, fmt.Errorf("%w: schedule is not allowed", ErrPlaylistScriptValidation)
	}

	return s.CreateScript(ctx, userID, req)
}

// GetUserScript retrieves a script only if owned by the given user.
func (s *PlaylistScriptService) GetUserScript(ctx context.Context, scriptID, userID uuid.UUID) (*models.PlaylistScript, error) {
	script, err := s.scriptRepo.GetByID(ctx, scriptID)
	if err != nil {
		return nil, err
	}
	if script == nil {
		return nil, ErrPlaylistScriptNotFound
	}
	if script.CreatedBy == nil || *script.CreatedBy != userID {
		return nil, ErrPlaylistScriptNotFound
	}
	return script, nil
}

// UpdateUserScript updates a script only if owned by the given user.
// Strategy changes are not allowed for user scripts.
func (s *PlaylistScriptService) UpdateUserScript(ctx context.Context, scriptID, userID uuid.UUID, req *models.UpdatePlaylistScriptRequest) (*models.PlaylistScript, error) {
	if _, err := s.GetUserScript(ctx, scriptID, userID); err != nil {
		return nil, err
	}

	if req.Strategy != nil && *req.Strategy != "standard" {
		return nil, fmt.Errorf("%w: user scripts only support the standard strategy", ErrPlaylistScriptValidation)
	}
	// The strategy is immutable on the user surface.
	req.Strategy = nil

	// Validate schedule if provided
	if req.Schedule != nil && !allowedUserSchedules[*req.Schedule] {
		return nil, fmt.Errorf("%w: schedule is not allowed", ErrPlaylistScriptValidation)
	}

	return s.UpdateScript(ctx, scriptID, req)
}

// DeleteUserScript deletes a script only if owned by the given user.
func (s *PlaylistScriptService) DeleteUserScript(ctx context.Context, scriptID, userID uuid.UUID) error {
	if _, err := s.GetUserScript(ctx, scriptID, userID); err != nil {
		return err
	}
	return s.DeleteScript(ctx, scriptID)
}

// GenerateUserPlaylist generates a playlist from a script only if owned by the given user.
func (s *PlaylistScriptService) GenerateUserPlaylist(ctx context.Context, scriptID, userID uuid.UUID) (*models.Playlist, error) {
	if _, err := s.GetUserScript(ctx, scriptID, userID); err != nil {
		return nil, err
	}
	return s.GeneratePlaylist(ctx, scriptID)
}

// CreateScript creates a new playlist script
func (s *PlaylistScriptService) CreateScript(ctx context.Context, userID uuid.UUID, req *models.CreatePlaylistScriptRequest) (*models.PlaylistScript, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", ErrPlaylistScriptValidation)
	}
	visibility := models.PlaylistVisibilityPublic
	if req.Visibility != nil {
		visibility = *req.Visibility
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	schedule := "daily"
	if req.Schedule != nil {
		schedule = *req.Schedule
	}
	strategy := "standard"
	if req.Strategy != nil {
		strategy = *req.Strategy
	}
	excludeNSFW := true
	if req.ExcludeNSFW != nil {
		excludeNSFW = *req.ExcludeNSFW
	}
	retentionDays := 30
	if req.RetentionDays != nil {
		retentionDays = *req.RetentionDays
	}

	var seedClipID *uuid.UUID
	if req.SeedClipID != nil {
		parsed, err := uuid.Parse(*req.SeedClipID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid seed_clip_id", ErrPlaylistScriptValidation)
		}
		seedClipID = &parsed
	}

	script := &models.PlaylistScript{
		ID:              uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		Sort:            req.Sort,
		Timeframe:       req.Timeframe,
		ClipLimit:       req.ClipLimit,
		Visibility:      visibility,
		IsActive:        isActive,
		Schedule:        schedule,
		Strategy:        strategy,
		GameID:          req.GameID,
		GameIDs:         req.GameIDs,
		BroadcasterID:   req.BroadcasterID,
		Tag:             req.Tag,
		ExcludeTags:     req.ExcludeTags,
		Language:        req.Language,
		MinVoteScore:    req.MinVoteScore,
		MinViewCount:    req.MinViewCount,
		ExcludeNSFW:     excludeNSFW,
		Top10kStreamers: req.Top10kStreamers != nil && *req.Top10kStreamers,
		SeedClipID:      seedClipID,
		RetentionDays:   retentionDays,
		TitleTemplate:   req.TitleTemplate,
		CreatedBy:       &userID,
	}
	if err := validatePlaylistScriptStrategy(script); err != nil {
		return nil, err
	}

	if err := s.scriptRepo.Create(ctx, script); err != nil {
		return nil, err
	}

	return script, nil
}

// UpdateScript updates an existing playlist script
func (s *PlaylistScriptService) UpdateScript(ctx context.Context, scriptID uuid.UUID, req *models.UpdatePlaylistScriptRequest) (*models.PlaylistScript, error) {
	if req == nil || !hasPlaylistScriptUpdate(req) {
		return nil, fmt.Errorf("%w: at least one field is required", ErrPlaylistScriptValidation)
	}
	script, err := s.scriptRepo.GetByID(ctx, scriptID)
	if err != nil {
		return nil, err
	}
	if script == nil {
		return nil, ErrPlaylistScriptNotFound
	}

	if req.Name != nil {
		script.Name = *req.Name
	}
	if req.Description != nil {
		script.Description = req.Description
	}
	if req.Sort != nil {
		script.Sort = *req.Sort
	}
	if req.Timeframe != nil {
		script.Timeframe = req.Timeframe
	}
	if req.ClipLimit != nil {
		script.ClipLimit = *req.ClipLimit
	}
	if req.Visibility != nil {
		script.Visibility = *req.Visibility
	}
	if req.IsActive != nil {
		script.IsActive = *req.IsActive
	}
	if req.Schedule != nil {
		script.Schedule = *req.Schedule
	}
	if req.Strategy != nil {
		script.Strategy = *req.Strategy
	}
	if req.GameID != nil {
		script.GameID = req.GameID
	}
	if req.GameIDs != nil {
		script.GameIDs = req.GameIDs
	}
	if req.BroadcasterID != nil {
		script.BroadcasterID = req.BroadcasterID
	}
	if req.Tag != nil {
		script.Tag = req.Tag
	}
	if req.ExcludeTags != nil {
		script.ExcludeTags = req.ExcludeTags
	}
	if req.Language != nil {
		script.Language = req.Language
	}
	if req.MinVoteScore != nil {
		script.MinVoteScore = req.MinVoteScore
	}
	if req.MinViewCount != nil {
		script.MinViewCount = req.MinViewCount
	}
	if req.ExcludeNSFW != nil {
		script.ExcludeNSFW = *req.ExcludeNSFW
	}
	if req.Top10kStreamers != nil {
		script.Top10kStreamers = *req.Top10kStreamers
	}
	if req.SeedClipID != nil {
		parsed, err := uuid.Parse(*req.SeedClipID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid seed_clip_id", ErrPlaylistScriptValidation)
		}
		script.SeedClipID = &parsed
	}
	if req.RetentionDays != nil {
		script.RetentionDays = *req.RetentionDays
	}
	if req.TitleTemplate != nil {
		script.TitleTemplate = req.TitleTemplate
	}
	if strings.TrimSpace(script.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", ErrPlaylistScriptValidation)
	}
	if err := validatePlaylistScriptStrategy(script); err != nil {
		return nil, err
	}

	if err := s.scriptRepo.Update(ctx, script); err != nil {
		return nil, err
	}

	return script, nil
}

func validatePlaylistScriptStrategy(script *models.PlaylistScript) error {
	switch script.Strategy {
	case "similar_vibes":
		if script.SeedClipID == nil {
			return fmt.Errorf("%w: similar_vibes requires seed_clip_id", ErrPlaylistScriptValidation)
		}
	case "cross_game_hits":
		if len(script.GameIDs) == 0 {
			return fmt.Errorf("%w: cross_game_hits requires game_ids", ErrPlaylistScriptValidation)
		}
	case "twitch_top_game":
		if script.GameID == nil || strings.TrimSpace(*script.GameID) == "" {
			return fmt.Errorf("%w: twitch_top_game requires game_id", ErrPlaylistScriptValidation)
		}
	case "twitch_top_broadcaster":
		if script.BroadcasterID == nil || strings.TrimSpace(*script.BroadcasterID) == "" {
			return fmt.Errorf("%w: twitch_top_broadcaster requires broadcaster_id", ErrPlaylistScriptValidation)
		}
	}
	return nil
}

func hasPlaylistScriptUpdate(req *models.UpdatePlaylistScriptRequest) bool {
	return req.Name != nil || req.Description != nil || req.Sort != nil || req.Timeframe != nil ||
		req.ClipLimit != nil || req.Visibility != nil || req.IsActive != nil || req.Schedule != nil ||
		req.Strategy != nil || req.GameID != nil || req.GameIDs != nil || req.BroadcasterID != nil ||
		req.Tag != nil || req.ExcludeTags != nil || req.Language != nil || req.MinVoteScore != nil ||
		req.MinViewCount != nil || req.ExcludeNSFW != nil || req.Top10kStreamers != nil ||
		req.SeedClipID != nil || req.RetentionDays != nil || req.TitleTemplate != nil
}

// DeleteScript removes a playlist script
func (s *PlaylistScriptService) DeleteScript(ctx context.Context, scriptID uuid.UUID) error {
	return s.scriptRepo.Delete(ctx, scriptID)
}

// GeneratePlaylist creates a playlist from a script using either standard filters or a curation strategy.
func (s *PlaylistScriptService) GeneratePlaylist(ctx context.Context, scriptID uuid.UUID) (*models.Playlist, error) {
	script, err := s.scriptRepo.GetByID(ctx, scriptID)
	if err != nil {
		return nil, err
	}
	if script == nil {
		return nil, ErrPlaylistScriptNotFound
	}
	if !script.IsActive {
		return nil, ErrPlaylistScriptInactive
	}

	var clips []models.Clip

	if script.Strategy == "" || script.Strategy == "standard" {
		// Standard strategy: use ClipFilters + ListWithFilters
		filters := buildFiltersFromScript(script)
		clips, _, err = s.clipRepo.ListWithFilters(ctx, filters, script.ClipLimit, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch clips: %w", err)
		}
	} else {
		// Advanced strategy: delegate to curation repository
		clips, err = s.executeStrategy(ctx, script)
		if err != nil {
			return nil, fmt.Errorf("strategy %s failed: %w", script.Strategy, err)
		}
	}

	ownerID := uuid.Nil
	if script.CreatedBy != nil {
		ownerID = *script.CreatedBy
	}
	if ownerID == uuid.Nil {
		return nil, fmt.Errorf("invalid script owner")
	}
	if len(clips) == 0 {
		return nil, ErrPlaylistGenerationEmpty
	}

	title := buildPlaylistTitle(script)
	presentation := generatedPlaylistPresentationForScript(script, ownerID)
	playlist := &models.Playlist{
		ID:           uuid.New(),
		UserID:       ownerID,
		Title:        title,
		Description:  script.Description,
		Visibility:   script.Visibility,
		IsCurated:    presentation.isCurated,
		IsFeatured:   presentation.isFeatured,
		DisplayOrder: presentation.displayOrder,
		ScriptID:     &script.ID,
	}

	if s.generationWriter == nil {
		return nil, fmt.Errorf("playlist generation persistence is unavailable")
	}
	if err := s.generationWriter.Persist(ctx, script, playlist, clips); err != nil {
		return nil, fmt.Errorf("failed to persist generated playlist: %w", err)
	}

	return playlist, nil
}

func generatedPlaylistPresentationForScript(script *models.PlaylistScript, ownerID uuid.UUID) generatedPlaylistPresentation {
	if script == nil || ownerID != BotUserID || script.Visibility != models.PlaylistVisibilityPublic {
		return generatedPlaylistPresentation{}
	}

	presentation := generatedPlaylistPresentation{isCurated: true, displayOrder: 999}
	if order, ok := siteFreshnessDisplayOrder[script.Name]; ok {
		presentation.isFeatured = true
		presentation.displayOrder = order
	}

	return presentation
}

func shouldAutoCurateGeneratedPlaylist(script *models.PlaylistScript, ownerID uuid.UUID) bool {
	return generatedPlaylistPresentationForScript(script, ownerID).isCurated
}

// ListDueForExecution returns scripts that are due for scheduled execution.
func (s *PlaylistScriptService) ListDueForExecution(ctx context.Context) ([]*models.PlaylistScript, error) {
	return s.scriptRepo.ListDueForExecution(ctx)
}

// DeleteStaleGeneratedPlaylists removes generated playlists past their retention period.
func (s *PlaylistScriptService) DeleteStaleGeneratedPlaylists(ctx context.Context) (int64, error) {
	return s.scriptRepo.DeleteStaleGeneratedPlaylists(ctx)
}

// buildFiltersFromScript converts script fields into ClipFilters for the standard strategy.
func buildFiltersFromScript(script *models.PlaylistScript) repository.ClipFilters {
	filters := repository.ClipFilters{
		Sort:            script.Sort,
		Timeframe:       script.Timeframe,
		GameID:          script.GameID,
		BroadcasterID:   script.BroadcasterID,
		Tag:             script.Tag,
		ExcludeTags:     script.ExcludeTags,
		Language:        script.Language,
		Top10kStreamers: script.Top10kStreamers,
	}
	return filters
}

// buildPlaylistTitle generates a title from the script's template or falls back to a default.
// Supported placeholders: {name}, {date}, {day}, {week_start}, {month}
func buildPlaylistTitle(script *models.PlaylistScript) string {
	now := time.Now()

	if script.TitleTemplate != nil && *script.TitleTemplate != "" {
		title := *script.TitleTemplate
		title = strings.ReplaceAll(title, "{name}", script.Name)
		title = strings.ReplaceAll(title, "{date}", now.Format("Jan 2, 2006"))
		title = strings.ReplaceAll(title, "{day}", now.Format("Monday"))
		// week_start = most recent Monday
		weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday+7)%7)
		title = strings.ReplaceAll(title, "{week_start}", weekStart.Format("Jan 2"))
		title = strings.ReplaceAll(title, "{month}", now.Format("January 2006"))
		return title
	}

	return fmt.Sprintf("%s \u2022 %s", script.Name, now.Format("Jan 2, 2006"))
}
