package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// ErrUnauthorized is returned when a user doesn't have permission to manage a clip
var ErrUnauthorized = errors.New("user does not have permission to manage this clip")

// ClipService handles business logic for clips
type ClipService struct {
	clipRepo            *repository.ClipRepository
	discoveryClipRepo   *repository.DiscoveryClipRepository
	voteRepo            *repository.VoteRepository
	favoriteRepo        *repository.FavoriteRepository
	userRepo            *repository.UserRepository
	watchHistoryRepo    *repository.WatchHistoryRepository
	redisClient         *redispkg.Client
	auditLogRepo        *repository.AuditLogRepository
	notificationService *NotificationService
}

// NewClipService creates a new ClipService
func NewClipService(
	clipRepo *repository.ClipRepository,
	discoveryClipRepo *repository.DiscoveryClipRepository,
	voteRepo *repository.VoteRepository,
	favoriteRepo *repository.FavoriteRepository,
	userRepo *repository.UserRepository,
	watchHistoryRepo *repository.WatchHistoryRepository,
	redisClient *redispkg.Client,
	auditLogRepo *repository.AuditLogRepository,
	notificationService *NotificationService,
) *ClipService {
	return &ClipService{
		clipRepo:            clipRepo,
		discoveryClipRepo:   discoveryClipRepo,
		voteRepo:            voteRepo,
		favoriteRepo:        favoriteRepo,
		userRepo:            userRepo,
		watchHistoryRepo:    watchHistoryRepo,
		redisClient:         redisClient,
		auditLogRepo:        auditLogRepo,
		notificationService: notificationService,
	}
}

// ClipWithUserData represents a clip with user-specific data
type ClipWithUserData struct {
	models.Clip
	UserVote      *int16                    `json:"user_vote,omitempty"`
	IsFavorited   bool                      `json:"is_favorited"`
	UpvoteCount   int                       `json:"upvote_count"`
	DownvoteCount int                       `json:"downvote_count"`
	SubmittedBy   *models.ClipSubmitterInfo `json:"submitted_by,omitempty"`
}

// buildWatchProgressInfo creates a WatchProgressInfo from resume position data
func (s *ClipService) buildWatchProgressInfo(progressSeconds int, completed bool, duration *float64) *models.WatchProgressInfo {
	if progressSeconds <= 0 {
		return nil
	}

	var progressPercent float64
	durationSeconds := 0
	if duration != nil && *duration > 0 {
		durationSeconds = int(*duration)
		progressPercent = (float64(progressSeconds) / float64(durationSeconds)) * 100
	}

	return &models.WatchProgressInfo{
		ProgressSeconds: progressSeconds,
		DurationSeconds: durationSeconds,
		ProgressPercent: progressPercent,
		Completed:       completed,
		// WatchedAt omitted for performance reasons
	}
}

