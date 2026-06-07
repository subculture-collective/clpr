package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockTagRepository is a mock implementation of TagRepository for testing
type MockTagRepository struct {
	tags     map[uuid.UUID]*models.Tag
	clipTags map[uuid.UUID]map[uuid.UUID]bool // clipID -> tagID -> exists
}

func NewMockTagRepository() *MockTagRepository {
	return &MockTagRepository{
		tags:     make(map[uuid.UUID]*models.Tag),
		clipTags: make(map[uuid.UUID]map[uuid.UUID]bool),
	}
}

func (m *MockTagRepository) Create(ctx context.Context, tag *models.Tag) error {
	m.tags[tag.ID] = tag
	return nil
}

func (m *MockTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	if tag, ok := m.tags[id]; ok {
		return tag, nil
	}
	return nil, nil
}

func (m *MockTagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	for _, tag := range m.tags {
		if tag.Slug == slug {
			return tag, nil
		}
	}
	return nil, nil
}

func (m *MockTagRepository) AddTagToClip(ctx context.Context, clipID, tagID uuid.UUID) error {
	if m.clipTags[clipID] == nil {
		m.clipTags[clipID] = make(map[uuid.UUID]bool)
	}
	m.clipTags[clipID][tagID] = true

	// Increment usage count
	if tag, ok := m.tags[tagID]; ok {
		tag.UsageCount++
	}
	return nil
}

func (m *MockTagRepository) RemoveTagFromClip(ctx context.Context, clipID, tagID uuid.UUID) error {
	if m.clipTags[clipID] != nil {
		delete(m.clipTags[clipID], tagID)
	}

	// Decrement usage count
	if tag, ok := m.tags[tagID]; ok && tag.UsageCount > 0 {
		tag.UsageCount--
	}
	return nil
}

func (m *MockTagRepository) GetClipTags(ctx context.Context, clipID uuid.UUID) ([]*models.Tag, error) {
	var tags []*models.Tag
	if tagMap, ok := m.clipTags[clipID]; ok {
		for tagID := range tagMap {
			if tag, exists := m.tags[tagID]; exists {
				tags = append(tags, tag)
			}
		}
	}
	return tags, nil
}

func TestMockTagRepository_Create(t *testing.T) {
	repo := NewMockTagRepository()
	ctx := context.Background()

	tag := &models.Tag{
		ID:          uuid.New(),
		Name:        "Ace",
		Slug:        "ace",
		Description: stringPtr("Team wipe achievement"),
		Color:       stringPtr("#FF0000"),
		UsageCount:  0,
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, tag.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve tag: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved tag is nil")
	}

	if retrieved.Name != tag.Name {
		t.Errorf("Expected tag name %s, got %s", tag.Name, retrieved.Name)
	}

	if retrieved.Slug != tag.Slug {
		t.Errorf("Expected tag slug %s, got %s", tag.Slug, retrieved.Slug)
	}
}

func TestMockTagRepository_GetBySlug(t *testing.T) {
	repo := NewMockTagRepository()
	ctx := context.Background()

	tag := &models.Tag{
		ID:         uuid.New(),
		Name:       "Clutch",
		Slug:       "clutch",
		UsageCount: 0,
		CreatedAt:  time.Now(),
	}

	err := repo.Create(ctx, tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	retrieved, err := repo.GetBySlug(ctx, "clutch")
	if err != nil {
		t.Fatalf("Failed to retrieve tag: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved tag is nil")
	}

	if retrieved.Name != "Clutch" {
		t.Errorf("Expected tag name Clutch, got %s", retrieved.Name)
	}
}

func TestMockTagRepository_AddTagToClip(t *testing.T) {
	repo := NewMockTagRepository()
	ctx := context.Background()

	tag := &models.Tag{
		ID:         uuid.New(),
		Name:       "Fail",
		Slug:       "fail",
		UsageCount: 0,
		CreatedAt:  time.Now(),
	}

	err := repo.Create(ctx, tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	clipID := uuid.New()
	err = repo.AddTagToClip(ctx, clipID, tag.ID)
	if err != nil {
		t.Fatalf("Failed to add tag to clip: %v", err)
	}

	tags, err := repo.GetClipTags(ctx, clipID)
	if err != nil {
		t.Fatalf("Failed to get clip tags: %v", err)
	}

	if len(tags) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0].Slug != "fail" {
		t.Errorf("Expected tag slug fail, got %s", tags[0].Slug)
	}

	// Check usage count was incremented
	retrieved, _ := repo.GetByID(ctx, tag.ID)
	if retrieved.UsageCount != 1 {
		t.Errorf("Expected usage count 1, got %d", retrieved.UsageCount)
	}
}

func TestMockTagRepository_RemoveTagFromClip(t *testing.T) {
	repo := NewMockTagRepository()
	ctx := context.Background()

	tag := &models.Tag{
		ID:         uuid.New(),
		Name:       "Funny",
		Slug:       "funny",
		UsageCount: 0,
		CreatedAt:  time.Now(),
	}

	err := repo.Create(ctx, tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	clipID := uuid.New()
	err = repo.AddTagToClip(ctx, clipID, tag.ID)
	if err != nil {
		t.Fatalf("Failed to add tag to clip: %v", err)
	}

	// Verify tag was added
	tags, _ := repo.GetClipTags(ctx, clipID)
	if len(tags) != 1 {
		t.Fatalf("Expected 1 tag after adding, got %d", len(tags))
	}

	// Remove tag
	err = repo.RemoveTagFromClip(ctx, clipID, tag.ID)
	if err != nil {
		t.Fatalf("Failed to remove tag from clip: %v", err)
	}

	// Verify tag was removed
	tags, _ = repo.GetClipTags(ctx, clipID)
	if len(tags) != 0 {
		t.Fatalf("Expected 0 tags after removing, got %d", len(tags))
	}

	// Check usage count was decremented
	retrieved, _ := repo.GetByID(ctx, tag.ID)
	if retrieved.UsageCount != 0 {
		t.Errorf("Expected usage count 0, got %d", retrieved.UsageCount)
	}
}

func TestMockTagRepository_MultipleTagsPerClip(t *testing.T) {
	repo := NewMockTagRepository()
	ctx := context.Background()

	// Create multiple tags
	tags := []*models.Tag{
		{
			ID:         uuid.New(),
			Name:       "Ace",
			Slug:       "ace",
			UsageCount: 0,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Name:       "Clutch",
			Slug:       "clutch",
			UsageCount: 0,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Name:       "Epic",
			Slug:       "epic",
			UsageCount: 0,
			CreatedAt:  time.Now(),
		},
	}

	for _, tag := range tags {
		err := repo.Create(ctx, tag)
		if err != nil {
			t.Fatalf("Failed to create tag: %v", err)
		}
	}

	// Add all tags to a clip
	clipID := uuid.New()
	for _, tag := range tags {
		err := repo.AddTagToClip(ctx, clipID, tag.ID)
		if err != nil {
			t.Fatalf("Failed to add tag to clip: %v", err)
		}
	}

	// Verify all tags were added
	clipTags, err := repo.GetClipTags(ctx, clipID)
	if err != nil {
		t.Fatalf("Failed to get clip tags: %v", err)
	}

	if len(clipTags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(clipTags))
	}
}

func stringPtr(s string) *string {
	return &s
}
