package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

// mockTwitchBanClient implements the TwitchBanClient interface for testing
type mockTwitchBanClient struct {
	banUserFunc   func(ctx context.Context, broadcasterID string, moderatorID string, userAccessToken string, request *twitch.BanUserRequest) (*twitch.BanUserResponse, error)
	unbanUserFunc func(ctx context.Context, broadcasterID string, moderatorID string, userID string, userAccessToken string) error
}

func (m *mockTwitchBanClient) BanUser(ctx context.Context, broadcasterID string, moderatorID string, userAccessToken string, request *twitch.BanUserRequest) (*twitch.BanUserResponse, error) {
	if m.banUserFunc != nil {
		return m.banUserFunc(ctx, broadcasterID, moderatorID, userAccessToken, request)
	}
	return &twitch.BanUserResponse{}, nil
}

func (m *mockTwitchBanClient) UnbanUser(ctx context.Context, broadcasterID string, moderatorID string, userID string, userAccessToken string) error {
	if m.unbanUserFunc != nil {
		return m.unbanUserFunc(ctx, broadcasterID, moderatorID, userID, userAccessToken)
	}
	return nil
}

// mockTwitchAuthRepo implements the TwitchAuthRepository interface for testing
type mockTwitchAuthRepo struct {
	getTwitchAuthFunc  func(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error)
	isTokenExpiredFunc func(auth *models.TwitchAuth) bool
}

func (m *mockTwitchAuthRepo) GetTwitchAuth(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error) {
	if m.getTwitchAuthFunc != nil {
		return m.getTwitchAuthFunc(ctx, userID)
	}
	return nil, errors.New("not found")
}

func (m *mockTwitchAuthRepo) IsTokenExpired(auth *models.TwitchAuth) bool {
	if m.isTokenExpiredFunc != nil {
		return m.isTokenExpiredFunc(auth)
	}
	return false
}

// mockUserRepo implements the ModerationUserRepo interface for testing
type mockUserRepoForTwitch struct {
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*models.User, error)
}

func (m *mockUserRepoForTwitch) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

// mockAuditLogRepo implements the AuditLogRepository interface for testing
type mockAuditLogRepoForTwitch struct {
	createFunc func(ctx context.Context, log *models.ModerationAuditLog) error
}

func (m *mockAuditLogRepoForTwitch) Create(ctx context.Context, log *models.ModerationAuditLog) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, log)
	}
	return nil
}

