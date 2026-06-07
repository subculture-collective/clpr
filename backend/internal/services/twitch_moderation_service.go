package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// Sentinel errors for Twitch moderation operations
var (
	ErrTwitchNotAuthenticated  = errors.New("user not authenticated with Twitch")
	ErrTwitchScopeInsufficient = errors.New("insufficient Twitch scopes for this action")
	ErrTwitchNotBroadcaster    = errors.New("user is not the broadcaster for this channel")
	ErrTwitchNotModerator      = errors.New("user is not a Twitch moderator for this channel")
	ErrSiteModeratorsReadOnly  = errors.New("site moderators cannot perform Twitch ban actions - Twitch channel moderator status required")
)

// TwitchAuthRepository defines the interface for Twitch auth operations
type TwitchAuthRepository interface {
	GetTwitchAuth(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error)
	IsTokenExpired(auth *models.TwitchAuth) bool
}

// TwitchBanClient defines the interface for Twitch ban/unban operations
type TwitchBanClient interface {
	BanUser(ctx context.Context, broadcasterID string, moderatorID string, userAccessToken string, request *twitch.BanUserRequest) (*twitch.BanUserResponse, error)
	UnbanUser(ctx context.Context, broadcasterID string, moderatorID string, userID string, userAccessToken string) error
}

// TwitchModerationService handles Twitch-specific moderation operations with scope enforcement
type TwitchModerationService struct {
	twitchClient   TwitchBanClient
	twitchAuthRepo TwitchAuthRepository
	userRepo       ModerationUserRepo
	auditLogRepo   ModerationAuditRepo
}

// NewTwitchModerationService creates a new TwitchModerationService
func NewTwitchModerationService(
	twitchClient TwitchBanClient,
	twitchAuthRepo TwitchAuthRepository,
	userRepo ModerationUserRepo,
	auditLogRepo ModerationAuditRepo,
) *TwitchModerationService {
	return &TwitchModerationService{
		twitchClient:   twitchClient,
		twitchAuthRepo: twitchAuthRepo,
		userRepo:       userRepo,
		auditLogRepo:   auditLogRepo,
	}
}

// ValidateTwitchBanScope validates that a user can perform Twitch ban/unban actions
// This enforces:
// - User must have valid Twitch OAuth token
// - Token must have moderator:manage:banned_users OR channel:manage:banned_users scope
// - User must be either:
//   - The broadcaster for the channel (twitchUserID == broadcasterID), OR
//   - A Twitch-recognized moderator for that channel
//
// - Site moderators are blocked (they have Clipper permissions but not Twitch permissions)
func (s *TwitchModerationService) ValidateTwitchBanScope(ctx context.Context, userID uuid.UUID, broadcasterID string) (*models.TwitchAuth, error) {
	// Get user to check if they're a site moderator
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Site moderators are explicitly blocked from Twitch actions
	// They have Clipper-level moderation powers but not necessarily Twitch mod status
	if user.AccountType == models.AccountTypeModerator && user.ModeratorScope == models.ModeratorScopeSite {
		return nil, ErrSiteModeratorsReadOnly
	}

	// Get user's Twitch auth
	auth, err := s.twitchAuthRepo.GetTwitchAuth(ctx, userID)
	if err != nil || auth == nil {
		return nil, ErrTwitchNotAuthenticated
	}

	// Check token expiry
	if s.twitchAuthRepo.IsTokenExpired(auth) {
		return nil, &twitch.AuthError{Message: "Twitch token expired, please re-authenticate"}
	}

	// Validate scopes - user must have at least one of the required scopes
	requiredScopes := map[string]bool{
		"moderator:manage:banned_users": true, // For moderators
		"channel:manage:banned_users":   true, // For broadcasters
	}

	hasRequiredScope := false
	scopes := strings.Split(auth.Scopes, " ")
	for _, scope := range scopes {
		if requiredScopes[scope] {
			hasRequiredScope = true
			break
		}
	}

	if !hasRequiredScope {
		return nil, ErrTwitchScopeInsufficient
	}

	// Check if user is the broadcaster
	if auth.TwitchUserID == broadcasterID {
		// User is the broadcaster, they can manage bans in their channel
		return auth, nil
	}

	// User is not the broadcaster, they must be a Twitch moderator for this channel
	// This would require checking Twitch's moderator list for the channel
	// For now, we enforce that only the broadcaster can ban (stricter than Twitch)
	// Future enhancement: Call Twitch API to verify moderator status
	// GET https://api.twitch.tv/helix/moderation/moderators?broadcaster_id={broadcaster_id}&user_id={user_id}

	// For P0 implementation, only broadcasters can use Twitch ban actions
	return nil, ErrTwitchNotBroadcaster
}