// GetClip retrieves a single clip with user data
func (s *ClipService) GetClip(ctx context.Context, clipID uuid.UUID, userID *uuid.UUID) (*ClipWithUserData, error) {
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return nil, err
	}

	clipWithData := &ClipWithUserData{
		Clip: *clip,
	}

	// Get vote counts
	upvotes, downvotes, err := s.voteRepo.GetVoteCounts(ctx, clipID)
	if err == nil {
		clipWithData.UpvoteCount = upvotes
		clipWithData.DownvoteCount = downvotes
	}

	// Get submitter information if available
	if clip.SubmittedByUserID != nil {
		submitter, err := s.userRepo.GetByID(ctx, *clip.SubmittedByUserID)
		if err == nil && submitter != nil {
			clipWithData.SubmittedBy = &models.ClipSubmitterInfo{
				ID:          submitter.ID,
				Username:    submitter.Username,
				DisplayName: submitter.DisplayName,
				AvatarURL:   submitter.AvatarURL,
			}
		}
	}

	// Get user-specific data if authenticated
	if userID != nil {
		vote, err := s.voteRepo.GetVote(ctx, *userID, clipID)
		if err == nil && vote != nil {
			clipWithData.UserVote = &vote.VoteType
		}

		isFavorited, err := s.favoriteRepo.IsFavorited(ctx, *userID, clipID)
		if err == nil {
			clipWithData.IsFavorited = isFavorited
		}

		// Get watch progress
		progressSeconds, completed, err := s.watchHistoryRepo.GetResumePosition(ctx, *userID, clipID)
		if err == nil {
			clipWithData.Clip.WatchProgress = s.buildWatchProgressInfo(progressSeconds, completed, clip.Duration)
		}
	}

	// Increment view count and check for threshold notifications (async, don't block on errors)
	go func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		newViewCount, err := s.clipRepo.IncrementViewCount(timeoutCtx, clipID)
		if err == nil && clip.CreatorID != nil && s.notificationService != nil {
			// Check if we reached a view threshold
			_ = s.notificationService.NotifyClipViewThreshold(timeoutCtx, clipID, newViewCount, *clip.CreatorID)
		}
	}()

	return clipWithData, nil
}

// GetClipByTwitchID retrieves a single clip by Twitch clip ID with user data
func (s *ClipService) GetClipByTwitchID(ctx context.Context, twitchClipID string, userID *uuid.UUID) (*ClipWithUserData, error) {
	clip, err := s.clipRepo.GetByTwitchClipID(ctx, twitchClipID)
	if err != nil {
		return nil, err
	}

	clipWithData := &ClipWithUserData{
		Clip: *clip,
	}

	// Get vote counts
	upvotes, downvotes, err := s.voteRepo.GetVoteCounts(ctx, clip.ID)
	if err == nil {
		clipWithData.UpvoteCount = upvotes
		clipWithData.DownvoteCount = downvotes
	}

	// Get submitter information if available
	if clip.SubmittedByUserID != nil {
		submitter, err := s.userRepo.GetByID(ctx, *clip.SubmittedByUserID)
		if err == nil && submitter != nil {
			clipWithData.SubmittedBy = &models.ClipSubmitterInfo{
				ID:          submitter.ID,
				Username:    submitter.Username,
				DisplayName: submitter.DisplayName,
				AvatarURL:   submitter.AvatarURL,
			}
		}
	}

	// Get user-specific data if authenticated
	if userID != nil {
		vote, err := s.voteRepo.GetVote(ctx, *userID, clip.ID)
		if err == nil && vote != nil {
			clipWithData.UserVote = &vote.VoteType
		}

		isFavorited, err := s.favoriteRepo.IsFavorited(ctx, *userID, clip.ID)
		if err == nil {
			clipWithData.IsFavorited = isFavorited
		}

		// Get watch progress
		progressSeconds, completed, err := s.watchHistoryRepo.GetResumePosition(ctx, *userID, clip.ID)
		if err == nil {
			clipWithData.Clip.WatchProgress = s.buildWatchProgressInfo(progressSeconds, completed, clip.Duration)
		}
	}

	// Increment view count and check for threshold notifications (async, don't block on errors)
	go func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		newViewCount, err := s.clipRepo.IncrementViewCount(timeoutCtx, clip.ID)
		if err == nil && clip.CreatorID != nil && s.notificationService != nil {
			// Check if we reached a view threshold
			_ = s.notificationService.NotifyClipViewThreshold(timeoutCtx, clip.ID, newViewCount, *clip.CreatorID)
		}
	}()

	return clipWithData, nil
}

