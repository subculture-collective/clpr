package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// ModerationEventType represents types of moderation events
type ModerationEventType string

const (
	// Submission events
	ModerationEventSubmissionReceived     ModerationEventType = "submission_received"
	ModerationEventSubmissionSuspicious   ModerationEventType = "submission_suspicious"
	ModerationEventSubmissionAutoRejected ModerationEventType = "submission_auto_rejected"
	ModerationEventSubmissionApproved     ModerationEventType = "submission_approved"
	ModerationEventSubmissionRejected     ModerationEventType = "submission_rejected"
	ModerationEventSubmissionDuplicate    ModerationEventType = "submission_duplicate"

	// Abuse events
	ModerationEventAbuseDetected         ModerationEventType = "abuse_detected"
	ModerationEventRateLimitExceeded     ModerationEventType = "rate_limit_exceeded"
	ModerationEventVelocityViolation     ModerationEventType = "velocity_violation"
	ModerationEventIPShareSuspicious     ModerationEventType = "ip_share_suspicious"
	ModerationEventUserCooldownActivated ModerationEventType = "user_cooldown_activated"

	// Queue size limits to prevent unbounded growth
	maxModerationQueueSize = 10000 // Maximum events in main moderation queue
	maxTypeEventListSize   = 1000  // Maximum events per type-based list
)

// ModerationEvent represents an event that requires moderation attention
type ModerationEvent struct {
	ID           uuid.UUID              `json:"id"`
	Type         ModerationEventType    `json:"type"`
	Severity     string                 `json:"severity"` // "info", "warning", "critical"
	UserID       uuid.UUID              `json:"user_id"`
	SubmissionID *uuid.UUID             `json:"submission_id,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	ReviewedBy   *uuid.UUID             `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time             `json:"reviewed_at,omitempty"`
	Status       string                 `json:"status"` // "pending", "reviewed", "actioned"
}

// ModerationEventService handles moderation events
type ModerationEventService struct {
	redisClient         *redispkg.Client
	notificationService *NotificationService
}

// NewModerationEventService creates a new moderation event service
func NewModerationEventService(redisClient *redispkg.Client, notificationService *NotificationService) *ModerationEventService {
	return &ModerationEventService{
		redisClient:         redisClient,
		notificationService: notificationService,
	}
}

