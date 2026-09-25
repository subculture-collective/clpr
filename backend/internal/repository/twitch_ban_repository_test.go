package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Note: These are unit tests for the repository layer.
// Integration tests with a real database belong in an integration-tagged file in
// this package that uses testutil.SetupTestDB.

func TestTwitchBan_StructFields(t *testing.T) {
	// This test validates the TwitchBan struct has all required fields
	channelID := uuid.New()
	bannedUserID := uuid.New()
	reason := "spam"
	twitchBanID := "ban123"
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	ban := TwitchBan{
		ID:               uuid.New(),
		ChannelID:        channelID,
		BannedUserID:     bannedUserID,
		Reason:           &reason,
		BannedAt:         now,
		ExpiresAt:        &expiresAt,
		SyncedFromTwitch: true,
		TwitchBanID:      &twitchBanID,
		LastSyncedAt:     &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Verify all fields are accessible
	if ban.ID == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if ban.ChannelID != channelID {
		t.Errorf("Expected ChannelID %v, got %v", channelID, ban.ChannelID)
	}
	if ban.BannedUserID != bannedUserID {
		t.Errorf("Expected BannedUserID %v, got %v", bannedUserID, ban.BannedUserID)
	}
	if ban.Reason == nil || *ban.Reason != reason {
		t.Errorf("Expected Reason %s, got %v", reason, ban.Reason)
	}
	if !ban.BannedAt.Equal(now) {
		t.Errorf("Expected BannedAt %v, got %v", now, ban.BannedAt)
	}
	if ban.ExpiresAt == nil || !ban.ExpiresAt.Equal(expiresAt) {
		t.Errorf("Expected ExpiresAt %v, got %v", expiresAt, ban.ExpiresAt)
	}
	if !ban.SyncedFromTwitch {
		t.Error("Expected SyncedFromTwitch to be true")
	}
	if ban.TwitchBanID == nil || *ban.TwitchBanID != twitchBanID {
		t.Errorf("Expected TwitchBanID %s, got %v", twitchBanID, ban.TwitchBanID)
	}
	if ban.LastSyncedAt == nil || !ban.LastSyncedAt.Equal(now) {
		t.Errorf("Expected LastSyncedAt %v, got %v", now, ban.LastSyncedAt)
	}
}

func TestTwitchBan_PermanentBan(t *testing.T) {
	// Test that permanent bans have nil ExpiresAt
	ban := TwitchBan{
		ID:               uuid.New(),
		ChannelID:        uuid.New(),
		BannedUserID:     uuid.New(),
		BannedAt:         time.Now(),
		ExpiresAt:        nil, // Permanent ban
		SyncedFromTwitch: true,
	}

	if ban.ExpiresAt != nil {
		t.Error("Permanent ban should have nil ExpiresAt")
	}
}

func TestTwitchBan_TemporaryBan(t *testing.T) {
	// Test that temporary bans have non-nil ExpiresAt
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	ban := TwitchBan{
		ID:               uuid.New(),
		ChannelID:        uuid.New(),
		BannedUserID:     uuid.New(),
		BannedAt:         now,
		ExpiresAt:        &expiresAt,
		SyncedFromTwitch: true,
	}

	if ban.ExpiresAt == nil {
		t.Error("Temporary ban should have non-nil ExpiresAt")
	}

	if ban.ExpiresAt.Before(ban.BannedAt) {
		t.Error("ExpiresAt should be after BannedAt")
	}
}

func TestNewTwitchBanRepository(t *testing.T) {
	// This test validates that the repository constructor works
	// In a real scenario, you'd pass a real pool here
	repo := NewTwitchBanRepository(nil)

	if repo == nil {
		t.Fatal("NewTwitchBanRepository should return a non-nil repository")
	}

	// Verify the pool field is set (even if nil for this test)
	if repo.pool != nil {
		t.Error("Expected pool to be nil in this test")
	}
}

// Integration tests would test actual database operations
// Example structure for integration tests (to be implemented in this package):
//
// func TestTwitchBanRepository_UpsertBan_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("Skipping integration test")
//     }
//     // Setup test database
//     // Create repository with real pool
//     // Test UpsertBan
//     // Verify data in database
//     // Cleanup
// }
//
// func TestTwitchBanRepository_BatchUpsertBans_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("Skipping integration test")
//     }
//     // Setup test database
//     // Test batch insert
//     // Verify transaction rollback on error
//     // Cleanup
// }
