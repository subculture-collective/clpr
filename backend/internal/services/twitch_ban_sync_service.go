package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// TwitchClientInterface defines the interface for Twitch client operations
type TwitchClientInterface interface {
	GetBannedUsers(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error)
}

// TwitchAuthRepositoryInterface defines the interface for Twitch auth operations
type TwitchAuthRepositoryInterface interface {
	GetTwitchAuth(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error)
	IsTokenExpired(auth *models.TwitchAuth) bool
}

// TwitchBanRepositoryInterface defines the interface for ban repository operations
type TwitchBanRepositoryInterface interface {
	BatchUpsertBans(ctx context.Context, bans []*repository.TwitchBan) error
}

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	GetByTwitchID(ctx context.Context, twitchID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

// TwitchBanSyncService handles syncing banned users from Twitch API
type TwitchBanSyncService struct {
	twitchClient   TwitchClientInterface
	twitchAuthRepo TwitchAuthRepositoryInterface
	banRepo        TwitchBanRepositoryInterface
	userRepo       UserRepositoryInterface
}

// NewTwitchBanSyncService creates a new TwitchBanSyncService
func NewTwitchBanSyncService(
	twitchClient TwitchClientInterface,
	twitchAuthRepo TwitchAuthRepositoryInterface,
	banRepo TwitchBanRepositoryInterface,
	userRepo UserRepositoryInterface,
) *TwitchBanSyncService {
	return &TwitchBanSyncService{
		twitchClient:   twitchClient,
		twitchAuthRepo: twitchAuthRepo,
		banRepo:        banRepo,
		userRepo:       userRepo,
	}
}

// SyncChannelBans fetches and syncs banned users for a specific channel
// userID: the ID of the user requesting the sync (must own the channel)
// channelID: the Twitch channel ID to sync bans for
func (s *TwitchBanSyncService) SyncChannelBans(ctx context.Context, userID, channelID string) error {
	start := time.Now()
	logger := utils.GetLogger()
	var syncErr error
	var errorType string
	
	// Defer metrics recording
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.ModerationSyncOperationDuration.WithLabelValues("full").Observe(duration)
		
		if syncErr == nil {
			metrics.ModerationSyncOperationsTotal.WithLabelValues("success", "").Inc()
		} else {
			// Determine error type
			var authErr *AuthenticationError
			var authzErr *AuthorizationError
			var apiErr *BanSyncTwitchAPIError
			var dbErr *DatabaseError
			
			if errors.As(syncErr, &authErr) {
				errorType = "auth_error"
			} else if errors.As(syncErr, &authzErr) {
				errorType = "authz_error"
			} else if errors.As(syncErr, &apiErr) {
				errorType = "api_error"
			} else if errors.As(syncErr, &dbErr) {
				errorType = "database_error"
			} else {
				errorType = "unknown_error"
			}
			metrics.ModerationSyncOperationsTotal.WithLabelValues("failed", errorType).Inc()
		}
	}()

	// Parse UUIDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		syncErr = fmt.Errorf("invalid user ID: %w", err)
		return syncErr
	}

	// Get user's Twitch auth credentials
	auth, err := s.twitchAuthRepo.GetTwitchAuth(ctx, userUUID)
	if err != nil {
		logger.Error("Failed to get Twitch auth", err, map[string]interface{}{
			"user_id": userID,
		})
		syncErr = &AuthenticationError{Message: "user not authenticated with Twitch"}
		return syncErr
	}

	if auth == nil {
		syncErr = &AuthenticationError{Message: "user not authenticated with Twitch"}
		return syncErr
	}

	// Check if token is expired and refresh if needed
	if s.twitchAuthRepo.IsTokenExpired(auth) {
		logger.Info("Twitch token expired, refreshing", map[string]interface{}{
			"user_id": userID,
		})

		// Refresh token logic would go here
		// For now, return an error if token is expired
		syncErr = &AuthenticationError{Message: "Twitch token expired, please re-authenticate"}
		return syncErr
	}

	// Validate user owns the channel by checking if their Twitch user ID matches the channel ID
	if auth.TwitchUserID != channelID {
		logger.Warn("User attempted to sync bans for channel they don't own", map[string]interface{}{
			"user_id":           userID,
			"user_twitch_id":    auth.TwitchUserID,
			"requested_channel": channelID,
		})
		syncErr = &AuthorizationError{Message: "user does not own the specified channel"}
		return syncErr
	}

	logger.Info("Starting ban sync", map[string]interface{}{
		"user_id":    userID,
		"channel_id": channelID,
	})

	// Fetch all banned users with pagination
	var allBans []twitch.BannedUser
	cursor := ""
	pageCount := 0

	for {
		pageCount++

		// Fetch banned users from Twitch with exponential backoff for rate limiting
		var bansResp *twitch.BannedUsersResponse
		err := s.retryWithBackoff(ctx, func() error {
			var fetchErr error
			bansResp, fetchErr = s.twitchClient.GetBannedUsers(ctx, channelID, auth.AccessToken, 100, cursor)
			return fetchErr
		})

		if err != nil {
			logger.Error("Failed to fetch banned users from Twitch", err, map[string]interface{}{
				"user_id":    userID,
				"channel_id": channelID,
				"page":       pageCount,
			})
			syncErr = &BanSyncTwitchAPIError{Message: "failed to fetch banned users from Twitch", Err: err}
			return syncErr
		}

		logger.Debug("Fetched page of banned users", map[string]interface{}{
			"page":     pageCount,
			"count":    len(bansResp.Data),
			"has_more": bansResp.Pagination.Cursor != "",
		})

		allBans = append(allBans, bansResp.Data...)

		// Check for more pages
		if bansResp.Pagination.Cursor == "" {
			break
		}
		cursor = bansResp.Pagination.Cursor
	}

	logger.Info("Fetched all banned users", map[string]interface{}{
		"user_id":    userID,
		"channel_id": channelID,
		"total_bans": len(allBans),
		"pages":      pageCount,
	})

	// Record bans processed
	metrics.ModerationSyncBansProcessed.WithLabelValues("fetched").Add(float64(len(allBans)))

	// Convert channel ID string to UUID
	channelUUID, err := s.getOrCreateChannelUUID(ctx, channelID)
	if err != nil {
		logger.Error("Failed to get/create channel UUID", err, map[string]interface{}{
			"channel_id": channelID,
		})
		syncErr = &DatabaseError{Message: "failed to process channel", Err: err}
		return syncErr
	}

	// Transform Twitch bans to database records
	now := time.Now()
	dbBans := make([]*repository.TwitchBan, 0, len(allBans))
	failedUserCreations := 0

	for _, twitchBan := range allBans {
		// Get or create user for banned user
		bannedUserUUID, err := s.getOrCreateUserByTwitchID(ctx, twitchBan.UserID, twitchBan.UserLogin, twitchBan.UserName)
		if err != nil {
			logger.Error("Failed to get/create banned user", err, map[string]interface{}{
				"twitch_user_id": twitchBan.UserID,
				"username":       twitchBan.UserLogin,
			})
			failedUserCreations++
			// Continue with other bans instead of failing completely
			continue
		}

		var expiresAt *time.Time
		// Check if ban has expiration (temporary ban)
		if !twitchBan.ExpiresAt.IsZero() {
			expiresAt = &twitchBan.ExpiresAt
		}

		var reason *string
		if twitchBan.Reason != "" {
			reason = &twitchBan.Reason
		}

		twitchBanID := fmt.Sprintf("%s:%s", channelID, twitchBan.UserID)

		dbBan := &repository.TwitchBan{
			ChannelID:        channelUUID,
			BannedUserID:     bannedUserUUID,
			Reason:           reason,
			BannedAt:         twitchBan.CreatedAt,
			ExpiresAt:        expiresAt,
			SyncedFromTwitch: true,
			TwitchBanID:      &twitchBanID,
			LastSyncedAt:     &now,
		}

		dbBans = append(dbBans, dbBan)
	}

	// Store bans in database atomically
	if len(dbBans) > 0 {
		err = s.banRepo.BatchUpsertBans(ctx, dbBans)
		if err != nil {
			logger.Error("Failed to store bans in database", err, map[string]interface{}{
				"user_id":    userID,
				"channel_id": channelID,
				"ban_count":  len(dbBans),
			})
			syncErr = &DatabaseError{Message: "failed to store bans in database", Err: err}
			return syncErr
		}
		// Record successfully processed bans
		metrics.ModerationSyncBansProcessed.WithLabelValues("stored").Add(float64(len(dbBans)))
	}

	logger.Info("Ban sync completed successfully", map[string]interface{}{
		"user_id":               userID,
		"channel_id":            channelID,
		"bans_synced":           len(dbBans),
		"pages":                 pageCount,
		"failed_user_creations": failedUserCreations,
	})

	return nil
}

