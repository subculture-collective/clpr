package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// UserSettingsService handles user settings operations
type UserSettingsService struct {
	userRepo            *repository.UserRepository
	userSettingsRepo    *repository.UserSettingsRepository
	accountDeletionRepo *repository.AccountDeletionRepository
	clipRepo            *repository.ClipRepository
	voteRepo            *repository.VoteRepository
	favoriteRepo        *repository.FavoriteRepository
	commentRepo         *repository.CommentRepository
	submissionRepo      *repository.SubmissionRepository
	subscriptionRepo    *repository.SubscriptionRepository
	consentRepo         *repository.ConsentRepository
	auditLogService     *AuditLogService
}

// NewUserSettingsService creates a new user settings service
func NewUserSettingsService(
	userRepo *repository.UserRepository,
	userSettingsRepo *repository.UserSettingsRepository,
	accountDeletionRepo *repository.AccountDeletionRepository,
	clipRepo *repository.ClipRepository,
	voteRepo *repository.VoteRepository,
	favoriteRepo *repository.FavoriteRepository,
	commentRepo *repository.CommentRepository,
	submissionRepo *repository.SubmissionRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	consentRepo *repository.ConsentRepository,
	auditLogService *AuditLogService,
) *UserSettingsService {
	return &UserSettingsService{
		userRepo:            userRepo,
		userSettingsRepo:    userSettingsRepo,
		accountDeletionRepo: accountDeletionRepo,
		clipRepo:            clipRepo,
		voteRepo:            voteRepo,
		favoriteRepo:        favoriteRepo,
		commentRepo:         commentRepo,
		submissionRepo:      submissionRepo,
		subscriptionRepo:    subscriptionRepo,
		consentRepo:         consentRepo,
		auditLogService:     auditLogService,
	}
}

// UpdateProfile updates user's display name and bio
func (s *UserSettingsService) UpdateProfile(ctx context.Context, userID uuid.UUID, displayName string, bio *string) error {
	return s.userRepo.UpdateProfile(ctx, userID, displayName, bio)
}

// GetSettings retrieves user settings
func (s *UserSettingsService) GetSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	return s.userSettingsRepo.GetByUserID(ctx, userID)
}

// UpdateSettings updates user settings
func (s *UserSettingsService) UpdateSettings(ctx context.Context, userID uuid.UUID, profileVisibility *string, showKarmaPublicly *bool) error {
	// Validate profile visibility if provided
	if profileVisibility != nil {
		validValues := map[string]bool{"public": true, "private": true, "followers": true}
		if !validValues[*profileVisibility] {
			return errors.New("invalid profile visibility value")
		}
	}

	return s.userSettingsRepo.Update(ctx, userID, profileVisibility, showKarmaPublicly)
}

