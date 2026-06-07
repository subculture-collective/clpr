package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	schedulertesting "git.subcult.tv/subculture-collective/clpr/internal/scheduler/testing"
)

// TestSchedulerWithMockClock demonstrates using the mock clock for deterministic testing
func TestSchedulerWithMockClock(t *testing.T) {
	// This test demonstrates how to use the mock clock to make scheduler tests deterministic

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := schedulertesting.NewMockClock(startTime)

	// Verify initial time
	if !clock.Now().Equal(startTime) {
		t.Errorf("Expected clock to start at %v, got %v", startTime, clock.Now())
	}

	// Create a ticker
	ticker := clock.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Advance clock and verify ticker fires
	clock.Advance(1 * time.Minute)

	select {
	case tickTime := <-ticker.C():
		expectedTime := startTime.Add(1 * time.Minute)
		if tickTime.Before(expectedTime.Add(-1*time.Second)) || tickTime.After(expectedTime.Add(1*time.Second)) {
			t.Errorf("Tick time %v not close to expected %v", tickTime, expectedTime)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Ticker did not fire after clock advance")
	}
}

// TestSchedulerWithJobExecutionHooks demonstrates capturing job execution events
func TestSchedulerWithJobExecutionHooks(t *testing.T) {
	hook := schedulertesting.NewJobExecutionHook()

	// Simulate job execution (no real-time delay needed)
	hook.OnJobStart("sync-job", map[string]interface{}{"batch": 1})
	hook.OnJobEnd("sync-job", map[string]interface{}{"items": 100})

	hook.OnJobStart("sync-job", map[string]interface{}{"batch": 2})
	hook.OnJobError("sync-job", errors.New("network error"), map[string]interface{}{"items": 50})

	// Verify events were recorded
	events := hook.GetEvents()
	if len(events) != 4 {
		t.Errorf("Expected 4 events, got %d", len(events))
	}

	// Verify event order
	expectedTypes := []string{"start", "end", "start", "error"}
	for i, expected := range expectedTypes {
		if events[i].EventType != expected {
			t.Errorf("Event %d: expected type %s, got %s", i, expected, events[i].EventType)
		}
	}

	// Filter by type
	errorEvents := hook.GetEventsByType("error")
	if len(errorEvents) != 1 {
		t.Errorf("Expected 1 error event, got %d", len(errorEvents))
	}

	if errorEvents[0].Error == nil {
		t.Error("Error event should have an error")
	}
}

// TestSchedulerWithFaultInjection demonstrates error injection for retry testing
func TestSchedulerWithFaultInjection(t *testing.T) {
	injector := schedulertesting.NewFaultInjector()

	// Configure to fail the first 3 attempts
	injector.FailNTimes(3)

	attempts := 0
	maxAttempts := 5

	for attempts < maxAttempts {
		attempts++

		shouldFail, err := injector.ShouldFail()
		if shouldFail {
			t.Logf("Attempt %d failed: %v", attempts, err)
			continue
		}

		// Success
		t.Logf("Attempt %d succeeded", attempts)
		break
	}

	if attempts != 4 {
		t.Errorf("Expected success on attempt 4, got success on attempt %d", attempts)
	}
}

// TestSchedulerWithRetryTracking demonstrates tracking retry behavior
func TestSchedulerWithRetryTracking(t *testing.T) {
	tracker := schedulertesting.NewRetryTracker()
	config := schedulertesting.DefaultRetryConfig()
	calc := schedulertesting.NewBackoffCalculator(config)

	operationID := "webhook-delivery-123"
	maxAttempts := 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		backoff := calc.CalculateBackoff(attempt)

		// Simulate retry
		var err error
		if attempt < maxAttempts {
			err = errors.New("delivery failed")
		}

		tracker.RecordAttempt(operationID, attempt, err, backoff)

		if err != nil && calc.ShouldRetry(attempt) {
			t.Logf("Retry attempt %d with backoff %dms", attempt, backoff)
			time.Sleep(time.Duration(backoff) * time.Millisecond / 100) // Shortened for testing
		}
	}

	attempts := tracker.GetAttempts(operationID)
	if len(attempts) != maxAttempts {
		t.Errorf("Expected %d retry attempts, got %d", maxAttempts, len(attempts))
	}

	// Verify exponential backoff
	for i := 1; i < len(attempts); i++ {
		if attempts[i].Backoff <= attempts[i-1].Backoff {
			t.Errorf("Backoff should increase: attempt %d (%d) <= attempt %d (%d)",
				i+1, attempts[i].Backoff, i, attempts[i-1].Backoff)
		}
	}
}

