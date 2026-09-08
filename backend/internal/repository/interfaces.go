package repository

import (
	"context"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

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
	UpdateList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured, isActive *bool) (*models.DiscoveryList, error)
	DeleteList(ctx context.Context, listID uuid.UUID) error
	AddClipToList(ctx context.Context, listID, clipID uuid.UUID) error
	RemoveClipFromList(ctx context.Context, listID, clipID uuid.UUID) error
	ReorderClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error
	ReorderListClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error
	GetListClipsCount(ctx context.Context, listID uuid.UUID) (int, error)
	ListAllDiscoveryLists(ctx context.Context, limit, offset int) ([]models.DiscoveryListWithStats, error)
	CreateDiscoveryList(ctx context.Context, name, slug, description string, isFeatured bool, createdBy uuid.UUID) (*models.DiscoveryList, error)
	UpdateDiscoveryList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured, isActive *bool) (*models.DiscoveryList, error)
	DeleteDiscoveryList(ctx context.Context, listID uuid.UUID) error
}
