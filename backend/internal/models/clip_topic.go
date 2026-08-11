package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ClipTopic struct {
	ClipID           uuid.UUID       `json:"clip_id"`
	TopicID          uuid.UUID       `json:"topic_id"`
	TopicName        string          `json:"topic_name"`
	TopicSlug        string          `json:"topic_slug"`
	Source           string          `json:"source"`
	Confidence       float64         `json:"confidence"`
	Evidence         json.RawMessage `json:"evidence"`
	AssignedByUserID *uuid.UUID      `json:"assigned_by_user_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type TopicCandidate struct {
	TopicID    uuid.UUID
	Source     string
	Confidence float64
	Evidence   json.RawMessage
}
type ClipTopicClassificationInput struct {
	ClipID                                                  uuid.UUID
	Title, TwitchCategoryID, TwitchCategoryName, Transcript string
	Tags                                                    []string
	MappedTopics                                            []Category
}
type ReplaceClipTopicsRequest struct {
	TopicIDs []uuid.UUID `json:"topic_ids" binding:"max=5"`
}
type MergeTopicsRequest struct {
	TargetTopicID uuid.UUID `json:"target_topic_id" binding:"required"`
}
type SplitTopicRequest struct {
	Name        string      `json:"name" binding:"required,max=100"`
	Slug        string      `json:"slug" binding:"required,max=100"`
	Description *string     `json:"description,omitempty" binding:"omitempty,max=1000"`
	ClipIDs     []uuid.UUID `json:"clip_ids" binding:"required,min=1,max=500"`
}