// ListClips retrieves clips with filters and pagination
func (s *ClipService) ListClips(ctx context.Context, filters repository.ClipFilters, page, limit int, userID *uuid.UUID) ([]ClipWithUserData, int, error) {
	// Check cache for non-user-specific queries
	cacheKey := s.buildCacheKey(filters, page, limit)
	var cachedClips []models.Clip
	var cachedTotal int

	if userID == nil {
		cached, err := s.redisClient.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var cacheData struct {
				Clips []models.Clip `json:"clips"`
				Total int           `json:"total"`
			}
			if json.Unmarshal([]byte(cached), &cacheData) == nil {
				cachedClips = cacheData.Clips
				cachedTotal = cacheData.Total
			}
		}
	}

	var clips []models.Clip
	var total int
	var err error

	if cachedClips != nil {
		clips = cachedClips
		total = cachedTotal
	} else {
		offset := (page - 1) * limit
		clips, total, err = s.clipRepo.ListWithFilters(ctx, filters, limit, offset)
		if err != nil {
			return nil, 0, err
		}

		// Cache non-user-specific results
		if userID == nil {
			cacheData := struct {
				Clips []models.Clip `json:"clips"`
				Total int           `json:"total"`
			}{
				Clips: clips,
				Total: total,
			}
			if data, err := json.Marshal(cacheData); err == nil {
				ttl := s.getCacheTTL(filters.Sort)
				_ = s.redisClient.Set(ctx, cacheKey, string(data), ttl)
			}
		}
	}

	// Collect unique submitter IDs for batch fetching
	submitterIDSet := make(map[uuid.UUID]struct{})
	for _, clip := range clips {
		if clip.SubmittedByUserID != nil {
			submitterIDSet[*clip.SubmittedByUserID] = struct{}{}
		}
	}

	// Convert set to slice for batch query
	submitterIDs := make([]uuid.UUID, 0, len(submitterIDSet))
	for id := range submitterIDSet {
		submitterIDs = append(submitterIDs, id)
	}

	// Batch fetch submitter information in a single query
	submitters := make(map[uuid.UUID]*models.ClipSubmitterInfo)
	if len(submitterIDs) > 0 {
		users, err := s.userRepo.GetByIDs(ctx, submitterIDs)
		if err == nil {
			for _, submitter := range users {
				submitters[submitter.ID] = &models.ClipSubmitterInfo{
					ID:          submitter.ID,
					Username:    submitter.Username,
					DisplayName: submitter.DisplayName,
					AvatarURL:   submitter.AvatarURL,
				}
			}
		}
	}

	// Batch fetch watch progress for all clips if user is authenticated
	var watchProgressMap map[uuid.UUID]*models.ResumePositionResponse
	if userID != nil && len(clips) > 0 {
		clipIDs := make([]uuid.UUID, len(clips))
		for i, clip := range clips {
			clipIDs[i] = clip.ID
		}
		watchProgressMap, _ = s.watchHistoryRepo.GetResumePositions(ctx, *userID, clipIDs)
	}

	// Enrich with user data
	clipsWithData := make([]ClipWithUserData, len(clips))
	for i, clip := range clips {
		clipsWithData[i] = ClipWithUserData{
			Clip: clip,
		}

		// Add submitter info if available
		if clip.SubmittedByUserID != nil {
			if submitter, ok := submitters[*clip.SubmittedByUserID]; ok {
				clipsWithData[i].SubmittedBy = submitter
			}
		}

		// Get vote counts
		upvotes, downvotes, err := s.voteRepo.GetVoteCounts(ctx, clip.ID)
		if err == nil {
			clipsWithData[i].UpvoteCount = upvotes
			clipsWithData[i].DownvoteCount = downvotes
		}

		// Get user-specific data if authenticated
		if userID != nil {
			vote, err := s.voteRepo.GetVote(ctx, *userID, clip.ID)
			if err == nil && vote != nil {
				clipsWithData[i].UserVote = &vote.VoteType
			}

			isFavorited, err := s.favoriteRepo.IsFavorited(ctx, *userID, clip.ID)
			if err == nil {
				clipsWithData[i].IsFavorited = isFavorited
			}

			// Add watch progress if available
			if watchProgress, ok := watchProgressMap[clip.ID]; ok && watchProgress.HasProgress {
				clipsWithData[i].Clip.WatchProgress = s.buildWatchProgressInfo(
					watchProgress.ProgressSeconds,
					watchProgress.Completed,
					clip.Duration,
				)
			}
		}
	}

	return clipsWithData, total, nil
}