// BanUserOnTwitch bans a user on Twitch via API
// Validates scope and permissions before calling Twitch API
func (s *TwitchModerationService) BanUserOnTwitch(ctx context.Context, moderatorUserID uuid.UUID, broadcasterID string, targetUserID string, reason *string, duration *int) error {
	// Record start time for latency metrics
	startTime := time.Now()
	action := "ban"
	var statusCode int
	var banErr error

	// Create a separate context for audit logging so it can complete even if the request context is cancelled
	auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ensure metrics and audit logs are emitted regardless of outcome
	defer func() {
		s.recordBanMetrics(action, startTime, statusCode, banErr)
		s.createBanAuditLog(auditCtx, action, moderatorUserID, broadcasterID, targetUserID, reason, duration, statusCode, banErr)
	}()

	// Validate scope and get auth
	auth, err := s.ValidateTwitchBanScope(ctx, moderatorUserID, broadcasterID)
	if err != nil {
		banErr = err
		return err
	}

	// Build ban request
	request := &twitch.BanUserRequest{
		UserID:   targetUserID,
		Duration: duration,
		Reason:   reason,
	}

	// Call Twitch API
	_, err = s.twitchClient.BanUser(ctx, broadcasterID, auth.TwitchUserID, auth.AccessToken, request)
	if err != nil {
		banErr = err
		// Extract status code from error
		statusCode = extractStatusCode(err)
		// Check for specific Twitch errors
		var authErr *twitch.AuthError
		if errors.As(err, &authErr) {
			return fmt.Errorf("Twitch authentication failed: %w", authErr)
		}
		return fmt.Errorf("failed to ban user on Twitch: %w", err)
	}

	statusCode = 200
	return nil
}

// UnbanUserOnTwitch unbans a user on Twitch via API
// Validates scope and permissions before calling Twitch API
func (s *TwitchModerationService) UnbanUserOnTwitch(ctx context.Context, moderatorUserID uuid.UUID, broadcasterID string, targetUserID string) error {
	// Record start time for latency metrics
	startTime := time.Now()
	action := "unban"
	var statusCode int
	var unbanErr error

	// Create a separate context for audit logging so it can complete even if the request context is cancelled
	auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ensure metrics and audit logs are emitted regardless of outcome
	defer func() {
		s.recordBanMetrics(action, startTime, statusCode, unbanErr)
		s.createBanAuditLog(auditCtx, action, moderatorUserID, broadcasterID, targetUserID, nil, nil, statusCode, unbanErr)
	}()

	// Validate scope and get auth
	auth, err := s.ValidateTwitchBanScope(ctx, moderatorUserID, broadcasterID)
	if err != nil {
		unbanErr = err
		return err
	}

	// Call Twitch API
	err = s.twitchClient.UnbanUser(ctx, broadcasterID, auth.TwitchUserID, targetUserID, auth.AccessToken)
	if err != nil {
		unbanErr = err
		// Extract status code from error
		statusCode = extractStatusCode(err)
		// Check for specific Twitch errors
		var authErr *twitch.AuthError
		if errors.As(err, &authErr) {
			return fmt.Errorf("Twitch authentication failed: %w", authErr)
		}
		return fmt.Errorf("failed to unban user on Twitch: %w", err)
	}

	statusCode = 200
	return nil
}

// categorizeError categorizes an error into a simple error code for metrics
func categorizeError(err error) string {
	if errors.Is(err, ErrTwitchNotAuthenticated) {
		return "not_authenticated"
	}
	if errors.Is(err, ErrTwitchScopeInsufficient) {
		return "insufficient_scope"
	}
	if errors.Is(err, ErrTwitchNotBroadcaster) {
		return "not_broadcaster"
	}
	if errors.Is(err, ErrTwitchNotModerator) {
		return "not_moderator"
	}
	if errors.Is(err, ErrSiteModeratorsReadOnly) {
		return "site_moderator_readonly"
	}

	var rateLimitErr *twitch.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return "rate_limited"
	}

	var modErr *twitch.ModerationError
	if errors.As(err, &modErr) {
		return string(modErr.Code)
	}

	var apiErr *twitch.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode >= 500 {
			return "server_error"
		}
		if apiErr.StatusCode >= 400 {
			return "client_error"
		}
	}

	return "unknown"
}

// extractStatusCode extracts HTTP status code from various error types
func extractStatusCode(err error) int {
	var apiErr *twitch.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}

	var modErr *twitch.ModerationError
	if errors.As(err, &modErr) {
		return modErr.StatusCode
	}

	return 0
}

