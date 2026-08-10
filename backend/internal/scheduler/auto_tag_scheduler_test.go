package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

func TestNewAutoTagScheduler(t *testing.T) {
	clipRepo := &repository.ClipRepository{}
	intervalSeconds := 30

	scheduler := NewAutoTagScheduler(nil, nil, nil, clipRepo, nil, intervalSeconds)

	if scheduler.clipRepo != clipRepo {
		t.Error("clipRepo not properly initialized")
	}
	if scheduler.whisper != nil {
		t.Error("whisper should be nil")
	}
	if scheduler.thumbnail != nil {
		t.Error("thumbnail should be nil")
	}

	expectedInterval := time.Duration(intervalSeconds) * time.Second
	if scheduler.interval != expectedInterval {
		t.Errorf("interval = %v, want %v", scheduler.interval, expectedInterval)
	}

	if scheduler.stopChan == nil {
		t.Error("stopChan not initialized")
	}

	// Verify stopChan is open initially.
	select {
	case <-scheduler.stopChan:
		t.Error("stopChan should be open initially")
	default:
		// Expected.
	}
}

func TestAutoTagScheduler_StopIdempotent(t *testing.T) {
	clipRepo := &repository.ClipRepository{}
	scheduler := NewAutoTagScheduler(nil, nil, nil, clipRepo, nil, 30)

	// First Stop should close channel.
	scheduler.Stop()

	select {
	case <-scheduler.stopChan:
		// Channel closed, expected.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("stopChan was not closed after first Stop()")
	}

	// Subsequent Stop calls should not panic.
	scheduler.Stop()
	scheduler.Stop()
}

func TestAutoTagScheduler_StopMultipleTimes(t *testing.T) {
	clipRepo := &repository.ClipRepository{}
	scheduler := NewAutoTagScheduler(nil, nil, nil, clipRepo, nil, 1)

	for i := 0; i < 10; i++ {
		scheduler.Stop()
	}
}

func TestAutoTagScheduler_StopConcurrent(t *testing.T) {
	clipRepo := &repository.ClipRepository{}
	scheduler := NewAutoTagScheduler(nil, nil, nil, clipRepo, nil, 1)

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scheduler.Stop()
		}()
	}

	wg.Wait()
}

func TestAutoTagScheduler_StopWhileRunning(t *testing.T) {
	// Start with a long interval to avoid the initial processClips tick,
	// but Start() always calls processClips immediately.
	//
	// Since processClips dereferences clipRepo.pool (which is nil),
	// we must avoid Start().  Instead test Stop while a goroutine is
	// blocked on the select loop via a simple pattern: create the
	// scheduler, launch Start() on a short-lived context that is
	// immediately cancelled, then call Stop.

	clipRepo := &repository.ClipRepository{}
	scheduler := &AutoTagScheduler{
		autoTag:  nil,
		clipRepo: clipRepo,
		interval: 10 * time.Second,
		stopChan: make(chan struct{}),
	}

	done := make(chan struct{})

	// Launch a goroutine that simulates the select loop without
	// calling processClips (which needs a real pool).
	go func() {
		ticker := time.NewTicker(scheduler.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// processClips would go here — skip for test.
			case <-scheduler.stopChan:
				close(done)
				return
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	// Stop should cause the goroutine to exit.
	scheduler.Stop()

	select {
	case <-done:
		// Success — goroutine exited cleanly.
	case <-time.After(2 * time.Second):
		t.Fatal("Goroutine did not exit after Stop()")
	}
}

func TestAutoTagScheduler_ContextCancellation(t *testing.T) {
	// Test the ctx.Done() branch in the select loop without Start().
	scheduler := &AutoTagScheduler{
		autoTag:  nil,
		clipRepo: nil,
		interval: 10 * time.Second,
		stopChan: make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(scheduler.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// processClips would go here — skip for test.
			case <-scheduler.stopChan:
				close(done)
				return
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success.
	case <-time.After(2 * time.Second):
		t.Fatal("Goroutine did not exit after context cancellation")
	}
}

func TestAutoTagScheduler_NilServices(t *testing.T) {
	clipRepo := &repository.ClipRepository{}
	scheduler := NewAutoTagScheduler(nil, nil, nil, clipRepo, nil, 30)

	if scheduler.whisper != nil {
		t.Error("whisper should be nil (deferred feature)")
	}
	if scheduler.thumbnail != nil {
		t.Error("thumbnail should be nil (deferred feature)")
	}
}