// EmitEvent emits a moderation event
func (s *ModerationEventService) EmitEvent(ctx context.Context, event *ModerationEvent) error {
	// Set ID and timestamp if not set
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.Status == "" {
		event.Status = "pending"
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Store event in Redis list (moderation queue)
	queueKey := "moderation:queue"
	if err := s.redisClient.ListPush(ctx, queueKey, string(eventJSON)); err != nil {
		return fmt.Errorf("failed to push event to queue: %w", err)
	}

	// Trim queue to prevent unbounded growth
	if err := s.redisClient.ListTrim(ctx, queueKey, -maxModerationQueueSize, -1); err != nil {
		log.Printf("Failed to trim moderation queue: %v", err)
	}

	// Store event by ID for retrieval
	eventKey := fmt.Sprintf("moderation:event:%s", event.ID.String())
	if err := s.redisClient.Set(ctx, eventKey, string(eventJSON), 30*24*time.Hour); err != nil {
		return fmt.Errorf("failed to store event by ID: %w", err)
	}

	// Store event by type for filtering
	typeKey := fmt.Sprintf("moderation:events:%s", event.Type)
	if err := s.redisClient.ListPush(ctx, typeKey, string(eventJSON)); err != nil {
		return fmt.Errorf("failed to store event by type: %w", err)
	}

	// Trim type-based event list to prevent unbounded growth
	if err := s.redisClient.ListTrim(ctx, typeKey, -maxTypeEventListSize, -1); err != nil {
		log.Printf("Failed to trim type-based event list: %v", err)
	}

	// Log the event
	s.logEvent(event)

	// Send notifications to moderators for critical events
	if event.Severity == "critical" || event.Type == ModerationEventSubmissionSuspicious {
		s.notifyModerators(ctx, event)
	}

	return nil
}

// EmitSubmissionEvent emits a submission-related moderation event
func (s *ModerationEventService) EmitSubmissionEvent(ctx context.Context, eventType ModerationEventType, submission *models.ClipSubmission, ip string, metadata map[string]interface{}) error {
	severity := "info"
	switch eventType {
	case ModerationEventSubmissionSuspicious, ModerationEventAbuseDetected:
		severity = "warning"
	case ModerationEventSubmissionAutoRejected, ModerationEventVelocityViolation:
		severity = "critical"
	}

	event := &ModerationEvent{
		Type:         eventType,
		Severity:     severity,
		UserID:       submission.UserID,
		SubmissionID: &submission.ID,
		IPAddress:    ip,
		Metadata:     metadata,
	}

	return s.EmitEvent(ctx, event)
}

// EmitAbuseEvent emits an abuse-related moderation event
func (s *ModerationEventService) EmitAbuseEvent(ctx context.Context, eventType ModerationEventType, userID uuid.UUID, ip string, metadata map[string]interface{}) error {
	event := &ModerationEvent{
		Type:      eventType,
		Severity:  "warning",
		UserID:    userID,
		IPAddress: ip,
		Metadata:  metadata,
	}

	return s.EmitEvent(ctx, event)
}

// GetPendingEvents retrieves pending moderation events
func (s *ModerationEventService) GetPendingEvents(ctx context.Context, limit int) ([]*ModerationEvent, error) {
	queueKey := "moderation:queue"

	// Get events from queue (non-destructive peek)
	items, err := s.redisClient.ListRange(ctx, queueKey, 0, int64(limit-1))
	if err != nil {
		return nil, fmt.Errorf("failed to get events from queue: %w", err)
	}

	events := make([]*ModerationEvent, 0, len(items))
	for _, item := range items {
		var event ModerationEvent
		if err := json.Unmarshal([]byte(item), &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		if event.Status == "pending" {
			events = append(events, &event)
		}
	}

	return events, nil
}

// GetEventsByType retrieves events filtered by type
func (s *ModerationEventService) GetEventsByType(ctx context.Context, eventType ModerationEventType, limit int) ([]*ModerationEvent, error) {
	typeKey := fmt.Sprintf("moderation:events:%s", eventType)

	items, err := s.redisClient.ListRange(ctx, typeKey, 0, int64(limit-1))
	if err != nil {
		return nil, fmt.Errorf("failed to get events by type: %w", err)
	}

	events := make([]*ModerationEvent, 0, len(items))
	for _, item := range items {
		var event ModerationEvent
		if err := json.Unmarshal([]byte(item), &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

// MarkEventReviewed marks an event as reviewed
func (s *ModerationEventService) MarkEventReviewed(ctx context.Context, eventID uuid.UUID, reviewerID uuid.UUID) error {
	eventKey := fmt.Sprintf("moderation:event:%s", eventID.String())

	// Get event
	eventJSON, err := s.redisClient.Get(ctx, eventKey)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	var event ModerationEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Update event
	now := time.Now()
	event.ReviewedBy = &reviewerID
	event.ReviewedAt = &now
	event.Status = "reviewed"

	// Save updated event
	updatedJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize updated event: %w", err)
	}

	return s.redisClient.Set(ctx, eventKey, string(updatedJSON), 30*24*time.Hour)
}

// GetEventStats returns statistics about moderation events
func (s *ModerationEventService) GetEventStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get queue length
	queueKey := "moderation:queue"
	queueLength, err := s.redisClient.ListLen(ctx, queueKey)
	if err != nil {
		return nil, err
	}
	stats["queue_length"] = queueLength

	// Count events by severity
	pendingEvents, err := s.GetPendingEvents(ctx, 1000)
	if err != nil {
		return nil, err
	}

	infoCount := 0
	warningCount := 0
	criticalCount := 0

	for _, event := range pendingEvents {
		switch event.Severity {
		case "info":
			infoCount++
		case "warning":
			warningCount++
		case "critical":
			criticalCount++
		}
	}

	stats["pending_info"] = infoCount
	stats["pending_warning"] = warningCount
	stats["pending_critical"] = criticalCount

	return stats, nil
}

// logEvent logs an event to application logs
func (s *ModerationEventService) logEvent(event *ModerationEvent) {
	metadataJSON, _ := json.Marshal(event.Metadata)
	log.Printf("[MODERATION EVENT] id=%s type=%s severity=%s user_id=%s ip=%s metadata=%s",
		event.ID, event.Type, event.Severity, event.UserID, event.IPAddress, string(metadataJSON))
}

// notifyModerators sends notifications to moderators about critical events
func (s *ModerationEventService) notifyModerators(ctx context.Context, event *ModerationEvent) {
	if s.notificationService == nil {
		return
	}

	// This would be implemented to query for moderators and send them notifications
	// For now, just log that notification would be sent
	log.Printf("[MODERATION NOTIFY] Would notify moderators about event: type=%s severity=%s user_id=%s",
		event.Type, event.Severity, event.UserID)
}

// ProcessEvent processes an event and removes it from the queue
func (s *ModerationEventService) ProcessEvent(ctx context.Context, eventID uuid.UUID, reviewerID uuid.UUID, action string) error {
	// Mark as reviewed
	if err := s.MarkEventReviewed(ctx, eventID, reviewerID); err != nil {
		return err
	}

	// Log the action
	log.Printf("[MODERATION ACTION] event_id=%s reviewer_id=%s action=%s", eventID, reviewerID, action)

	return nil
}
