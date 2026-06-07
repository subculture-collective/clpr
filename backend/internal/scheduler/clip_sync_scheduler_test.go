package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// MockClipSyncService is a mock implementation for testing
type MockClipSyncService struct{}

// SyncTrendingClips is a mock implementation that matches the interface
func (m *MockClipSyncService) SyncTrendingClips(ctx context.Context, hours int, opts *services.TrendingSyncOptions) (*services.SyncStats, error) {
	return &services.SyncStats{
		ClipsFetched: 5,
		ClipsCreated: 5,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Second),
	}, nil
}

// TestStopMultipleTimes verifies that calling Stop() multiple times doesn't panic
func TestStopMultipleTimes(t *testing.T) {
	mockService := &MockClipSyncService{}
	scheduler := NewClipSyncScheduler(mockService, 1)

	// Call Stop() multiple times - should not panic
	for i := 0; i < 10; i++ {
		scheduler.Stop()
	}
}

// TestConcurrentStopCalls verifies thread-safety of Stop() with concurrent calls
func TestConcurrentStopCalls(t *testing.T) {
	mockService := &MockClipSyncService{}
	scheduler := NewClipSyncScheduler(mockService, 1)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Launch multiple goroutines calling Stop() concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scheduler.Stop()
		}()
	}

	wg.Wait()
}

// TestStopWhileRunning verifies that Stop() works correctly while Start() is running
func TestStopWhileRunning(t *testing.T) {
	mockService := &MockClipSyncService{}
	scheduler := NewClipSyncScheduler(mockService, 10) // 10 minute interval

	ctx := context.Background()

	// Start the scheduler in a goroutine
	done := make(chan bool)
	go func() {
		scheduler.Start(ctx)
		done <- true
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Stop the scheduler
	scheduler.Stop()

	// Wait for Start() to finish
	select {
	case <-done:
		// Success - Start() exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not exit after Stop() was called")
	}
}

// TestMultipleStopWhileRunning verifies calling Stop() multiple times while Start() is running
func TestMultipleStopWhileRunning(t *testing.T) {
	mockService := &MockClipSyncService{}
	scheduler := NewClipSyncScheduler(mockService, 10)

	ctx := context.Background()

	// Start the scheduler
	done := make(chan bool)
	go func() {
		scheduler.Start(ctx)
		done <- true
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Call Stop() multiple times concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scheduler.Stop()
		}()
	}
	wg.Wait()

	// Wait for Start() to finish
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not exit after Stop() was called")
	}
}

// TestStopIdempotency verifies that Stop() is idempotent and safe to call multiple times
func TestStopIdempotency(t *testing.T) {
	mockService := &MockClipSyncService{}
	scheduler := NewClipSyncScheduler(mockService, 1)

	// First Stop() should close the channel
	scheduler.Stop()

	// Verify channel is closed by trying to read from it
	select {
	case <-scheduler.stopChan:
		// Channel is closed, as expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("stopChan was not closed after first Stop()")
	}

	// Subsequent Stop() calls should be no-ops and not panic
	scheduler.Stop()
	scheduler.Stop()
}

// TestContextCancellation verifies that Start() respects context cancellation
func TestContextCancellation(t *testing.T) {
	mockService := &MockClipSyncService{}
	scheduler := NewClipSyncScheduler(mockService, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)
	go func() {
		scheduler.Start(ctx)
		done <- true
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context
	cancel()

	// Wait for Start() to finish
	select {
	case <-done:
		// Success - Start() exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not exit after context was cancelled")
	}
}

// TestNewClipSyncScheduler verifies proper initialization
func TestNewClipSyncScheduler(t *testing.T) {
	mockService := &MockClipSyncService{}
	intervalMinutes := 15

	scheduler := NewClipSyncScheduler(mockService, intervalMinutes)

	if scheduler.syncService != mockService {
		t.Error("syncService not properly initialized")
	}

	expectedInterval := time.Duration(intervalMinutes) * time.Minute
	if scheduler.interval != expectedInterval {
		t.Errorf("interval = %v, want %v", scheduler.interval, expectedInterval)
	}

	if scheduler.stopChan == nil {
		t.Error("stopChan not initialized")
	}

	// Verify stopChan is open initially
	select {
	case <-scheduler.stopChan:
		t.Error("stopChan should be open initially")
	default:
		// Expected - channel is open
	}
}
