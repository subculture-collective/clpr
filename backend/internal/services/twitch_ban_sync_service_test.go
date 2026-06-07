package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

// Mock implementations

type mockTwitchClient struct {
	getBannedUsersFunc func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error)
}

func (m *mockTwitchClient) GetBannedUsers(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
	if m.getBannedUsersFunc != nil {
		return m.getBannedUsersFunc(ctx, broadcasterID, userAccessToken, first, after)
	}
	return &twitch.BannedUsersResponse{Data: []twitch.BannedUser{}}, nil
}

type mockTwitchAuthRepository struct {
	getTwitchAuthFunc  func(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error)
	isTokenExpiredFunc func(auth *models.TwitchAuth) bool
}

func (m *mockTwitchAuthRepository) GetTwitchAuth(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error) {
	if m.getTwitchAuthFunc != nil {
		return m.getTwitchAuthFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockTwitchAuthRepository) IsTokenExpired(auth *models.TwitchAuth) bool {
	if m.isTokenExpiredFunc != nil {
		return m.isTokenExpiredFunc(auth)
	}
	return false
}

type mockBanRepository struct {
	batchUpsertBansFunc func(ctx context.Context, bans []*repository.TwitchBan) error
}

func (m *mockBanRepository) BatchUpsertBans(ctx context.Context, bans []*repository.TwitchBan) error {
	if m.batchUpsertBansFunc != nil {
		return m.batchUpsertBansFunc(ctx, bans)
	}
	return nil
}

type mockUserRepository struct {
	getByTwitchIDFunc func(ctx context.Context, twitchID string) (*models.User, error)
	createFunc        func(ctx context.Context, user *models.User) error
}

func (m *mockUserRepository) GetByTwitchID(ctx context.Context, twitchID string) (*models.User, error) {
	if m.getByTwitchIDFunc != nil {
		return m.getByTwitchIDFunc(ctx, twitchID)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

// Tests

func TestSyncChannelBans_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	channelID := twitchUserID

	// Setup mocks
	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			if uid == userID {
				return &models.TwitchAuth{
					UserID:       userID,
					TwitchUserID: twitchUserID,
					AccessToken:  "valid_token",
					ExpiresAt:    time.Now().Add(time.Hour),
				}, nil
			}
			return nil, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchClient{
		getBannedUsersFunc: func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
			return &twitch.BannedUsersResponse{
				Data: []twitch.BannedUser{
					{
						UserID:    "banned1",
						UserLogin: "banneduser1",
						UserName:  "BannedUser1",
						CreatedAt: time.Now(),
						Reason:    "spam",
					},
				},
				Pagination: twitch.Pagination{Cursor: ""},
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		getByTwitchIDFunc: func(ctx context.Context, twitchID string) (*models.User, error) {
			// Return existing user for channel
			if twitchID == channelID {
				return &models.User{
					ID:       uuid.New(),
					TwitchID: &twitchID,
					Username: "channel_user",
				}, nil
			}
			// Return user not found for banned users (they will be created)
			return nil, repository.ErrUserNotFound
		},
		createFunc: func(ctx context.Context, user *models.User) error {
			// Successfully create user
			return nil
		},
	}

	bansCaptured := []*repository.TwitchBan{}
	banRepo := &mockBanRepository{
		batchUpsertBansFunc: func(ctx context.Context, bans []*repository.TwitchBan) error {
			bansCaptured = bans
			return nil
		},
	}

	service := NewTwitchBanSyncService(twitchClient, authRepo, banRepo, userRepo)

	// Execute
	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(bansCaptured) != 1 {
		t.Fatalf("Expected 1 ban to be saved, got %d", len(bansCaptured))
	}

	ban := bansCaptured[0]
	if ban.Reason == nil || *ban.Reason != "spam" {
		t.Errorf("Expected ban reason 'spam', got: %v", ban.Reason)
	}
	if !ban.SyncedFromTwitch {
		t.Error("Expected ban to be marked as synced from Twitch")
	}
}

func TestSyncChannelBans_AuthenticationError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := "123456"

	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return nil, nil // No auth found
		},
	}

	service := NewTwitchBanSyncService(nil, authRepo, nil, nil)

	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	if err == nil {
		t.Fatal("Expected authentication error, got nil")
	}

	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Errorf("Expected AuthenticationError, got: %T", err)
	}
}

