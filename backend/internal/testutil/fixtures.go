package testutil

import (
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// Helper functions for creating pointer values
func StringPtr(s string) *string {
	return &s
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

// TestUser creates a test user with default values
func TestUser() *models.User {
	email := "test@example.com"
	avatarURL := "https://example.com/avatar.png"
	bio := "Test user bio"
	twitchID := "test_twitch_123"

	return &models.User{
		ID:          uuid.New(),
		TwitchID:    &twitchID,
		Username:    "testuser",
		DisplayName: "Test User",
		Email:       &email,
		AvatarURL:   &avatarURL,
		Bio:         &bio,
		Role:        "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TestAdminUser creates a test admin user
func TestAdminUser() *models.User {
	user := TestUser()
	adminTwitchID := "admin_twitch_456"
	user.TwitchID = &adminTwitchID
	user.Username = "adminuser"
	user.DisplayName = "Admin User"
	email := "admin@example.com"
	user.Email = &email
	user.Role = "admin"
	return user
}

// TestClip creates a test clip with default values
func TestClip() *models.Clip {
	thumbnailURL := "https://example.com/thumbnail.jpg"
	duration := 30.5
	gameID := "game_123"
	gameName := "Test Game"
	language := "en"

	return &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    "test_clip_123",
		Title:           "Test Clip",
		TwitchClipURL:   "https://clips.twitch.tv/test_clip_123",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=test_clip_123",
		ThumbnailURL:    &thumbnailURL,
		ViewCount:       1000,
		Duration:        &duration,
		BroadcasterID:   StringPtr("broadcaster_123"),
		BroadcasterName: "TestBroadcaster",
		CreatorID:       StringPtr("creator_123"),
		CreatorName:     "TestCreator",
		GameID:          &gameID,
		GameName:        &gameName,
		Language:        &language,
		CreatedAt:       time.Now().Add(-24 * time.Hour),
		ImportedAt:      time.Now(),
	}
}

// TestComment creates a test comment with default values
func TestComment(userID, clipID uuid.UUID) *models.Comment {
	return &models.Comment{
		ID:        uuid.New(),
		UserID:    userID,
		ClipID:    clipID,
		Content:   "This is a test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestTag creates a test tag with default values
func TestTag() *models.Tag {
	description := "Test tag description"
	return &models.Tag{
		ID:          uuid.New(),
		Name:        "test-tag",
		Slug:        "test-tag",
		Description: &description,
		UsageCount:  0,
		CreatedAt:   time.Now(),
	}
}

// TestReport creates a test report with default values
func TestReport(reporterID uuid.UUID, reportableType string, reportableID uuid.UUID) *models.Report {
	description := "Test description"
	return &models.Report{
		ID:             uuid.New(),
		ReporterID:     reporterID,
		ReportableType: reportableType,
		ReportableID:   reportableID,
		Reason:         "Test reason",
		Description:    &description,
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
}

// TestClipSubmission creates a test submission with default values
func TestClipSubmission(userID uuid.UUID, twitchClipURL string) *models.ClipSubmission {
	return &models.ClipSubmission{
		ID:            uuid.New(),
		UserID:        userID,
		TwitchClipID:  "test_clip_submission_123",
		TwitchClipURL: twitchClipURL,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