// ListScrapedClips retrieves discovery clips (not yet claimed by users) with filters and pagination.
// Now reads from the discovery_clips staging table via DiscoveryClipRepository.
func (s *ClipService) ListScrapedClips(ctx context.Context, filters repository.ClipFilters, page, limit int, userID *uuid.UUID) ([]ClipWithUserData, int, error) {
	// Convert ClipFilters → DiscoveryClipFilters
	dcFilters := repository.DiscoveryClipFilters{
		GameID:          filters.GameID,
		BroadcasterID:   filters.BroadcasterID,
		CreatorID:       filters.CreatorID,
		Tag:             filters.Tag,
		ExcludeTags:     filters.ExcludeTags,
		Search:          filters.Search,
		Language:        filters.Language,
		Timeframe:       filters.Timeframe,
		DateFrom:        filters.DateFrom,
		DateTo:          filters.DateTo,
		Sort:            filters.Sort,
		Top10kStreamers: filters.Top10kStreamers,
	}

	// Check cache for non-user-specific queries
	cacheKey := s.buildCacheKey(filters, page, limit) + ":discovery"
	var cachedClips []models.DiscoveryClip
	var cachedTotal int

	if userID == nil {
		cached, err := s.redisClient.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var cacheData struct {
				Clips []models.DiscoveryClip `json:"clips"`
				Total int                    `json:"total"`
			}
			if json.Unmarshal([]byte(cached), &cacheData) == nil {
				cachedClips = cacheData.Clips
				cachedTotal = cacheData.Total
			}
		}
	}

	var discoveryClips []models.DiscoveryClip
	var total int
	var err error

	if cachedClips != nil {
		discoveryClips = cachedClips
		total = cachedTotal
	} else {
		offset := (page - 1) * limit
		discoveryClips, total, err = s.discoveryClipRepo.ListWithFilters(ctx, dcFilters, limit, offset)
		if err != nil {
			return nil, 0, err
		}

		// Cache non-user-specific results
		if userID == nil {
			cacheData := struct {
				Clips []models.DiscoveryClip `json:"clips"`
				Total int                    `json:"total"`
			}{
				Clips: discoveryClips,
				Total: total,
			}
			if data, err := json.Marshal(cacheData); err == nil {
				ttl := s.getCacheTTL(filters.Sort)
				_ = s.redisClient.Set(ctx, cacheKey, string(data), ttl)
			}
		}
	}

	// Convert DiscoveryClip → ClipWithUserData for backward-compatible API response
	clipsWithData := make([]ClipWithUserData, len(discoveryClips))
	for i, dc := range discoveryClips {
		clipsWithData[i] = ClipWithUserData{
			Clip: models.Clip{
				ID:              dc.ID,
				TwitchClipID:    dc.TwitchClipID,
				TwitchClipURL:   dc.TwitchClipURL,
				EmbedURL:        dc.EmbedURL,
				Title:           dc.Title,
				CreatorName:     dc.CreatorName,
				CreatorID:       dc.CreatorID,
				BroadcasterName: dc.BroadcasterName,
				BroadcasterID:   dc.BroadcasterID,
				GameID:          dc.GameID,
				GameName:        dc.GameName,
				Language:        dc.Language,
				ThumbnailURL:    dc.ThumbnailURL,
				Duration:        dc.Duration,
				ViewCount:       dc.ViewCount,
				CreatedAt:       dc.CreatedAt,
				ImportedAt:      dc.ImportedAt,
				IsNSFW:          dc.IsNSFW,
				IsRemoved:       dc.IsRemoved,
				IsHidden:        dc.IsHidden,
			},
		}
	}

	return clipsWithData, total, nil
}

