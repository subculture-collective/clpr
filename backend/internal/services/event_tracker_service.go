package services

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// EventTracker handles batching and writing of feed analytics events
type EventTracker struct {
	db            *pgxpool.Pool
	eventBatch    chan models.Event
	batchSize     int
	flushInterval time.Duration
	mu            sync.Mutex
	lastEvents    map[string]time.Time // For deduplication
	dedupWindow   time.Duration
}

// NewEventTracker creates a new event tracker with batching
func NewEventTracker(db *pgxpool.Pool, batchSize int, flushInterval time.Duration) *EventTracker {
	if batchSize <= 0 {
		batchSize = 100
	}
	if flushInterval <= 0 {
		flushInterval = 5 * time.Second
	}

	return &EventTracker{
		db:            db,
		eventBatch:    make(chan models.Event, batchSize*2), // Buffer for 2x batch size
		batchSize:     batchSize,
		flushInterval: flushInterval,
		lastEvents:    make(map[string]time.Time),
		dedupWindow:   1 * time.Second,
	}
}

// Start begins the background event processing loop
func (et *EventTracker) Start(ctx context.Context) {
	ticker := time.NewTicker(et.flushInterval)
	defer ticker.Stop()

	var batch []models.Event

	for {
		select {
		case event := <-et.eventBatch:
			// Check for duplicate
			if et.isDuplicate(event) {
				continue
			}

			batch = append(batch, event)
			if len(batch) >= et.batchSize {
				et.flush(ctx, batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				et.flush(ctx, batch)
				batch = nil
			}
			// Cleanup old dedup entries
			et.cleanupDedupMap()
		case <-ctx.Done():
			// Flush remaining events before shutdown
			if len(batch) > 0 {
				et.flush(ctx, batch)
			}
			return
		}
	}
}

// TrackEvent adds an event to the batch queue
func (et *EventTracker) TrackEvent(event models.Event) error {
	event.Timestamp = time.Now()
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Non-blocking send
	select {
	case et.eventBatch <- event:
		return nil
	default:
		// Channel is full, log warning and return error for visibility
		log.Printf("Warning: Event batch channel full, dropping event: %s (type: %s, session: %s)",
			event.ID, event.EventType, event.SessionID)
		return nil // Still return nil to not break caller flow, but event is logged as dropped
	}
}

// flush writes a batch of events to the database
func (et *EventTracker) flush(ctx context.Context, events []models.Event) error {
	if len(events) == 0 {
		return nil
	}

	query := `
		INSERT INTO events (id, event_type, user_id, session_id, timestamp, properties, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Use a transaction for better performance
	tx, err := et.db.Begin(ctx)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback(ctx)

	successCount := 0
	failedCount := 0
	for _, event := range events {
		propsJSON, err := json.Marshal(event.Properties)
		if err != nil {
			log.Printf("Error marshaling event properties for event %s: %v", event.ID, err)
			failedCount++
			continue
		}

		_, err = tx.Exec(ctx, query,
			event.ID, event.EventType, event.UserID, event.SessionID,
			event.Timestamp, propsJSON, event.CreatedAt)

		if err != nil {
			// Log error with event details for troubleshooting
			// Note: Failed events are dropped to maintain throughput.
			// Consider implementing a dead letter queue or retry mechanism for production use.
			log.Printf("Error writing event %s (type: %s, session: %s): %v",
				event.ID, event.EventType, event.SessionID, err)
			failedCount++
		} else {
			successCount++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	if successCount > 0 {
		log.Printf("Successfully flushed %d/%d events to database (failed: %d)",
			successCount, len(events), failedCount)
	}

	return nil
}

// isDuplicate checks if an event is a duplicate within the dedup window
func (et *EventTracker) isDuplicate(event models.Event) bool {
	et.mu.Lock()
	defer et.mu.Unlock()

	// Create a unique key for this event
	key := et.eventKey(event)

	lastTime, exists := et.lastEvents[key]
	if exists && time.Since(lastTime) < et.dedupWindow {
		return true
	}

	et.lastEvents[key] = time.Now()
	return false
}

// eventKey creates a unique key for event deduplication
func (et *EventTracker) eventKey(event models.Event) string {
	userIDStr := "anonymous"
	if event.UserID != nil {
		userIDStr = event.UserID.String()
	}
	return event.EventType + "_" + userIDStr + "_" + event.SessionID
}

// cleanupDedupMap removes old entries from the dedup map
func (et *EventTracker) cleanupDedupMap() {
	et.mu.Lock()
	defer et.mu.Unlock()

	cutoff := time.Now().Add(-et.dedupWindow * 2) // Keep 2x the window
	for key, timestamp := range et.lastEvents {
		if timestamp.Before(cutoff) {
			delete(et.lastEvents, key)
		}
	}
}

// GetHourlyMetrics retrieves hourly aggregated metrics for a specific event type
func (et *EventTracker) GetHourlyMetrics(ctx context.Context, eventType string, hours int) ([]models.HourlyMetric, error) {
	query := `
		SELECT hour, event_type, count, unique_users, unique_sessions
		FROM events_hourly_metrics
		WHERE event_type = $1 AND hour > NOW() - INTERVAL '1 hour' * $2
		ORDER BY hour DESC
	`

	rows, err := et.db.Query(ctx, query, eventType, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.HourlyMetric
	for rows.Next() {
		var metric models.HourlyMetric
		if err := rows.Scan(&metric.Hour, &metric.EventType, &metric.Count, &metric.UniqueUsers, &metric.UniqueSessions); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

// GetFeedMetrics retrieves feed analytics metrics
func (et *EventTracker) GetFeedMetrics(ctx context.Context, hours int) (map[string]interface{}, error) {
	query := `
		SELECT 
			event_type,
			COUNT(*) as total_count,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(DISTINCT session_id) as unique_sessions
		FROM events
		WHERE timestamp > NOW() - INTERVAL '1 hour' * $1
		GROUP BY event_type
		ORDER BY total_count DESC
	`

	rows, err := et.db.Query(ctx, query, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[string]interface{})
	var eventMetrics []map[string]interface{}

	for rows.Next() {
		var eventType string
		var totalCount, uniqueUsers, uniqueSessions int64

		if err := rows.Scan(&eventType, &totalCount, &uniqueUsers, &uniqueSessions); err != nil {
			return nil, err
		}

		eventMetrics = append(eventMetrics, map[string]interface{}{
			"event_type":      eventType,
			"total_count":     totalCount,
			"unique_users":    uniqueUsers,
			"unique_sessions": uniqueSessions,
		})
	}

	metrics["events"] = eventMetrics
	metrics["period_hours"] = hours

	return metrics, rows.Err()
}

// RefreshHourlyMetrics refreshes the materialized view (call via cron job)
func (et *EventTracker) RefreshHourlyMetrics(ctx context.Context) error {
	query := `SELECT refresh_events_hourly_metrics()`
	_, err := et.db.Exec(ctx, query)
	if err != nil {
		log.Printf("Error refreshing hourly metrics: %v", err)
		return err
	}
	log.Println("Successfully refreshed events hourly metrics")
	return nil
}