func TestSyncChannelBans_AuthorizationError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	differentChannelID := "999999"

	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "valid_token",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	service := NewTwitchBanSyncService(nil, authRepo, nil, nil)

	err := service.SyncChannelBans(ctx, userID.String(), differentChannelID)

	if err == nil {
		t.Fatal("Expected authorization error, got nil")
	}

	var authzErr *AuthorizationError
	if !errors.As(err, &authzErr) {
		t.Errorf("Expected AuthorizationError, got: %T", err)
	}
}

func TestSyncChannelBans_TokenExpired(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"

	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "expired_token",
				ExpiresAt:    time.Now().Add(-time.Hour), // Expired
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return true // Token is expired
		},
	}

	service := NewTwitchBanSyncService(nil, authRepo, nil, nil)

	err := service.SyncChannelBans(ctx, userID.String(), twitchUserID)

	if err == nil {
		t.Fatal("Expected authentication error for expired token, got nil")
	}

	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Errorf("Expected AuthenticationError, got: %T", err)
	}
}

func TestSyncChannelBans_Pagination(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	channelID := twitchUserID

	callCount := 0
	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "valid_token",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchClient{
		getBannedUsersFunc: func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
			callCount++

			if after == "" {
				// First page
				return &twitch.BannedUsersResponse{
					Data: []twitch.BannedUser{
						{UserID: "banned1", UserLogin: "user1", UserName: "User1", CreatedAt: time.Now()},
						{UserID: "banned2", UserLogin: "user2", UserName: "User2", CreatedAt: time.Now()},
					},
					Pagination: twitch.Pagination{Cursor: "page2"},
				}, nil
			} else if after == "page2" {
				// Second page
				return &twitch.BannedUsersResponse{
					Data: []twitch.BannedUser{
						{UserID: "banned3", UserLogin: "user3", UserName: "User3", CreatedAt: time.Now()},
					},
					Pagination: twitch.Pagination{Cursor: ""}, // No more pages
				}, nil
			}

			return &twitch.BannedUsersResponse{Data: []twitch.BannedUser{}}, nil
		},
	}

	userRepo := &mockUserRepository{
		getByTwitchIDFunc: func(ctx context.Context, twitchID string) (*models.User, error) {
			if twitchID == channelID {
				return &models.User{ID: uuid.New(), TwitchID: &twitchID}, nil
			}
			return nil, repository.ErrUserNotFound
		},
		createFunc: func(ctx context.Context, user *models.User) error {
			return nil
		},
	}

	bansCaptured := []*repository.TwitchBan{}
	banRepo := &mockBanRepository{
		batchUpsertBansFunc: func(ctx context.Context, bans []*repository.TwitchBan) error {
			bansCaptured = bans
			return nil
		},
	}

	service := NewTwitchBanSyncService(twitchClient, authRepo, banRepo, userRepo)

	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", callCount)
	}

	if len(bansCaptured) != 3 {
		t.Errorf("Expected 3 total bans, got %d", len(bansCaptured))
	}
}