// ExportUserData exports all user data as a JSON structure
func (s *UserSettingsService) ExportUserData(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	// Get user data
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user settings
	settings, err := s.userSettingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		settings = nil // Settings might not exist
	}

	// Get user's favorites (fetch all, paginated)
	var favorites []models.Favorite
	page := 1
	pageSize := 1000 // reasonable page size to avoid memory issues
	for {
		pageFavorites, err := s.favoriteRepo.GetByUserID(ctx, userID, page, pageSize)
		if err != nil {
			favorites = nil
			break
		}
		if len(pageFavorites) == 0 {
			break
		}
		favorites = append(favorites, pageFavorites...)
		if len(pageFavorites) < pageSize {
			break
		}
		page++
	}

	// Get user's comments (fetch all, paginated)
	var comments []interface{}
	var commentsError error
	offset := 0
	limit := 1000
	for {
		pageComments, _, err := s.commentRepo.ListByUserID(ctx, userID, limit, offset)
		if err != nil {
			log.Printf("Error fetching user comments during export: %v", err)
			commentsError = err
			break
		}
		if len(pageComments) == 0 {
			break
		}
		for _, c := range pageComments {
			comments = append(comments, c)
		}
		if len(pageComments) < limit {
			break
		}
		offset += limit
	}

	// Get user's submissions (fetch all, paginated)
	var submissions []interface{}
	var submissionsError error
	submissionPage := 1
	submissionLimit := 1000
	for {
		pageSubmissions, _, err := s.submissionRepo.ListByUser(ctx, userID, submissionPage, submissionLimit)
		if err != nil {
			log.Printf("Error fetching user submissions during export: %v", err)
			submissionsError = err
			break
		}
		if len(pageSubmissions) == 0 {
			break
		}
		for _, sub := range pageSubmissions {
			submissions = append(submissions, sub)
		}
		if len(pageSubmissions) < submissionLimit {
			break
		}
		submissionPage++
	}

	// Get user's subscription data
	subscription, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Only log if it's not a "no rows" error (user may not have a subscription)
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Error fetching user subscription during export: %v", err)
		}
		subscription = nil
	}

	// Get user's cookie consent preferences
	consent, err := s.consentRepo.GetConsent(ctx, userID)
	if err != nil {
		// Only log if it's not a "not found" error (user may not have set consent preferences)
		if !errors.Is(err, repository.ErrConsentNotFound) {
			log.Printf("Error fetching user consent during export: %v", err)
		}
		consent = nil
	}

	// Build export metadata to indicate data completeness
	exportMeta := map[string]interface{}{
		"exported_at": time.Now(),
		"complete":    true,
		"warnings":    []string{},
	}

	// Track any partial data warnings
	warnings := []string{}
	if commentsError != nil {
		warnings = append(warnings, "Comments data may be incomplete due to an error during export")
		exportMeta["complete"] = false
	}
	if submissionsError != nil {
		warnings = append(warnings, "Submissions data may be incomplete due to an error during export")
		exportMeta["complete"] = false
	}
	if len(warnings) > 0 {
		exportMeta["warnings"] = warnings
	}

	// Create export structure with all user data
	export := map[string]interface{}{
		"user": map[string]interface{}{
			"id":              user.ID,
			"twitch_id":       user.TwitchID,
			"username":        user.Username,
			"display_name":    user.DisplayName,
			"email":           user.Email,
			"bio":             user.Bio,
			"avatar_url":      user.AvatarURL,
			"social_links":    user.SocialLinks,
			"karma_points":    user.KarmaPoints,
			"trust_score":     user.TrustScore,
			"role":            user.Role,
			"account_type":    user.AccountType,
			"follower_count":  user.FollowerCount,
			"following_count": user.FollowingCount,
			"created_at":      user.CreatedAt,
			"updated_at":      user.UpdatedAt,
			"last_login_at":   user.LastLoginAt,
		},
		"settings":     settings,
		"favorites":    favorites,
		"comments":     comments,
		"submissions":  submissions,
		"subscription": subscription,
		"consent":      consent,
		"export_meta":  exportMeta,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, err
	}

	// Create a ZIP file
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add JSON file to ZIP
	jsonFile, err := zipWriter.Create("user_data.json")
	if err != nil {
		return nil, err
	}
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return nil, err
	}

	// Add comprehensive README explaining the export structure
	readmeContent := []byte(`Clipper User Data Export
========================

This archive contains all your personal data from Clipper in accordance with 
GDPR Article 15 (Right to Access) and Article 20 (Right to Data Portability).

Export Date: ` + time.Now().Format(time.RFC3339) + `
User ID: ` + userID.String() + `

Contents
--------

user_data.json - Complete export of your data in JSON format, including:

1. User Profile
   - Account information (ID, username, email, display name)
   - Profile details (bio, avatar, social links)
   - Account metrics (karma points, trust score, follower counts)
   - Account type and role
   - Timestamps (created, updated, last login)

2. Settings
   - Profile visibility preferences (public/private/followers)
   - Karma display preferences
   - Other privacy and display settings

3. Favorites
   - All clips you have favorited
   - Includes clip IDs and timestamps

4. Comments
   - All comments you have posted on clips
   - Includes comment content, timestamps, vote scores
   - Parent/reply relationships preserved

5. Submissions
   - All clip submissions you have made
   - Includes submission status, metadata, timestamps

6. Subscription
   - Premium subscription details (if applicable)
   - Subscription tier, status, and billing information
   - Payment method information (last 4 digits only)

7. Cookie Consent
   - Your cookie consent preferences
   - Categories: Essential, Functional, Analytics, Advertising
   - Consent date and IP address (for verification)

8. Export Metadata
   - Export timestamp
   - Data completeness status
   - Warnings (if any data collection encountered errors)

Data Completeness
-----------------

The export metadata includes information about data completeness. If any warnings
are present, it means some data could not be retrieved due to temporary errors.
In such cases, you may request a new export or contact privacy@clpr.gg for
assistance.

Data Not Included
-----------------

The following data is intentionally not included in the automated export:
- Search history (privacy-sensitive data, not required under GDPR portability)
- IP addresses from activity logs (privacy-sensitive, available via support if needed)
- Moderation actions taken against your account (available via support if needed)
- Internal system identifiers and technical metadata

For any additional data requests, contact privacy@clpr.gg

Your Rights
-----------

Under GDPR, you have the following rights regarding your data:

1. Right to Access (Article 15) - This export
2. Right to Rectification (Article 16) - Update your profile in Settings
3. Right to Erasure (Article 17) - Request account deletion in Settings
4. Right to Restriction (Article 18) - Contact support@clpr.gg
5. Right to Data Portability (Article 20) - This export
6. Right to Object (Article 21) - Manage in Cookie Settings

For questions or additional data requests, contact: privacy@clpr.gg

Data Format
-----------

The data is provided in JSON format, which is:
- Machine-readable
- Easily portable to other services
- Human-readable with a text editor
- Compatible with most data processing tools

Technical Details
-----------------

- Character Encoding: UTF-8
- Date Format: ISO 8601 (RFC3339)
- File Format: JSON (JavaScript Object Notation)
- Archive Format: ZIP

Legal
-----

This export is provided in compliance with:
- GDPR (General Data Protection Regulation) - EU Regulation 2016/679
- CCPA (California Consumer Privacy Act) - California Civil Code Section 1798.100
- Other applicable data protection laws

Clipper is committed to protecting your privacy and ensuring compliance with 
all applicable data protection regulations.

For our complete Privacy Policy, visit: https://clpr.gg/privacy
For our Terms of Service, visit: https://clpr.gg/terms

© 2024 Clipper. All rights reserved.
`)
	readmeFile, err := zipWriter.Create("README.txt")
	if err != nil {
		return nil, err
	}
	_, err = readmeFile.Write(readmeContent)
	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RequestAccountDeletion requests account deletion with a grace period
