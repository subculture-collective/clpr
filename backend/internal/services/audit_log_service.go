package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// AuditLogRepository defines the interface for audit log repository operations
type AuditLogRepository interface {
	List(ctx context.Context, filters repository.AuditLogFilters, page, limit int) ([]*models.ModerationAuditLogWithUser, int, error)
	Create(ctx context.Context, log *models.ModerationAuditLog) error
	Export(ctx context.Context, filters repository.AuditLogFilters) ([]*models.ModerationAuditLogWithUser, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.ModerationAuditLogWithUser, error)
}

// AuditLogService handles audit log business logic
type AuditLogService struct {
	auditLogRepo AuditLogRepository
}

// NewAuditLogService creates a new AuditLogService
func NewAuditLogService(auditLogRepo AuditLogRepository) *AuditLogService {
	return &AuditLogService{
		auditLogRepo: auditLogRepo,
	}
}

// GetAuditLogs retrieves audit logs with optional filters
func (s *AuditLogService) GetAuditLogs(ctx context.Context, filters repository.AuditLogFilters, page, limit int) ([]*models.ModerationAuditLogWithUser, int, error) {
	return s.auditLogRepo.List(ctx, filters, page, limit)
}

// GetAuditLogByID retrieves a single audit log entry by ID
func (s *AuditLogService) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*models.ModerationAuditLogWithUser, error) {
	return s.auditLogRepo.GetByID(ctx, id)
}

// AuditLogOptions contains optional fields for logging moderation actions
type AuditLogOptions struct {
	Channel   *uuid.UUID
	Reason    *string
	Metadata  map[string]interface{}
	IPAddress *string
	UserAgent *string
}

// LogAction logs a moderation action with comprehensive context
// This is a generic method for logging any moderation action with full audit trail support
func (s *AuditLogService) LogAction(ctx context.Context, action string, actor uuid.UUID, target uuid.UUID, entityType string, opts AuditLogOptions) error {
	log := &models.ModerationAuditLog{
		Action:      action,
		EntityType:  entityType,
		EntityID:    target,
		ModeratorID: actor,
		Reason:      opts.Reason,
		Metadata:    opts.Metadata,
		IPAddress:   opts.IPAddress,
		UserAgent:   opts.UserAgent,
		ChannelID:   opts.Channel,
	}

	return s.auditLogRepo.Create(ctx, log)
}

// ExportAuditLogsCSV exports audit logs to CSV format
func (s *AuditLogService) ExportAuditLogsCSV(ctx context.Context, filters repository.AuditLogFilters, writer io.Writer) error {
	// Get all logs matching filters
	logs, err := s.auditLogRepo.Export(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to export audit logs: %w", err)
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"ID",
		"Action",
		"Entity Type",
		"Entity ID",
		"Moderator ID",
		"Moderator Username",
		"Reason",
		"Metadata",
		"IP Address",
		"User Agent",
		"Channel ID",
		"Created At",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, log := range logs {
		var reason string
		if log.Reason != nil {
			reason = *log.Reason
		}

		var moderatorUsername string
		if log.Moderator != nil {
			moderatorUsername = log.Moderator.Username
		}

		metadata := ""
		if log.Metadata != nil {
			metadata = fmt.Sprintf("%v", log.Metadata)
		}

		ipAddress := ""
		if log.IPAddress != nil {
			ipAddress = *log.IPAddress
		}

		userAgent := ""
		if log.UserAgent != nil {
			userAgent = *log.UserAgent
		}

		channelID := ""
		if log.ChannelID != nil {
			channelID = log.ChannelID.String()
		}

		row := []string{
			log.ID.String(),
			log.Action,
			log.EntityType,
			log.EntityID.String(),
			log.ModeratorID.String(),
			moderatorUsername,
			reason,
			metadata,
			ipAddress,
			userAgent,
			channelID,
			log.CreatedAt.Format(time.RFC3339),
		}

		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ParseFiltersFromQuery parses audit log filters from query parameters
func ParseAuditLogFilters(moderatorID, action, entityType, entityID, channelID, startDate, endDate, search string) (repository.AuditLogFilters, error) {
	filters := repository.AuditLogFilters{}

	if moderatorID != "" {
		id, err := uuid.Parse(moderatorID)
		if err != nil {
			return filters, fmt.Errorf("invalid moderator_id: %w", err)
		}
		filters.ModeratorID = &id
	}

	if action != "" {
		filters.Action = action
	}

	if entityType != "" {
		filters.EntityType = entityType
	}

	if entityID != "" {
		id, err := uuid.Parse(entityID)
		if err != nil {
			return filters, fmt.Errorf("invalid entity_id: %w", err)
		}
		filters.EntityID = &id
	}

	if channelID != "" {
		id, err := uuid.Parse(channelID)
		if err != nil {
			return filters, fmt.Errorf("invalid channel_id: %w", err)
		}
		filters.ChannelID = &id
	}

	if startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			return filters, fmt.Errorf("invalid start_date format (use RFC3339): %w", err)
		}
		filters.StartDate = &t
	}

	if endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			return filters, fmt.Errorf("invalid end_date format (use RFC3339): %w", err)
		}
		filters.EndDate = &t
	}

	if search != "" {
		filters.Search = search
	}

	return filters, nil
}