// VoteOnClip handles voting on a clip
func (s *ClipService) VoteOnClip(ctx context.Context, userID, clipID uuid.UUID, voteType int16) error {
	// Validate vote type
	if voteType != -1 && voteType != 0 && voteType != 1 {
		return fmt.Errorf("invalid vote type: must be -1, 0, or 1")
	}

	// Check if clip exists
	_, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return err
	}

	// Get old vote if exists
	oldVote, _ := s.voteRepo.GetVote(ctx, userID, clipID)

	if voteType == 0 {
		if oldVote != nil {
			if err := s.voteRepo.DeleteVote(ctx, userID, clipID); err != nil {
				return err
			}
		}
		return nil
	}

	// Calculate if this vote will increase the score (only notify on increases)
	scoreWillIncrease := false
	if oldVote == nil {
		scoreWillIncrease = (voteType == 1)
	} else if oldVote.VoteType != voteType {
		scoreWillIncrease = (oldVote.VoteType == -1 && voteType == 1)
	}

	// Upsert vote
	err = s.voteRepo.UpsertVote(ctx, userID, clipID, voteType)
	if err != nil {
		return err
	}

	// Only check for vote thresholds if the score increased
	if scoreWillIncrease {
		clip, err := s.clipRepo.GetByID(ctx, clipID)
		if err == nil && clip.CreatorID != nil && s.notificationService != nil {
			// Check if we reached a vote threshold (async with timeout)
			go func() {
				timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_ = s.notificationService.NotifyClipVoteThreshold(timeoutCtx, clipID, clip.VoteScore, *clip.CreatorID)
			}()
		}
	}

	// Update user karma (async)
	go func() {
		karmaChange := 0
		if oldVote == nil {
			// New vote
			if voteType == 1 {
				karmaChange = 1
			} else {
				karmaChange = -1
			}
		} else {
			// Changed vote
			if oldVote.VoteType == 1 && voteType == -1 {
				karmaChange = -2
			} else if oldVote.VoteType == -1 && voteType == 1 {
				karmaChange = 2
			}
		}

		if karmaChange != 0 {
			_ = s.userRepo.UpdateKarma(context.Background(), userID, karmaChange)
		}
	}()

	// Invalidate cache
	s.invalidateCache(ctx)

	return nil
}

// AddFavorite adds a clip to user's favorites
func (s *ClipService) AddFavorite(ctx context.Context, userID, clipID uuid.UUID) error {
	// Check if clip exists
	_, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return err
	}

	return s.favoriteRepo.Create(ctx, userID, clipID)
}

// RemoveFavorite removes a clip from user's favorites
func (s *ClipService) RemoveFavorite(ctx context.Context, userID, clipID uuid.UUID) error {
	return s.favoriteRepo.Delete(ctx, userID, clipID)
}

// GetRelatedClips retrieves related clips
func (s *ClipService) GetRelatedClips(ctx context.Context, clipID uuid.UUID, limit int) ([]models.Clip, error) {
	return s.clipRepo.GetRelated(ctx, clipID, limit)
}

// UpdateClip updates clip properties (admin only)
func (s *ClipService) UpdateClip(ctx context.Context, clipID uuid.UUID, updates map[string]interface{}) error {
	// Validate allowed fields
	allowedFields := map[string]bool{
		"is_featured":    true,
		"is_nsfw":        true,
		"is_removed":     true,
		"removed_reason": true,
	}

	for field := range updates {
		if !allowedFields[field] {
			return fmt.Errorf("field '%s' is not allowed to be updated", field)
		}
	}

	err := s.clipRepo.Update(ctx, clipID, updates)
	if err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateCache(ctx)

	return nil
}

// DeleteClip soft deletes a clip (admin only)
func (s *ClipService) DeleteClip(ctx context.Context, clipID uuid.UUID, reason string) error {
	err := s.clipRepo.SoftDelete(ctx, clipID, reason)
	if err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateCache(ctx)

	return nil
}

