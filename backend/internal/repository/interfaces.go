package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// SubscriptionRepositoryInterface defines the interface for subscription repository operations
type SubscriptionRepositoryInterface interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	GetByStripeCustomerID(ctx context.Context, customerID string) (*models.Subscription, error)
	GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*models.Subscription, error)
	Create(ctx context.Context, sub *models.Subscription) error
	Update(ctx context.Context, sub *models.Subscription) error
	GetEventByStripeEventID(ctx context.Context, eventID string) (*models.SubscriptionEvent, error)
	LogSubscriptionEvent(ctx context.Context, subscriptionID *uuid.UUID, eventType string, stripeEventID *string, eventData interface{}) error
}

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

// WebhookRepositoryInterface defines the interface for webhook repository operations
type WebhookRepositoryInterface interface {
	AddToRetryQueue(ctx context.Context, stripeEventID string, eventType string, payload interface{}, maxRetries int) error
	GetPendingRetries(ctx context.Context, limit int) ([]*models.WebhookRetryQueue, error)
	UpdateRetryQueueItem(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt *time.Time, lastError string) error
	RemoveFromRetryQueue(ctx context.Context, stripeEventID string) error
}

// DiscoveryListRepositoryInterface defines the interface for discovery list repository operations
type DiscoveryListRepositoryInterface interface {
	ListDiscoveryLists(ctx context.Context, featuredOnly bool, userID *uuid.UUID, limit, offset int) ([]models.DiscoveryListWithStats, error)
	GetDiscoveryList(ctx context.Context, idOrSlug string, userID *uuid.UUID) (*models.DiscoveryListWithStats, error)
	GetListClips(ctx context.Context, listID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]models.ClipWithSubmitter, int, error)
	GetListClipCount(ctx context.Context, listID uuid.UUID) (int, error)
	GetListClipsForExport(ctx context.Context, listID uuid.UUID, limit int) ([]models.ClipWithSubmitter, error)
	FollowList(ctx context.Context, userID, listID uuid.UUID) error
	UnfollowList(ctx context.Context, userID, listID uuid.UUID) error
	BookmarkList(ctx context.Context, userID, listID uuid.UUID) error
	UnbookmarkList(ctx context.Context, userID, listID uuid.UUID) error
	GetUserFollowedLists(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.DiscoveryListWithStats, error)
	CreateList(ctx context.Context, name, slug, description string, isFeatured bool, createdBy uuid.UUID) (*models.DiscoveryList, error)
	UpdateList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured *bool) (*models.DiscoveryList, error)
	DeleteList(ctx context.Context, listID uuid.UUID) error
	AddClipToList(ctx context.Context, listID, clipID uuid.UUID) error
	RemoveClipFromList(ctx context.Context, listID, clipID uuid.UUID) error
	ReorderClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error
	ReorderListClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error
	GetListClipsCount(ctx context.Context, listID uuid.UUID) (int, error)
	ListAllDiscoveryLists(ctx context.Context, limit, offset int) ([]models.DiscoveryListWithStats, error)
	CreateDiscoveryList(ctx context.Context, name, slug, description string, isFeatured bool, createdBy uuid.UUID) (*models.DiscoveryList, error)
	UpdateDiscoveryList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured *bool) (*models.DiscoveryList, error)
	DeleteDiscoveryList(ctx context.Context, listID uuid.UUID) error
}