func TestSyncChannelBans_RateLimitRetry(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	channelID := twitchUserID

	callCount := 0
	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "valid_token",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchClient{
		getBannedUsersFunc: func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
			callCount++

			// First call returns rate limit error
			if callCount == 1 {
				return nil, &twitch.RateLimitError{
					Message:    "rate limited",
					RetryAfter: 1,
				}
			}

			// Second call succeeds
			return &twitch.BannedUsersResponse{
				Data: []twitch.BannedUser{
					{UserID: "banned1", UserLogin: "user1", UserName: "User1", CreatedAt: time.Now()},
				},
				Pagination: twitch.Pagination{Cursor: ""},
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		getByTwitchIDFunc: func(ctx context.Context, twitchID string) (*models.User, error) {
			if twitchID == channelID {
				return &models.User{ID: uuid.New(), TwitchID: &twitchID}, nil
			}
			return nil, repository.ErrUserNotFound
		},
		createFunc: func(ctx context.Context, user *models.User) error {
			return nil
		},
	}

	bansCaptured := []*repository.TwitchBan{}
	banRepo := &mockBanRepository{
		batchUpsertBansFunc: func(ctx context.Context, bans []*repository.TwitchBan) error {
			bansCaptured = bans
			return nil
		},
	}

	service := NewTwitchBanSyncService(twitchClient, authRepo, banRepo, userRepo)

	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	if err != nil {
		t.Fatalf("Expected no error after retry, got: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 API calls (1 failed, 1 retry), got %d", callCount)
	}

	if len(bansCaptured) != 1 {
		t.Errorf("Expected 1 ban to be saved, got %d", len(bansCaptured))
	}
}

func TestSyncChannelBans_DatabaseError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	channelID := twitchUserID

	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "valid_token",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchClient{
		getBannedUsersFunc: func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
			return &twitch.BannedUsersResponse{
				Data: []twitch.BannedUser{
					{UserID: "banned1", UserLogin: "user1", UserName: "User1", CreatedAt: time.Now()},
				},
				Pagination: twitch.Pagination{Cursor: ""},
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		getByTwitchIDFunc: func(ctx context.Context, twitchID string) (*models.User, error) {
			if twitchID == channelID {
				return &models.User{ID: uuid.New(), TwitchID: &twitchID}, nil
			}
			return nil, repository.ErrUserNotFound
		},
		createFunc: func(ctx context.Context, user *models.User) error {
			return nil
		},
	}

	banRepo := &mockBanRepository{
		batchUpsertBansFunc: func(ctx context.Context, bans []*repository.TwitchBan) error {
			return errors.New("database connection failed")
		},
	}

	service := NewTwitchBanSyncService(twitchClient, authRepo, banRepo, userRepo)

	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	if err == nil {
		t.Fatal("Expected database error, got nil")
	}

	var dbErr *DatabaseError
	if !errors.As(err, &dbErr) {
		t.Errorf("Expected DatabaseError, got: %T", err)
	}
}

// TestSyncChannelBans_UserCreation tests that users are created for banned users
func TestSyncChannelBans_UserCreation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	channelID := twitchUserID

	createdUsers := []string{}

	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "valid_token",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchClient{
		getBannedUsersFunc: func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
			return &twitch.BannedUsersResponse{
				Data: []twitch.BannedUser{
					{UserID: "banned1", UserLogin: "user1", UserName: "User1", CreatedAt: time.Now()},
					{UserID: "banned2", UserLogin: "user2", UserName: "User2", CreatedAt: time.Now()},
				},
				Pagination: twitch.Pagination{Cursor: ""},
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		getByTwitchIDFunc: func(ctx context.Context, twitchID string) (*models.User, error) {
			if twitchID == channelID {
				return &models.User{ID: uuid.New(), TwitchID: &twitchID}, nil
			}
			// All banned users are new
			return nil, repository.ErrUserNotFound
		},
		createFunc: func(ctx context.Context, user *models.User) error {
			createdUsers = append(createdUsers, user.Username)
			return nil
		},
	}

	banRepo := &mockBanRepository{
		batchUpsertBansFunc: func(ctx context.Context, bans []*repository.TwitchBan) error {
			return nil
		},
	}

	service := NewTwitchBanSyncService(twitchClient, authRepo, banRepo, userRepo)

	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(createdUsers) != 2 {
		t.Errorf("Expected 2 users to be created, got %d", len(createdUsers))
	}

	if !testContainsString(createdUsers, "user1") || !testContainsString(createdUsers, "user2") {
		t.Errorf("Expected users 'user1' and 'user2' to be created, got: %v", createdUsers)
	}
}

// TestSyncChannelBans_EmptyBanList tests syncing with no bans
func TestSyncChannelBans_EmptyBanList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	channelID := twitchUserID

	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "valid_token",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchClient{
		getBannedUsersFunc: func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
			return &twitch.BannedUsersResponse{
				Data:       []twitch.BannedUser{}, // No bans
				Pagination: twitch.Pagination{Cursor: ""},
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		getByTwitchIDFunc: func(ctx context.Context, twitchID string) (*models.User, error) {
			if twitchID == channelID {
				return &models.User{ID: uuid.New(), TwitchID: &twitchID}, nil
			}
			return nil, repository.ErrUserNotFound
		},
	}

	banRepo := &mockBanRepository{}

	service := NewTwitchBanSyncService(twitchClient, authRepo, banRepo, userRepo)

	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	if err != nil {
		t.Fatalf("Expected no error with empty ban list, got: %v", err)
	}
}

// TestSyncChannelBans_WithExpiringBans tests syncing temporary bans with expiration
func TestSyncChannelBans_WithExpiringBans(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	twitchUserID := "123456"
	channelID := twitchUserID

	expirationTime := time.Now().Add(24 * time.Hour)

	authRepo := &mockTwitchAuthRepository{
		getTwitchAuthFunc: func(ctx context.Context, uid uuid.UUID) (*models.TwitchAuth, error) {
			return &models.TwitchAuth{
				UserID:       userID,
				TwitchUserID: twitchUserID,
				AccessToken:  "valid_token",
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
		isTokenExpiredFunc: func(auth *models.TwitchAuth) bool {
			return false
		},
	}

	twitchClient := &mockTwitchClient{
		getBannedUsersFunc: func(ctx context.Context, broadcasterID string, userAccessToken string, first int, after string) (*twitch.BannedUsersResponse, error) {
			return &twitch.BannedUsersResponse{
				Data: []twitch.BannedUser{
					{
						UserID:    "banned1",
						UserLogin: "user1",
						UserName:  "User1",
						CreatedAt: time.Now(),
						ExpiresAt: expirationTime,
						Reason:    "timeout",
					},
				},
				Pagination: twitch.Pagination{Cursor: ""},
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		getByTwitchIDFunc: func(ctx context.Context, twitchID string) (*models.User, error) {
			if twitchID == channelID {
				return &models.User{ID: uuid.New(), TwitchID: &twitchID}, nil
			}
			return nil, repository.ErrUserNotFound
		},
		createFunc: func(ctx context.Context, user *models.User) error {
			return nil
		},
	}

	bansCaptured := []*repository.TwitchBan{}
	banRepo := &mockBanRepository{
		batchUpsertBansFunc: func(ctx context.Context, bans []*repository.TwitchBan) error {
			bansCaptured = bans
			return nil
		},
	}

	service := NewTwitchBanSyncService(twitchClient, authRepo, banRepo, userRepo)

	err := service.SyncChannelBans(ctx, userID.String(), channelID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(bansCaptured) != 1 {
		t.Fatalf("Expected 1 ban to be saved, got %d", len(bansCaptured))
	}

	ban := bansCaptured[0]
	if ban.ExpiresAt == nil {
		t.Error("Expected ban to have expiration time")
	} else if !ban.ExpiresAt.Equal(expirationTime) {
		t.Errorf("Expected expiration time %v, got %v", expirationTime, *ban.ExpiresAt)
	}

	if ban.Reason == nil || *ban.Reason != "timeout" {
		t.Errorf("Expected ban reason 'timeout', got: %v", ban.Reason)
	}
}

// testContainsString is a helper function to check if a slice contains a string
func testContainsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
