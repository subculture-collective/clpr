//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
)

func TestDiscoveryListRepository_CreateList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	// Create a test user for created_by
	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Test creating a discovery list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "A test discovery list", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	if list.ID == uuid.Nil {
		t.Error("Expected non-nil ID")
	}
	if list.Name != "Test List" {
		t.Errorf("Expected name 'Test List', got %s", list.Name)
	}
	if list.Slug != "test-list" {
		t.Errorf("Expected slug 'test-list', got %s", list.Slug)
	}
	if list.Description == nil || *list.Description != "A test discovery list" {
		t.Error("Expected description to be set")
	}
	if list.IsFeatured {
		t.Error("Expected is_featured to be false")
	}
	if !list.IsActive {
		t.Error("Expected is_active to be true by default")
	}
	if list.CreatedBy == nil || *list.CreatedBy != userID {
		t.Error("Expected created_by to be set")
	}
}

func TestDiscoveryListRepository_GetDiscoveryList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Test getting by ID
	retrieved, err := repo.GetDiscoveryList(ctx, list.ID.String(), nil)
	if err != nil {
		t.Fatalf("Failed to get discovery list by ID: %v", err)
	}

	if retrieved.ID != list.ID {
		t.Errorf("Expected ID %s, got %s", list.ID, retrieved.ID)
	}

	// Test getting by slug
	retrievedBySlug, err := repo.GetDiscoveryList(ctx, "test-list", nil)
	if err != nil {
		t.Fatalf("Failed to get discovery list by slug: %v", err)
	}

	if retrievedBySlug.ID != list.ID {
		t.Errorf("Expected ID %s, got %s", list.ID, retrievedBySlug.ID)
	}

	// Test getting non-existent list
	_, err = repo.GetDiscoveryList(ctx, uuid.New().String(), nil)
	if err != ErrDiscoveryListNotFound {
		t.Errorf("Expected ErrDiscoveryListNotFound, got %v", err)
	}
}

func TestDiscoveryListRepository_UpdateList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Update the list
	newName := "Updated List"
	newDesc := "Updated description"
	isFeatured := true

	updated, err := repo.UpdateList(ctx, list.ID, &newName, &newDesc, &isFeatured)
	if err != nil {
		t.Fatalf("Failed to update discovery list: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("Expected name '%s', got '%s'", newName, updated.Name)
	}
	if updated.Description == nil || *updated.Description != newDesc {
		t.Errorf("Expected description '%s', got %v", newDesc, updated.Description)
	}
	if !updated.IsFeatured {
		t.Error("Expected is_featured to be true")
	}

	// Test updating non-existent list
	_, err = repo.UpdateList(ctx, uuid.New(), &newName, nil, nil)
	if err != ErrDiscoveryListNotFound {
		t.Errorf("Expected ErrDiscoveryListNotFound, got %v", err)
	}

	// Test updating with no fields (should return error)
	_, err = repo.UpdateList(ctx, list.ID, nil, nil, nil)
	if err == nil {
		t.Error("Expected error when updating with no fields, got nil")
	}
}

func TestDiscoveryListRepository_DeleteList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Delete the list
	err = repo.DeleteList(ctx, list.ID)
	if err != nil {
		t.Fatalf("Failed to delete discovery list: %v", err)
	}

	// Verify it's deleted
	_, err = repo.GetDiscoveryList(ctx, list.ID.String(), nil)
	if err != ErrDiscoveryListNotFound {
		t.Errorf("Expected ErrDiscoveryListNotFound after deletion, got %v", err)
	}

	// Test deleting non-existent list
	err = repo.DeleteList(ctx, uuid.New())
	if err != ErrDiscoveryListNotFound {
		t.Errorf("Expected ErrDiscoveryListNotFound, got %v", err)
	}
}

func TestDiscoveryListRepository_AddAndRemoveClip(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "clips", "top_streamers")

	repo := NewDiscoveryListRepository(pool)
	clipRepo := NewClipRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Create a test clip
	clip := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    "test-clip-" + uuid.NewString(),
		TwitchClipURL:   "https://clips.twitch.tv/test",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=test",
		Title:           "Test Clip",
		CreatorName:     "creator",
		BroadcasterName: "broadcaster",
		CreatedAt:       time.Now(),
		ImportedAt:      time.Now(),
	}

	err = clipRepo.Create(ctx, clip)
	if err != nil {
		t.Fatalf("Failed to create clip: %v", err)
	}

	// Add clip to list
	err = repo.AddClipToList(ctx, list.ID, clip.ID)
	if err != nil {
		t.Fatalf("Failed to add clip to list: %v", err)
	}

	// Verify clip was added
	count, err := repo.GetListClipCount(ctx, list.ID)
	if err != nil {
		t.Fatalf("Failed to get clip count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected clip count 1, got %d", count)
	}

	// Remove clip from list
	err = repo.RemoveClipFromList(ctx, list.ID, clip.ID)
	if err != nil {
		t.Fatalf("Failed to remove clip from list: %v", err)
	}

	// Verify clip was removed
	count, err = repo.GetListClipCount(ctx, list.ID)
	if err != nil {
		t.Fatalf("Failed to get clip count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected clip count 0, got %d", count)
	}

	// Test removing non-existent clip
	err = repo.RemoveClipFromList(ctx, list.ID, uuid.New())
	if err != ErrClipNotFoundInList {
		t.Errorf("Expected ErrClipNotFoundInList, got %v", err)
	}
}

