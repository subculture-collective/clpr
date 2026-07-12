package testing

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// FaultInjector provides utilities for injecting errors and faults in tests
type FaultInjector struct {
	mu              sync.RWMutex
	failureRate     float64 // 0.0 to 1.0
	failAfterN      int     // Fail after N successful calls
	failCount       int32   // Number of times to fail before succeeding
	callCount       int32   // Current call count
	customErrorFunc func(callNum int) error
}

// NewFaultInjector creates a new fault injector
func NewFaultInjector() *FaultInjector {
	return &FaultInjector{}
}

// SetFailureRate sets the probability of failure (0.0 to 1.0)
func (f *FaultInjector) SetFailureRate(rate float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failureRate = rate
}

// FailAfterN configures the injector to fail after N successful calls
func (f *FaultInjector) FailAfterN(n int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failAfterN = n
}

// FailNTimes configures the injector to fail the next N calls
func (f *FaultInjector) FailNTimes(n int) {
	if n < 0 {
		n = 0
	}
	if n > int(^uint32(0)>>1) {
		n = int(^uint32(0) >> 1)
	}
	// n is explicitly bounded to the int32 range above.
	// #nosec G115 -- bounded conversion is part of this test-only fault injector.
	atomic.StoreInt32(&f.failCount, int32(n))
}

// SetCustomErrorFunc sets a custom function to determine errors based on call number
func (f *FaultInjector) SetCustomErrorFunc(fn func(callNum int) error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.customErrorFunc = fn
}

// ShouldFail returns whether the current call should fail and the error to return
func (f *FaultInjector) ShouldFail() (bool, error) {
	callNum := int(atomic.AddInt32(&f.callCount, 1))

	// Check custom error function first
	f.mu.RLock()
	customFunc := f.customErrorFunc
	f.mu.RUnlock()

	if customFunc != nil {
		if err := customFunc(callNum); err != nil {
			return true, err
		}
	}

	// Check fail count
	if failCount := atomic.LoadInt32(&f.failCount); failCount > 0 {
		atomic.AddInt32(&f.failCount, -1)
		return true, errors.New("injected fault: fail count")
	}

	// Check fail after N - fail exactly once on the (N+1)th call
	f.mu.RLock()
	failAfterN := f.failAfterN
	f.mu.RUnlock()

	if failAfterN > 0 && callNum == failAfterN+1 {
		return true, errors.New("injected fault: fail after N")
	}

	// Check failure rate - use proper randomization
	f.mu.RLock()
	failureRate := f.failureRate
	f.mu.RUnlock()

	if failureRate > 0 {
		// Use hash-based pseudo-randomization for deterministic but varied patterns
		// The signed-to-unsigned wrap is intentional for Knuth's hash and only
		// controls deterministic test fault distribution.
		hash := uint32(callNum) * 2654435761 // #nosec G115 -- intentional hash wrap.
		if float64(hash%10000)/10000.0 < failureRate {
			return true, errors.New("injected fault: random failure")
		}
	}

	return false, nil
}

// Reset resets the fault injector state
func (f *FaultInjector) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.failureRate = 0
	f.failAfterN = 0
	atomic.StoreInt32(&f.failCount, 0)
	atomic.StoreInt32(&f.callCount, 0)
	f.customErrorFunc = nil
}

// GetCallCount returns the number of times ShouldFail has been called
func (f *FaultInjector) GetCallCount() int {
	return int(atomic.LoadInt32(&f.callCount))
}

// FaultInjectableService wraps a service with fault injection capabilities
type FaultInjectableService struct {
	injector *FaultInjector
	wrapped  interface{}
}

// NewFaultInjectableService creates a new service wrapper with fault injection
func NewFaultInjectableService(service interface{}) *FaultInjectableService {
	return &FaultInjectableService{
		injector: NewFaultInjector(),
		wrapped:  service,
	}
}

// GetInjector returns the fault injector for configuration
func (s *FaultInjectableService) GetInjector() *FaultInjector {
	return s.injector
}

// GetWrappedService returns the wrapped service
func (s *FaultInjectableService) GetWrappedService() interface{} {
	return s.wrapped
}

