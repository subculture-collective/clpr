package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ApplicationLog represents a log entry from frontend or mobile clients
type ApplicationLog struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	Level      string          `json:"level" db:"level"` // debug, info, warn, error
	Message    string          `json:"message" db:"message"`
	Timestamp  time.Time       `json:"timestamp" db:"timestamp"`
	Service    string          `json:"service" db:"service"`             // clpr-frontend, clpr-mobile
	Platform   *string         `json:"platform,omitempty" db:"platform"` // web, ios, android
	UserID     *uuid.UUID      `json:"user_id,omitempty" db:"user_id"`
	SessionID  *string         `json:"session_id,omitempty" db:"session_id"`
	TraceID    *string         `json:"trace_id,omitempty" db:"trace_id"`
	URL        *string         `json:"url,omitempty" db:"url"`
	UserAgent  *string         `json:"user_agent,omitempty" db:"user_agent"`
	DeviceID   *string         `json:"device_id,omitempty" db:"device_id"`
	AppVersion *string         `json:"app_version,omitempty" db:"app_version"`
	Error      *string         `json:"error,omitempty" db:"error"`
	Stack      *string         `json:"stack,omitempty" db:"stack"`
	Context    json.RawMessage `json:"context,omitempty" db:"context"` // JSONB
	IPAddress  *string         `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// CreateApplicationLogRequest represents the request to create a log entry
type CreateApplicationLogRequest struct {
	Level      string                 `json:"level" binding:"required,oneof=debug info warn error"`
	Message    string                 `json:"message" binding:"required,max=10000"`
	Timestamp  *time.Time             `json:"timestamp,omitempty"`
	Service    string                 `json:"service" binding:"omitempty,max=50"`
	Platform   string                 `json:"platform,omitempty" binding:"omitempty,oneof=web ios android"`
	UserID     *string                `json:"user_id,omitempty"`
	SessionID  *string                `json:"session_id,omitempty" binding:"omitempty,max=255"`
	TraceID    *string                `json:"trace_id,omitempty" binding:"omitempty,max=255"`
	URL        *string                `json:"url,omitempty" binding:"omitempty,max=2048"`
	UserAgent  *string                `json:"user_agent,omitempty" binding:"omitempty,max=1000"`
	DeviceID   *string                `json:"device_id,omitempty" binding:"omitempty,max=255"`
	AppVersion *string                `json:"app_version,omitempty" binding:"omitempty,max=50"`
	Error      *string                `json:"error,omitempty" binding:"omitempty,max=10000"`
	Stack      *string                `json:"stack,omitempty" binding:"omitempty,max=50000"`
	Context    map[string]interface{} `json:"context,omitempty"`
}