// retryWithBackoff retries a function with exponential backoff
func (s *TwitchBanSyncService) retryWithBackoff(ctx context.Context, fn func() error) error {
	maxRetries := 3
	baseDelay := time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error
		var rateLimitErr *twitch.RateLimitError
		if errors.As(err, &rateLimitErr) {
			delay := time.Duration(rateLimitErr.RetryAfter) * time.Second
			if delay == 0 {
				delay = baseDelay * time.Duration(1<<uint(attempt))
			}

			logger := utils.GetLogger()
			logger.Warn("Rate limited by Twitch, retrying", map[string]interface{}{
				"attempt":     attempt + 1,
				"max_retries": maxRetries,
				"delay":       delay.String(),
			})

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// For other errors, use exponential backoff
		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(attempt))

			logger := utils.GetLogger()
			logger.Warn("Request failed, retrying", map[string]interface{}{
				"attempt":     attempt + 1,
				"max_retries": maxRetries,
				"delay":       delay.String(),
				"error":       err.Error(),
			})

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

// getOrCreateChannelUUID converts a Twitch channel ID to a UUID
// For now, this creates a deterministic UUID from the channel ID
func (s *TwitchBanSyncService) getOrCreateChannelUUID(ctx context.Context, channelID string) (uuid.UUID, error) {
	// Try to find user by Twitch ID
	user, err := s.userRepo.GetByTwitchID(ctx, channelID)
	if err == nil && user != nil {
		return user.ID, nil
	}

	// If user not found, we need to create a placeholder or return error
	// For this implementation, we'll return an error as channels should be users
	if err == repository.ErrUserNotFound {
		return uuid.Nil, fmt.Errorf("channel user not found for Twitch ID: %s", channelID)
	}

	return uuid.Nil, err
}

// getOrCreateUserByTwitchID gets or creates a user by their Twitch ID
func (s *TwitchBanSyncService) getOrCreateUserByTwitchID(ctx context.Context, twitchID, login, displayName string) (uuid.UUID, error) {
	// Try to find existing user
	user, err := s.userRepo.GetByTwitchID(ctx, twitchID)
	if err == nil && user != nil {
		return user.ID, nil
	}

	// User not found, create unclaimed account
	if err == repository.ErrUserNotFound {
		logger := utils.GetLogger()
		logger.Info("Creating unclaimed user account", map[string]interface{}{
			"twitch_id": twitchID,
			"login":     login,
		})

		// Create unclaimed user
		newUser := &models.User{
			ID:            uuid.New(),
			TwitchID:      &twitchID,
			Username:      login,
			DisplayName:   displayName,
			Role:          "user",
			AccountType:   models.AccountTypeMember,
			AccountStatus: "unclaimed",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err = s.userRepo.Create(ctx, newUser)
		if err != nil {
			// Try to get the user in case it was created in a race condition
			user, getErr := s.userRepo.GetByTwitchID(ctx, twitchID)
			if getErr == nil && user != nil {
				return user.ID, nil
			}
			return uuid.Nil, fmt.Errorf("failed to create unclaimed user: %w", err)
		}

		return newUser.ID, nil
	}

	return uuid.Nil, err
}

// Error types
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %s", e.Message)
}

type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return fmt.Sprintf("authorization error: %s", e.Message)
}

type BanSyncTwitchAPIError struct {
	Message string
	Err     error
}

func (e *BanSyncTwitchAPIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Twitch API error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("Twitch API error: %s", e.Message)
}

func (e *BanSyncTwitchAPIError) Unwrap() error {
	return e.Err
}

type DatabaseError struct {
	Message string
	Err     error
}

func (e *DatabaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("database error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("database error: %s", e.Message)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}