// Helper functions

func (s *ClipService) buildCacheKey(filters repository.ClipFilters, page, limit int) string {
	key := fmt.Sprintf("clips:list:%s:page:%d:limit:%d", filters.Sort, page, limit)

	if filters.GameID != nil {
		key += fmt.Sprintf(":game:%s", *filters.GameID)
	}
	if filters.BroadcasterID != nil {
		key += fmt.Sprintf(":broadcaster:%s", *filters.BroadcasterID)
	}
	if filters.CreatorID != nil {
		key += fmt.Sprintf(":creator:%s", *filters.CreatorID)
	}
	if filters.Tag != nil {
		key += fmt.Sprintf(":tag:%s", *filters.Tag)
	}
	if filters.Search != nil {
		key += fmt.Sprintf(":search:%s", *filters.Search)
	}
	if filters.Timeframe != nil {
		key += fmt.Sprintf(":timeframe:%s", *filters.Timeframe)
	}
	if filters.Language != nil {
		key += fmt.Sprintf(":language:%s", *filters.Language)
	}

	key += fmt.Sprintf(":top10k:%t", filters.Top10kStreamers)
	key += fmt.Sprintf(":show_hidden:%t", filters.ShowHidden)
	key += fmt.Sprintf(":user_submitted_only:%t", filters.UserSubmittedOnly)

	return key
}

func (s *ClipService) getCacheTTL(sort string) time.Duration {
	switch sort {
	case "hot":
		return 5 * time.Minute
	case "new":
		return 2 * time.Minute
	case "top":
		return 15 * time.Minute
	case "rising":
		return 3 * time.Minute
	default:
		return 5 * time.Minute
	}
}

func (s *ClipService) invalidateCache(ctx context.Context) {
	// Invalidate all clip list caches
	pattern := "clips:list:*"
	_ = s.redisClient.DeletePattern(ctx, pattern)
}

// CanManageClip checks if a user can manage a specific clip
func (s *ClipService) CanManageClip(ctx context.Context, userID uuid.UUID, clipID uuid.UUID) (bool, error) {
	// Get user to check role
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// Admins and moderators can manage any clip
	if user.Role == "admin" || user.Role == "moderator" {
		return true, nil
	}

	// Get clip to check creator
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return false, fmt.Errorf("failed to get clip: %w", err)
	}

	// Check if user is the creator (by matching Twitch ID)
	if clip.CreatorID != nil && user.TwitchID != nil && *user.TwitchID == *clip.CreatorID {
		return true, nil
	}

	return false, nil
}

// UpdateClipMetadata updates clip metadata (title) - only accessible by creator or admin
func (s *ClipService) UpdateClipMetadata(ctx context.Context, userID uuid.UUID, clipID uuid.UUID, title *string) error {
	// Check authorization
	canManage, err := s.CanManageClip(ctx, userID, clipID)
	if err != nil {
		return err
	}
	if !canManage {
		return ErrUnauthorized
	}

	// Update metadata
	err = s.clipRepo.UpdateMetadata(ctx, clipID, title)
	if err != nil {
		return err
	}

	// Log the change
	changes := make(map[string]interface{})
	if title != nil {
		changes["title"] = *title
	}

	if len(changes) > 0 {
		auditLog := &models.ModerationAuditLog{
			Action:      "clip_metadata_updated",
			EntityType:  "clip",
			EntityID:    clipID,
			ModeratorID: userID,
			Metadata:    changes,
		}
		_ = s.auditLogRepo.Create(ctx, auditLog)
	}

	// Invalidate cache
	s.invalidateCache(ctx)

	return nil
}