// TestSchedulerWithWorkerPool demonstrates concurrent job execution testing
func TestSchedulerWithWorkerPool(t *testing.T) {
	numWorkers := 5
	pool := schedulertesting.NewWorkerPool(numWorkers)
	pool.Start()
	defer pool.Shutdown()

	numJobs := 20
	var mu sync.Mutex
	jobResults := make(map[string]bool)

	// Collect results in background
	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for result := range pool.Results() {
			mu.Lock()
			jobResults[result.JobID] = result.Success
			mu.Unlock()
		}
	}()

	// Submit jobs
	for i := 0; i < numJobs; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		job := schedulertesting.Job{
			ID: jobID,
			WorkFunc: func(ctx context.Context) error {
				// Simulate minimal CPU work
				sum := 0
				for j := 0; j < 1000; j++ {
					sum += j
				}
				return nil
			},
		}

		if !pool.Submit(job) {
			t.Errorf("Failed to submit job %s", jobID)
		}
	}

	// Wait for completion
	pool.Stop()
	resultsWg.Wait() // Wait for all results to be collected

	// Verify all jobs completed
	mu.Lock()
	defer mu.Unlock()

	if len(jobResults) != numJobs {
		t.Errorf("Expected %d job results, got %d", numJobs, len(jobResults))
	}

	for jobID, success := range jobResults {
		if !success {
			t.Errorf("Job %s failed", jobID)
		}
	}

	// Check stats
	stats := pool.GetStats()
	t.Logf("Worker pool stats: completed=%d, failed=%d, avg_duration=%v",
		stats.CompletedJobs, stats.FailedJobs, stats.AverageDuration)
}

// TestSchedulerWithQueueMonitoring demonstrates queue behavior monitoring
func TestSchedulerWithQueueMonitoring(t *testing.T) {
	monitor := schedulertesting.NewQueueMonitor(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start monitoring
	go monitor.StartSampling(ctx)

	// Simulate queue operations
	for i := 0; i < 100; i++ {
		monitor.RecordEnqueue()
		time.Sleep(2 * time.Millisecond)

		if i%3 == 0 {
			monitor.RecordDequeue()
		}
	}

	time.Sleep(100 * time.Millisecond) // Allow final samples
	monitor.Stop()

	samples := monitor.GetSamples()
	if len(samples) == 0 {
		t.Error("Expected queue samples to be recorded")
	}

	maxLength := monitor.GetMaxQueueLength()
	t.Logf("Max queue length observed: %d", maxLength)

	if maxLength == 0 {
		t.Error("Expected non-zero max queue length")
	}
}

// TestSchedulerWithConcurrencyTester demonstrates concurrency testing
func TestSchedulerWithConcurrencyTester(t *testing.T) {
	tester := schedulertesting.NewConcurrencyTester()

	// Test concurrent operations
	numGoroutines := 10
	errors := tester.ExecuteConcurrent(numGoroutines, func(threadID int) error {
		// Simulate operation
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	// Verify all succeeded
	for i, err := range errors {
		if err != nil {
			t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}

	ops := tester.GetOperations()
	if len(ops) != numGoroutines {
		t.Errorf("Expected %d operations, got %d", numGoroutines, len(ops))
	}
}

// TestSchedulerStressTest demonstrates stress testing
func TestSchedulerStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tester := schedulertesting.NewConcurrencyTester()

	testFunc := func() error {
		// Simulate work
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	result := tester.StressTest(100*time.Millisecond, 5, testFunc)

	t.Logf("Stress test results:")
	t.Logf("  Duration: %v", result.Duration)
	t.Logf("  Goroutines: %d", result.Goroutines)
	t.Logf("  Total operations: %d", result.TotalOps)
	t.Logf("  Success count: %d", result.SuccessCount)
	t.Logf("  Error count: %d", result.ErrorCount)
	t.Logf("  Ops/sec: %.2f", result.OpsPerSecond)

	if result.TotalOps == 0 {
		t.Error("Expected some operations to be performed")
	}

	// Verify all operations succeeded
	if result.ErrorCount > 0 {
		t.Errorf("Expected no errors, got %d", result.ErrorCount)
	}
}

// TestSchedulerWithMetrics demonstrates metrics collection
func TestSchedulerWithMetrics(t *testing.T) {
	metrics := schedulertesting.NewJobMetrics()

	// Simulate job executions
	metrics.RecordExecution(100*time.Millisecond, true, 50, 0)
	metrics.RecordExecution(150*time.Millisecond, true, 75, 0)
	metrics.RecordExecution(200*time.Millisecond, false, 30, 5)

	// Verify metrics
	if metrics.GetExecutionCount() != 3 {
		t.Errorf("Expected 3 executions, got %d", metrics.GetExecutionCount())
	}

	if metrics.GetSuccessCount() != 2 {
		t.Errorf("Expected 2 successes, got %d", metrics.GetSuccessCount())
	}

	if metrics.GetErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", metrics.GetErrorCount())
	}

	if metrics.GetItemsProcessed() != 155 {
		t.Errorf("Expected 155 items processed, got %d", metrics.GetItemsProcessed())
	}

	if metrics.GetItemsFailed() != 5 {
		t.Errorf("Expected 5 items failed, got %d", metrics.GetItemsFailed())
	}

	avgDuration := metrics.GetAverageDuration()
	expectedAvg := 150 * time.Millisecond
	if avgDuration != expectedAvg {
		t.Errorf("Expected average duration %v, got %v", expectedAvg, avgDuration)
	}
}