func (s *UserSettingsService) RequestAccountDeletion(ctx context.Context, userID uuid.UUID, reason *string) (*models.AccountDeletion, error) {
	// Check if there's already a pending deletion
	existing, err := s.accountDeletionRepo.GetPendingByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("account deletion already requested")
	}

	// Create deletion request with 30-day grace period
	deletion := &models.AccountDeletion{
		ID:           uuid.New(),
		UserID:       userID,
		ScheduledFor: time.Now().Add(30 * 24 * time.Hour), // 30 days
		Reason:       reason,
	}

	err = s.accountDeletionRepo.Create(ctx, deletion)
	if err != nil {
		return nil, err
	}

	// Log audit event
	if err := s.auditLogService.LogAccountDeletionRequested(ctx, userID, reason); err != nil {
		log.Printf("Failed to log account deletion request: %v", err)
	}

	return deletion, nil
}

// CancelAccountDeletion cancels a pending account deletion
func (s *UserSettingsService) CancelAccountDeletion(ctx context.Context, userID uuid.UUID) error {
	// Get pending deletion
	deletion, err := s.accountDeletionRepo.GetPendingByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if deletion == nil {
		return errors.New("no pending account deletion found")
	}

	// Cancel the deletion
	err = s.accountDeletionRepo.Cancel(ctx, deletion.ID)
	if err != nil {
		return err
	}

	// Log audit event
	if err := s.auditLogService.LogAccountDeletionCancelled(ctx, userID); err != nil {
		log.Printf("Failed to log account deletion cancellation: %v", err)
	}

	return nil
}

// GetPendingDeletion retrieves pending deletion for a user
func (s *UserSettingsService) GetPendingDeletion(ctx context.Context, userID uuid.UUID) (*models.AccountDeletion, error) {
	return s.accountDeletionRepo.GetPendingByUserID(ctx, userID)
}

// UpdateSocialLinks updates user's social media links
func (s *UserSettingsService) UpdateSocialLinks(ctx context.Context, userID uuid.UUID, req *models.UpdateSocialLinksRequest) error {
	// Convert request to JSON
	socialLinks := models.SocialLinks{
		Twitter: req.Twitter,
		Twitch:  req.Twitch,
		Discord: req.Discord,
		YouTube: req.YouTube,
		Website: req.Website,
	}

	jsonData, err := json.Marshal(socialLinks)
	if err != nil {
		return err
	}

	// Update in database
	return s.userRepo.UpdateSocialLinks(ctx, userID, string(jsonData))
}