// LogSubscriptionEvent logs a subscription-related event for audit purposes
func (s *AuditLogService) LogSubscriptionEvent(ctx context.Context, userID uuid.UUID, action string, metadata map[string]interface{}) error {
	log := &models.ModerationAuditLog{
		Action:      action,
		EntityType:  "subscription",
		EntityID:    userID, // Use user ID as entity ID for subscription events
		ModeratorID: userID, // For subscription events, moderator is the user themselves
		Metadata:    metadata,
	}

	return s.auditLogRepo.Create(ctx, log)
}

// LogAccountDeletionRequested logs when a user requests account deletion
func (s *AuditLogService) LogAccountDeletionRequested(ctx context.Context, userID uuid.UUID, reason *string) error {
	metadata := make(map[string]interface{})
	if reason != nil {
		metadata["reason"] = *reason
	}

	log := &models.ModerationAuditLog{
		Action:      "account_deletion_requested",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: userID,
		Reason:      reason,
		Metadata:    metadata,
	}

	return s.auditLogRepo.Create(ctx, log)
}

// LogAccountDeletionCancelled logs when a user cancels account deletion
func (s *AuditLogService) LogAccountDeletionCancelled(ctx context.Context, userID uuid.UUID) error {
	log := &models.ModerationAuditLog{
		Action:      "account_deletion_cancelled",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: userID,
	}

	return s.auditLogRepo.Create(ctx, log)
}

// LogEntitlementDenial logs when a user is denied access to a feature due to lack of entitlement
func (s *AuditLogService) LogEntitlementDenial(ctx context.Context, userID uuid.UUID, action string, metadata map[string]interface{}) error {
	log := &models.ModerationAuditLog{
		Action:      action,
		EntityType:  "entitlement",
		EntityID:    userID,
		ModeratorID: userID,
		Metadata:    metadata,
	}

	return s.auditLogRepo.Create(ctx, log)
}

// LogClipMetadataUpdate logs when a creator updates clip metadata
func (s *AuditLogService) LogClipMetadataUpdate(ctx context.Context, userID uuid.UUID, clipID uuid.UUID, changes map[string]interface{}) error {
	log := &models.ModerationAuditLog{
		Action:      "clip_metadata_updated",
		EntityType:  "clip",
		EntityID:    clipID,
		ModeratorID: userID,
		Metadata:    changes,
	}

	return s.auditLogRepo.Create(ctx, log)
}

// LogClipVisibilityChange logs when a creator changes clip visibility
func (s *AuditLogService) LogClipVisibilityChange(ctx context.Context, userID uuid.UUID, clipID uuid.UUID, isHidden bool) error {
	metadata := map[string]interface{}{
		"is_hidden": isHidden,
	}

	action := "clip_hidden"
	if !isHidden {
		action = "clip_unhidden"
	}

	log := &models.ModerationAuditLog{
		Action:      action,
		EntityType:  "clip",
		EntityID:    clipID,
		ModeratorID: userID,
		Metadata:    metadata,
	}

	return s.auditLogRepo.Create(ctx, log)
}
