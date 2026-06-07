package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/storage"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	pkgutils "git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
)

// SubmissionService handles clip submission business logic
type SubmissionService struct {
	submissionRepo          *repository.SubmissionRepository
	clipRepo                *repository.ClipRepository
	discoveryClipRepo       *repository.DiscoveryClipRepository
	userRepo                *repository.UserRepository
	voteRepo                *repository.VoteRepository
	auditLogRepo            *repository.AuditLogRepository
	twitchClient            *twitch.Client
	redisClient             *redispkg.Client
	notificationService     *NotificationService
	creatorModeration       CreatorModerationChecker
	abuseDetector           *SubmissionAbuseDetector
	moderationEvents        *ModerationEventService
	externalMetadataFetcher ExternalMetadataFetcher
	webhookService          *OutboundWebhookService
	cacheService            *CacheService
	clipStorage             storage.ClipStorage
	cfg                     *config.Config
	logger                  *pkgutils.StructuredLogger

	// Test fixture controls
	testFixturesEnabled      bool
	bypassRateLimits         bool
	allowDuplicateSubmission bool

	userLookupFn               func(context.Context, uuid.UUID) (*models.User, error)
	submissionCreateFn         func(context.Context, *models.ClipSubmission) error
	submissionLookupBySourceFn func(context.Context, string, string) (*models.ClipSubmission, error)
	clipCreateFn               func(context.Context, *models.Clip) error
}

// NewSubmissionService creates a new SubmissionService
func NewSubmissionService(
	submissionRepo *repository.SubmissionRepository,
	clipRepo *repository.ClipRepository,
	discoveryClipRepo *repository.DiscoveryClipRepository,
	userRepo *repository.UserRepository,
	voteRepo *repository.VoteRepository,
	auditLogRepo *repository.AuditLogRepository,
	twitchClient *twitch.Client,
	notificationService *NotificationService,
	redisClient *redispkg.Client,
	webhookService *OutboundWebhookService,
	cacheService *CacheService,
	cfg *config.Config,
) *SubmissionService {
	var abuseDetector *SubmissionAbuseDetector
	var moderationEvents *ModerationEventService
	testFixturesEnabled := strings.EqualFold(cfg.Server.Environment, "test") || cfg.Server.GinMode != "release" || strings.EqualFold(os.Getenv("ENABLE_TEST_FIXTURES"), "true") || strings.EqualFold(os.Getenv("E2E_TEST_MODE"), "true")
	bypassRateLimits := testFixturesEnabled && strings.EqualFold(os.Getenv("SUBMISSION_BYPASS_RATE_LIMIT"), "true")
	allowDuplicateSubmission := testFixturesEnabled && strings.EqualFold(os.Getenv("SUBMISSION_ALLOW_DUPLICATES"), "true")

	if redisClient != nil {
		abuseDetector = NewSubmissionAbuseDetector(redisClient)
		moderationEvents = NewModerationEventService(redisClient, notificationService)
	}

	return &SubmissionService{
		submissionRepo:           submissionRepo,
		clipRepo:                 clipRepo,
		discoveryClipRepo:        discoveryClipRepo,
		userRepo:                 userRepo,
		voteRepo:                 voteRepo,
		auditLogRepo:             auditLogRepo,
		twitchClient:             twitchClient,
		redisClient:              redisClient,
		notificationService:      notificationService,
		abuseDetector:            abuseDetector,
		moderationEvents:         moderationEvents,
		externalMetadataFetcher:  NewExternalMetadataFetcher(nil),
		webhookService:           webhookService,
		cacheService:             cacheService,
		cfg:                      cfg,
		logger:                   pkgutils.GetLogger(),
		testFixturesEnabled:      testFixturesEnabled,
		bypassRateLimits:         bypassRateLimits,
		allowDuplicateSubmission: allowDuplicateSubmission,
	}
}

// GetAbuseDetector returns the abuse detector instance
func (s *SubmissionService) GetAbuseDetector() *SubmissionAbuseDetector {
	return s.abuseDetector
}

// GetModerationEventService returns the moderation event service instance
func (s *SubmissionService) GetModerationEventService() *ModerationEventService {
	return s.moderationEvents
}

// SetClipStorage configures the storage backend used for upload approvals.
func (s *SubmissionService) SetClipStorage(clipStorage storage.ClipStorage) {
	s.clipStorage = clipStorage
}

// SetCreatorModerationService configures creator-scoped moderation checks.
func (s *SubmissionService) SetCreatorModerationService(creatorModeration CreatorModerationChecker) {
	s.creatorModeration = creatorModeration
}

func (s *SubmissionService) requireCreatorSubmissionPermission(ctx context.Context, creatorAccountID *uuid.UUID, userID uuid.UUID) error {
	if s.creatorModeration == nil || creatorAccountID == nil || *creatorAccountID == uuid.Nil || userID == uuid.Nil {
		return nil
	}

	allowed, message, err := s.creatorModeration.CanSubmit(ctx, *creatorAccountID, userID)
	if err != nil {
		return err
	}
	if !allowed {
		return &CreatorModerationError{Message: message}
	}
	return nil
}

func (s *SubmissionService) shouldAutoUpvoteClaimedClip(ctx context.Context, creatorAccountID *uuid.UUID, userID uuid.UUID) bool {
	if s.creatorModeration == nil || creatorAccountID == nil || *creatorAccountID == uuid.Nil || userID == uuid.Nil {
		return true
	}

	allowed, _, err := s.creatorModeration.CanInteract(ctx, *creatorAccountID, userID)
	if err != nil {
		log.Printf("Warning: failed to check claimed clip interaction permission for user %s: %v\n", userID, err)
		return false
	}

	return allowed
}

func (s *SubmissionService) getUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if s.userLookupFn != nil {
		return s.userLookupFn(ctx, userID)
	}
	if s.userRepo == nil {
		return nil, fmt.Errorf("user repository is not configured")
	}
	return s.userRepo.GetByID(ctx, userID)
}

func (s *SubmissionService) createSubmissionRecord(ctx context.Context, submission *models.ClipSubmission) error {
	if s.submissionCreateFn != nil {
		return s.submissionCreateFn(ctx, submission)
	}
	if s.submissionRepo == nil {
		return fmt.Errorf("submission repository is not configured")
	}
	return s.submissionRepo.Create(ctx, submission)
}

func (s *SubmissionService) createClipRecord(ctx context.Context, clip *models.Clip) error {
	if s.clipCreateFn != nil {
		return s.clipCreateFn(ctx, clip)
	}
	if s.clipRepo == nil {
		return fmt.Errorf("clip repository is not configured")
	}
	return s.clipRepo.Create(ctx, clip)
}

func (s *SubmissionService) getSubmissionBySourceIdentity(ctx context.Context, sourcePlatform, sourceID string) (*models.ClipSubmission, error) {
	if s.submissionLookupBySourceFn != nil {
		return s.submissionLookupBySourceFn(ctx, sourcePlatform, sourceID)
	}
	if s.submissionRepo == nil {
		return nil, nil
	}
	return s.submissionRepo.GetBySourcePlatformAndID(ctx, sourcePlatform, sourceID)
}

func (s *SubmissionService) trackDuplicateAttempt(ctx context.Context, userID uuid.UUID, ip string, clipID string) {
	if s.abuseDetector != nil {
		if err := s.abuseDetector.TrackDuplicateAttempt(ctx, userID, ip, clipID); err != nil {
			log.Printf("Failed to track duplicate attempt: %v", err)
		}
	}
}

func (s *SubmissionService) emitDuplicateModerationEvent(ctx context.Context, userID uuid.UUID, ip string, metadata map[string]interface{}) {
	if s.moderationEvents != nil {
		if err := s.moderationEvents.EmitAbuseEvent(ctx, ModerationEventSubmissionDuplicate, userID, ip, metadata); err != nil {
			log.Printf("Failed to emit duplicate event: %v", err)
		}
	}
}

func (s *SubmissionService) rejectDuplicateExternalSubmission(ctx context.Context, userID uuid.UUID, ip string, existing *models.ClipSubmission) error {
	sourceID := ""
	if existing.SourceID != nil {
		sourceID = *existing.SourceID
	}

	s.trackDuplicateAttempt(ctx, userID, ip, sourceID)

	switch existing.Status {
	case "pending":
		metadata := map[string]interface{}{
			"source_platform": existing.SourcePlatform,
			"source_id":       sourceID,
			"reason":          "submission_pending",
			"submission_id":   existing.ID.String(),
		}
		s.emitDuplicateModerationEvent(ctx, userID, ip, metadata)
		return &ValidationError{
			Field:   "clip_url",
			Message: "This clip is already pending review. You'll be notified once it's been reviewed by our moderators.",
		}
	case "approved":
		metadata := map[string]interface{}{
			"source_platform": existing.SourcePlatform,
			"source_id":       sourceID,
			"reason":          "source_already_exists",
			"submission_id":   existing.ID.String(),
		}
		if existing.ClipID != nil {
			metadata["existing_clip_id"] = existing.ClipID.String()
		}
		s.emitDuplicateModerationEvent(ctx, userID, ip, metadata)

		message := "This clip has already been approved and added to our database"
		if existing.ClipID != nil {
			message = fmt.Sprintf("%s (clip %s)", message, existing.ClipID.String())
		}
		return &ValidationError{
			Field:   "clip_url",
			Message: message,
		}
	case "rejected":
		if time.Since(existing.CreatedAt) < 7*24*time.Hour {
			hoursRemaining := 7*24 - int(time.Since(existing.CreatedAt).Hours())
			if hoursRemaining < 24 {
				return &ValidationError{
					Field:   "clip_url",
					Message: "This clip was recently rejected. You can resubmit it in less than 24 hours",
				}
			}
			daysRemaining := hoursRemaining / 24
			return &ValidationError{
				Field:   "clip_url",
				Message: fmt.Sprintf("This clip was recently rejected. You can resubmit it in %d days", daysRemaining),
			}
		}
	}

	return nil
}

func (s *SubmissionService) ensureSubmissionCanBeApproved(submission *models.ClipSubmission) error {
	switch submission.SourceType {
	case string(SourceTypeTwitch), string(SourceTypeExternal), "upload":
		return nil
	default:
		if submission.SourceType == "" {
			return fmt.Errorf("submission source type is required")
		}
		return fmt.Errorf("unsupported submission source type %q", submission.SourceType)
	}
}

