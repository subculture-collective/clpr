package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// Constants for DMCA strike removal reasons
const (
	StrikeRemovalReasonCounterNotice = "counter_notice_successful"
	StrikeRemovalReasonAdminOverride = "admin_override"
	StrikeRemovalReasonExpired       = "expired"
)

// DMCAService handles DMCA takedown notices, counter-notices, and strike management
type DMCAService struct {
	repo           *repository.DMCARepository
	clipRepo       *repository.ClipRepository
	userRepo       *repository.UserRepository
	auditLogRepo   *repository.AuditLogRepository
	emailService   *EmailService
	searchIndexer  SearchIndexer // Interface for search index operations
	db             *pgxpool.Pool
	baseURL        string
	dmcaAgentEmail string
	logger         *utils.StructuredLogger
}

// SearchIndexer interface for removing content from search index
type SearchIndexer interface {
	DeleteClipFromIndex(ctx context.Context, clipID uuid.UUID) error
}

// DMCAServiceConfig holds configuration for DMCA service
type DMCAServiceConfig struct {
	BaseURL        string
	DMCAAgentEmail string
}

// NewDMCAService creates a new DMCA service
func NewDMCAService(
	repo *repository.DMCARepository,
	clipRepo *repository.ClipRepository,
	userRepo *repository.UserRepository,
	auditLogRepo *repository.AuditLogRepository,
	emailService *EmailService,
	searchIndexer SearchIndexer,
	db *pgxpool.Pool,
	cfg *DMCAServiceConfig,
) *DMCAService {
	return &DMCAService{
		repo:           repo,
		clipRepo:       clipRepo,
		userRepo:       userRepo,
		auditLogRepo:   auditLogRepo,
		emailService:   emailService,
		searchIndexer:  searchIndexer,
		db:             db,
		baseURL:        cfg.BaseURL,
		dmcaAgentEmail: cfg.DMCAAgentEmail,
		logger:         utils.GetLogger(),
	}
}

// ==============================================================================
// DMCA Notice Submission and Validation
// ==============================================================================

