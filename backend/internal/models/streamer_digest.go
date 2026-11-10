package models

import "github.com/google/uuid"

// Streamer digest send statuses (#85).
const (
	StreamerDigestSent    = "sent"
	StreamerDigestSkipped = "skipped"
	StreamerDigestFailed  = "failed"
)

// StreamerDigestRecipient is a broadcaster who features their clips and
// gets the weekly digest email.
type StreamerDigestRecipient struct {
	UserID        uuid.UUID
	BroadcasterID string
	Name          string
	Email         string
}

// StreamerDigest is what one weekly digest email shows.
type StreamerDigest struct {
	Name          string
	BroadcasterID string
	// WeekStart is the Monday (UTC) that starts the week, YYYY-MM-DD.
	WeekStart  string
	TopClips   []StreamerDigestClip
	Placements []CountdownFeature
	// PackStatus is the week's clip pack status, or "" without a pack.
	PackStatus    string
	PackClipCount int
}

// StreamerDigestClip is one of the broadcaster's top clips of the week.
type StreamerDigestClip struct {
	ClipID    uuid.UUID
	Title     string
	ViewCount int64
}

// Empty reports whether there is nothing to tell the broadcaster this week.
// A pack that failed or expired is not news on its own.
func (d *StreamerDigest) Empty() bool {
	packNews := d.PackStatus == ClipPackReady || d.PackStatus == ClipPackPending || d.PackStatus == ClipPackBuilding
	return len(d.TopClips) == 0 && len(d.Placements) == 0 && !packNews
}