// SubmitClipRequest represents a clip submission request
type SubmitClipRequest struct {
	ClipURL                 string   `json:"clip_url" binding:"required"`
	CustomTitle             *string  `json:"custom_title,omitempty"`
	BroadcasterNameOverride *string  `json:"broadcaster_name_override,omitempty"`
	Tags                    []string `json:"tags,omitempty"`
	IsNSFW                  bool     `json:"is_nsfw"`
	SubmissionReason        *string  `json:"submission_reason,omitempty"`
}

// SubmitUploadRequest represents a hosted upload submission request.
type SubmitUploadRequest struct {
	SubmissionID     uuid.UUID
	CustomTitle      *string
	IsNSFW           bool
	SubmissionReason *string
	OriginalFilename string
	MimeType         string
	FileSizeBytes    int64
	DurationSeconds  int64
	DurationVerified bool
	StorageProvider  string
	StorageBucket    string
	StorageKey       string
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// TwitchAPIError represents an error from the Twitch API
type TwitchAPIError struct {
	Message string
}

func (e *TwitchAPIError) Error() string {
	return e.Message
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	Message    string `json:"error"`
	Limit      int    `json:"limit"`
	Window     int    `json:"window"`      // Window in seconds
	RetryAfter int64  `json:"retry_after"` // Unix timestamp when user can retry
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %d submissions per %d seconds", e.Limit, e.Window)
}

// ClipMetadata represents the metadata returned from the Twitch API for a clip
type ClipMetadata struct {
	ClipID       string    `json:"clip_id"`
	Title        string    `json:"title"`
	StreamerName string    `json:"streamer_name"`
	GameName     string    `json:"game_name,omitempty"`
	ViewCount    int       `json:"view_count"`
	CreatedAt    time.Time `json:"created_at"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Duration     float64   `json:"duration"`
	URL          string    `json:"url"`
}

const (
	clipMetadataCacheKeyPrefix = "twitch:clip:metadata:"
	clipMetadataCacheTTL       = 1 * time.Hour
)

// GetClipMetadata fetches clip metadata from Twitch API with Redis caching
func (s *SubmissionService) GetClipMetadata(ctx context.Context, clipURLOrID string) (*ClipMetadata, error) {
	// Validate input
	if strings.TrimSpace(clipURLOrID) == "" {
		return nil, &ValidationError{
			Field:   "url",
			Message: "Clip URL or ID is required",
		}
	}

	// Normalize and extract clip ID
	clipID, normalizedURL := s.normalizeClipURL(clipURLOrID)
	if clipID == "" {
		return nil, &ValidationError{
			Field:   "url",
			Message: "Invalid Twitch clip URL. Please provide a valid URL like 'https://clips.twitch.tv/ClipID' or 'https://www.twitch.tv/username/clip/ClipID'",
		}
	}

	// Check cache first
	if s.redisClient != nil {
		cacheKey := clipMetadataCacheKeyPrefix + clipID
		var cachedMetadata ClipMetadata
		err := s.redisClient.GetJSON(ctx, cacheKey, &cachedMetadata)
		if err == nil {
			// Cache hit
			return &cachedMetadata, nil
		}
		// Cache miss or error, continue to fetch from Twitch
	}

	// Check Twitch client is configured
	if s.twitchClient == nil {
		return nil, fmt.Errorf("Twitch API is not configured")
	}

	// Fetch from Twitch API
	params := &twitch.ClipParams{
		ClipIDs: []string{clipID},
	}

	resp, err := s.twitchClient.GetClips(ctx, params)
	if err != nil {
		return nil, &TwitchAPIError{
			Message: "Unable to fetch clip information from Twitch. Please verify the URL is correct and try again later.",
		}
	}

	if len(resp.Data) == 0 {
		return nil, &ValidationError{
			Field:   "url",
			Message: "This clip was not found on Twitch. It may have been deleted or the URL is incorrect.",
		}
	}

	clip := resp.Data[0]

	// Resolve game name if game ID is present
	gameName := ""
	if clip.GameID != "" {
		gamesResp, err := s.twitchClient.GetGames(ctx, []string{clip.GameID}, nil)
		if err == nil && len(gamesResp.Data) > 0 {
			gameName = gamesResp.Data[0].Name
		}
		// If game lookup fails, continue without game name (optional field)
	}

	metadata := &ClipMetadata{
		ClipID:       clip.ID,
		Title:        clip.Title,
		StreamerName: clip.BroadcasterName,
		GameName:     gameName,
		ViewCount:    clip.ViewCount,
		CreatedAt:    clip.CreatedAt,
		ThumbnailURL: clip.ThumbnailURL,
		Duration:     clip.Duration,
		URL:          normalizedURL,
	}

	// Cache the result
	if s.redisClient != nil {
		cacheKey := clipMetadataCacheKeyPrefix + clipID
		if cacheErr := s.redisClient.SetJSON(ctx, cacheKey, metadata, clipMetadataCacheTTL); cacheErr != nil {
			// Log cache error but don't fail the request
			log.Printf("Failed to cache clip metadata: %v", cacheErr)
		}
	}

	return metadata, nil
}

// SubmitClip handles clip submission with validation and duplicate detection
func (s *SubmissionService) SubmitClip(ctx context.Context, userID uuid.UUID, req *SubmitClipRequest, ip string, deviceFingerprint string) (*models.ClipSubmission, error) {
	// Validate and normalize input fields first
	if err := s.validateSubmissionInput(req); err != nil {
		return nil, err
	}

	// Check user permissions and rate limits
	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	isAdmin := user.Role == models.RoleAdmin

	if user.IsBanned {
		return nil, &ValidationError{Field: "user", Message: "Your account has been banned and cannot submit clips. Please contact support if you believe this is an error."}
	}

	// Check minimum karma requirement (configurable, can be disabled)
	if s.cfg.Karma.RequireKarmaForSubmission && user.KarmaPoints < s.cfg.Karma.SubmissionKarmaRequired {
		return nil, &ValidationError{Field: "karma", Message: fmt.Sprintf("You need at least %d karma points to submit clips. Earn karma by participating in the community through voting and commenting.", s.cfg.Karma.SubmissionKarmaRequired)}
	}

	// Perform abuse detection checks (skip for admins)
	if s.abuseDetector != nil && !isAdmin {
		abuseCheck, err := s.abuseDetector.CheckSubmissionAbuse(ctx, userID, ip, deviceFingerprint)
		if err != nil {
			log.Printf("Error checking abuse: %v", err)
		} else if !abuseCheck.Allowed {
			// Emit abuse event
			if s.moderationEvents != nil {
				metadata := map[string]interface{}{
					"reason":         abuseCheck.Reason,
					"severity":       abuseCheck.Severity,
					"cooldown_until": abuseCheck.CooldownUntil,
				}
				_ = s.moderationEvents.EmitAbuseEvent(ctx, ModerationEventUserCooldownActivated, userID, ip, metadata)
			}

			return nil, &ValidationError{
				Field:   "rate_limit",
				Message: abuseCheck.Reason,
			}
		} else if abuseCheck.Severity == "warning" {
			// Log warning but allow submission
			if s.moderationEvents != nil {
				metadata := map[string]interface{}{
					"warning": "IP sharing detected",
				}
				_ = s.moderationEvents.EmitAbuseEvent(ctx, ModerationEventIPShareSuspicious, userID, ip, metadata)
			}
		}
	}

	// Check rate limits (5 per hour, 20 per day) — admins are bypassed inside checkRateLimits
	if s.bypassRateLimits {
		log.Printf("SubmissionService: bypassing rate limits for test fixtures")
	} else if err := s.checkRateLimits(ctx, userID); err != nil {
		// Emit rate limit event
		if s.moderationEvents != nil {
			metadata := map[string]interface{}{
				"error": err.Error(),
			}
			_ = s.moderationEvents.EmitAbuseEvent(ctx, ModerationEventRateLimitExceeded, userID, ip, metadata)
		}
		return nil, err
	}

	clipInput := strings.TrimSpace(req.ClipURL)
	detectedSource, err := s.resolveSubmissionClipInput(clipInput)
	if err != nil {
		return nil, err
	}

	if detectedSource.SourceType == SourceTypeExternal {
		existingSubmission, err := s.getSubmissionBySourceIdentity(ctx, string(detectedSource.Platform), detectedSource.SourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to check external submission duplicates: %w", err)
		}
		if existingSubmission != nil {
			if err := s.rejectDuplicateExternalSubmission(ctx, userID, ip, existingSubmission); err != nil {
				return nil, err
			}
		}

		metadata, err := s.fetchExternalMetadata(ctx, detectedSource)
		if err != nil {
			return nil, err
		}

		if metadata.DurationVerified && metadata.DurationSeconds != nil && *metadata.DurationSeconds > s.cfg.ClipSource.MaxDurationSeconds {
			return nil, &ValidationError{Field: "clip_url", Message: fmt.Sprintf("External source duration exceeds the maximum of %d seconds", s.cfg.ClipSource.MaxDurationSeconds)}
		}

		submission, err := s.buildExternalSubmissionRecord(userID, req, detectedSource, metadata, time.Now())
		if err != nil {
			return nil, err
		}

		if err := s.createSubmissionRecord(ctx, submission); err != nil {
			return nil, fmt.Errorf("failed to create external submission: %w", err)
		}

		if s.webhookService != nil {
			webhookData := map[string]interface{}{
				"submission_id":     submission.ID.String(),
				"user_id":           userID.String(),
				"source_type":       submission.SourceType,
				"source_platform":   submission.SourcePlatform,
				"clip_id":           submission.TwitchClipID,
				"clip_url":          submission.TwitchClipURL,
				"duration_verified": submission.DurationVerified,
				"status":            submission.Status,
				"is_nsfw":           submission.IsNSFW,
				"created_at":        submission.CreatedAt,
			}
			if submission.CustomTitle != nil {
				webhookData["custom_title"] = *submission.CustomTitle
			}
			if submission.SubmissionReason != nil {
				webhookData["submission_reason"] = *submission.SubmissionReason
			}
			if err := s.webhookService.TriggerEvent(ctx, models.WebhookEventClipSubmitted, submission.ID, webhookData); err != nil {
				log.Printf("Failed to trigger webhook event: %v", err)
			}
		}

		if s.moderationEvents != nil {
			metadata := map[string]interface{}{
				"submission_id":     submission.ID.String(),
				"clip_id":           submission.TwitchClipID,
				"clip_url":          submission.TwitchClipURL,
				"source_type":       submission.SourceType,
				"source_platform":   submission.SourcePlatform,
				"duration_verified": submission.DurationVerified,
				"status":            submission.Status,
				"is_nsfw":           submission.IsNSFW,
			}
			if submission.CustomTitle != nil {
				metadata["custom_title"] = *submission.CustomTitle
			}
			if submission.SubmissionReason != nil {
				metadata["submission_reason"] = *submission.SubmissionReason
			}
			if err := s.moderationEvents.EmitSubmissionEvent(ctx, ModerationEventSubmissionReceived, submission, ip, metadata); err != nil {
				log.Printf("Failed to emit external submission event: %v", err)
			}
		}

		return submission, nil
	}

	// Check if clip exists and whether it can be claimed
	clipExistence, err := s.checkClipExistence(ctx, detectedSource.SourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check clip existence: %w", err)
	}

	// If clip exists and can be claimed (scraped clip), claim it directly
	if clipExistence.Exists && clipExistence.CanBeClaimed {
		now := time.Now()
		title := req.CustomTitle
		broadcasterName := req.BroadcasterNameOverride

		// Claim the discovery clip (atomically moves from discovery_clips to clips)
		if err := s.requireCreatorSubmissionPermission(ctx, clipExistence.Clip.CreatorAccountID, userID); err != nil {
			return nil, err
		}
		_, err := s.discoveryClipRepo.ClaimDiscoveryClip(ctx, clipExistence.Clip.TwitchClipID, userID, title, req.IsNSFW, broadcasterName)
		if err != nil {
			return nil, fmt.Errorf("failed to claim discovery clip: %w", err)
		}

		// Auto-upvote the claimed clip when interaction is allowed
		if s.voteRepo != nil && s.shouldAutoUpvoteClaimedClip(ctx, clipExistence.Clip.CreatorAccountID, userID) {
			if err := s.voteRepo.UpsertVote(ctx, userID, clipExistence.Clip.ID, 1); err != nil {
				// Log error but don't fail
				log.Printf("Warning: failed to auto-upvote claimed clip for user %s: %v\n", userID, err)
			}
		}

		// Award karma for claiming
		if err := s.awardKarma(ctx, userID, 10); err != nil {
			// Log error but don't fail
			log.Printf("Failed to award karma: %v\n", err)
		}

		// Create a real submission record for audit trail and consistency
		submission := &models.ClipSubmission{
			ID:                      uuid.New(),
			UserID:                  userID,
			TwitchClipID:            clipExistence.Clip.TwitchClipID,
			TwitchClipURL:           clipExistence.Clip.TwitchClipURL,
			CustomTitle:             title,
			Title:                   &clipExistence.Clip.Title,
			IsNSFW:                  req.IsNSFW,
			Tags:                    req.Tags,
			SubmissionReason:        req.SubmissionReason,
			BroadcasterNameOverride: req.BroadcasterNameOverride,
			Status:                  "approved", // Claimed clips are immediately approved
			CreatedAt:               now,
			UpdatedAt:               now,
			SourceType:              "twitch",
			SourcePlatform:          "twitch",
			SourceURL:               &clipExistence.Clip.TwitchClipURL,
			SourceID:                &clipExistence.Clip.TwitchClipID,
			SourceMetadata:          []byte(`{}`),
			DurationVerified:        true,
			UploadStatus:            "none",
			StorageVisibility:       "private",
			ReviewedAt:              &now,
			ReviewedBy:              &userID,
			// Copy metadata from existing clip
			CreatorName:      &clipExistence.Clip.CreatorName,
			CreatorID:        clipExistence.Clip.CreatorID,
			CreatorAccountID: clipExistence.Clip.CreatorAccountID,
			BroadcasterName:  &clipExistence.Clip.BroadcasterName,
			BroadcasterID:    clipExistence.Clip.BroadcasterID,
			GameID:           clipExistence.Clip.GameID,
			GameName:         clipExistence.Clip.GameName,
			ThumbnailURL:     clipExistence.Clip.ThumbnailURL,
			Duration:         clipExistence.Clip.Duration,
			ViewCount:        clipExistence.Clip.ViewCount,
		}

		// Save submission to database for audit trail
		if err := s.submissionRepo.Create(ctx, submission); err != nil {
			return nil, fmt.Errorf("failed to create submission record for claimed clip: %w", err)
		}

		// Trigger webhook events for integrations
		if s.webhookService != nil {
			webhookData := map[string]interface{}{
				"submission_id":   submission.ID.String(),
				"user_id":         userID.String(),
				"twitch_clip_id":  submission.TwitchClipID,
				"twitch_clip_url": submission.TwitchClipURL,
				"clip_id":         clipExistence.Clip.ID.String(),
				"claimed":         true, // Distinguish from normal submissions
			}
			if submission.CustomTitle != nil {
				webhookData["custom_title"] = *submission.CustomTitle
			}
			if len(submission.Tags) > 0 {
				webhookData["tags"] = submission.Tags
			}

			// Trigger clip.submitted event
			if err := s.webhookService.TriggerEvent(ctx, models.WebhookEventClipSubmitted, submission.ID, webhookData); err != nil {
				log.Printf("Warning: failed to trigger clip.submitted webhook for claimed clip: %v\n", err)
			}

			// Trigger clip.approved event (claimed clips are auto-approved)
			webhookDataApproved := map[string]interface{}{
				"submission_id":   submission.ID.String(),
				"user_id":         userID.String(),
				"twitch_clip_id":  submission.TwitchClipID,
				"twitch_clip_url": submission.TwitchClipURL,
				"clip_id":         clipExistence.Clip.ID.String(),
				"claimed":         true,
				"reviewer_id":     userID.String(),
				"approved_at":     now,
			}
			if err := s.webhookService.TriggerEvent(ctx, models.WebhookEventClipApproved, submission.ID, webhookDataApproved); err != nil {
				log.Printf("Warning: failed to trigger clip.approved webhook for claimed clip: %v\n", err)
			}
		}

		return submission, nil
	}

	// If clip exists but cannot be claimed (already claimed), return error
	if clipExistence.Exists && !clipExistence.CanBeClaimed {
		// Track duplicate attempt
		if s.abuseDetector != nil {
			if err := s.abuseDetector.TrackDuplicateAttempt(ctx, userID, ip, detectedSource.SourceID); err != nil {
				log.Printf("Failed to track duplicate attempt: %v", err)
			}
		}

		return nil, &ValidationError{
			Field:   "clip_url",
			Message: "This clip has already been posted by another user",
		}
	}

	// Check for duplicates in submissions table
	if err := s.checkDuplicates(ctx, detectedSource.SourceID, userID, ip); err != nil {
		return nil, err
	}

	// Fetch clip metadata from Twitch
	twitchClip, err := s.fetchClipFromTwitch(ctx, detectedSource.SourceID)
	if err != nil {
		return nil, err
	}

	// Validate clip quality
	if err := s.validateClipQuality(twitchClip); err != nil {
		return nil, err
	}

	// Use normalized URL
	if detectedSource.NormalizedURL != "" {
		twitchClip.URL = detectedSource.NormalizedURL
	}

	// Create submission
	submission := &models.ClipSubmission{
		ID:                      uuid.New(),
		UserID:                  userID,
		TwitchClipID:            twitchClip.ID,
		TwitchClipURL:           twitchClip.URL,
		Title:                   &twitchClip.Title,
		CustomTitle:             req.CustomTitle,
		BroadcasterNameOverride: req.BroadcasterNameOverride,
		Tags:                    req.Tags,
		IsNSFW:                  req.IsNSFW,
		SubmissionReason:        req.SubmissionReason,
		Status:                  "pending",
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		SourceType:              "twitch",
		SourcePlatform:          "twitch",
		SourceURL:               &twitchClip.URL,
		SourceID:                &twitchClip.ID,
		SourceMetadata:          []byte(`{}`),
		DurationVerified:        true,
		UploadStatus:            "none",
		StorageVisibility:       "private",
		// Metadata from Twitch
		CreatorName:      &twitchClip.CreatorName,
		CreatorID:        utils.StringPtr(twitchClip.CreatorID),
		CreatorAccountID: nil,
		BroadcasterName:  &twitchClip.BroadcasterName,
		BroadcasterID:    utils.StringPtr(twitchClip.BroadcasterID),
		GameID:           utils.StringPtr(twitchClip.GameID),
		ThumbnailURL:     utils.StringPtr(twitchClip.ThumbnailURL),
		Duration:         utils.Float64Ptr(twitchClip.Duration),
		ViewCount:        twitchClip.ViewCount,
	}

	// Check for auto-approval
	if s.shouldAutoApprove(user) {
		submission.Status = "approved"
		submission.ReviewedBy = &userID
		submission.ReviewedAt = &submission.CreatedAt

		// Create clip immediately
		clipID, err := s.createClipFromSubmission(ctx, submission)
		if err != nil {
			return nil, fmt.Errorf("failed to create clip: %w", err)
		}

		// Store clip ID
		submission.ClipID = &clipID

		// Award karma
		if err := s.awardKarma(ctx, userID, 10); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to award karma: %v\n", err)
		}
	}

	// Save submission
	if err := s.submissionRepo.Create(ctx, submission); err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	// Trigger webhook for clip submission
	if s.webhookService != nil {
		webhookData := map[string]interface{}{
			"submission_id":   submission.ID.String(),
			"user_id":         submission.UserID.String(),
			"twitch_clip_id":  submission.TwitchClipID,
			"twitch_clip_url": submission.TwitchClipURL,
			"status":          submission.Status,
			"is_nsfw":         submission.IsNSFW,
			"created_at":      submission.CreatedAt,
		}
		if submission.CustomTitle != nil {
			webhookData["custom_title"] = *submission.CustomTitle
		}
		if len(submission.Tags) > 0 {
			webhookData["tags"] = submission.Tags
		}

		// Always send clip.submitted event
		if err := s.webhookService.TriggerEvent(ctx, models.WebhookEventClipSubmitted, submission.ID, webhookData); err != nil {
			log.Printf("Failed to trigger webhook event: %v", err)
		}

		// If auto-approved, also send clip.approved event with auto_approved field
		if submission.Status == "approved" {
			webhookDataApproved := make(map[string]interface{})
			for k, v := range webhookData {
				webhookDataApproved[k] = v
			}
			webhookDataApproved["auto_approved"] = true
			if err := s.webhookService.TriggerEvent(ctx, models.WebhookEventClipApproved, submission.ID, webhookDataApproved); err != nil {
				log.Printf("Failed to trigger webhook event: %v", err)
			}
		}
	}

	// Emit moderation event for new submission
	if s.moderationEvents != nil {
		eventType := ModerationEventSubmissionReceived
		if submission.Status == "approved" {
			eventType = ModerationEventSubmissionApproved
		}

		metadata := map[string]interface{}{
			"submission_id": submission.ID.String(),
			"clip_id":       submission.TwitchClipID,
			"clip_url":      submission.TwitchClipURL,
			"status":        submission.Status,
			"is_nsfw":       submission.IsNSFW,
			"auto_approved": submission.Status == "approved",
		}

		if submission.CustomTitle != nil {
			metadata["custom_title"] = *submission.CustomTitle
		}
		if len(submission.Tags) > 0 {
			metadata["tags"] = submission.Tags
		}

		if err := s.moderationEvents.EmitSubmissionEvent(ctx, eventType, submission, ip, metadata); err != nil {
			log.Printf("Failed to emit submission event: %v", err)
		}
	}

	return submission, nil
}

// SubmitUpload handles hosted video submissions.
func (s *SubmissionService) SubmitUpload(ctx context.Context, userID uuid.UUID, req *SubmitUploadRequest, ip string, deviceFingerprint string) (*models.ClipSubmission, error) {
	if err := s.validateUploadSubmissionInput(req); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	isAdmin := user.Role == models.RoleAdmin
	if user.IsBanned {
		return nil, &ValidationError{Field: "user", Message: "Your account has been banned and cannot submit clips. Please contact support if you believe this is an error."}
	}

	if s.cfg.Karma.RequireKarmaForSubmission && user.KarmaPoints < s.cfg.Karma.SubmissionKarmaRequired {
		return nil, &ValidationError{Field: "karma", Message: fmt.Sprintf("You need at least %d karma points to submit clips. Earn karma by participating in the community through voting and commenting.", s.cfg.Karma.SubmissionKarmaRequired)}
	}

	if s.abuseDetector != nil && !isAdmin {
		abuseCheck, err := s.abuseDetector.CheckSubmissionAbuse(ctx, userID, ip, deviceFingerprint)
		if err != nil {
			log.Printf("Error checking abuse: %v", err)
		} else if !abuseCheck.Allowed {
			if s.moderationEvents != nil {
				metadata := map[string]interface{}{
					"reason":         abuseCheck.Reason,
					"severity":       abuseCheck.Severity,
					"cooldown_until": abuseCheck.CooldownUntil,
				}
				_ = s.moderationEvents.EmitAbuseEvent(ctx, ModerationEventUserCooldownActivated, userID, ip, metadata)
			}

			return nil, &ValidationError{
				Field:   "rate_limit",
				Message: abuseCheck.Reason,
			}
		} else if abuseCheck.Severity == "warning" {
			if s.moderationEvents != nil {
				metadata := map[string]interface{}{
					"warning": "IP sharing detected",
				}
				_ = s.moderationEvents.EmitAbuseEvent(ctx, ModerationEventIPShareSuspicious, userID, ip, metadata)
			}
		}
	}

	if s.bypassRateLimits {
		log.Printf("SubmissionService: bypassing rate limits for test fixtures")
	} else if err := s.checkRateLimits(ctx, userID); err != nil {
		if s.moderationEvents != nil {
			metadata := map[string]interface{}{"error": err.Error()}
			_ = s.moderationEvents.EmitAbuseEvent(ctx, ModerationEventRateLimitExceeded, userID, ip, metadata)
		}
		return nil, err
	}

	now := time.Now()
	baseName := strings.TrimSpace(req.OriginalFilename)
	if baseName == "" {
		baseName = "Uploaded clip"
	}

	customTitle := req.CustomTitle
	if customTitle != nil {
		title := strings.TrimSpace(*customTitle)
		if title == "" {
			customTitle = nil
		} else {
			*customTitle = title
		}
	}

	reason := req.SubmissionReason
	if reason != nil {
		trimmed := strings.TrimSpace(*reason)
		if trimmed == "" {
			reason = nil
		} else {
			*reason = trimmed
		}
	}

	title := customTitle
	if title == nil {
		title = &baseName
	}

	submissionID := req.SubmissionID
	if submissionID == uuid.Nil {
		submissionID = uuid.New()
	}
	submission, err := s.buildUploadSubmissionRecord(userID, submissionID, req, title, customTitle, reason, now)
	if err != nil {
		return nil, err
	}
	userIDStr := userID.String()

	if err := s.submissionRepo.Create(ctx, submission); err != nil {
		return nil, fmt.Errorf("failed to create upload submission: %w", err)
	}

	if s.moderationEvents != nil {
		metadata := map[string]interface{}{
			"submission_id":      submission.ID.String(),
			"user_id":            userIDStr,
			"source_type":        "upload",
			"storage_key":        req.StorageKey,
			"storage_visibility": "private",
			"upload_status":      "validated",
			"duration_seconds":   req.DurationSeconds,
		}
		if submission.CustomTitle != nil {
			metadata["custom_title"] = *submission.CustomTitle
		}
		if submission.SubmissionReason != nil {
			metadata["submission_reason"] = *submission.SubmissionReason
		}
		if err := s.moderationEvents.EmitSubmissionEvent(ctx, ModerationEventSubmissionReceived, submission, ip, metadata); err != nil {
			log.Printf("Failed to emit upload submission event: %v", err)
		}
	}

	return submission, nil
}

func (s *SubmissionService) buildUploadSubmissionRecord(userID, submissionID uuid.UUID, req *SubmitUploadRequest, title, customTitle, reason *string, now time.Time) (*models.ClipSubmission, error) {
	fileSize := req.FileSizeBytes
	sourceID := submissionID.String()
	resolvedTitle := title
	if resolvedTitle == nil {
		fallback := strings.TrimSpace(req.OriginalFilename)
		if fallback == "" {
			fallback = "Uploaded clip"
		}
		resolvedTitle = &fallback
	}
	metadata := map[string]interface{}{
		"original_filename": req.OriginalFilename,
		"mime_type":         req.MimeType,
		"duration_seconds":  req.DurationSeconds,
		"duration_verified": req.DurationVerified,
		"storage_provider":  req.StorageProvider,
		"storage_bucket":    req.StorageBucket,
		"storage_key":       req.StorageKey,
		"file_size_bytes":   fileSize,
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to encode upload metadata: %w", err)
	}

	fileSizePtr := fileSize
	durationSeconds := int(req.DurationSeconds)
	return &models.ClipSubmission{
		ID:                      submissionID,
		UserID:                  userID,
		ClipID:                  nil,
		TwitchClipID:            legacyUploadTwitchClipID(sourceID, req.StorageKey),
		TwitchClipURL:           "",
		Title:                   resolvedTitle,
		CustomTitle:             customTitle,
		IsNSFW:                  req.IsNSFW,
		SubmissionReason:        reason,
		Status:                  "pending",
		CreatedAt:               now,
		UpdatedAt:               now,
		SourceType:              "upload",
		SourcePlatform:          "upload",
		SourceURL:               nil,
		SourceID:                &sourceID,
		SourceMetadata:          metadataBytes,
		DurationSeconds:         &durationSeconds,
		DurationVerified:        req.DurationVerified,
		StorageProvider:         &req.StorageProvider,
		StorageBucket:           &req.StorageBucket,
		StorageKey:              &req.StorageKey,
		OriginalFilename:        &req.OriginalFilename,
		MimeType:                &req.MimeType,
		FileSizeBytes:           &fileSizePtr,
		UploadStatus:            "validated",
		StorageVisibility:       "private",
		DurationValidationError: nil,
		CreatorAccountID:        nil,
	}, nil
}

func (s *SubmissionService) fetchExternalMetadata(ctx context.Context, source DetectedSource) (ExternalMetadata, error) {
	fetcher := s.externalMetadataFetcher
	if fetcher == nil {
		fetcher = NewExternalMetadataFetcher(nil)
	}
	return fetcher.Fetch(ctx, source)
}

func (s *SubmissionService) buildExternalSubmissionRecord(userID uuid.UUID, req *SubmitClipRequest, source DetectedSource, metadata ExternalMetadata, now time.Time) (*models.ClipSubmission, error) {
	sourceID := source.SourceID
	sourceURL := source.NormalizedURL
	if sourceURL == "" {
		sourceURL = source.RawURL
	}

	title := strings.TrimSpace(metadata.Title)
	if title == "" {
		title = sourceID
	}

	authorName := strings.TrimSpace(metadata.AuthorName)
	resolvedThumbnail := strings.TrimSpace(metadata.ThumbnailURL)
	metadataBytes, err := encodeExternalMetadata(source, metadata)
	if err != nil {
		return nil, err
	}

	var durationSeconds *int
	if metadata.DurationSeconds != nil {
		duration := int(*metadata.DurationSeconds)
		durationSeconds = &duration
	}

	return &models.ClipSubmission{
		ID:                uuid.New(),
		UserID:            userID,
		ClipID:            nil,
		TwitchClipID:      legacyExternalTwitchClipID(string(source.Platform), sourceID, sourceURL),
		TwitchClipURL:     sourceURL,
		Title:             &title,
		CustomTitle:       req.CustomTitle,
		Tags:              req.Tags,
		IsNSFW:            req.IsNSFW,
		SubmissionReason:  req.SubmissionReason,
		Status:            "pending",
		CreatedAt:         now,
		UpdatedAt:         now,
		SourceType:        string(SourceTypeExternal),
		SourcePlatform:    string(source.Platform),
		SourceURL:         &sourceURL,
		SourceID:          &sourceID,
		SourceMetadata:    metadataBytes,
		DurationSeconds:   durationSeconds,
		DurationVerified:  metadata.DurationVerified,
		UploadStatus:      "none",
		StorageVisibility: "private",
		CreatorName: func() *string {
			if authorName == "" {
				return nil
			}
			return &authorName
		}(),
		CreatorID:       nil,
		BroadcasterName: nil,
		BroadcasterID:   nil,
		GameID:          nil,
		GameName:        nil,
		ThumbnailURL: func() *string {
			if resolvedThumbnail == "" {
				return nil
			}
			return &resolvedThumbnail
		}(),
		ViewCount:               0,
		DurationValidationError: nil,
	}, nil
}

// BuildUploadStorageKey constructs the pending storage key for an uploaded clip.
func BuildUploadStorageKey(userID, submissionID uuid.UUID, ext string) string {
	return fmt.Sprintf("uploads/pending/%s/%s/original%s", userID.String(), submissionID.String(), ext)
}

// validateSubmissionInput validates and normalizes submission request fields
func (s *SubmissionService) validateSubmissionInput(req *SubmitClipRequest) error {
	// Validate clip URL (non-empty is already enforced by binding:"required")
	if len(req.ClipURL) > 500 {
		return &ValidationError{
			Field:   "clip_url",
			Message: "Clip URL is too long (maximum 500 characters)",
		}
	}

	// Validate custom title if provided
	if req.CustomTitle != nil {
		title := strings.TrimSpace(*req.CustomTitle)
		if title != "" {
			if len(title) < 3 {
				return &ValidationError{
					Field:   "custom_title",
					Message: "Custom title must be at least 3 characters long",
				}
			}
			if len(title) > 200 {
				return &ValidationError{
					Field:   "custom_title",
					Message: "Custom title is too long (maximum 200 characters)",
				}
			}
			// Normalize: update the pointer to the trimmed value
			*req.CustomTitle = title
		} else {
			// If it's empty after trimming, set to nil
			req.CustomTitle = nil
		}
	}

	// Validate broadcaster name override if provided
	if req.BroadcasterNameOverride != nil {
		broadcaster := strings.TrimSpace(*req.BroadcasterNameOverride)
		if broadcaster != "" {
			if len(broadcaster) < 2 {
				return &ValidationError{
					Field:   "broadcaster_name_override",
					Message: "Broadcaster name must be at least 2 characters long",
				}
			}
			if len(broadcaster) > 100 {
				return &ValidationError{
					Field:   "broadcaster_name_override",
					Message: "Broadcaster name is too long (maximum 100 characters)",
				}
			}
			// Validate broadcaster name format (alphanumeric + underscores only)
			if !isValidUsername(broadcaster) {
				return &ValidationError{
					Field:   "broadcaster_name_override",
					Message: "Broadcaster name can only contain letters, numbers, and underscores",
				}
			}
			// Normalize: update the pointer to the trimmed value
			*req.BroadcasterNameOverride = broadcaster
		} else {
			// If it's empty after trimming, set to nil
			req.BroadcasterNameOverride = nil
		}
	}

	// Validate tags
	if len(req.Tags) > 10 {
		return &ValidationError{
			Field:   "tags",
			Message: "Too many tags (no more than 10 tags allowed)",
		}
	}

	// Normalize and validate each tag
	normalizedTags := make([]string, 0, len(req.Tags))
	seenTags := make(map[string]bool)
	for _, tag := range req.Tags {
		// Trim and lowercase for normalization
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue // Skip empty tags
		}

		// Check for duplicates (case-insensitive)
		if seenTags[normalized] {
			continue // Skip duplicate tags
		}

		if len(normalized) < 2 {
			return &ValidationError{
				Field:   "tags",
				Message: fmt.Sprintf("Tag '%s' is too short (minimum 2 characters)", tag),
			}
		}
		if len(normalized) > 50 {
			return &ValidationError{
				Field:   "tags",
				Message: fmt.Sprintf("Tag '%s' is too long (maximum 50 characters)", tag),
			}
		}

		// Validate tag format (alphanumeric + hyphens only)
		if !isValidTag(normalized) {
			return &ValidationError{
				Field:   "tags",
				Message: fmt.Sprintf("Tag '%s' contains invalid characters (only letters, numbers, and hyphens allowed)", tag),
			}
		}

		normalizedTags = append(normalizedTags, normalized)
		seenTags[normalized] = true
	}
	req.Tags = normalizedTags

	// Validate submission reason if provided
	if req.SubmissionReason != nil {
		reason := strings.TrimSpace(*req.SubmissionReason)
		if reason != "" {
			if len(reason) > 1000 {
				return &ValidationError{
					Field:   "submission_reason",
					Message: "Submission reason is too long (maximum 1000 characters)",
				}
			}
			// Normalize: update the pointer to the trimmed value
			*req.SubmissionReason = reason
		} else {
			// If it's empty after trimming, set to nil
			req.SubmissionReason = nil
		}
	}

	return nil
}

// validateUploadSubmissionInput validates hosted upload submission metadata.
func (s *SubmissionService) validateUploadSubmissionInput(req *SubmitUploadRequest) error {
	if req == nil {
		return &ValidationError{Field: "request", Message: "Upload request is required"}
	}
	if strings.TrimSpace(req.OriginalFilename) == "" {
		return &ValidationError{Field: "file", Message: "Original filename is required"}
	}
	if strings.TrimSpace(req.MimeType) == "" {
		return &ValidationError{Field: "mime_type", Message: "Mime type is required"}
	}
	if strings.TrimSpace(req.StorageProvider) == "" {
		return &ValidationError{Field: "storage_provider", Message: "Storage provider is required"}
	}
	if strings.TrimSpace(req.StorageBucket) == "" {
		return &ValidationError{Field: "storage_bucket", Message: "Storage bucket is required"}
	}
	if strings.TrimSpace(req.StorageKey) == "" {
		return &ValidationError{Field: "storage_key", Message: "Storage key is required"}
	}
	if req.FileSizeBytes <= 0 {
		return &ValidationError{Field: "file_size_bytes", Message: "File size must be greater than zero"}
	}
	if !req.DurationVerified {
		return &ValidationError{Field: "duration_seconds", Message: "Upload duration must be verified"}
	}
	if req.DurationSeconds < 0 {
		return &ValidationError{Field: "duration_seconds", Message: "Duration cannot be negative"}
	}

	if req.CustomTitle != nil {
		title := strings.TrimSpace(*req.CustomTitle)
		if title != "" {
			if len(title) < 3 {
				return &ValidationError{Field: "custom_title", Message: "Custom title must be at least 3 characters long"}
			}
			if len(title) > 200 {
				return &ValidationError{Field: "custom_title", Message: "Custom title is too long (maximum 200 characters)"}
			}
			*req.CustomTitle = title
		} else {
			req.CustomTitle = nil
		}
	}

	if req.SubmissionReason != nil {
		reason := strings.TrimSpace(*req.SubmissionReason)
		if reason != "" {
			if len(reason) > 1000 {
				return &ValidationError{Field: "submission_reason", Message: "Submission reason is too long (maximum 1000 characters)"}
			}
			*req.SubmissionReason = reason
		} else {
			req.SubmissionReason = nil
		}
	}

	return nil
}

// isValidUsername checks if a username contains only valid characters
func isValidUsername(username string) bool {
	if username == "" {
		return false
	}
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

// isValidTag checks if a tag contains only valid characters
func isValidTag(tag string) bool {
	if tag == "" {
		return false
	}
	for _, r := range tag {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}

// normalizeClipURL extracts clip ID and returns normalized URL
func (s *SubmissionService) normalizeClipURL(clipURLOrID string) (clipID string, normalizedURL string) {
	clipID = extractClipIDFromURL(clipURLOrID)
	if clipID == "" {
		return "", ""
	}
	// Return canonical clips.twitch.tv URL
	normalizedURL = fmt.Sprintf("https://clips.twitch.tv/%s", clipID)
	return clipID, normalizedURL
}

func (s *SubmissionService) resolveSubmissionClipInput(clipInput string) (DetectedSource, error) {
	if strings.HasPrefix(strings.ToLower(clipInput), "http://") || strings.HasPrefix(strings.ToLower(clipInput), "https://") {
		detectedSource, err := DetectClipSource(clipInput)
		if err != nil {
			return DetectedSource{}, &ValidationError{Field: "clip_url", Message: err.Error()}
		}
		return detectedSource, nil
	}

	// Preserve the existing direct-ID Twitch path.
	clipID, normalizedURL := s.normalizeClipURL(clipInput)
	if clipID == "" {
		return DetectedSource{}, &ValidationError{
			Field:   "clip_url",
			Message: "Invalid Twitch clip URL. Please provide a valid URL like 'https://clips.twitch.tv/ClipID' or 'https://www.twitch.tv/username/clip/ClipID'",
		}
	}

	return DetectedSource{
		RawURL:        clipInput,
		NormalizedURL: normalizedURL,
		Platform:      SourcePlatformTwitch,
		SourceType:    SourceTypeTwitch,
		SourceID:      clipID,
	}, nil
}

// checkRateLimits validates rate limits for submissions
func (s *SubmissionService) checkRateLimits(ctx context.Context, userID uuid.UUID) error {
	// Admins should not be rate limited for submissions
	if s.userRepo != nil {
		if user, err := s.userRepo.GetByID(ctx, userID); err == nil {
			if user.Role == models.RoleAdmin {
				return nil
			}
		}
	}

	// Check hourly limit (10 per hour) - using E2E test expectations
	hourlyCount, err := s.submissionRepo.CountUserSubmissions(ctx, userID, time.Now().Add(-1*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to check hourly rate limit: %w", err)
	}
	if hourlyCount >= 10 {
		// TODO: Calculate retry_after based on oldest submission timestamp + window
		// For now, using a conservative 1-hour wait from current time
		// This matches E2E test expectations (simple cooldown period)
		retryAfter := time.Now().Add(1 * time.Hour).Unix()
		return &RateLimitError{
			Message:    "rate_limit_exceeded",
			Limit:      10,
			Window:     3600, // 1 hour in seconds
			RetryAfter: retryAfter,
		}
	}

	// Check daily limit (20 per day)
	dailyCount, err := s.submissionRepo.CountUserSubmissions(ctx, userID, time.Now().Add(-24*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to check daily rate limit: %w", err)
	}
	if dailyCount >= 20 {
		// TODO: Calculate retry_after based on oldest submission timestamp + window
		// For now, using a conservative 24-hour wait from current time
		retryAfter := time.Now().Add(24 * time.Hour).Unix()
		return &RateLimitError{
			Message:    "rate_limit_exceeded",
			Limit:      20,
			Window:     86400, // 24 hours in seconds
			RetryAfter: retryAfter,
		}
	}

	return nil
}

// ClipExistenceResult represents the result of checking if a clip exists
type ClipExistenceResult struct {
	Exists       bool
	Clip         *models.Clip
	CanBeClaimed bool // True if clip exists but submitted_by_user_id is NULL
}

// CheckClipExistence checks if a clip already exists in the database and whether it can be claimed by a user.
// This is a public wrapper for the internal checkClipExistence method.
//
// Parameters:
//   - ctx: Context for the operation
//   - twitchClipID: The Twitch clip ID to check
//
// Returns:
//   - ClipExistenceResult: Contains information about the clip's existence and claimability
//   - error: Any error that occurred during the check
//
// The CanBeClaimed field in the result will be true when:
//   - The clip exists in the database (Exists = true)
//   - The clip has no submitted_by_user_id (it's a scraped/imported clip)
//
// Example usage:
//
//	result, err := service.CheckClipExistence(ctx, "AwesomeClipID123")
//	if err != nil {
//	    return err
//	}
//	if result.CanBeClaimed {
//	    // User can claim this scraped clip
//	}
func (s *SubmissionService) CheckClipExistence(ctx context.Context, twitchClipID string) (*ClipExistenceResult, error) {
	return s.checkClipExistence(ctx, twitchClipID)
}

// checkClipExistence is the internal implementation that checks if a clip exists and whether it can be claimed.
// It first checks the discovery_clips staging table (claimable), then the main clips table (not claimable).
func (s *SubmissionService) checkClipExistence(ctx context.Context, twitchClipID string) (*ClipExistenceResult, error) {
	// Check discovery_clips first — these are claimable
	if s.discoveryClipRepo != nil {
		dc, err := s.discoveryClipRepo.GetByTwitchClipID(ctx, twitchClipID)
		if err == nil && dc != nil {
			// Convert DiscoveryClip to Clip for backward-compatible result
			clip := &models.Clip{
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
			}
			return &ClipExistenceResult{
				Exists:       true,
				Clip:         clip,
				CanBeClaimed: true,
			}, nil
		}
		// If error is "no rows", continue to check clips table
		if err != nil && !strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("failed to check discovery clip existence: %w", err)
		}
	}

	// Check main clips table — these are already posted and not claimable
	clip, err := s.clipRepo.GetByTwitchClipID(ctx, twitchClipID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return &ClipExistenceResult{Exists: false, CanBeClaimed: false}, nil
		}
		return nil, fmt.Errorf("failed to check clip existence: %w", err)
	}

	return &ClipExistenceResult{
		Exists:       true,
		Clip:         clip,
		CanBeClaimed: false,
	}, nil
}

// checkDuplicates checks if clip already exists or was submitted
func (s *SubmissionService) checkDuplicates(ctx context.Context, twitchClipID string, userID uuid.UUID, ip string) error {
	if s.allowDuplicateSubmission {
		log.Printf("SubmissionService: allowing duplicate submissions (test fixtures enabled)")
		return nil
	}

	// Check if clip already exists in clips table
	exists, err := s.clipRepo.ExistsByTwitchClipID(ctx, twitchClipID)
	if err != nil {
		return fmt.Errorf("failed to check clip existence: %w", err)
	}
	// Also check discovery_clips table
	if !exists && s.discoveryClipRepo != nil {
		exists, err = s.discoveryClipRepo.ExistsByTwitchClipID(ctx, twitchClipID)
		if err != nil {
			return fmt.Errorf("failed to check discovery clip existence: %w", err)
		}
	}
	if exists {
		// Track duplicate attempt
		s.trackDuplicateAttempt(ctx, userID, ip, twitchClipID)

		// Emit moderation event
		s.emitDuplicateModerationEvent(ctx, userID, ip, map[string]interface{}{
			"clip_id": twitchClipID,
			"reason":  "clip_already_exists",
		})

		return &ValidationError{
			Field:   "clip_url",
			Message: "This clip has already been added to our database and cannot be submitted again",
		}
	}

	// Check if clip was already submitted
	submission, err := s.submissionRepo.GetByTwitchClipID(ctx, twitchClipID)
	if err != nil {
		return fmt.Errorf("failed to check submission existence: %w", err)
	}
	if submission != nil {
		// Track duplicate attempt
		s.trackDuplicateAttempt(ctx, userID, ip, twitchClipID)

		if submission.Status == "pending" {
			// Emit moderation event for duplicate pending submission
			s.emitDuplicateModerationEvent(ctx, userID, ip, map[string]interface{}{
				"clip_id":       twitchClipID,
				"reason":        "submission_pending",
				"submission_id": submission.ID.String(),
			})

			return &ValidationError{
				Field:   "clip_url",
				Message: "This clip is already pending review. You'll be notified once it's been reviewed by our moderators.",
			}
		}
		if submission.Status == "approved" {
			return &ValidationError{
				Field:   "clip_url",
				Message: "This clip has already been approved and added to our database",
			}
		}
		// If rejected, allow resubmission after some time
		if submission.Status == "rejected" && time.Since(submission.CreatedAt) < 7*24*time.Hour {
			hoursRemaining := 7*24 - int(time.Since(submission.CreatedAt).Hours())
			if hoursRemaining < 24 {
				return &ValidationError{
					Field:   "clip_url",
					Message: "This clip was recently rejected. You can resubmit it in less than 24 hours",
				}
			}
			daysRemaining := hoursRemaining / 24
			return &ValidationError{
				Field:   "clip_url",
				Message: fmt.Sprintf("This clip was recently rejected. You can resubmit it in %d days", daysRemaining),
			}
		}
	}

	return nil
}

// fetchClipFromTwitch fetches clip metadata from Twitch API
func (s *SubmissionService) fetchClipFromTwitch(ctx context.Context, clipID string) (*twitch.Clip, error) {
	if s.twitchClient == nil {
		if s.testFixturesEnabled {
			// Try to hydrate from existing clip fixtures in the database
			if clip, err := s.clipRepo.GetByTwitchClipID(ctx, clipID); err == nil {
				broadcasterID := ""
				if clip.BroadcasterID != nil {
					broadcasterID = *clip.BroadcasterID
				}

				creatorID := ""
				if clip.CreatorID != nil {
					creatorID = *clip.CreatorID
				}

				gameID := ""
				if clip.GameID != nil {
					gameID = *clip.GameID
				}

				language := ""
				if clip.Language != nil {
					language = *clip.Language
				}

				thumbnailURL := ""
				if clip.ThumbnailURL != nil {
					thumbnailURL = *clip.ThumbnailURL
				}

				duration := 0.0
				if clip.Duration != nil {
					duration = *clip.Duration
				}

				return &twitch.Clip{
					ID:              clip.TwitchClipID,
					URL:             clip.TwitchClipURL,
					EmbedURL:        clip.EmbedURL,
					BroadcasterID:   broadcasterID,
					BroadcasterName: clip.BroadcasterName,
					CreatorID:       creatorID,
					CreatorName:     clip.CreatorName,
					GameID:          gameID,
					Language:        language,
					Title:           clip.Title,
					ViewCount:       clip.ViewCount,
					CreatedAt:       clip.CreatedAt,
					ThumbnailURL:    thumbnailURL,
					Duration:        duration,
				}, nil
			}
		}
		return nil, fmt.Errorf("Twitch API is not configured")
	}

	params := &twitch.ClipParams{
		ClipIDs: []string{clipID},
	}

	resp, err := s.twitchClient.GetClips(ctx, params)
	if err != nil {
		return nil, &ValidationError{
			Field:   "clip_url",
			Message: "Unable to fetch clip information from Twitch. Please verify the URL is correct and the clip exists.",
		}
	}

	if len(resp.Data) == 0 {
		return nil, &ValidationError{
			Field:   "clip_url",
			Message: "This clip was not found on Twitch. It may have been deleted or the URL is incorrect.",
		}
	}

	return &resp.Data[0], nil
}

// validateClipQuality validates clip meets quality requirements
func (s *SubmissionService) validateClipQuality(clip *twitch.Clip) error {
	// Check if clip is too old (>6 months)
	if time.Since(clip.CreatedAt) > 6*30*24*time.Hour {
		now := time.Now()
		years := now.Year() - clip.CreatedAt.Year()
		months := int(now.Month()) - int(clip.CreatedAt.Month())
		ageInMonths := years*12 + months
		if now.Day() < clip.CreatedAt.Day() {
			ageInMonths--
		}
		return &ValidationError{
			Field:   "clip",
			Message: fmt.Sprintf("This clip is too old (%d months). Only clips less than 6 months old can be submitted.", ageInMonths),
		}
	}

	// Check if clip is too short (<5 seconds)
	if clip.Duration < 5.0 {
		return &ValidationError{
			Field:   "clip",
			Message: fmt.Sprintf("This clip is too short (%.1f seconds). Clips must be at least 5 seconds long.", clip.Duration),
		}
	}

	// Check if clip has valid metadata
	if clip.Title == "" || clip.BroadcasterName == "" {
		return &ValidationError{
			Field:   "clip",
			Message: "This clip has missing or invalid metadata from Twitch. Please try a different clip.",
		}
	}

	return nil
}

// shouldAutoApprove determines if a submission should be auto-approved
func (s *SubmissionService) shouldAutoApprove(user *models.User) bool {
	// Admins and moderators are auto-approved
	if user.Role == "admin" || user.Role == "moderator" {
		return true
	}

	// High karma users (>1000) are auto-approved
	if user.KarmaPoints >= 1000 {
		return true
	}

	return false
}

func clipSourceMetadata(submission *models.ClipSubmission) json.RawMessage {
	if len(submission.SourceMetadata) == 0 {
		return json.RawMessage(`{}`)
	}
	return submission.SourceMetadata
}

func clipSourceString(values ...any) string {
	for _, value := range values {
		switch v := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				return trimmed
			}
		case *string:
			if v != nil && strings.TrimSpace(*v) != "" {
				return strings.TrimSpace(*v)
			}
		}
	}
	return ""
}

func legacyTwitchClipIDHash(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func legacyExternalTwitchClipID(platform, sourceID, sourceURL string) string {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		platform = "unknown"
	}
	identity := strings.TrimSpace(sourceID)
	if identity == "" {
		identity = strings.TrimSpace(sourceURL)
	}
	return fmt.Sprintf("external:%s:%s", platform, legacyTwitchClipIDHash(identity))
}

func legacyUploadTwitchClipID(submissionID, storageKey string) string {
	if id := strings.TrimSpace(submissionID); id != "" {
		return "upload:" + id
	}
	return fmt.Sprintf("upload:%s", legacyTwitchClipIDHash(storageKey))
}

func clipDurationSeconds(submission *models.ClipSubmission) *int {
	if submission.DurationSeconds != nil {
		v := *submission.DurationSeconds
		return &v
	}
	if submission.Duration != nil {
		v := int(*submission.Duration)
		return &v
	}
	return nil
}

func clipSourceFloat(duration *float64) *float64 {
	if duration == nil {
		return nil
	}
	v := *duration
	return &v
}

func clipSourcePlatformValue(submission *models.ClipSubmission) string {
	platform := strings.TrimSpace(submission.SourcePlatform)
	if platform == "" {
		return string(SourcePlatformTwitch)
	}
	return platform
}

func clipSourceURLFromSubmission(submission *models.ClipSubmission) *string {
	if submission.SourceURL != nil && strings.TrimSpace(*submission.SourceURL) != "" {
		url := strings.TrimSpace(*submission.SourceURL)
		return &url
	}
	if strings.TrimSpace(submission.TwitchClipURL) != "" {
		url := strings.TrimSpace(submission.TwitchClipURL)
		return &url
	}
	return nil
}

func (s *SubmissionService) buildClipFromSubmission(ctx context.Context, submission *models.ClipSubmission) (*models.Clip, error) {
	emptyStr := ""
	title := utils.StringOrDefault(submission.CustomTitle, submission.Title)
	creatorName := utils.StringOrDefault(submission.CreatorName, &emptyStr)
	broadcasterNameFallback := utils.StringOrDefault(submission.BroadcasterName, &creatorName)
	broadcasterName := utils.StringOrDefault(submission.BroadcasterNameOverride, &broadcasterNameFallback)
	metadata := clipSourceMetadata(submission)
	now := time.Now()

	buildCommon := func(sourceType string, sourcePlatform string, sourceID string, sourceURL *string) *models.Clip {
		var clipSourceID *string
		if strings.TrimSpace(sourceID) != "" {
			clipSourceID = utils.StringPtr(strings.TrimSpace(sourceID))
		}
		return &models.Clip{
			ID:                uuid.New(),
			TwitchClipID:      submission.TwitchClipID,
			TwitchClipURL:     submission.TwitchClipURL,
			EmbedURL:          clipEmbedURL(sourceType, submission, sourceURL),
			Title:             title,
			CreatorName:       creatorName,
			CreatorID:         submission.CreatorID,
			BroadcasterName:   broadcasterName,
			BroadcasterID:     submission.BroadcasterID,
			GameID:            submission.GameID,
			GameName:          submission.GameName,
			Language:          nil,
			ThumbnailURL:      submission.ThumbnailURL,
			Duration:          clipSourceFloat(submission.Duration),
			ViewCount:         submission.ViewCount,
			CreatedAt:         now,
			ImportedAt:        now,
			IsNSFW:            submission.IsNSFW,
			IsRemoved:         false,
			SubmittedByUserID: &submission.UserID,
			SubmittedAt:       &submission.CreatedAt,
			SourceType:        sourceType,
			SourcePlatform:    sourcePlatform,
			SourceURL:         sourceURL,
			SourceID:          clipSourceID,
			SourceMetadata:    metadata,
			DurationSeconds:   clipDurationSeconds(submission),
			DurationVerified:  submission.DurationVerified,
			StorageProvider:   submission.StorageProvider,
			StorageBucket:     submission.StorageBucket,
			StorageKey:        submission.StorageKey,
			OriginalFilename:  submission.OriginalFilename,
			MimeType:          submission.MimeType,
			FileSizeBytes:     submission.FileSizeBytes,
			StreamSource:      utils.StringPtr(sourceType),
			Status:            utils.StringPtr("ready"),
		}
	}

	switch strings.TrimSpace(submission.SourceType) {
	case string(SourceTypeTwitch):
		clip := buildCommon(string(SourceTypeTwitch), string(SourcePlatformTwitch), submission.TwitchClipID, clipSourceURLFromSubmission(submission))
		clip.SourceID = utils.StringPtr(submission.TwitchClipID)
		clip.EmbedURL = fmt.Sprintf("https://clips.twitch.tv/embed?clip=%s", submission.TwitchClipID)
		clip.SourceURL = clipSourceURLFromSubmission(submission)
		clip.StorageProvider = nil
		clip.StorageBucket = nil
		clip.StorageKey = nil
		clip.OriginalFilename = nil
		clip.MimeType = nil
		clip.FileSizeBytes = nil
		clip.VideoURL = nil
		return clip, nil
	case string(SourceTypeExternal):
		clip := buildCommon(string(SourceTypeExternal), clipSourcePlatformValue(submission), clipSourceString(submission.SourceID, submission.TwitchClipID), clipSourceURLFromSubmission(submission))
		clip.SourceURL = clipSourceURLFromSubmission(submission)
		clip.SourceID = utils.StringPtr(clipSourceString(submission.SourceID, submission.TwitchClipID))
		clip.TwitchClipURL = clipSourceString(clip.SourceURL, submission.TwitchClipURL)
		clip.EmbedURL = clipMetadataEmbedURL(submission)
		if clip.EmbedURL == "" {
			if clip.SourceURL != nil {
				clip.EmbedURL = *clip.SourceURL
			} else {
				clip.EmbedURL = clip.TwitchClipURL
			}
		}
		clip.VideoURL = nil
		clip.StorageProvider = nil
		clip.StorageBucket = nil
		clip.StorageKey = nil
		clip.OriginalFilename = nil
		clip.MimeType = nil
		clip.FileSizeBytes = nil
		return clip, nil
	case "upload":
		if s.clipStorage == nil {
			return nil, fmt.Errorf("clip storage is not configured")
		}
		if submission.StorageKey == nil || strings.TrimSpace(*submission.StorageKey) == "" {
			return nil, fmt.Errorf("upload submission is missing storage key")
		}
		storageKey := strings.TrimSpace(*submission.StorageKey)
		if !storage.IsPublicClipStorageKey(storageKey) {
			publicKey, ok := storage.PublicClipStorageKeyFromPendingKey(storageKey)
			if !ok {
				return nil, fmt.Errorf("upload submission storage key must use the pending prefix")
			}
			contentType := ""
			if submission.MimeType != nil {
				contentType = strings.TrimSpace(*submission.MimeType)
			}
			if _, err := s.clipStorage.CopyObject(ctx, storageKey, publicKey, contentType); err != nil {
				return nil, fmt.Errorf("failed to promote uploaded clip: %w", err)
			}
			storageKey = publicKey
		}
		publicURL := strings.TrimSpace(s.clipStorage.PublicURL(storageKey))
		if publicURL == "" {
			return nil, fmt.Errorf("failed to resolve public url for upload")
		}
		clip := buildCommon("upload", "upload", clipSourceString(submission.SourceID, submission.TwitchClipID), utils.StringPtr(publicURL))
		clip.TwitchClipURL = publicURL
		clip.EmbedURL = publicURL
		clip.VideoURL = utils.StringPtr(publicURL)
		clip.SourceURL = utils.StringPtr(publicURL)
		clip.StorageProvider = submission.StorageProvider
		clip.StorageBucket = submission.StorageBucket
		clip.StorageKey = utils.StringPtr(storageKey)
		clip.OriginalFilename = submission.OriginalFilename
		clip.MimeType = submission.MimeType
		clip.FileSizeBytes = submission.FileSizeBytes
		return clip, nil
	default:
		return nil, fmt.Errorf("unsupported submission source type %q", submission.SourceType)
	}
}

func clipEmbedURL(sourceType string, submission *models.ClipSubmission, sourceURL *string) string {
	switch sourceType {
	case string(SourceTypeTwitch):
		return fmt.Sprintf("https://clips.twitch.tv/embed?clip=%s", submission.TwitchClipID)
	case string(SourceTypeExternal):
		if embedURL := clipMetadataEmbedURL(submission); embedURL != "" {
			return embedURL
		}
		if sourceURL != nil {
			return *sourceURL
		}
		return submission.TwitchClipURL
	case "upload":
		if sourceURL != nil {
			return *sourceURL
		}
		return submission.TwitchClipURL
	default:
		if sourceURL != nil {
			return *sourceURL
		}
		return submission.TwitchClipURL
	}
}

func clipMetadataEmbedURL(submission *models.ClipSubmission) string {
	if len(submission.SourceMetadata) == 0 {
		return ""
	}
	var metadata map[string]any
	if err := json.Unmarshal(submission.SourceMetadata, &metadata); err != nil {
		return ""
	}
	if raw, ok := metadata["embed_url"]; ok && raw != nil {
		return strings.TrimSpace(fmt.Sprint(raw))
	}
	return ""
}

// createClipFromSubmission creates a clip in the main clips table
func (s *SubmissionService) createClipFromSubmission(ctx context.Context, submission *models.ClipSubmission) (uuid.UUID, error) {
	clip, err := s.buildClipFromSubmission(ctx, submission)
	if err != nil {
		return uuid.Nil, err
	}

	if clip.SourceType == string(SourceTypeTwitch) {
		clip.StreamSource = utils.StringPtr(string(SourceTypeTwitch))
		clip.Status = utils.StringPtr("ready")
	}

	if err := s.createClipRecord(ctx, clip); err != nil {
		return uuid.Nil, err
	}

	// Auto-upvote: Create an upvote from the submitter
	// This encourages engagement and shows creator approval
	if s.voteRepo != nil {
		if err := s.voteRepo.UpsertVote(ctx, submission.UserID, clip.ID, 1); err != nil {
			// Log error but don't fail the clip creation
			s.logger.Warn("Failed to auto-upvote clip for user", map[string]interface{}{
				"user_id": submission.UserID,
				"clip_id": clip.ID,
				"error":   err.Error(),
			})
		}
	}

	// Invalidate feed caches so the new clip appears immediately
	if s.cacheService != nil {
		if err := s.cacheService.InvalidateOnNewClip(ctx, clip); err != nil {
			// Log error but don't fail the clip creation
			s.logger.Warn("Failed to invalidate feed caches for clip", map[string]interface{}{
				"clip_id": clip.ID,
				"error":   err.Error(),
			})
		}
	}

	return clip.ID, nil
}

// awardKarma awards karma points to a user
func (s *SubmissionService) awardKarma(ctx context.Context, userID uuid.UUID, points int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.KarmaPoints += points
	return s.userRepo.Update(ctx, user)
}

// getClipTitle returns the clip title, preferring custom title over original title
func getClipTitle(submission *models.ClipSubmission) string {
	if submission.CustomTitle != nil && *submission.CustomTitle != "" {
		return *submission.CustomTitle
	}
	if submission.Title != nil {
		return *submission.Title
	}
	return ""
}

// ApproveSubmission approves a submission and creates the clip
func (s *SubmissionService) ApproveSubmission(ctx context.Context, submissionID, reviewerID uuid.UUID) error {
	submission, err := s.submissionRepo.GetByID(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	if submission.Status != "pending" {
		return fmt.Errorf("submission is not pending")
	}
	if err := s.ensureSubmissionCanBeApproved(submission); err != nil {
		return err
	}

	// Create clip
	clipID, err := s.createClipFromSubmission(ctx, submission)
	if err != nil {
		return fmt.Errorf("failed to create clip: %w", err)
	}

	// Update submission status and clip ID
	if err := s.submissionRepo.UpdateStatus(ctx, submissionID, "approved", reviewerID, nil); err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	// Update submission with clip ID
	if err := s.submissionRepo.UpdateClipID(ctx, submissionID, clipID); err != nil {
		return fmt.Errorf("failed to update submission clip ID: %w", err)
	}
	if s.clipStorage != nil && submission.StorageKey != nil {
		if err := s.clipStorage.DeleteObject(ctx, strings.TrimSpace(*submission.StorageKey)); err != nil {
			log.Printf("Failed to delete pending upload after approval: %v\n", err)
		}
	}

	// Create audit log
	if s.auditLogRepo != nil {
		auditLog := &models.ModerationAuditLog{
			ID:          uuid.New(),
			Action:      "approve",
			EntityType:  "clip_submission",
			EntityID:    submissionID,
			ModeratorID: reviewerID,
			CreatedAt:   time.Now(),
		}
		if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to create audit log: %v\n", err)
		}
	}

	// Award karma to submitter
	if err := s.awardKarma(ctx, submission.UserID, 10); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to award karma: %v\n", err)
	}

	// Send notification to submitter
	if s.notificationService != nil {
		clipTitle := getClipTitle(submission)
		if err := s.notificationService.NotifySubmissionApproved(ctx, submission.UserID, submissionID, clipTitle); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to send notification: %v\n", err)
		}
	}

	// Trigger webhook for approval
	if s.webhookService != nil {
		webhookData := map[string]interface{}{
			"submission_id":   submissionID.String(),
			"user_id":         submission.UserID.String(),
			"twitch_clip_id":  submission.TwitchClipID,
			"twitch_clip_url": submission.TwitchClipURL,
			"reviewer_id":     reviewerID.String(),
			"approved_at":     time.Now(),
		}
		if submission.CustomTitle != nil {
			webhookData["custom_title"] = *submission.CustomTitle
		}

		if err := s.webhookService.TriggerEvent(ctx, models.WebhookEventClipApproved, submissionID, webhookData); err != nil {
			log.Printf("Failed to trigger webhook event: %v", err)
		}
	}

	return nil
}

// RejectSubmission rejects a submission
func (s *SubmissionService) RejectSubmission(ctx context.Context, submissionID, reviewerID uuid.UUID, reason string) error {
	submission, err := s.submissionRepo.GetByID(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	if submission.Status != "pending" {
		return fmt.Errorf("submission is not pending")
	}

	// Update submission status
	if err := s.submissionRepo.UpdateStatus(ctx, submissionID, "rejected", reviewerID, &reason); err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	// Create audit log
	if s.auditLogRepo != nil {
		auditLog := &models.ModerationAuditLog{
			ID:          uuid.New(),
			Action:      "reject",
			EntityType:  "clip_submission",
			EntityID:    submissionID,
			ModeratorID: reviewerID,
			Reason:      &reason,
			CreatedAt:   time.Now(),
		}
		if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to create audit log: %v\n", err)
		}
	}

	// Penalize karma
	if err := s.awardKarma(ctx, submission.UserID, -5); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to penalize karma: %v\n", err)
	}

	// Send notification to submitter
	if s.notificationService != nil {
		clipTitle := getClipTitle(submission)
		if err := s.notificationService.NotifySubmissionRejected(ctx, submission.UserID, submissionID, clipTitle, reason); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to send notification: %v\n", err)
		}
	}

	// Trigger webhook for rejection
	if s.webhookService != nil {
		webhookData := map[string]interface{}{
			"submission_id":    submissionID.String(),
			"user_id":          submission.UserID.String(),
			"twitch_clip_id":   submission.TwitchClipID,
			"twitch_clip_url":  submission.TwitchClipURL,
			"reviewer_id":      reviewerID.String(),
			"rejection_reason": reason,
			"rejected_at":      time.Now(),
		}
		if submission.CustomTitle != nil {
			webhookData["custom_title"] = *submission.CustomTitle
		}

		if err := s.webhookService.TriggerEvent(ctx, models.WebhookEventClipRejected, submissionID, webhookData); err != nil {
			log.Printf("Failed to trigger webhook event: %v", err)
		}
	}

	return nil
}

// BulkApproveSubmissions approves multiple submissions
func (s *SubmissionService) BulkApproveSubmissions(ctx context.Context, submissionIDs []uuid.UUID, reviewerID uuid.UUID) error {
	// Get submissions to validate they're all pending
	submissions, err := s.submissionRepo.GetByIDs(ctx, submissionIDs)
	if err != nil {
		return fmt.Errorf("failed to get submissions: %w", err)
	}

	// Validate all are pending
	for _, submission := range submissions {
		if submission.Status != "pending" {
			return fmt.Errorf("submission %s is not pending", submission.ID)
		}
		if err := s.ensureSubmissionCanBeApproved(submission); err != nil {
			return err
		}
	}

	// Create clips for all submissions
	for _, submission := range submissions {
		if _, err := s.createClipFromSubmission(ctx, submission); err != nil {
			return fmt.Errorf("failed to create clip for submission %s: %w", submission.ID, err)
		}
	}

	// Bulk update status
	if err := s.submissionRepo.BulkUpdateStatus(ctx, submissionIDs, "approved", reviewerID, nil); err != nil {
		return fmt.Errorf("failed to bulk update submission status: %w", err)
	}
	if s.clipStorage != nil {
		for _, submission := range submissions {
			if submission.StorageKey == nil {
				continue
			}
			if err := s.clipStorage.DeleteObject(ctx, strings.TrimSpace(*submission.StorageKey)); err != nil {
				log.Printf("Failed to delete pending upload after bulk approval for submission %s: %v\n", submission.ID, err)
			}
		}
	}

	// Create audit log
	if s.auditLogRepo != nil {
		metadata := map[string]interface{}{
			"submission_count": len(submissionIDs),
			"submission_ids":   submissionIDs,
		}
		auditLog := &models.ModerationAuditLog{
			ID:          uuid.New(),
			Action:      "bulk_approve",
			EntityType:  "clip_submission",
			EntityID:    uuid.Nil, // Use uuid.Nil as entity ID for bulk actions
			ModeratorID: reviewerID,
			Metadata:    metadata,
			CreatedAt:   time.Now(),
		}
		if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to create audit log: %v\n", err)
		}
	}

	// Award karma to submitters
	for _, submission := range submissions {
		if err := s.awardKarma(ctx, submission.UserID, 10); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to award karma: %v\n", err)
		}
	}

	return nil
}

// BulkRejectSubmissions rejects multiple submissions
func (s *SubmissionService) BulkRejectSubmissions(ctx context.Context, submissionIDs []uuid.UUID, reviewerID uuid.UUID, reason string) error {
	// Get submissions to validate they're all pending
	submissions, err := s.submissionRepo.GetByIDs(ctx, submissionIDs)
	if err != nil {
		return fmt.Errorf("failed to get submissions: %w", err)
	}

	// Validate all are pending
	for _, submission := range submissions {
		if submission.Status != "pending" {
			return fmt.Errorf("submission %s is not pending", submission.ID)
		}
	}

	// Bulk update status
	if err := s.submissionRepo.BulkUpdateStatus(ctx, submissionIDs, "rejected", reviewerID, &reason); err != nil {
		return fmt.Errorf("failed to bulk update submission status: %w", err)
	}

	// Create audit log
	if s.auditLogRepo != nil {
		metadata := map[string]interface{}{
			"submission_count": len(submissionIDs),
			"submission_ids":   submissionIDs,
		}
		auditLog := &models.ModerationAuditLog{
			ID:          uuid.New(),
			Action:      "bulk_reject",
			EntityType:  "clip_submission",
			EntityID:    uuid.Nil, // No single entity; use Nil UUID for bulk actions
			ModeratorID: reviewerID,
			Reason:      &reason,
			Metadata:    metadata,
			CreatedAt:   time.Now(),
		}
		if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to create audit log: %v\n", err)
		}
	}

	// Penalize karma
	for _, submission := range submissions {
		if err := s.awardKarma(ctx, submission.UserID, -5); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to penalize karma: %v\n", err)
		}
	}

	return nil
}

// GetUserSubmissions retrieves submissions for a user
func (s *SubmissionService) GetUserSubmissions(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.ClipSubmission, int, error) {
	return s.submissionRepo.ListByUser(ctx, userID, page, limit)
}

// GetPendingSubmissions retrieves pending submissions for moderation
func (s *SubmissionService) GetPendingSubmissions(ctx context.Context, page, limit int) ([]*models.ClipSubmissionWithUser, int, error) {
	return s.submissionRepo.ListPending(ctx, page, limit)
}

// GetPendingSubmissionsWithFilters retrieves pending submissions with filters
func (s *SubmissionService) GetPendingSubmissionsWithFilters(ctx context.Context, filters repository.SubmissionFilters, page, limit int) ([]*models.ClipSubmissionWithUser, int, error) {
	return s.submissionRepo.ListPendingWithFilters(ctx, filters, page, limit)
}

// GetSubmissionStats retrieves submission statistics for a user
func (s *SubmissionService) GetSubmissionStats(ctx context.Context, userID uuid.UUID) (*models.SubmissionStats, error) {
	return s.submissionRepo.GetUserStats(ctx, userID)
}

// extractClipIDFromURL extracts the clip ID from a Twitch clip URL or returns the ID if already provided
func extractClipIDFromURL(clipURLOrID string) string {
	// Trim whitespace first
	clipURLOrID = strings.TrimSpace(clipURLOrID)

	// Handle trailing slashes
	clipURLOrID = strings.TrimSuffix(clipURLOrID, "/")

	// If it's empty, return empty
	if len(clipURLOrID) == 0 {
		return ""
	}

	// If it's already just an ID (not starting with http), return it
	if !strings.HasPrefix(clipURLOrID, "http") {
		// Still need to strip query params and fragments if someone passes "ClipID?param=value"
		clipID := clipURLOrID
		if idx := strings.IndexAny(clipID, "?#"); idx != -1 {
			clipID = clipID[:idx]
		}
		return clipID
	}

	// Handle full URLs - find the last path segment
	parts := []rune(clipURLOrID)
	lastSlash := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || lastSlash == len(parts)-1 {
		return ""
	}

	clipID := string(parts[lastSlash+1:])

	// Remove query parameters and fragment identifiers if present
	if idx := strings.IndexAny(clipID, "?#"); idx != -1 {
		clipID = clipID[:idx]
	}

	return clipID
}