// TestValidateTwitchBanScope_Broadcaster tests that broadcaster is allowed
func TestValidateTwitchBanScope_Broadcaster(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:          userID,
				AccountType: models.AccountTypeMember,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{
		getTwitchAuthFunc: func(ctx context.Context, id uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: broadcasterTwitchID,
				AccessToken:  "test_token",
				Scopes:       "chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	service := NewTwitchModerationService(&mockTwitchBanClient{}, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	auth, err := service.ValidateTwitchBanScope(ctx, userID, broadcasterTwitchID)
	if err != nil {
		t.Errorf("Expected broadcaster to be allowed, got error: %v", err)
	}
	if auth == nil {
		t.Error("Expected auth to be returned")
	}
	if auth != nil && auth.TwitchUserID != broadcasterTwitchID {
		t.Errorf("Expected TwitchUserID %s, got %s", broadcasterTwitchID, auth.TwitchUserID)
	}
}

// TestValidateTwitchBanScope_SiteModeratorDenied tests that site moderators are blocked
func TestValidateTwitchBanScope_SiteModeratorDenied(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:             userID,
				AccountType:    models.AccountTypeModerator,
				ModeratorScope: models.ModeratorScopeSite,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{}
	service := NewTwitchModerationService(&mockTwitchBanClient{}, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	_, err := service.ValidateTwitchBanScope(ctx, userID, broadcasterTwitchID)
	if err == nil {
		t.Error("Expected site moderator to be denied")
	}
	if !errors.Is(err, ErrSiteModeratorsReadOnly) {
		t.Errorf("Expected ErrSiteModeratorsReadOnly, got: %v", err)
	}
}

// TestValidateTwitchBanScope_NotAuthenticated tests that non-authenticated users are denied
func TestValidateTwitchBanScope_NotAuthenticated(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:          userID,
				AccountType: models.AccountTypeMember,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{
		getTwitchAuthFunc: func(ctx context.Context, id uuid.UUID) (*models.TwitchAuth, error) {
			return nil, errors.New("not found")
		},
	}

	service := NewTwitchModerationService(&mockTwitchBanClient{}, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	_, err := service.ValidateTwitchBanScope(ctx, userID, broadcasterTwitchID)
	if err == nil {
		t.Error("Expected unauthenticated user to be denied")
	}
	if !errors.Is(err, ErrTwitchNotAuthenticated) {
		t.Errorf("Expected ErrTwitchNotAuthenticated, got: %v", err)
	}
}

// TestValidateTwitchBanScope_InsufficientScopes tests that users without required scopes are denied
func TestValidateTwitchBanScope_InsufficientScopes(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:          userID,
				AccountType: models.AccountTypeMember,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{
		getTwitchAuthFunc: func(ctx context.Context, id uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: broadcasterTwitchID,
				AccessToken:  "test_token",
				Scopes:       "chat:read chat:edit", // Missing ban scopes
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	service := NewTwitchModerationService(&mockTwitchBanClient{}, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	_, err := service.ValidateTwitchBanScope(ctx, userID, broadcasterTwitchID)
	if err == nil {
		t.Error("Expected user without scopes to be denied")
	}
	if !errors.Is(err, ErrTwitchScopeInsufficient) {
		t.Errorf("Expected ErrTwitchScopeInsufficient, got: %v", err)
	}
}

// TestValidateTwitchBanScope_NotBroadcaster tests that non-broadcasters are denied
func TestValidateTwitchBanScope_NotBroadcaster(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"
	userTwitchID := "67890" // Different from broadcaster

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:          userID,
				AccountType: models.AccountTypeMember,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{
		getTwitchAuthFunc: func(ctx context.Context, id uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: userTwitchID,
				AccessToken:  "test_token",
				Scopes:       "chat:read chat:edit moderator:manage:banned_users",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	service := NewTwitchModerationService(&mockTwitchBanClient{}, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	_, err := service.ValidateTwitchBanScope(ctx, userID, broadcasterTwitchID)
	if err == nil {
		t.Error("Expected non-broadcaster to be denied (in P0, only broadcasters allowed)")
	}
	if !errors.Is(err, ErrTwitchNotBroadcaster) {
		t.Errorf("Expected ErrTwitchNotBroadcaster, got: %v", err)
	}
}

// TestValidateTwitchBanScope_ExpiredToken tests that expired tokens are rejected
func TestValidateTwitchBanScope_ExpiredToken(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:          userID,
				AccountType: models.AccountTypeMember,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{
		getTwitchAuthFunc: func(ctx context.Context, id uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: broadcasterTwitchID,
				AccessToken:  "test_token",
				Scopes:       "chat:read chat:edit moderator:manage:banned_users",
				ExpiresAt:    time.Now().Add(-time.Hour), // Expired
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return true // Token is expired
		},
	}

	service := NewTwitchModerationService(&mockTwitchBanClient{}, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	_, err := service.ValidateTwitchBanScope(ctx, userID, broadcasterTwitchID)
	if err == nil {
		t.Error("Expected expired token to be rejected")
	}
	// Check for AuthError
	var authErr *twitch.AuthError
	if !errors.As(err, &authErr) {
		t.Errorf("Expected twitch.AuthError for expired token, got: %v", err)
	}
}

// TestBanUserOnTwitch_Success tests successful ban
func TestBanUserOnTwitch_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"
	targetUserID := "target123"

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:          userID,
				AccountType: models.AccountTypeMember,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{
		getTwitchAuthFunc: func(ctx context.Context, id uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: broadcasterTwitchID,
				AccessToken:  "test_token",
				Scopes:       "channel:manage:banned_users",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchBanClient{
		banUserFunc: func(ctx context.Context, broadcasterID string, moderatorID string, userAccessToken string, request *twitch.BanUserRequest) (*twitch.BanUserResponse, error) {
			if broadcasterID != broadcasterTwitchID {
				t.Errorf("Expected broadcaster ID %s, got %s", broadcasterTwitchID, broadcasterID)
			}
			if request.UserID != targetUserID {
				t.Errorf("Expected target user ID %s, got %s", targetUserID, request.UserID)
			}
			return &twitch.BanUserResponse{}, nil
		},
	}

	service := NewTwitchModerationService(twitchClient, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	reason := "Test reason"
	err := service.BanUserOnTwitch(ctx, userID, broadcasterTwitchID, targetUserID, &reason, nil)
	if err != nil {
		t.Errorf("Expected successful ban, got error: %v", err)
	}
}

// TestUnbanUserOnTwitch_Success tests successful unban
func TestUnbanUserOnTwitch_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	broadcasterTwitchID := "12345"
	targetUserID := "target123"

	userRepo := &mockUserRepoForTwitch{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{
				ID:          userID,
				AccountType: models.AccountTypeMember,
			}, nil
		},
	}

	authRepo := &mockTwitchAuthRepo{
		getTwitchAuthFunc: func(ctx context.Context, id uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: broadcasterTwitchID,
				AccessToken:  "test_token",
				Scopes:       "channel:manage:banned_users",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchBanClient{
		unbanUserFunc: func(ctx context.Context, broadcasterID string, moderatorID string, userID string, userAccessToken string) error {
			if broadcasterID != broadcasterTwitchID {
				t.Errorf("Expected broadcaster ID %s, got %s", broadcasterTwitchID, broadcasterID)
			}
			if userID != targetUserID {
				t.Errorf("Expected target user ID %s, got %s", targetUserID, userID)
			}
			return nil
		},
	}

	service := NewTwitchModerationService(twitchClient, authRepo, userRepo, &mockAuditLogRepoForTwitch{})

	err := service.UnbanUserOnTwitch(ctx, userID, broadcasterTwitchID, targetUserID)
	if err != nil {
		t.Errorf("Expected successful unban, got error: %v", err)
	}
}
