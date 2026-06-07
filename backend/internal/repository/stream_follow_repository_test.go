package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockStreamFollowRepository is a mock implementation of StreamFollowRepository for testing
type MockStreamFollowRepository struct {
	follows map[string]*models.StreamFollow // key: userID_streamer
}

func NewMockStreamFollowRepository() *MockStreamFollowRepository {
	return &MockStreamFollowRepository{
		follows: make(map[string]*models.StreamFollow),
	}
}

func (m *MockStreamFollowRepository) FollowStreamer(ctx context.Context, userID uuid.UUID, streamerUsername string, notificationsEnabled bool) (*models.StreamFollow, error) {
	key := userID.String() + "_" + streamerUsername

	// Check if already exists
	if existing, ok := m.follows[key]; ok {
		existing.NotificationsEnabled = notificationsEnabled
		existing.UpdatedAt = time.Now()
		return existing, nil
	}

	follow := &models.StreamFollow{
		ID:                   uuid.New(),
		UserID:               userID,
		StreamerUsername:     streamerUsername,
		NotificationsEnabled: notificationsEnabled,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	m.follows[key] = follow
	return follow, nil
}

func (m *MockStreamFollowRepository) UnfollowStreamer(ctx context.Context, userID uuid.UUID, streamerUsername string) error {
	key := userID.String() + "_" + streamerUsername
	delete(m.follows, key)
	return nil
}

func (m *MockStreamFollowRepository) IsFollowing(ctx context.Context, userID uuid.UUID, streamerUsername string) (bool, error) {
	key := userID.String() + "_" + streamerUsername
	_, exists := m.follows[key]
	return exists, nil
}

func (m *MockStreamFollowRepository) GetFollowedStreamers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.StreamFollow, error) {
	var follows []models.StreamFollow
	for _, follow := range m.follows {
		if follow.UserID == userID {
			follows = append(follows, *follow)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(follows) {
		return []models.StreamFollow{}, nil
	}
	if end > len(follows) {
		end = len(follows)
	}

	return follows[start:end], nil
}

func (m *MockStreamFollowRepository) GetFollowersForStreamer(ctx context.Context, streamerUsername string) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	for _, follow := range m.follows {
		if follow.StreamerUsername == streamerUsername && follow.NotificationsEnabled {
			userIDs = append(userIDs, follow.UserID)
		}
	}
	return userIDs, nil
}

func (m *MockStreamFollowRepository) GetFollow(ctx context.Context, userID uuid.UUID, streamerUsername string) (*models.StreamFollow, error) {
	key := userID.String() + "_" + streamerUsername
	if follow, ok := m.follows[key]; ok {
		return follow, nil
	}
	return nil, nil
}

func (m *MockStreamFollowRepository) UpdateNotificationPreference(ctx context.Context, userID uuid.UUID, streamerUsername string, enabled bool) error {
	key := userID.String() + "_" + streamerUsername
	if follow, ok := m.follows[key]; ok {
		follow.NotificationsEnabled = enabled
		follow.UpdatedAt = time.Now()
	}
	return nil
}

func (m *MockStreamFollowRepository) GetFollowCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, follow := range m.follows {
		if follow.UserID == userID {
			count++
		}
	}
	return count, nil
}

// Test FollowStreamer functionality
func TestMockStreamFollowRepository_FollowStreamer(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	userID := uuid.New()
	streamer := "teststreamer"

	follow, err := repo.FollowStreamer(ctx, userID, streamer, true)
	if err != nil {
		t.Fatalf("Failed to follow streamer: %v", err)
	}

	if follow == nil {
		t.Fatal("Follow is nil")
	}

	if follow.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, follow.UserID)
	}

	if follow.StreamerUsername != streamer {
		t.Errorf("Expected streamer %s, got %s", streamer, follow.StreamerUsername)
	}

	if !follow.NotificationsEnabled {
		t.Error("Expected notifications to be enabled")
	}
}

// Test UnfollowStreamer functionality
func TestMockStreamFollowRepository_UnfollowStreamer(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	userID := uuid.New()
	streamer := "teststreamer"

	// Follow first
	_, err := repo.FollowStreamer(ctx, userID, streamer, true)
	if err != nil {
		t.Fatalf("Failed to follow streamer: %v", err)
	}

	// Verify following
	isFollowing, err := repo.IsFollowing(ctx, userID, streamer)
	if err != nil {
		t.Fatalf("Failed to check follow status: %v", err)
	}
	if !isFollowing {
		t.Error("Expected to be following after follow")
	}

	// Unfollow
	err = repo.UnfollowStreamer(ctx, userID, streamer)
	if err != nil {
		t.Fatalf("Failed to unfollow streamer: %v", err)
	}

	// Verify not following
	isFollowing, err = repo.IsFollowing(ctx, userID, streamer)
	if err != nil {
		t.Fatalf("Failed to check follow status: %v", err)
	}
	if isFollowing {
		t.Error("Expected to not be following after unfollow")
	}
}

