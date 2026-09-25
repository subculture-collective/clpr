package scheduler

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fakeLiveStatusService struct {
	calls [][]string
}

func (f *fakeLiveStatusService) UpdateLiveStatusForBroadcasters(_ context.Context, ids []string) error {
	f.calls = append(f.calls, append([]string(nil), ids...))
	return nil
}

type fakeLiveBroadcasterRepo struct {
	followed      []string
	followedErr   error
	candidates    []string
	candidateErr  error
	candidateHits int
	gotLimit      int
	gotWindow     time.Duration
}

func (f *fakeLiveBroadcasterRepo) GetAllFollowedBroadcasterIDs(context.Context) ([]string, error) {
	return f.followed, f.followedErr
}

func (f *fakeLiveBroadcasterRepo) GetLiveStatusCandidateBroadcasterIDs(_ context.Context, limit int, window time.Duration) ([]string, error) {
	f.candidateHits++
	f.gotLimit, f.gotWindow = limit, window
	if f.candidateErr != nil {
		return nil, f.candidateErr
	}
	if len(f.candidates) > limit {
		return f.candidates[:limit], nil
	}
	return f.candidates, nil
}

func TestLiveStatusSchedulerChecksFollowedAndPopularBroadcasters(t *testing.T) {
	svc := &fakeLiveStatusService{}
	repo := &fakeLiveBroadcasterRepo{
		followed:   []string{"100", "200"},
		candidates: []string{"200", "300", "", "400"},
	}
	s := NewLiveStatusScheduler(svc, repo, 30, LiveStatusCandidateConfig{Limit: 10, Window: 7 * 24 * time.Hour})

	s.updateLiveStatuses(context.Background())

	if want := [][]string{{"100", "200", "300", "400"}}; !reflect.DeepEqual(svc.calls, want) {
		t.Fatalf("checked %v, want %v", svc.calls, want)
	}
	if repo.gotLimit != 10 || repo.gotWindow != 7*24*time.Hour {
		t.Fatalf("candidate query got limit %d window %v", repo.gotLimit, repo.gotWindow)
	}
}

func TestLiveStatusSchedulerCachesCandidatesUntilRefresh(t *testing.T) {
	svc := &fakeLiveStatusService{}
	repo := &fakeLiveBroadcasterRepo{candidates: []string{"300"}}
	s := NewLiveStatusScheduler(svc, repo, 30, LiveStatusCandidateConfig{Limit: 5, Refresh: 10 * time.Minute})
	now := time.Date(2026, 9, 25, 5, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }

	s.updateLiveStatuses(context.Background())
	now = now.Add(5 * time.Minute)
	repo.candidates = []string{"999"}
	s.updateLiveStatuses(context.Background())
	if repo.candidateHits != 1 {
		t.Fatalf("candidate query ran %d times within refresh interval, want 1", repo.candidateHits)
	}

	now = now.Add(6 * time.Minute)
	s.updateLiveStatuses(context.Background())
	if repo.candidateHits != 2 {
		t.Fatalf("candidate query ran %d times after refresh interval, want 2", repo.candidateHits)
	}
	if got := svc.calls[len(svc.calls)-1]; !reflect.DeepEqual(got, []string{"999"}) {
		t.Fatalf("after refresh checked %v, want [999]", got)
	}
}

func TestLiveStatusSchedulerKeepsCandidatesWhenQueryFails(t *testing.T) {
	svc := &fakeLiveStatusService{}
	repo := &fakeLiveBroadcasterRepo{candidates: []string{"300", "400"}}
	s := NewLiveStatusScheduler(svc, repo, 30, LiveStatusCandidateConfig{Limit: 5, Refresh: time.Minute})
	now := time.Now()
	s.now = func() time.Time { return now }

	s.updateLiveStatuses(context.Background())
	now = now.Add(2 * time.Minute)
	repo.candidateErr = errors.New("statement timeout")
	repo.followedErr = errors.New("statement timeout")
	s.updateLiveStatuses(context.Background())

	if got := svc.calls[len(svc.calls)-1]; !reflect.DeepEqual(got, []string{"300", "400"}) {
		t.Fatalf("after failures checked %v, want previous candidates", got)
	}
}

func TestLiveStatusSchedulerCandidateLimitZeroChecksFollowedOnly(t *testing.T) {
	svc := &fakeLiveStatusService{}
	repo := &fakeLiveBroadcasterRepo{followed: []string{"100"}, candidates: []string{"300"}}
	s := NewLiveStatusScheduler(svc, repo, 30, LiveStatusCandidateConfig{})

	s.updateLiveStatuses(context.Background())

	if repo.candidateHits != 0 {
		t.Fatalf("candidate query ran with limit 0")
	}
	if want := [][]string{{"100"}}; !reflect.DeepEqual(svc.calls, want) {
		t.Fatalf("checked %v, want %v", svc.calls, want)
	}
}

func TestLiveStatusSchedulerSkipsEmptyRun(t *testing.T) {
	svc := &fakeLiveStatusService{}
	s := NewLiveStatusScheduler(svc, &fakeLiveBroadcasterRepo{}, 30, LiveStatusCandidateConfig{Limit: 5})
	s.updateLiveStatuses(context.Background())
	if len(svc.calls) != 0 {
		t.Fatalf("service called with no broadcasters: %v", svc.calls)
	}
}