func TestDiscoveryListRepository_GetListClips(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "clips", "top_streamers", "votes")

	repo := NewDiscoveryListRepository(pool)
	clipRepo := NewClipRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Create test clips
	for i := 0; i < 5; i++ {
		clip := &models.Clip{
			ID:              uuid.New(),
			TwitchClipID:    "test-clip-" + uuid.NewString(),
			TwitchClipURL:   "https://clips.twitch.tv/test",
			EmbedURL:        "https://clips.twitch.tv/embed?clip=test",
			Title:           "Test Clip",
			CreatorName:     "creator",
			BroadcasterName: "broadcaster",
			CreatedAt:       time.Now(),
			ImportedAt:      time.Now(),
		}

		err = clipRepo.Create(ctx, clip)
		if err != nil {
			t.Fatalf("Failed to create clip: %v", err)
		}

		err = repo.AddClipToList(ctx, list.ID, clip.ID)
		if err != nil {
			t.Fatalf("Failed to add clip to list: %v", err)
		}
	}

	// Get clips with pagination
	clips, total, err := repo.GetListClips(ctx, list.ID, nil, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get list clips: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if len(clips) != 5 {
		t.Errorf("Expected 5 clips, got %d", len(clips))
	}

	// Test pagination
	clips, total, err = repo.GetListClips(ctx, list.ID, nil, 2, 0)
	if err != nil {
		t.Fatalf("Failed to get list clips with limit: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if len(clips) != 2 {
		t.Errorf("Expected 2 clips, got %d", len(clips))
	}
}

func TestDiscoveryListRepository_FollowAndUnfollow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)
	followerID := uuid.New()
	insertTestUser(t, pool, followerID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Follow the list
	err = repo.FollowList(ctx, followerID, list.ID)
	if err != nil {
		t.Fatalf("Failed to follow list: %v", err)
	}

	// Verify follow status
	retrieved, err := repo.GetDiscoveryList(ctx, list.ID.String(), &followerID)
	if err != nil {
		t.Fatalf("Failed to get discovery list: %v", err)
	}
	if !retrieved.IsFollowing {
		t.Error("Expected IsFollowing to be true")
	}

	// Follow again (should not error due to ON CONFLICT)
	err = repo.FollowList(ctx, followerID, list.ID)
	if err != nil {
		t.Fatalf("Failed to follow list again: %v", err)
	}

	// Unfollow the list
	err = repo.UnfollowList(ctx, followerID, list.ID)
	if err != nil {
		t.Fatalf("Failed to unfollow list: %v", err)
	}

	// Verify unfollow status
	retrieved, err = repo.GetDiscoveryList(ctx, list.ID.String(), &followerID)
	if err != nil {
		t.Fatalf("Failed to get discovery list: %v", err)
	}
	if retrieved.IsFollowing {
		t.Error("Expected IsFollowing to be false")
	}

	// Unfollow again (should error)
	err = repo.UnfollowList(ctx, followerID, list.ID)
	if err == nil {
		t.Error("Expected error when unfollowing a list not followed")
	}
}

func TestDiscoveryListRepository_BookmarkAndUnbookmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)
	bookmarkerID := uuid.New()
	insertTestUser(t, pool, bookmarkerID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Bookmark the list
	err = repo.BookmarkList(ctx, bookmarkerID, list.ID)
	if err != nil {
		t.Fatalf("Failed to bookmark list: %v", err)
	}

	// Verify bookmark status
	retrieved, err := repo.GetDiscoveryList(ctx, list.ID.String(), &bookmarkerID)
	if err != nil {
		t.Fatalf("Failed to get discovery list: %v", err)
	}
	if !retrieved.IsBookmarked {
		t.Error("Expected IsBookmarked to be true")
	}

	// Bookmark again (should not error due to ON CONFLICT)
	err = repo.BookmarkList(ctx, bookmarkerID, list.ID)
	if err != nil {
		t.Fatalf("Failed to bookmark list again: %v", err)
	}

	// Unbookmark the list
	err = repo.UnbookmarkList(ctx, bookmarkerID, list.ID)
	if err != nil {
		t.Fatalf("Failed to unbookmark list: %v", err)
	}

	// Verify unbookmark status
	retrieved, err = repo.GetDiscoveryList(ctx, list.ID.String(), &bookmarkerID)
	if err != nil {
		t.Fatalf("Failed to get discovery list: %v", err)
	}
	if retrieved.IsBookmarked {
		t.Error("Expected IsBookmarked to be false")
	}

	// Unbookmark again (should error)
	err = repo.UnbookmarkList(ctx, bookmarkerID, list.ID)
	if err == nil {
		t.Error("Expected error when unbookmarking a list not bookmarked")
	}
}