// RetryConfig defines retry behavior for testing
type RetryConfig struct {
	MaxAttempts     int
	InitialBackoff  int64 // milliseconds
	MaxBackoff      int64 // milliseconds
	BackoffMultiple float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  100,
		MaxBackoff:      5000,
		BackoffMultiple: 2.0,
	}
}

// BackoffCalculator calculates backoff durations for retry testing
type BackoffCalculator struct {
	config RetryConfig
}

// NewBackoffCalculator creates a new backoff calculator
func NewBackoffCalculator(config RetryConfig) *BackoffCalculator {
	return &BackoffCalculator{config: config}
}

// CalculateBackoff calculates the backoff duration for a given attempt number
func (b *BackoffCalculator) CalculateBackoff(attempt int) int64 {
	if attempt <= 0 {
		return 0
	}

	backoff := float64(b.config.InitialBackoff)
	for i := 1; i < attempt; i++ {
		backoff *= b.config.BackoffMultiple
		if backoff > float64(b.config.MaxBackoff) {
			backoff = float64(b.config.MaxBackoff)
			break
		}
	}

	return int64(backoff)
}

// ShouldRetry returns whether to retry based on the attempt number
func (b *BackoffCalculator) ShouldRetry(attempt int) bool {
	return attempt < b.config.MaxAttempts
}

// RetryTracker tracks retry attempts for testing
type RetryTracker struct {
	mu       sync.RWMutex
	attempts map[string][]RetryAttempt
}

// RetryAttempt represents a single retry attempt
type RetryAttempt struct {
	Attempt   int
	Error     error
	Backoff   int64
	Timestamp int64
}

// NewRetryTracker creates a new retry tracker
func NewRetryTracker() *RetryTracker {
	return &RetryTracker{
		attempts: make(map[string][]RetryAttempt),
	}
}

// RecordAttempt records a retry attempt
func (t *RetryTracker) RecordAttempt(operationID string, attempt int, err error, backoff int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.attempts[operationID]; !exists {
		t.attempts[operationID] = make([]RetryAttempt, 0)
	}

	t.attempts[operationID] = append(t.attempts[operationID], RetryAttempt{
		Attempt:   attempt,
		Error:     err,
		Backoff:   backoff,
		Timestamp: time.Now().UnixNano(),
	})
}

// GetAttempts returns all retry attempts for an operation
func (t *RetryTracker) GetAttempts(operationID string) []RetryAttempt {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if attempts, exists := t.attempts[operationID]; exists {
		result := make([]RetryAttempt, len(attempts))
		copy(result, attempts)
		return result
	}
	return nil
}

// GetAttemptCount returns the number of attempts for an operation
func (t *RetryTracker) GetAttemptCount(operationID string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if attempts, exists := t.attempts[operationID]; exists {
		return len(attempts)
	}
	return 0
}

// Clear clears all tracked attempts
func (t *RetryTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.attempts = make(map[string][]RetryAttempt)
}

// MockServiceWithFaults provides a mock service with configurable fault injection
type MockServiceWithFaults struct {
	injector *FaultInjector
	callFunc func(ctx context.Context) error
}

// NewMockServiceWithFaults creates a new mock service with fault injection
func NewMockServiceWithFaults() *MockServiceWithFaults {
	return &MockServiceWithFaults{
		injector: NewFaultInjector(),
		callFunc: func(ctx context.Context) error { return nil },
	}
}

// SetCallFunc sets the function to call when the service is invoked
func (m *MockServiceWithFaults) SetCallFunc(fn func(ctx context.Context) error) {
	m.callFunc = fn
}

// GetInjector returns the fault injector for configuration
func (m *MockServiceWithFaults) GetInjector() *FaultInjector {
	return m.injector
}

// Call invokes the service, potentially injecting faults
func (m *MockServiceWithFaults) Call(ctx context.Context) error {
	if shouldFail, err := m.injector.ShouldFail(); shouldFail {
		return err
	}

	if m.callFunc != nil {
		return m.callFunc(ctx)
	}

	return nil
}
