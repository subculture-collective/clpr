package models

import (
	"time"

	"github.com/google/uuid"
)

// Clip pack statuses (#86).
const (
	ClipPackPending  = "pending"
	ClipPackBuilding = "building"
	ClipPackReady    = "ready"
	ClipPackFailed   = "failed"
	ClipPackExpired  = "expired"
)

// ClipPack is one broadcaster's weekly pack of their top clips.
type ClipPack struct {
	ID              uuid.UUID      `json:"id"`
	BroadcasterID   string         `json:"broadcaster_id"`
	BroadcasterName string         `json:"broadcaster_name"`
	WeekStart       string         `json:"week_start"`
	Status          string         `json:"status"`
	FileName        *string        `json:"-"`
	SHA256          *string        `json:"sha256,omitempty"`
	Bytes           *int64         `json:"bytes,omitempty"`
	ClipCount       *int           `json:"clip_count,omitempty"`
	FailureReason   *string        `json:"-"`
	ReadyAt         *time.Time     `json:"ready_at,omitempty"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	Clips           []ClipPackClip `json:"-"`
	// DownloadURL is a signed link to the zip, set by the API for ready packs.
	DownloadURL *string `json:"download_url,omitempty"`
	// Withdrawn: the broadcaster stopped featuring their clips or was
	// excluded, or a clip in the pack was removed, so it is not served.
	Withdrawn bool `json:"withdrawn"`
}

// ClipPackClip is one clip chosen for a pack. Position 1 has the most views.
type ClipPackClip struct {
	Position        int
	ClipID          uuid.UUID
	Title           string
	BroadcasterName string
	BroadcasterID   *string
	TwitchClipID    string
	TwitchClipURL   string
	Duration        *float64
	ViewCount       int64
}

// ClipPackFile is what the pack's zip server needs to serve it.
type ClipPackFile struct {
	FileName  string
	Status    string
	ExpiresAt *time.Time
	Deleted   bool
	Withdrawn bool
}