// recordBanMetrics records Prometheus metrics for ban/unban operations
func (s *TwitchModerationService) recordBanMetrics(action string, startTime time.Time, statusCode int, err error) {
	// Record latency
	metrics.TwitchBanActionDuration.WithLabelValues(action).Observe(time.Since(startTime).Seconds())

	// Determine status and error code for metrics
	status := "success"
	errorCode := "none"

	if err != nil {
		status = "failed"
		errorCode = categorizeError(err)

		// Track specific error types
		if errors.Is(err, ErrTwitchScopeInsufficient) {
			metrics.TwitchBanPermissionErrors.WithLabelValues(action, "insufficient_scope").Inc()
		} else if errors.Is(err, ErrTwitchNotBroadcaster) {
			metrics.TwitchBanPermissionErrors.WithLabelValues(action, "not_broadcaster").Inc()
		} else if errors.Is(err, ErrTwitchNotAuthenticated) {
			metrics.TwitchBanPermissionErrors.WithLabelValues(action, "not_authenticated").Inc()
		} else if errors.Is(err, ErrSiteModeratorsReadOnly) {
			metrics.TwitchBanPermissionErrors.WithLabelValues(action, "site_moderator_readonly").Inc()
		}

		// Track rate limits - check both RateLimitError and ModerationError types
		var rateLimitErr *twitch.RateLimitError
		if errors.As(err, &rateLimitErr) {
			metrics.TwitchBanRateLimitHits.WithLabelValues(action).Inc()
		} else {
			// Only check ModerationError if it's not already a RateLimitError
			var modErr *twitch.ModerationError
			if errors.As(err, &modErr) && modErr.Code == twitch.ModerationErrorCodeRateLimited {
				metrics.TwitchBanRateLimitHits.WithLabelValues(action).Inc()
			}
		}

		// Track server errors
		if statusCode >= 500 {
			metrics.TwitchBanServerErrors.WithLabelValues(action, strconv.Itoa(statusCode)).Inc()
		}
	}

	// Record total action count
	metrics.TwitchBanActionTotal.WithLabelValues(action, status, errorCode).Inc()

	// Record HTTP status
	if statusCode > 0 {
		statusClass := fmt.Sprintf("%dxx", statusCode/100)
		metrics.TwitchBanHTTPStatus.WithLabelValues(action, strconv.Itoa(statusCode), statusClass).Inc()
	}
}

// createBanAuditLog creates an audit log entry for ban/unban operations
func (s *TwitchModerationService) createBanAuditLog(ctx context.Context, action string, moderatorUserID uuid.UUID, broadcasterID string, targetUserID string, reason *string, duration *int, statusCode int, err error) {
	// Build action name for audit log
	auditAction := "twitch_" + action

	// Create audit metadata
	auditMetadata := map[string]interface{}{
		"action":         auditAction,
		"broadcaster_id": broadcasterID,
		"target_user_id": targetUserID,
		"success":        err == nil,
	}

	if reason != nil {
		auditMetadata["reason"] = *reason
	}
	if duration != nil {
		auditMetadata["duration_seconds"] = *duration
	}
	if statusCode > 0 {
		auditMetadata["http_status"] = statusCode
	}
	if err != nil {
		auditMetadata["error"] = err.Error()
		auditMetadata["error_code"] = categorizeError(err)
	}

	// Create audit log entry
	// IMPORTANT DESIGN NOTE: EntityID uses moderatorUserID instead of targetUserID
	// This is due to a schema limitation - entity_id requires UUID but Twitch user IDs are strings.
	// The actual target Twitch user ID is stored in metadata["target_user_id"] for full traceability.
	// This means queries for "actions against user X" must filter on metadata, not entity_id.
	auditLog := &models.ModerationAuditLog{
		Action:      auditAction,
		EntityType:  "twitch_user",
		EntityID:    moderatorUserID, // Schema constraint: must be UUID; target ID in metadata
		ModeratorID: moderatorUserID,
		Reason:      reason,
		Metadata:    auditMetadata,
	}

	// Attempt to create audit log, but don't fail the operation if it fails
	if auditErr := s.auditLogRepo.Create(ctx, auditLog); auditErr != nil {
		// Log audit failures for visibility into the observability layer itself
		logger := utils.GetLogger()
		logger.Error("Failed to create audit log for Twitch ban action", auditErr, map[string]interface{}{
			"action":         auditAction,
			"moderator_id":   moderatorUserID.String(),
			"broadcaster_id": broadcasterID,
			"target_user_id": targetUserID,
			"error_type":     fmt.Sprintf("%T", auditErr),
		})
	}
}
