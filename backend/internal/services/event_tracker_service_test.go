package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestEventTracker_EventKey(t *testing.T) {
	et := &EventTracker{}

	userID := uuid.New()
	sessionID := "session-123"

	tests := []struct {
		name     string
		event    models.Event
		expected string
	}{
		{
			name: "event with user ID",
			event: models.Event{
				EventType: models.EventFeedViewed,
				UserID:    &userID,
				SessionID: sessionID,
			},
			expected: models.EventFeedViewed + "_" + userID.String() + "_" + sessionID,
		},
		{
			name: "event without user ID (anonymous)",
			event: models.Event{
				EventType: models.EventFilterApplied,
				UserID:    nil,
				SessionID: sessionID,
			},
			expected: models.EventFilterApplied + "_anonymous_" + sessionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := et.eventKey(tt.event)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestEventTracker_IsDuplicate(t *testing.T) {
	et := NewEventTracker(nil, 100, 5*time.Second)
	et.dedupWindow = 100 * time.Millisecond // Short window for testing

	userID := uuid.New()
	sessionID := "session-123"

	event1 := models.Event{
		EventType: models.EventFeedViewed,
		UserID:    &userID,
		SessionID: sessionID,
	}

	event2 := models.Event{
		EventType: models.EventFeedViewed,
		UserID:    &userID,
		SessionID: sessionID,
	}

	// First event should not be duplicate
	assert.False(t, et.isDuplicate(event1))

	// Same event immediately after should be duplicate
	assert.True(t, et.isDuplicate(event2))

	// Wait for dedup window to expire
	time.Sleep(150 * time.Millisecond)

	// Should not be duplicate after window expires
	assert.False(t, et.isDuplicate(event2))
}

func TestEventTracker_CleanupDedupMap(t *testing.T) {
	et := NewEventTracker(nil, 100, 5*time.Second)
	et.dedupWindow = 50 * time.Millisecond

	// Add some old events
	et.lastEvents["event1"] = time.Now().Add(-200 * time.Millisecond)
	et.lastEvents["event2"] = time.Now().Add(-10 * time.Millisecond)
	et.lastEvents["event3"] = time.Now()

	// Cleanup should remove old entries
	et.cleanupDedupMap()

	// event1 should be removed (older than 2x window)
	_, exists := et.lastEvents["event1"]
	assert.False(t, exists)

	// event2 and event3 should remain
	_, exists = et.lastEvents["event2"]
	assert.True(t, exists)
	_, exists = et.lastEvents["event3"]
	assert.True(t, exists)
}

func TestNewEventTracker_DefaultValues(t *testing.T) {
	// Test with zero values
	et := NewEventTracker(nil, 0, 0)

	assert.Equal(t, 100, et.batchSize, "should use default batch size")
	assert.Equal(t, 5*time.Second, et.flushInterval, "should use default flush interval")
	assert.NotNil(t, et.eventBatch, "event batch channel should be initialized")
	assert.NotNil(t, et.lastEvents, "lastEvents map should be initialized")
	assert.Equal(t, 1*time.Second, et.dedupWindow, "should use default dedup window")
}

func TestNewEventTracker_CustomValues(t *testing.T) {
	batchSize := 50
	flushInterval := 10 * time.Second

	et := NewEventTracker(nil, batchSize, flushInterval)

	assert.Equal(t, batchSize, et.batchSize)
	assert.Equal(t, flushInterval, et.flushInterval)
}

func TestEventTracker_TrackEvent_Queueing(t *testing.T) {
	et := NewEventTracker(nil, 100, 5*time.Second)

	userID := uuid.New()
	event := models.Event{
		EventType: models.EventFeedViewed,
		UserID:    &userID,
		SessionID: "test-session",
		Properties: map[string]interface{}{
			"clips_count": 10,
		},
	}

	// Test verifies queueing mechanism without database persistence
	// For database integration tests, see integration test suite
	err := et.TrackEvent(event)
	assert.NoError(t, err)

	// TrackEvent modifies the event internally but doesn't return it,
	// so we just verify no error occurred which means the event was queued
}