// Test GetFollowedStreamers functionality
func TestMockStreamFollowRepository_GetFollowedStreamers(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	userID := uuid.New()
	streamers := []string{"streamer1", "streamer2", "streamer3"}

	// Follow multiple streamers
	for _, streamer := range streamers {
		_, err := repo.FollowStreamer(ctx, userID, streamer, true)
		if err != nil {
			t.Fatalf("Failed to follow streamer %s: %v", streamer, err)
		}
	}

	// Get followed streamers
	follows, err := repo.GetFollowedStreamers(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get followed streamers: %v", err)
	}

	if len(follows) != 3 {
		t.Errorf("Expected 3 follows, got %d", len(follows))
	}
}

// Test GetFollowersForStreamer functionality
func TestMockStreamFollowRepository_GetFollowersForStreamer(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	streamer := "popularstreamer"
	users := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Have multiple users follow the same streamer
	for _, userID := range users {
		_, err := repo.FollowStreamer(ctx, userID, streamer, true)
		if err != nil {
			t.Fatalf("Failed to follow streamer: %v", err)
		}
	}

	// Get followers
	followers, err := repo.GetFollowersForStreamer(ctx, streamer)
	if err != nil {
		t.Fatalf("Failed to get followers: %v", err)
	}

	if len(followers) != 3 {
		t.Errorf("Expected 3 followers, got %d", len(followers))
	}
}

// Test GetFollowersForStreamer with notifications disabled
func TestMockStreamFollowRepository_GetFollowersForStreamer_NotificationsDisabled(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	streamer := "streamer"
	user1 := uuid.New()
	user2 := uuid.New()

	// User 1 follows with notifications enabled
	_, err := repo.FollowStreamer(ctx, user1, streamer, true)
	if err != nil {
		t.Fatalf("Failed to follow streamer: %v", err)
	}

	// User 2 follows with notifications disabled
	_, err = repo.FollowStreamer(ctx, user2, streamer, false)
	if err != nil {
		t.Fatalf("Failed to follow streamer: %v", err)
	}

	// Get followers (should only return user1)
	followers, err := repo.GetFollowersForStreamer(ctx, streamer)
	if err != nil {
		t.Fatalf("Failed to get followers: %v", err)
	}

	if len(followers) != 1 {
		t.Errorf("Expected 1 follower with notifications enabled, got %d", len(followers))
	}

	if len(followers) > 0 && followers[0] != user1 {
		t.Errorf("Expected follower to be user1, got %s", followers[0])
	}
}

// Test UpdateNotificationPreference functionality
func TestMockStreamFollowRepository_UpdateNotificationPreference(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	userID := uuid.New()
	streamer := "streamer"

	// Follow with notifications enabled
	_, err := repo.FollowStreamer(ctx, userID, streamer, true)
	if err != nil {
		t.Fatalf("Failed to follow streamer: %v", err)
	}

	// Update notification preference to disabled
	err = repo.UpdateNotificationPreference(ctx, userID, streamer, false)
	if err != nil {
		t.Fatalf("Failed to update notification preference: %v", err)
	}

	// Get follow to verify
	follow, err := repo.GetFollow(ctx, userID, streamer)
	if err != nil {
		t.Fatalf("Failed to get follow: %v", err)
	}

	if follow.NotificationsEnabled {
		t.Error("Expected notifications to be disabled after update")
	}
}

// Test GetFollowCount functionality
func TestMockStreamFollowRepository_GetFollowCount(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	userID := uuid.New()

	// Initially should have 0 follows
	count, err := repo.GetFollowCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get follow count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 follows initially, got %d", count)
	}

	// Follow 5 streamers
	for i := 1; i <= 5; i++ {
		streamer := "streamer" + string(rune(i))
		_, err := repo.FollowStreamer(ctx, userID, streamer, true)
		if err != nil {
			t.Fatalf("Failed to follow streamer: %v", err)
		}
	}

	// Should now have 5 follows
	count, err = repo.GetFollowCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get follow count: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 follows, got %d", count)
	}
}

// Test follow update (re-follow with different notification preference)
func TestMockStreamFollowRepository_FollowUpdate(t *testing.T) {
	repo := NewMockStreamFollowRepository()
	ctx := context.Background()

	userID := uuid.New()
	streamer := "streamer"

	// Follow with notifications enabled
	follow1, err := repo.FollowStreamer(ctx, userID, streamer, true)
	if err != nil {
		t.Fatalf("Failed to follow streamer: %v", err)
	}

	if !follow1.NotificationsEnabled {
		t.Error("Expected notifications to be enabled")
	}

	// Follow again with notifications disabled (should update)
	follow2, err := repo.FollowStreamer(ctx, userID, streamer, false)
	if err != nil {
		t.Fatalf("Failed to re-follow streamer: %v", err)
	}

	if follow2.NotificationsEnabled {
		t.Error("Expected notifications to be disabled after update")
	}

	// Should still only have 1 follow
	count, err := repo.GetFollowCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get follow count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 follow after re-follow, got %d", count)
	}
}
