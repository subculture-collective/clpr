package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type promotionServiceStub struct{ calls atomic.Int32 }

func (s *promotionServiceStub) CheckPromotionCandidates(context.Context) ([]string, error) {
	s.calls.Add(1)
	return []string{"community/example"}, nil
}

func (s *promotionServiceStub) PendingPromotionCount(context.Context) (int, error) { return 1, nil }

func TestTagPromotionSchedulerHonorsInitialDelay(t *testing.T) {
	service := &promotionServiceStub{}
	scheduler := NewTagPromotionScheduler(service, time.Hour, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		scheduler.Start(ctx)
	}()
	time.Sleep(20 * time.Millisecond)
	if got := service.calls.Load(); got != 0 {
		t.Fatalf("calls before initial delay = %d", got)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after cancellation")
	}
}