// UpdateClipVisibility updates clip visibility (hidden status) - only accessible by creator or admin
func (s *ClipService) UpdateClipVisibility(ctx context.Context, userID uuid.UUID, clipID uuid.UUID, isHidden bool) error {
	// Check authorization
	canManage, err := s.CanManageClip(ctx, userID, clipID)
	if err != nil {
		return err
	}
	if !canManage {
		return ErrUnauthorized
	}

	// Update visibility
	err = s.clipRepo.UpdateVisibility(ctx, clipID, isHidden)
	if err != nil {
		return err
	}

	// Log the change
	action := "clip_hidden"
	if !isHidden {
		action = "clip_unhidden"
	}

	auditLog := &models.ModerationAuditLog{
		Action:      action,
		EntityType:  "clip",
		EntityID:    clipID,
		ModeratorID: userID,
		Metadata:    map[string]interface{}{"is_hidden": isHidden},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	// Invalidate cache
	s.invalidateCache(ctx)

	return nil
}

// ListCreatorClips retrieves clips created by a specific creator (including hidden ones if the user is the creator)
func (s *ClipService) ListCreatorClips(ctx context.Context, creatorTwitchID string, userID *uuid.UUID, page, limit int) ([]ClipWithUserData, int, error) {
	offset := (page - 1) * limit

	// Determine if we should show hidden clips
	showHidden := false
	if userID != nil {
		user, err := s.userRepo.GetByID(ctx, *userID)
		if err == nil {
			// Show hidden clips if user is the creator or an admin/moderator
			if (user.TwitchID != nil && *user.TwitchID == creatorTwitchID) || user.Role == "admin" || user.Role == "moderator" {
				showHidden = true
			}
		}
	}

	// Get clips with creator filter
	filters := repository.ClipFilters{
		CreatorID:  &creatorTwitchID,
		Sort:       "new",
		ShowHidden: showHidden,
	}

	clips, total, err := s.clipRepo.ListWithFilters(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convert to ClipWithUserData
	clipsWithData := make([]ClipWithUserData, len(clips))
	for i, clip := range clips {
		clipsWithData[i] = ClipWithUserData{
			Clip: clip,
		}

		// Get vote counts
		upvotes, downvotes, err := s.voteRepo.GetVoteCounts(ctx, clip.ID)
		if err == nil {
			clipsWithData[i].UpvoteCount = upvotes
			clipsWithData[i].DownvoteCount = downvotes
		}

		// Get user-specific data if authenticated
		if userID != nil {
			vote, err := s.voteRepo.GetVote(ctx, *userID, clip.ID)
			if err == nil && vote != nil {
				clipsWithData[i].UserVote = &vote.VoteType
			}

			isFavorited, err := s.favoriteRepo.IsFavorited(ctx, *userID, clip.ID)
			if err == nil {
				clipsWithData[i].IsFavorited = isFavorited
			}
		}
	}

	return clipsWithData, total, nil
}

// ClipMediaInfo contains the media URLs for a clip
type ClipMediaInfo struct {
	ID           uuid.UUID `json:"id"`
	EmbedURL     string    `json:"embed_url"`
	ThumbnailURL *string   `json:"thumbnail_url,omitempty"`
}

// BatchGetClipMedia retrieves media URLs for multiple clips efficiently
func (s *ClipService) BatchGetClipMedia(ctx context.Context, clipIDs []uuid.UUID) ([]ClipMediaInfo, error) {
	if len(clipIDs) == 0 {
		return []ClipMediaInfo{}, nil
	}

	// Limit batch size to prevent abuse
	if len(clipIDs) > 100 {
		return nil, errors.New("batch size exceeds maximum of 100 clips")
	}

	clips, err := s.clipRepo.GetClipsByIDs(ctx, clipIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clips: %w", err)
	}

	// Build result slice with media info
	result := make([]ClipMediaInfo, 0, len(clips))
	for _, clip := range clips {
		result = append(result, ClipMediaInfo{
			ID:           clip.ID,
			EmbedURL:     clip.EmbedURL,
			ThumbnailURL: clip.ThumbnailURL,
		})
	}

	return result, nil
}
