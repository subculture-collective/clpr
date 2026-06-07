package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockWatchPartyRepository is a mock implementation for testing
type MockWatchPartyRepository struct {
	parties      map[uuid.UUID]*models.WatchParty
	participants map[uuid.UUID][]models.WatchPartyParticipant
	messages     map[uuid.UUID][]models.WatchPartyMessage
	reactions    map[uuid.UUID][]models.WatchPartyReaction
}

func NewMockWatchPartyRepository() *MockWatchPartyRepository {
	return &MockWatchPartyRepository{
		parties:      make(map[uuid.UUID]*models.WatchParty),
		participants: make(map[uuid.UUID][]models.WatchPartyParticipant),
		messages:     make(map[uuid.UUID][]models.WatchPartyMessage),
		reactions:    make(map[uuid.UUID][]models.WatchPartyReaction),
	}
}

func (m *MockWatchPartyRepository) CreateMessage(ctx context.Context, message *models.WatchPartyMessage) error {
	message.CreatedAt = time.Now()
	m.messages[message.WatchPartyID] = append(m.messages[message.WatchPartyID], *message)
	return nil
}

func (m *MockWatchPartyRepository) GetMessages(ctx context.Context, partyID uuid.UUID, limit int) ([]models.WatchPartyMessage, error) {
	msgs := m.messages[partyID]
	if len(msgs) > limit {
		return msgs[len(msgs)-limit:], nil
	}
	return msgs, nil
}

func (m *MockWatchPartyRepository) CreateReaction(ctx context.Context, reaction *models.WatchPartyReaction) error {
	reaction.CreatedAt = time.Now()
	m.reactions[reaction.WatchPartyID] = append(m.reactions[reaction.WatchPartyID], *reaction)
	return nil
}

func (m *MockWatchPartyRepository) GetRecentReactions(ctx context.Context, partyID uuid.UUID, since time.Time) ([]models.WatchPartyReaction, error) {
	var reactions []models.WatchPartyReaction
	for _, reaction := range m.reactions[partyID] {
		if reaction.CreatedAt.After(since) || reaction.CreatedAt.Equal(since) {
			reactions = append(reactions, reaction)
		}
	}
	return reactions, nil
}

// Test creating and retrieving messages
func TestWatchPartyMessageLifecycle(t *testing.T) {
	repo := NewMockWatchPartyRepository()
	ctx := context.Background()

	partyID := uuid.New()
	userID := uuid.New()

	// Create a message
	message := &models.WatchPartyMessage{
		ID:           uuid.New(),
		WatchPartyID: partyID,
		UserID:       userID,
		Message:      "Hello, world!",
	}

	err := repo.CreateMessage(ctx, message)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Retrieve messages
	messages, err := repo.GetMessages(ctx, partyID, 100)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Message != "Hello, world!" {
		t.Fatalf("Expected message 'Hello, world!', got '%s'", messages[0].Message)
	}
}

// Test creating and retrieving reactions
func TestWatchPartyReactionLifecycle(t *testing.T) {
	repo := NewMockWatchPartyRepository()
	ctx := context.Background()

	partyID := uuid.New()
	userID := uuid.New()

	// Create a reaction
	videoTimestamp := 42.5
	reaction := &models.WatchPartyReaction{
		ID:             uuid.New(),
		WatchPartyID:   partyID,
		UserID:         userID,
		Emoji:          "🔥",
		VideoTimestamp: &videoTimestamp,
	}

	err := repo.CreateReaction(ctx, reaction)
	if err != nil {
		t.Fatalf("Failed to create reaction: %v", err)
	}

	// Retrieve reactions
	since := time.Now().Add(-1 * time.Minute)
	reactions, err := repo.GetRecentReactions(ctx, partyID, since)
	if err != nil {
		t.Fatalf("Failed to get reactions: %v", err)
	}

	if len(reactions) != 1 {
		t.Fatalf("Expected 1 reaction, got %d", len(reactions))
	}

	if reactions[0].Emoji != "🔥" {
		t.Fatalf("Expected emoji '🔥', got '%s'", reactions[0].Emoji)
	}

	if reactions[0].VideoTimestamp == nil || *reactions[0].VideoTimestamp != 42.5 {
		t.Fatalf("Expected video timestamp 42.5, got %v", reactions[0].VideoTimestamp)
	}
}

// Test message limit
func TestWatchPartyMessageLimit(t *testing.T) {
	repo := NewMockWatchPartyRepository()
	ctx := context.Background()

	partyID := uuid.New()
	userID := uuid.New()

	// Create multiple messages
	for i := 0; i < 150; i++ {
		message := &models.WatchPartyMessage{
			ID:           uuid.New(),
			WatchPartyID: partyID,
			UserID:       userID,
			Message:      "Test message",
		}
		err := repo.CreateMessage(ctx, message)
		if err != nil {
			t.Fatalf("Failed to create message %d: %v", i, err)
		}
	}

	// Retrieve with limit
	messages, err := repo.GetMessages(ctx, partyID, 100)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 100 {
		t.Fatalf("Expected 100 messages (limit), got %d", len(messages))
	}
}
