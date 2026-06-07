package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StreamerClipRoomApprovalManual = "manual"
	StreamerClipRoomApprovalAuto   = "auto"

	StreamerClipRoomItemStatusPending  = "pending"
	StreamerClipRoomItemStatusApproved = "approved"
	StreamerClipRoomItemStatusRejected = "rejected"
	StreamerClipRoomItemStatusSkipped  = "skipped"

	StreamerClipRoomSourceClpr   = "clpr"
	StreamerClipRoomSourceTwitch = "twitch"
)

type StreamerClipRoom struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	OwnerUserID       uuid.UUID  `json:"owner_user_id" db:"owner_user_id"`
	TwitchChannel     string     `json:"twitch_channel" db:"twitch_channel"`
	ApprovalMode      string     `json:"approval_mode" db:"approval_mode"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	LastListenerError *string    `json:"last_listener_error,omitempty" db:"last_listener_error"`
	ListenerStartedAt *time.Time `json:"listener_started_at,omitempty" db:"listener_started_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type StreamerClipRoomItem struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	RoomID           uuid.UUID  `json:"room_id" db:"room_id"`
	ClipID           *uuid.UUID `json:"clip_id,omitempty" db:"clip_id"`
	Clip             *Clip      `json:"clip,omitempty" db:"-"`
	SourceURL        string     `json:"source_url" db:"source_url"`
	SourceType       string     `json:"source_type" db:"source_type"`
	Status           string     `json:"status" db:"status"`
	Position         *int       `json:"position,omitempty" db:"position"`
	TwitchMessageID  *string    `json:"twitch_message_id,omitempty" db:"twitch_message_id"`
	TwitchUserID     *string    `json:"twitch_user_id,omitempty" db:"twitch_user_id"`
	TwitchUsername   *string    `json:"twitch_username,omitempty" db:"twitch_username"`
	MessageText      *string    `json:"message_text,omitempty" db:"message_text"`
	SkipReason       *string    `json:"skip_reason,omitempty" db:"skip_reason"`
	DetectedAt       time.Time  `json:"detected_at" db:"detected_at"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	ApprovedByUserID *uuid.UUID `json:"approved_by_user_id,omitempty" db:"approved_by_user_id"`
	RejectedAt       *time.Time `json:"rejected_at,omitempty" db:"rejected_at"`
	RejectedByUserID *uuid.UUID `json:"rejected_by_user_id,omitempty" db:"rejected_by_user_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type StreamerClipRoomWithItems struct {
	Room          StreamerClipRoom       `json:"room"`
	PendingItems  []StreamerClipRoomItem `json:"pending_items"`
	ApprovedItems []StreamerClipRoomItem `json:"approved_items"`
	SkippedItems  []StreamerClipRoomItem `json:"skipped_items"`
}

type ReorderStreamerClipRoomItemsRequest struct {
	ItemIDs []string `json:"item_ids" binding:"required,min=1"`
}

type StreamerClipRoomEvent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
