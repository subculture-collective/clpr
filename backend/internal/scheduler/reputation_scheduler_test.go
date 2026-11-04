package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockReputationService struct {
	mu           sync.Mutex
	statsUpdated []uuid.UUID
}

func (m *mockReputationService) CheckAndAwardBadges(context.Context, uuid.UUID) ([]string, error) {
	return nil, nil
}

func (m *mockReputationService) UpdateUserStats(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statsUpdated = append(m.statsUpdated, userID)
	return nil
}

type mockActiveUserRepository struct {
	since time.Time
	users []uuid.UUID
}

func (m *mockActiveUserRepository) GetUserIDsActiveSince(_ context.Context, since time.Time) ([]uuid.UUID, error) {
	m.since = since
	return m.users, nil
}

func TestReputationSchedulerRecalculatesOnlyRecentlyActiveUsers(t *testing.T) {
	users := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &mockActiveUserRepository{users: users}
	service := &mockReputationService{}
	scheduler := NewReputationScheduler(service, repo, 6)

	before := time.Now()
	scheduler.runTasks(context.Background())

	wantSince := before.Add(-reputationActivityWindow)
	if repo.since.Before(wantSince.Add(-time.Minute)) || repo.since.After(wantSince.Add(time.Minute)) {
		t.Fatalf("active-since cutoff = %s, want about %s", repo.since, wantSince)
	}
	if len(service.statsUpdated) != len(users) {
		t.Fatalf("updated stats for %d users, want %d", len(service.statsUpdated), len(users))
	}
}
