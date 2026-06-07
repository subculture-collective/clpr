package services

import (
	"context"

	"github.com/google/uuid"
)

// CreatorModerationChecker is the minimal moderation surface used by content services.
type CreatorModerationChecker interface {
	CanInteract(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error)
	CanSubmit(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error)
	CanComment(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error)
}

// CreatorModerationError marks moderation denials so handlers can return 403.
type CreatorModerationError struct {
	Message string
}

func (e *CreatorModerationError) Error() string {
	if e == nil {
		return "creator moderation denied"
	}
	return e.Message
}