func TestDiscoveryListRepository_ReorderClips(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "clips", "top_streamers")

	repo := NewDiscoveryListRepository(pool)
	clipRepo := NewClipRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Create a test list
	list, err := repo.CreateList(ctx, "Test List", "test-list", "Test description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create discovery list: %v", err)
	}

	// Create test clips
	clipIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		clip := &models.Clip{
			ID:              uuid.New(),
			TwitchClipID:    "test-clip-" + uuid.NewString(),
			TwitchClipURL:   "https://clips.twitch.tv/test",
			EmbedURL:        "https://clips.twitch.tv/embed?clip=test",
			Title:           "Test Clip",
			CreatorName:     "creator",
			BroadcasterName: "broadcaster",
			CreatedAt:       time.Now(),
			ImportedAt:      time.Now(),
		}

		err = clipRepo.Create(ctx, clip)
		if err != nil {
			t.Fatalf("Failed to create clip: %v", err)
		}

		err = repo.AddClipToList(ctx, list.ID, clip.ID)
		if err != nil {
			t.Fatalf("Failed to add clip to list: %v", err)
		}

		clipIDs[i] = clip.ID
	}

	// Reorder clips (reverse order)
	reversedIDs := []uuid.UUID{clipIDs[2], clipIDs[1], clipIDs[0]}
	err = repo.ReorderClips(ctx, list.ID, reversedIDs)
	if err != nil {
		t.Fatalf("Failed to reorder clips: %v", err)
	}

	// Verify order by fetching clips
	clips, _, err := repo.GetListClips(ctx, list.ID, nil, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get list clips: %v", err)
	}

	if len(clips) != 3 {
		t.Errorf("Expected 3 clips, got %d", len(clips))
	}

	// Verify the order matches the reversed order
	for i, clip := range clips {
		if clip.ID != reversedIDs[i] {
			t.Errorf("Expected clip at position %d to be %s, got %s", i, reversedIDs[i], clip.ID)
		}
	}
}

func TestDiscoveryListRepository_ListDiscoveryLists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)

	// Create test lists
	_, err := repo.CreateList(ctx, "Featured List 1", "featured-1", "Featured", true, userID)
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	_, err = repo.CreateList(ctx, "Featured List 2", "featured-2", "Featured", true, userID)
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	_, err = repo.CreateList(ctx, "Regular List", "regular", "Regular", false, userID)
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// List all lists
	lists, err := repo.ListDiscoveryLists(ctx, false, nil, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list discovery lists: %v", err)
	}

	if len(lists) < 3 {
		t.Errorf("Expected at least 3 lists, got %d", len(lists))
	}

	// List only featured
	featuredLists, err := repo.ListDiscoveryLists(ctx, true, nil, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list featured discovery lists: %v", err)
	}

	if len(featuredLists) < 2 {
		t.Errorf("Expected at least 2 featured lists, got %d", len(featuredLists))
	}

	for _, list := range featuredLists {
		if !list.IsFeatured {
			t.Error("Expected all lists to be featured")
		}
	}
}

func TestDiscoveryListRepository_GetUserFollowedLists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "discovery_lists", "discovery_list_clips", "discovery_list_follows", "discovery_list_bookmarks")

	repo := NewDiscoveryListRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)
	followerID := uuid.New()
	insertTestUser(t, pool, followerID)

	// Create test lists
	list1, err := repo.CreateList(ctx, "List 1", "list-1", "Description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create list 1: %v", err)
	}

	list2, err := repo.CreateList(ctx, "List 2", "list-2", "Description", false, userID)
	if err != nil {
		t.Fatalf("Failed to create list 2: %v", err)
	}

	// Follow the lists
	err = repo.FollowList(ctx, followerID, list1.ID)
	if err != nil {
		t.Fatalf("Failed to follow list 1: %v", err)
	}

	err = repo.FollowList(ctx, followerID, list2.ID)
	if err != nil {
		t.Fatalf("Failed to follow list 2: %v", err)
	}

	// Get followed lists
	followedLists, err := repo.GetUserFollowedLists(ctx, followerID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get followed lists: %v", err)
	}

	if len(followedLists) != 2 {
		t.Errorf("Expected 2 followed lists, got %d", len(followedLists))
	}

	// Verify all returned lists have IsFollowing = true
	for _, list := range followedLists {
		if !list.IsFollowing {
			t.Error("Expected all lists to have IsFollowing = true")
		}
	}
}
