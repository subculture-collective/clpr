package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestModerationEventService_EmitEvent(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()
	userID := uuid.New()

	event := &ModerationEvent{
		Type:      ModerationEventSubmissionReceived,
		Severity:  "info",
		UserID:    userID,
		IPAddress: "192.168.1.1",
		Metadata: map[string]interface{}{
			"test": "data",
		},
	}

	err := service.EmitEvent(ctx, event)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, event.ID)
	assert.False(t, event.CreatedAt.IsZero())
	assert.Equal(t, "pending", event.Status)
}

func TestModerationEventService_EmitSubmissionEvent(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()

	submission := &models.ClipSubmission{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		TwitchClipID:  "TestClip123",
		TwitchClipURL: "https://clips.twitch.tv/TestClip123",
		Status:        "pending",
	}

	err := service.EmitSubmissionEvent(ctx, ModerationEventSubmissionReceived, submission, "192.168.1.1", map[string]interface{}{
		"test": "metadata",
	})
	require.NoError(t, err)
}

func TestModerationEventService_EmitAbuseEvent(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()
	userID := uuid.New()

	err := service.EmitAbuseEvent(ctx, ModerationEventAbuseDetected, userID, "192.168.1.1", map[string]interface{}{
		"reason": "test abuse",
	})
	require.NoError(t, err)
}

func TestModerationEventService_GetPendingEvents(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()
	userID := uuid.New()

	// Emit some events
	for i := 0; i < 3; i++ {
		event := &ModerationEvent{
			Type:      ModerationEventSubmissionReceived,
			Severity:  "info",
			UserID:    userID,
			IPAddress: "192.168.1.1",
			Metadata:  map[string]interface{}{},
		}
		err := service.EmitEvent(ctx, event)
		require.NoError(t, err)
	}

	// Get pending events
	events, err := service.GetPendingEvents(ctx, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(events), 3)

	// Verify all are pending
	for _, event := range events {
		assert.Equal(t, "pending", event.Status)
	}
}

func TestModerationEventService_GetEventsByType(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()
	userID := uuid.New()

	// Emit events of specific type
	eventType := ModerationEventSubmissionSuspicious
	for i := 0; i < 2; i++ {
		event := &ModerationEvent{
			Type:      eventType,
			Severity:  "warning",
			UserID:    userID,
			IPAddress: "192.168.1.1",
			Metadata:  map[string]interface{}{},
		}
		err := service.EmitEvent(ctx, event)
		require.NoError(t, err)
	}

	// Get events by type
	events, err := service.GetEventsByType(ctx, eventType, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(events), 2)

	// Verify all are of the correct type
	for _, event := range events {
		assert.Equal(t, eventType, event.Type)
	}
}

func TestModerationEventService_MarkEventReviewed(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()
	userID := uuid.New()
	reviewerID := uuid.New()

	// Emit an event
	event := &ModerationEvent{
		Type:      ModerationEventSubmissionReceived,
		Severity:  "info",
		UserID:    userID,
		IPAddress: "192.168.1.1",
		Metadata:  map[string]interface{}{},
	}
	err := service.EmitEvent(ctx, event)
	require.NoError(t, err)

	// Mark as reviewed
	err = service.MarkEventReviewed(ctx, event.ID, reviewerID)
	require.NoError(t, err)

	// Verify it was marked as reviewed
	eventKey := "moderation:event:" + event.ID.String()
	exists, _ := redisClient.Exists(ctx, eventKey)
	assert.True(t, exists)
}

func TestModerationEventService_GetEventStats(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()
	userID := uuid.New()

	// Emit events with different severities
	severities := []string{"info", "warning", "critical"}
	for _, severity := range severities {
		event := &ModerationEvent{
			Type:      ModerationEventSubmissionReceived,
			Severity:  severity,
			UserID:    userID,
			IPAddress: "192.168.1.1",
			Metadata:  map[string]interface{}{},
		}
		err := service.EmitEvent(ctx, event)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := service.GetEventStats(ctx)
	require.NoError(t, err)
	assert.Contains(t, stats, "queue_length")
	assert.Contains(t, stats, "pending_info")
	assert.Contains(t, stats, "pending_warning")
	assert.Contains(t, stats, "pending_critical")
}

func TestModerationEventService_ProcessEvent(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		return
	}
	defer redisClient.Close()

	service := NewModerationEventService(redisClient, nil)
	ctx := context.Background()
	userID := uuid.New()
	reviewerID := uuid.New()

	// Emit an event
	event := &ModerationEvent{
		Type:      ModerationEventSubmissionReceived,
		Severity:  "info",
		UserID:    userID,
		IPAddress: "192.168.1.1",
		Metadata:  map[string]interface{}{},
	}
	err := service.EmitEvent(ctx, event)
	require.NoError(t, err)

	// Process the event
	err = service.ProcessEvent(ctx, event.ID, reviewerID, "approved")
	require.NoError(t, err)
}