// SubmitTakedownNotice validates and creates a new DMCA takedown notice
func (s *DMCAService) SubmitTakedownNotice(ctx context.Context, req *models.SubmitDMCANoticeRequest, ipAddress, userAgent string) (*models.DMCANotice, error) {
	// Validate the notice
	if err := s.validateTakedownNotice(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create the notice
	notice := &models.DMCANotice{
		ComplainantName:            req.ComplainantName,
		ComplainantEmail:           req.ComplainantEmail,
		ComplainantAddress:         req.ComplainantAddress,
		ComplainantPhone:           req.ComplainantPhone,
		Relationship:               req.Relationship,
		CopyrightedWorkDescription: req.CopyrightedWorkDescription,
		InfringingURLs:             req.InfringingURLs,
		GoodFaithStatement:         req.GoodFaithStatement,
		AccuracyStatement:          req.AccuracyStatement,
		Signature:                  req.Signature,
		SubmittedAt:                time.Now(),
		Status:                     "pending",
		IPAddress:                  &ipAddress,
		UserAgent:                  &userAgent,
	}

	if err := s.repo.CreateNotice(ctx, notice); err != nil {
		return nil, fmt.Errorf("failed to create notice: %w", err)
	}

	// Send confirmation email to complainant
	if err := s.sendTakedownNoticeConfirmation(ctx, notice); err != nil {
		s.logger.Error("Failed to send takedown notice confirmation email", nil, map[string]interface{}{
			"notice_id": notice.ID,
			"error":     err.Error(),
		})
		// Don't fail the request if email fails
	}

	// Send notification email to DMCA agent
	if err := s.notifyDMCAAgent(ctx, notice); err != nil {
		s.logger.Error("Failed to notify DMCA agent", nil, map[string]interface{}{
			"notice_id": notice.ID,
			"error":     err.Error(),
		})
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:     "dmca_notice_received",
		EntityType: "dmca_notice",
		EntityID:   notice.ID,
		Metadata: map[string]interface{}{
			"complainant_email": notice.ComplainantEmail,
			"urls_count":        len(notice.InfringingURLs),
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return notice, nil
}

// validateTakedownNotice validates a takedown notice for completeness
func (s *DMCAService) validateTakedownNotice(req *models.SubmitDMCANoticeRequest) error {
	// Parse and validate base URL once
	baseURL, err := url.Parse(s.baseURL)
	if err != nil {
		return fmt.Errorf("service configuration error: invalid base URL")
	}

	// Validate URLs are from this platform
	for _, urlStr := range req.InfringingURLs {
		// Limit URL length to prevent log injection
		if len(urlStr) > 500 {
			return fmt.Errorf("URL exceeds maximum length of 500 characters")
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL format in infringing URLs")
		}

		// Check if URL is from our domain
		if parsedURL.Host != baseURL.Host {
			return fmt.Errorf("provided URL is not from this platform")
		}
	}

	// Validate signature matches complainant name (fuzzy match)
	if !s.fuzzyMatchSignature(req.Signature, req.ComplainantName) {
		return fmt.Errorf("signature does not match complainant name")
	}

	// Ensure required statements are checked
	if !req.GoodFaithStatement {
		return fmt.Errorf("good faith statement must be accepted")
	}
	if !req.AccuracyStatement {
		return fmt.Errorf("accuracy statement must be accepted")
	}

	return nil
}

// fuzzyMatchSignature checks if signature roughly matches the name (case-insensitive, ignoring punctuation)
func (s *DMCAService) fuzzyMatchSignature(signature, name string) bool {
	normalize := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", "")
		s = strings.ReplaceAll(s, "-", "")
		s = strings.TrimSpace(s)
		return s
	}

	normSig := normalize(signature)
	normName := normalize(name)

	// Check if signature contains most of the name words
	nameWords := strings.Fields(normName)

	// Edge case: if name has no valid words, require exact match
	if len(nameWords) == 0 {
		return normSig == normName
	}

	matchCount := 0
	validWordCount := 0 // Count words longer than 2 chars

	for _, word := range nameWords {
		if len(word) > 2 {
			validWordCount++
			if strings.Contains(normSig, word) {
				matchCount++
			}
		}
	}

	// If all words are <= 2 chars, fall back to substring match
	if validWordCount == 0 {
		return strings.Contains(normSig, normName)
	}

	// At least 50% of valid words should be in signature
	return matchCount >= (validWordCount+1)/2
}

// ==============================================================================
// Notice Review and Processing (Admin)
// ==============================================================================

// ReviewNotice allows admin to review and mark notice as valid or invalid
func (s *DMCAService) ReviewNotice(ctx context.Context, noticeID, reviewerID uuid.UUID, status string, notes *string) error {
	if status != "valid" && status != "invalid" {
		return fmt.Errorf("invalid status: must be 'valid' or 'invalid'")
	}

	// Update notice status
	if err := s.repo.UpdateNoticeStatus(ctx, noticeID, status, reviewerID, notes); err != nil {
		return fmt.Errorf("failed to update notice status: %w", err)
	}

	// Get notice details
	notice, err := s.repo.GetNoticeByID(ctx, noticeID)
	if err != nil {
		return fmt.Errorf("failed to get notice: %w", err)
	}

	// Send email based on status
	if status == "invalid" {
		if err := s.sendNoticeIncompleteEmail(ctx, notice); err != nil {
			s.logger.Error("Failed to send incomplete notice email", nil, map[string]interface{}{
				"notice_id": noticeID,
				"error":     err.Error(),
			})
		}
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:      "dmca_notice_reviewed",
		EntityType:  "dmca_notice",
		EntityID:    noticeID,
		ModeratorID: reviewerID,
		Metadata: map[string]interface{}{
			"status": status,
			"notes":  notes,
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// ProcessTakedown processes a valid takedown notice and removes content
func (s *DMCAService) ProcessTakedown(ctx context.Context, noticeID, adminID uuid.UUID) error {
	// Get notice
	notice, err := s.repo.GetNoticeByID(ctx, noticeID)
	if err != nil {
		return fmt.Errorf("failed to get notice: %w", err)
	}

	if notice.Status != "valid" {
		return fmt.Errorf("notice must be validated before processing")
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Process each infringing URL
	affectedUserIDs := make(map[uuid.UUID]bool)
	removedClipIDs := []uuid.UUID{}

	for _, urlStr := range notice.InfringingURLs {
		clipID, err := s.extractClipIDFromURL(urlStr)
		if err != nil {
			s.logger.Warn("Failed to extract clip ID from URL", map[string]interface{}{
				"url":   urlStr,
				"error": err.Error(),
			})
			continue
		}

		// Get clip to find the user
		clip, err := s.clipRepo.GetByID(ctx, clipID)
		if err != nil {
			if err == pgx.ErrNoRows {
				s.logger.Warn("Clip not found", map[string]interface{}{"clip_id": clipID})
				continue
			}
			return fmt.Errorf("failed to get clip: %w", err)
		}

		// Mark clip as DMCA removed
		if err := s.removeClipForDMCA(ctx, tx, clipID, notice.ID); err != nil {
			return fmt.Errorf("failed to remove clip: %w", err)
		}

		removedClipIDs = append(removedClipIDs, clipID)

		// Track affected user
		if clip.SubmittedByUserID != nil {
			affectedUserIDs[*clip.SubmittedByUserID] = true
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update notice status to processed (after transaction commits)
	if err := s.repo.UpdateNoticeStatus(ctx, noticeID, "processed", adminID, nil); err != nil {
		s.logger.Error("Failed to update notice status after processing", nil, map[string]interface{}{
			"notice_id": noticeID,
			"error":     err.Error(),
		})
		// Don't return error - content is already removed
	}

	// Issue strikes and notify users (outside transaction)
	for userID := range affectedUserIDs {
		if err := s.issueStrikeAndNotify(ctx, userID, notice.ID); err != nil {
			s.logger.Error("Failed to issue strike", nil, map[string]interface{}{
				"user_id":   userID,
				"notice_id": noticeID,
				"error":     err.Error(),
			})
		}
	}

	// Send confirmation to complainant
	if err := s.sendTakedownProcessedEmail(ctx, notice, removedClipIDs); err != nil {
		s.logger.Error("Failed to send takedown processed email", nil, map[string]interface{}{
			"notice_id": noticeID,
			"error":     err.Error(),
		})
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:      "dmca_takedown_processed",
		EntityType:  "dmca_notice",
		EntityID:    noticeID,
		ModeratorID: adminID,
		Metadata: map[string]interface{}{
			"clips_removed":  len(removedClipIDs),
			"users_affected": len(affectedUserIDs),
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// removeClipForDMCA marks a clip as DMCA removed
func (s *DMCAService) removeClipForDMCA(ctx context.Context, tx pgx.Tx, clipID, noticeID uuid.UUID) error {
	query := `
		UPDATE clips
		SET dmca_removed = true,
		    dmca_notice_id = $1,
		    dmca_removed_at = NOW(),
		    is_hidden = true
		WHERE id = $2`

	_, err := tx.Exec(ctx, query, noticeID, clipID)
	if err != nil {
		return err
	}

	// Remove from search index (best effort)
	if s.searchIndexer != nil {
		if err := s.searchIndexer.DeleteClipFromIndex(ctx, clipID); err != nil {
			s.logger.Warn("Failed to remove clip from search index", map[string]interface{}{
				"clip_id": clipID,
				"error":   err.Error(),
			})
		}
	}

	return nil
}

// extractClipIDFromURL extracts clip ID from a platform URL
func (s *DMCAService) extractClipIDFromURL(urlStr string) (uuid.UUID, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return uuid.Nil, err
	}

	// Extract ID from path like /clip/{id}
	// Remove query params and fragments, then split path
	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")

	// Expect exactly 2 parts: "clip" and the UUID
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("invalid clip URL format: expected /clip/{id}")
	}

	if parts[0] != "clip" {
		return uuid.Nil, fmt.Errorf("invalid clip URL format: path must start with /clip/")
	}

	// Parse and validate UUID
	clipID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid clip ID format: %w", err)
	}

	return clipID, nil
}

// ==============================================================================
// Strike Management
// ==============================================================================

// issueStrikeAndNotify issues a copyright strike to a user and applies penalties
func (s *DMCAService) issueStrikeAndNotify(ctx context.Context, userID, noticeID uuid.UUID) error {
	// Get user's current active strikes
	activeStrikes, err := s.repo.GetUserActiveStrikes(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user strikes: %w", err)
	}

	strikeNumber := len(activeStrikes) + 1

	// Create new strike
	strike := &models.DMCAStrike{
		UserID:       userID,
		DMCANoticeID: noticeID,
		StrikeNumber: strikeNumber,
		IssuedAt:     time.Now(),
		ExpiresAt:    time.Now().AddDate(1, 0, 0), // Expires in 12 months
		Status:       "active",
	}

	if err := s.repo.CreateStrike(ctx, strike); err != nil {
		return fmt.Errorf("failed to create strike: %w", err)
	}

	// Apply penalties based on strike number
	switch strikeNumber {
	case 1:
		// Strike 1: Warning only
		if err := s.sendStrike1WarningEmail(ctx, userID, strike); err != nil {
			s.logger.Error("Failed to send strike 1 email", nil, map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
		}

	case 2:
		// Strike 2: 7-day suspension
		suspendUntil := time.Now().AddDate(0, 0, 7)
		if err := s.suspendUser(ctx, userID, suspendUntil); err != nil {
			return fmt.Errorf("failed to suspend user: %w", err)
		}
		if err := s.sendStrike2SuspensionEmail(ctx, userID, strike, suspendUntil); err != nil {
			s.logger.Error("Failed to send strike 2 email", nil, map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
		}

	case 3:
		// Strike 3: Permanent termination
		if err := s.terminateUser(ctx, userID); err != nil {
			return fmt.Errorf("failed to terminate user: %w", err)
		}
		if err := s.sendStrike3TerminationEmail(ctx, userID, strike); err != nil {
			s.logger.Error("Failed to send strike 3 email", nil, map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
		}
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:     "dmca_strike_issued",
		EntityType: "dmca_strike",
		EntityID:   strike.ID,
		Metadata: map[string]interface{}{
			"user_id":       userID,
			"strike_number": strikeNumber,
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// suspendUser temporarily suspends a user's account
func (s *DMCAService) suspendUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	query := `
		UPDATE users
		SET dmca_suspended_until = $1
		WHERE id = $2`

	_, err := s.db.Exec(ctx, query, until, userID)
	if err != nil {
		return err
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:     "dmca_user_suspended",
		EntityType: "user",
		EntityID:   userID,
		Metadata: map[string]interface{}{
			"suspended_until": until,
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// terminateUser permanently terminates a user's account
func (s *DMCAService) terminateUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET dmca_terminated = true,
		    dmca_terminated_at = NOW(),
		    is_banned = true
		WHERE id = $1`

	_, err := s.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	// Hide all user's clips (soft delete)
	hideQuery := `
		UPDATE clips
		SET is_hidden = true
		WHERE submitted_by_user_id = $1`

	_, err = s.db.Exec(ctx, hideQuery, userID)
	if err != nil {
		s.logger.Error("Failed to hide user clips", nil, map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:     "dmca_user_terminated",
		EntityType: "user",
		EntityID:   userID,
		Metadata: map[string]interface{}{
			"reason": "repeat_copyright_infringement",
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// Helper methods for email sending
func (s *DMCAService) sendTakedownNoticeConfirmation(ctx context.Context, notice *models.DMCANotice) error {
	subject := fmt.Sprintf("DMCA Takedown Notice Received - Notice #%s", s.formatNoticeID(notice.ID))

	data := map[string]interface{}{
		"NoticeID":        s.formatNoticeID(notice.ID),
		"ComplainantName": notice.ComplainantName,
		"SubmittedAt":     notice.SubmittedAt.Format("January 2, 2006 at 3:04 PM MST"),
		"URLCount":        len(notice.InfringingURLs),
	}

	htmlBody, textBody := s.emailService.prepareDMCATakedownConfirmationEmail(data)

	req := EmailRequest{
		To:      []string{notice.ComplainantEmail},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "takedown-notice", "confirmation"},
	}

	return s.emailService.SendEmail(ctx, req)
}

// formatNoticeID formats a UUID for display (first 8 characters)
func (s *DMCAService) formatNoticeID(id uuid.UUID) string {
	str := id.String()
	if len(str) < 8 {
		return str
	}
	return str[:8]
}

func (s *DMCAService) notifyDMCAAgent(ctx context.Context, notice *models.DMCANotice) error {
	subject := fmt.Sprintf("🚨 New DMCA Notice #%s - Review Required", s.formatNoticeID(notice.ID))

	data := map[string]interface{}{
		"NoticeID":         s.formatNoticeID(notice.ID),
		"ComplainantName":  notice.ComplainantName,
		"ComplainantEmail": notice.ComplainantEmail,
		"SubmittedAt":      notice.SubmittedAt.Format("January 2, 2006 at 3:04 PM MST"),
		"URLCount":         len(notice.InfringingURLs),
		"ReviewURL":        fmt.Sprintf("%s/admin/dmca/notices/%s", s.baseURL, notice.ID),
	}

	htmlBody, textBody := s.emailService.prepareDMCAAgentNotificationEmail(data)

	req := EmailRequest{
		To:      []string{s.dmcaAgentEmail},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "agent-notification", "admin"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendNoticeIncompleteEmail(ctx context.Context, notice *models.DMCANotice) error {
	subject := fmt.Sprintf("DMCA Notice Incomplete - Notice #%s", s.formatNoticeID(notice.ID))

	notes := "Your notice did not meet the requirements for a valid DMCA takedown request."
	if notice.Notes != nil && *notice.Notes != "" {
		notes = *notice.Notes
	}

	data := map[string]interface{}{
		"NoticeID":        s.formatNoticeID(notice.ID),
		"ComplainantName": notice.ComplainantName,
		"Notes":           notes,
	}

	htmlBody, textBody := s.emailService.prepareDMCANoticeIncompleteEmail(data)

	req := EmailRequest{
		To:      []string{notice.ComplainantEmail},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "notice-incomplete"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendTakedownProcessedEmail(ctx context.Context, notice *models.DMCANotice, clipIDs []uuid.UUID) error {
	subject := fmt.Sprintf("DMCA Takedown Processed - Notice #%s", s.formatNoticeID(notice.ID))

	data := map[string]interface{}{
		"NoticeID":        s.formatNoticeID(notice.ID),
		"ComplainantName": notice.ComplainantName,
		"ClipsRemoved":    len(clipIDs),
		"ProcessedAt":     time.Now().Format("January 2, 2006 at 3:04 PM MST"),
	}

	htmlBody, textBody := s.emailService.prepareDMCATakedownProcessedEmail(data)

	req := EmailRequest{
		To:      []string{notice.ComplainantEmail},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "takedown-processed"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendStrike1WarningEmail(ctx context.Context, userID uuid.UUID, strike *models.DMCAStrike) error {
	// Get user information
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Email == nil {
		s.logger.Warn("Cannot send strike 1 email - user has no email address", map[string]interface{}{
			"user_id":   userID.String(),
			"strike_id": strike.ID.String(),
		})
		return nil
	}

	subject := "⚠️ Copyright Strike Warning - Strike 1 of 3"

	data := map[string]interface{}{
		"UserName":         user.Username,
		"StrikeID":         s.formatNoticeID(strike.ID),
		"NoticeID":         s.formatNoticeID(strike.DMCANoticeID),
		"IssuedAt":         strike.IssuedAt.Format("January 2, 2006 at 3:04 PM MST"),
		"ExpiresAt":        strike.ExpiresAt.Format("January 2, 2006"),
		"CounterNoticeURL": fmt.Sprintf("%s/dmca/counter-notice/%s", s.baseURL, strike.DMCANoticeID),
	}

	htmlBody, textBody := s.emailService.prepareDMCAStrike1Email(data)

	req := EmailRequest{
		To:      []string{*user.Email},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "strike-1", "warning"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendStrike2SuspensionEmail(ctx context.Context, userID uuid.UUID, strike *models.DMCAStrike, suspendUntil time.Time) error {
	// Get user information
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Email == nil {
		s.logger.Warn("Cannot send strike 2 email - user has no email address", map[string]interface{}{
			"user_id":   userID.String(),
			"strike_id": strike.ID.String(),
		})
		return nil
	}

	subject := "🚫 Account Suspended - Copyright Strike 2 of 3"

	data := map[string]interface{}{
		"UserName":         user.Username,
		"StrikeID":         s.formatNoticeID(strike.ID),
		"NoticeID":         s.formatNoticeID(strike.DMCANoticeID),
		"IssuedAt":         strike.IssuedAt.Format("January 2, 2006 at 3:04 PM MST"),
		"SuspendUntil":     suspendUntil.Format("January 2, 2006 at 3:04 PM MST"),
		"CounterNoticeURL": fmt.Sprintf("%s/dmca/counter-notice/%s", s.baseURL, strike.DMCANoticeID),
	}

	htmlBody, textBody := s.emailService.prepareDMCAStrike2Email(data)

	req := EmailRequest{
		To:      []string{*user.Email},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "strike-2", "suspension"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendStrike3TerminationEmail(ctx context.Context, userID uuid.UUID, strike *models.DMCAStrike) error {
	// Get user information
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Email == nil {
		s.logger.Warn("Cannot send strike 3 email - user has no email address", map[string]interface{}{
			"user_id":   userID.String(),
			"strike_id": strike.ID.String(),
		})
		return nil
	}

	subject := "❌ Account Terminated - Copyright Strike 3 of 3"

	data := map[string]interface{}{
		"UserName": user.Username,
		"StrikeID": s.formatNoticeID(strike.ID),
		"NoticeID": s.formatNoticeID(strike.DMCANoticeID),
		"IssuedAt": strike.IssuedAt.Format("January 2, 2006 at 3:04 PM MST"),
	}

	htmlBody, textBody := s.emailService.prepareDMCAStrike3Email(data)

	req := EmailRequest{
		To:      []string{*user.Email},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "strike-3", "termination"},
	}

	return s.emailService.SendEmail(ctx, req)
}

// ==============================================================================
// Counter-Notice Submission and Processing
// ==============================================================================

// SubmitCounterNotice validates and creates a new DMCA counter-notice
func (s *DMCAService) SubmitCounterNotice(ctx context.Context, req *models.SubmitDMCACounterNoticeRequest, userID *uuid.UUID, ipAddress, userAgent string) (*models.DMCACounterNotice, error) {
	// Validate counter-notice
	if err := s.validateCounterNotice(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify the DMCA notice exists and content was actually removed
	notice, err := s.repo.GetNoticeByID(ctx, req.DMCANoticeID)
	if err != nil {
		return nil, fmt.Errorf("DMCA notice not found: %w", err)
	}

	if notice.Status != "processed" {
		return nil, fmt.Errorf("DMCA notice has not been processed yet")
	}

	// Create counter-notice with 10-14 business day waiting period
	waitingPeriodEnd := s.calculateWaitingPeriodEnd(time.Now())

	counterNotice := &models.DMCACounterNotice{
		DMCANoticeID:               req.DMCANoticeID,
		UserID:                     userID, // Optional: nil if user not logged in
		UserName:                   req.UserName,
		UserEmail:                  req.UserEmail,
		UserAddress:                req.UserAddress,
		UserPhone:                  req.UserPhone,
		RemovedMaterialURL:         req.RemovedMaterialURL,
		RemovedMaterialDescription: req.RemovedMaterialDescription,
		GoodFaithStatement:         req.GoodFaithStatement,
		ConsentToJurisdiction:      req.ConsentToJurisdiction,
		ConsentToService:           req.ConsentToService,
		Signature:                  req.Signature,
		SubmittedAt:                time.Now(),
		WaitingPeriodEnds:          &waitingPeriodEnd,
		Status:                     "pending",
		IPAddress:                  &ipAddress,
		UserAgent:                  &userAgent,
	}

	if err := s.repo.CreateCounterNotice(ctx, counterNotice); err != nil {
		return nil, fmt.Errorf("failed to create counter-notice: %w", err)
	}

	// Send confirmation email to user
	if err := s.sendCounterNoticeConfirmationEmail(ctx, counterNotice); err != nil {
		s.logger.Error("Failed to send counter-notice confirmation", nil, map[string]interface{}{
			"counter_notice_id": counterNotice.ID,
			"error":             err.Error(),
		})
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:     "dmca_counter_notice_received",
		EntityType: "dmca_counter_notice",
		EntityID:   counterNotice.ID,
		Metadata: map[string]interface{}{
			"dmca_notice_id": req.DMCANoticeID,
			"user_id":        userID,
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return counterNotice, nil
}

// validateCounterNotice validates a counter-notice for completeness
func (s *DMCAService) validateCounterNotice(req *models.SubmitDMCACounterNoticeRequest) error {
	// Validate URL
	if _, err := url.Parse(req.RemovedMaterialURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Validate signature matches user name
	if !s.fuzzyMatchSignature(req.Signature, req.UserName) {
		return fmt.Errorf("signature does not match user name")
	}

	// Ensure required statements are checked
	if !req.GoodFaithStatement {
		return fmt.Errorf("good faith statement must be accepted")
	}
	if !req.ConsentToJurisdiction {
		return fmt.Errorf("consent to jurisdiction must be accepted")
	}
	if !req.ConsentToService {
		return fmt.Errorf("consent to service must be accepted")
	}

	return nil
}

// calculateWaitingPeriodEnd calculates the end of the 10-14 business day waiting period
// Note: Calculations are performed in UTC for consistency across timezones
func (s *DMCAService) calculateWaitingPeriodEnd(start time.Time) time.Time {
	// Use 14 business days (approximately 20 calendar days accounting for weekends)
	// Standardize to UTC for consistent legal compliance
	businessDays := 14
	daysAdded := 0
	current := start.UTC()

	for daysAdded < businessDays {
		current = current.AddDate(0, 0, 1)
		// Skip weekends (Saturday and Sunday in UTC)
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			daysAdded++
		}
	}

	return current
}

// ForwardCounterNoticeToComplainant forwards counter-notice to original complainant
func (s *DMCAService) ForwardCounterNoticeToComplainant(ctx context.Context, counterNoticeID, adminID uuid.UUID) error {
	// Get counter-notice
	cn, err := s.repo.GetCounterNoticeByID(ctx, counterNoticeID)
	if err != nil {
		return fmt.Errorf("failed to get counter-notice: %w", err)
	}

	if cn.Status != "pending" {
		return fmt.Errorf("counter-notice must be pending")
	}

	// Get original notice
	notice, err := s.repo.GetNoticeByID(ctx, cn.DMCANoticeID)
	if err != nil {
		return fmt.Errorf("failed to get original notice: %w", err)
	}

	// Send counter-notice to complainant
	if err := s.sendCounterNoticeToComplainantEmail(ctx, cn, notice); err != nil {
		return fmt.Errorf("failed to send counter-notice to complainant: %w", err)
	}

	// Mark as forwarded and update status to waiting
	if err := s.repo.MarkCounterNoticeForwarded(ctx, counterNoticeID); err != nil {
		return fmt.Errorf("failed to mark counter-notice as forwarded: %w", err)
	}

	// Update status to waiting
	if err := s.repo.UpdateCounterNoticeStatus(ctx, counterNoticeID, "waiting", nil); err != nil {
		return fmt.Errorf("failed to update counter-notice status: %w", err)
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:      "dmca_counter_notice_forwarded",
		EntityType:  "dmca_counter_notice",
		EntityID:    counterNoticeID,
		ModeratorID: adminID,
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// ProcessExpiredWaitingPeriods restores content for counter-notices past their waiting period
func (s *DMCAService) ProcessExpiredWaitingPeriods(ctx context.Context) error {
	// Get counter-notices awaiting restore
	counterNotices, err := s.repo.GetCounterNoticesAwaitingRestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to get counter-notices: %w", err)
	}

	for _, cn := range counterNotices {
		if err := s.reinstateContent(ctx, &cn); err != nil {
			s.logger.Error("Failed to reinstate content", nil, map[string]interface{}{
				"counter_notice_id": cn.ID,
				"error":             err.Error(),
			})
			continue
		}
	}

	return nil
}

// reinstateContent reinstates removed content after counter-notice waiting period
func (s *DMCAService) reinstateContent(ctx context.Context, cn *models.DMCACounterNotice) error {
	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Extract clip ID from URL
	clipID, err := s.extractClipIDFromURL(cn.RemovedMaterialURL)
	if err != nil {
		return fmt.Errorf("failed to extract clip ID: %w", err)
	}

	// Reinstate clip
	reinstateQuery := `
		UPDATE clips
		SET dmca_removed = false,
		    dmca_reinstated_at = NOW(),
		    is_hidden = false
		WHERE id = $1 AND dmca_removed = true`

	_, err = tx.Exec(ctx, reinstateQuery, clipID)
	if err != nil {
		return fmt.Errorf("failed to reinstate clip: %w", err)
	}

	// Remove strike if user has one related to this notice
	if cn.UserID != nil {
		removeStrikeQuery := `
			UPDATE dmca_strikes
			SET status = 'removed', 
			    removal_reason = $3,
			    removed_at = NOW()
			WHERE user_id = $1 AND dmca_notice_id = $2 AND status = 'active'`

		_, err = tx.Exec(ctx, removeStrikeQuery, cn.UserID, cn.DMCANoticeID, StrikeRemovalReasonCounterNotice)
		if err != nil {
			s.logger.Error("Failed to remove strike", nil, map[string]interface{}{
				"user_id":   cn.UserID,
				"notice_id": cn.DMCANoticeID,
				"error":     err.Error(),
			})
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update counter-notice status (after transaction commits)
	if err := s.repo.UpdateCounterNoticeStatus(ctx, cn.ID, "reinstated", nil); err != nil {
		s.logger.Error("Failed to update counter-notice status after reinstatement", nil, map[string]interface{}{
			"counter_notice_id": cn.ID,
			"error":             err.Error(),
		})
		// Don't return error - content is already reinstated
	}

	// Send notification emails
	if cn.UserID != nil {
		if err := s.sendContentReinstatedEmail(ctx, *cn.UserID, cn); err != nil {
			s.logger.Error("Failed to send reinstatement email", nil, map[string]interface{}{
				"user_id": cn.UserID,
				"error":   err.Error(),
			})
		}
	}

	// Notify complainant
	notice, err := s.repo.GetNoticeByID(ctx, cn.DMCANoticeID)
	if err == nil {
		if err := s.sendComplainantReinstatedEmail(ctx, notice, cn); err != nil {
			s.logger.Error("Failed to send complainant reinstatement email", nil, map[string]interface{}{
				"notice_id": cn.DMCANoticeID,
				"error":     err.Error(),
			})
		}
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:     "dmca_content_reinstated",
		EntityType: "clip",
		EntityID:   clipID,
		Metadata: map[string]interface{}{
			"counter_notice_id": cn.ID,
			"dmca_notice_id":    cn.DMCANoticeID,
		},
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// MarkLawsuitFiled marks a counter-notice as having lawsuit filed (content stays removed)
func (s *DMCAService) MarkLawsuitFiled(ctx context.Context, counterNoticeID, adminID uuid.UUID) error {
	query := `
		UPDATE dmca_counter_notices
		SET lawsuit_filed = true,
		    lawsuit_filed_at = NOW(),
		    status = 'rejected',
		    updated_at = NOW()
		WHERE id = $1`

	_, err := s.db.Exec(ctx, query, counterNoticeID)
	if err != nil {
		return fmt.Errorf("failed to mark lawsuit filed: %w", err)
	}

	// Audit log
	if err := s.auditLogRepo.Create(ctx, &models.ModerationAuditLog{
		Action:      "dmca_lawsuit_filed",
		EntityType:  "dmca_counter_notice",
		EntityID:    counterNoticeID,
		ModeratorID: adminID,
	}); err != nil {
		s.logger.Error("Failed to create audit log", nil, map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// ==============================================================================
// Scheduled Jobs / Maintenance
// ==============================================================================

// ExpireOldStrikes marks strikes older than 12 months as expired
func (s *DMCAService) ExpireOldStrikes(ctx context.Context) error {
	count, err := s.repo.ExpireOldStrikes(ctx)
	if err != nil {
		return fmt.Errorf("failed to expire strikes: %w", err)
	}

	s.logger.Info("Expired old DMCA strikes", map[string]interface{}{
		"count": count,
	})

	return nil
}

// GetUserStrikes retrieves all strikes for a user
func (s *DMCAService) GetUserStrikes(ctx context.Context, userID uuid.UUID) ([]models.DMCAStrike, error) {
	return s.repo.GetUserAllStrikes(ctx, userID)
}

// ==============================================================================
// Email Helper Methods (Templates)
// ==============================================================================

func (s *DMCAService) sendCounterNoticeConfirmationEmail(ctx context.Context, cn *models.DMCACounterNotice) error {
	subject := fmt.Sprintf("DMCA Counter-Notice Received - #%s", s.formatNoticeID(cn.ID))

	data := map[string]interface{}{
		"UserName":        cn.UserName,
		"CounterNoticeID": s.formatNoticeID(cn.ID),
		"NoticeID":        s.formatNoticeID(cn.DMCANoticeID),
		"SubmittedAt":     cn.SubmittedAt.Format("January 2, 2006 at 3:04 PM MST"),
	}

	htmlBody, textBody := s.emailService.prepareDMCACounterNoticeConfirmationEmail(data)

	req := EmailRequest{
		To:      []string{cn.UserEmail},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "counter-notice", "confirmation"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendCounterNoticeToComplainantEmail(ctx context.Context, cn *models.DMCACounterNotice, notice *models.DMCANotice) error {
	subject := fmt.Sprintf("DMCA Counter-Notice Filed - Notice #%s", s.formatNoticeID(notice.ID))

	waitingPeriodEnds := time.Now().AddDate(0, 0, 14) // 14 days from now
	if cn.WaitingPeriodEnds != nil {
		waitingPeriodEnds = *cn.WaitingPeriodEnds
	}

	data := map[string]interface{}{
		"ComplainantName":   notice.ComplainantName,
		"NoticeID":          s.formatNoticeID(notice.ID),
		"CounterNoticeID":   s.formatNoticeID(cn.ID),
		"UserName":          cn.UserName,
		"UserAddress":       cn.UserAddress,
		"ForwardedAt":       time.Now().Format("January 2, 2006 at 3:04 PM MST"),
		"WaitingPeriodEnds": waitingPeriodEnds.Format("January 2, 2006"),
	}

	htmlBody, textBody := s.emailService.prepareDMCACounterNoticeToComplainantEmail(data)

	req := EmailRequest{
		To:      []string{notice.ComplainantEmail},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "counter-notice", "complainant"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendContentReinstatedEmail(ctx context.Context, userID uuid.UUID, cn *models.DMCACounterNotice) error {
	// Get user information
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Email == nil {
		s.logger.Warn("Cannot send content reinstated email - user has no email address", map[string]interface{}{
			"user_id":           userID.String(),
			"counter_notice_id": cn.ID.String(),
		})
		return nil
	}

	subject := "✅ Content Reinstated - Counter-Notice Successful"

	// Parse and validate the URL
	contentURL := s.buildContentURL(cn.RemovedMaterialURL)

	data := map[string]interface{}{
		"UserName":        user.Username,
		"CounterNoticeID": s.formatNoticeID(cn.ID),
		"ReinstatedAt":    time.Now().Format("January 2, 2006 at 3:04 PM MST"),
		"ContentURL":      contentURL,
	}

	htmlBody, textBody := s.emailService.prepareDMCAContentReinstatedEmail(data)

	req := EmailRequest{
		To:      []string{*user.Email},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "content-reinstated", "user"},
	}

	return s.emailService.SendEmail(ctx, req)
}

func (s *DMCAService) sendComplainantReinstatedEmail(ctx context.Context, notice *models.DMCANotice, cn *models.DMCACounterNotice) error {
	subject := fmt.Sprintf("Content Reinstated - Notice #%s", s.formatNoticeID(notice.ID))

	data := map[string]interface{}{
		"ComplainantName": notice.ComplainantName,
		"NoticeID":        s.formatNoticeID(notice.ID),
		"CounterNoticeID": s.formatNoticeID(cn.ID),
		"ReinstatedAt":    time.Now().Format("January 2, 2006 at 3:04 PM MST"),
	}

	htmlBody, textBody := s.emailService.prepareDMCAComplainantReinstatedEmail(data)

	req := EmailRequest{
		To:      []string{notice.ComplainantEmail},
		Subject: subject,
		Data: map[string]interface{}{
			"HTML": htmlBody,
			"Text": textBody,
		},
		Tags: []string{"dmca", "content-reinstated", "complainant"},
	}

	return s.emailService.SendEmail(ctx, req)
}

// buildContentURL safely constructs a content URL from the provided URL string
func (s *DMCAService) buildContentURL(materialURL string) string {
	// Parse the URL to validate it
	parsedURL, err := url.Parse(materialURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		// If invalid or relative URL, treat the input as a single path segment.
		// Trim leading slashes and whitespace, then URL-escape it so it cannot
		// affect the path structure (prevents traversal and segment injection).
		segment := strings.TrimSpace(materialURL)
		segment = strings.TrimPrefix(segment, "/")
		safeSegment := url.PathEscape(segment)
		return fmt.Sprintf("%s/content/%s", s.baseURL, safeSegment)
	}

	// Ensure the URL uses a safe scheme (http or https)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		// Invalid scheme, sanitize and construct safe URL from base as above
		segment := strings.TrimSpace(materialURL)
		segment = strings.TrimPrefix(segment, "/")
		safeSegment := url.PathEscape(segment)
		return fmt.Sprintf("%s/content/%s", s.baseURL, safeSegment)
	}

	// Validate that the URL belongs to a trusted domain (our baseURL)
	baseURL, err := url.Parse(s.baseURL)
	if err == nil && parsedURL.Host != baseURL.Host {
		// External URL detected - treat as untrusted and construct safe URL
		segment := strings.TrimSpace(materialURL)
		segment = strings.TrimPrefix(segment, "/")
		safeSegment := url.PathEscape(segment)
		return fmt.Sprintf("%s/content/%s", s.baseURL, safeSegment)
	}

	return materialURL
}
