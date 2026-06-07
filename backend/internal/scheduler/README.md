# Scheduler Package

This package contains background job schedulers for the Clipper application. Schedulers run periodic tasks such as syncing clips, generating embeddings, refreshing scores, and retrying failed webhooks.

## Architecture

All schedulers follow a consistent pattern:

1. **Interface-based dependencies**: Schedulers depend on interfaces (e.g., `ClipSyncServiceInterface`, `EmbeddingServiceInterface`) rather than concrete implementations, enabling easy mocking in tests.

2. **Context-aware**: All schedulers respect context cancellation, allowing for graceful shutdown.

3. **Thread-safe Stop()**: Using `sync.Once` ensures that stopping a scheduler multiple times is safe and idempotent.

4. **Ticker-based scheduling**: Schedulers use `time.Ticker` for periodic execution at configurable intervals.

5. **Metrics integration**: Schedulers emit Prometheus metrics for monitoring job execution, duration, success/failure rates, and items processed.

## Available Schedulers

### ClipSyncScheduler
Periodically syncs trending clips from Twitch API.

**Configuration:**
- Interval: Configurable in minutes (default: production uses longer intervals)
- Sync window: Last 24 hours of trending clips
- Page limit: Fixed window to control API usage

**Metrics:**
- `job_execution_total{job="clip_sync",status="success|failed"}`
- `job_execution_duration_seconds{job="clip_sync"}`
- `job_last_success_timestamp_seconds{job="clip_sync"}`
- `job_items_processed{job="clip_sync",result="success|skipped|failed"}`

### EmbeddingScheduler
Generates embeddings for clips that don't have them yet.

**Configuration:**
- Interval: Configurable in minutes
- Batch size: 100 clips per run
- Age limit: Only processes clips from last 7 days

**Metrics:**
- `indexing_jobs_total{status="success|failed|partial"}`
- `indexing_job_duration_seconds`
- `clips_with_embeddings`
- `clips_without_embeddings`

### HotScoreScheduler
Refreshes hot scores for clips using a time-decay algorithm.

**Configuration:**
- Interval: Configurable in minutes (typically 5-15 minutes)

**Metrics:**
- `job_execution_total{job="hot_score_refresh",status="success|failed"}`
- `job_execution_duration_seconds{job="hot_score_refresh"}`
- `job_last_success_timestamp_seconds{job="hot_score_refresh"}`

### TrendingScoreScheduler
Updates trending scores for clips based on engagement metrics.

**Configuration:**
- Interval: Configurable in minutes (typically hourly)

**Metrics:**
- `job_execution_total{job="trending_score_refresh",status="success|failed"}`
- `job_execution_duration_seconds{job="trending_score_refresh"}`
- `job_last_success_timestamp_seconds{job="trending_score_refresh"}`

### WebhookRetryScheduler
Retries failed webhook deliveries with exponential backoff.

**Configuration:**
- Interval: Configurable in minutes
- Batch size: Number of webhook retries to process per run

**Logging prefix:** `[WEBHOOK_SCHEDULER]` for easy log filtering

### CDNScheduler
Ensures clip assets are synchronized to and refreshed in the CDN.

**Configuration:**
- Interval: Configurable in minutes or hours
- Max items per run: Limits the number of assets processed per execution

**Metrics:**
- Uses the standard scheduler metrics described above for job execution, duration, and last success time.

### DunningScheduler
Runs dunning workflows for users with failed or overdue payments.

**Configuration:**
- Interval: Configurable in minutes or hours
- Batch size: Number of accounts to process per run

**Metrics:**
- Uses the standard scheduler metrics described above for job execution, duration, and last success time.

### EmailMetricsScheduler
Collects and aggregates metrics related to outbound emails (opens, clicks, bounces).

**Configuration:**
- Interval: Configurable in minutes
- Lookback window: Time window for pulling fresh email events

**Metrics:**
- Uses the standard scheduler metrics described above, plus email-specific counters as implemented in the email service.

### ExportScheduler
Triggers generation and delivery of scheduled data exports.

**Configuration:**
- Interval: Configurable in minutes or hours
- Max exports per run: Limits the number of exports processed per execution

**Metrics:**
- Uses the standard scheduler metrics described above for job execution, duration, and last success time.

### LiveStatusScheduler
Refreshes live status information (e.g., whether channels or streams are currently live).

**Configuration:**
- Interval: Configurable in minutes (typically short to keep status fresh)
- Page size: Number of channels/status entries fetched per run

**Metrics:**
- Uses the standard scheduler metrics described above for job execution, duration, and last success time.

### MirrorScheduler
Mirrors or replicates clip data and metadata to secondary storage or services.

**Configuration:**
- Interval: Configurable in minutes or hours
- Max items per run: Limits the number of records mirrored per execution

**Metrics:**
- Uses the standard scheduler metrics described above for job execution, duration, and last success time.

### OutboundWebhookScheduler
Manages scheduled or batched outbound webhooks that are not covered by the retry scheduler.

**Configuration:**
- Interval: Configurable in minutes
- Batch size: Number of webhooks to send per run

**Metrics:**
- Uses the standard scheduler metrics described above for job execution, duration, and last success time.

### ReputationScheduler
Recalculates and updates reputation scores for users or channels based on recent activity.

**Configuration:**
- Interval: Configurable in minutes or hours
- Lookback window: Time range of activity considered per run

**Metrics:**
- Uses the standard scheduler metrics described above for job execution, duration, and last success time.

## Testing Framework

A comprehensive testing framework is available in the [`testing/`](./testing/) subdirectory to enhance scheduler test coverage, determinism, and performance validation.

### Features

The testing framework provides:

1. **Time Mocking** - Deterministic clock control for testing without real time delays
2. **Job Execution Hooks** - Capture and verify job execution events and outcomes
3. **Fault Injection** - Inject errors to test retry logic and backoff behavior
4. **Concurrency Testing** - Worker pools and queue monitoring for concurrent execution validation
5. **Performance Benchmarks** - Throughput and latency measurements with minimal overhead

### Quick Example

```go
import schedulertesting "git.subcult.tv/subculture-collective/clpr/internal/scheduler/testing"

func TestSchedulerWithFramework(t *testing.T) {
    // Use mock clock for deterministic timing
    clock := schedulertesting.NewMockClock(time.Now())
    
    // Set up execution hooks
    hook := schedulertesting.NewJobExecutionHook()
    
    // Configure fault injection
    injector := schedulertesting.NewFaultInjector()
    injector.FailNTimes(2) // First 2 attempts fail
    
    // Create scheduler and advance time
    scheduler := NewScheduler(mockService, clock)
    clock.Advance(1 * time.Minute)
    
    // Verify execution
    events := hook.GetEvents()
    assert.Equal(t, 3, len(events)) // 2 failures + 1 success
}
```

### Documentation

See [Testing Framework README](./testing/README.md) for complete documentation including:

- Clock interface usage
- Job execution hooks and metrics
- Fault injection patterns
- Concurrency testing utilities
- Performance benchmarking
- Best practices and examples

## Testing Patterns

### Mock Services
Each scheduler test uses mock implementations of the required service interfaces:

```go
type MockClipSyncService struct{}

func (m *MockClipSyncService) SyncTrendingClips(ctx context.Context, hours int, opts *services.TrendingSyncOptions) (*services.SyncStats, error) {
    return &services.SyncStats{
        ClipsFetched: 5,
        ClipsCreated: 5,
        StartTime:    time.Now(),
        EndTime:      time.Now().Add(time.Second),
    }, nil
}
```

### Test Categories

#### 1. Initialization Tests
Verify that schedulers are created with correct configuration:
```go
func TestNewClipSyncScheduler(t *testing.T) {
    mockService := &MockClipSyncService{}
    scheduler := NewClipSyncScheduler(mockService, 15)
    
    if scheduler.interval != 15*time.Minute {
        t.Errorf("interval = %v, want %v", scheduler.interval, 15*time.Minute)
    }
}
```

#### 2. Stop Safety Tests
Ensure `Stop()` can be called multiple times safely:
```go
func TestStopMultipleTimes(t *testing.T) {
    scheduler := NewClipSyncScheduler(mockService, 1)
    
    for i := 0; i < 10; i++ {
        scheduler.Stop() // Should not panic
    }
}
```

#### 3. Concurrent Stop Tests
Verify thread-safety with concurrent calls:
```go
func TestConcurrentStopCalls(t *testing.T) {
    scheduler := NewClipSyncScheduler(mockService, 1)
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            scheduler.Stop()
        }()
    }
    
    wg.Wait()
}
```

#### 4. Graceful Shutdown Tests
Test that `Start()` exits cleanly when `Stop()` is called:
```go
func TestStopWhileRunning(t *testing.T) {
    scheduler := NewClipSyncScheduler(mockService, 10)
    ctx := context.Background()
    
    done := make(chan bool)
    go func() {
        scheduler.Start(ctx)
        done <- true
    }()
    
    time.Sleep(100 * time.Millisecond) // Allow scheduler to start
    scheduler.Stop()
    
    select {
    case <-done:
        // Success
    case <-time.After(2 * time.Second):
        t.Fatal("Start() did not exit after Stop()")
    }
}
```

#### 5. Context Cancellation Tests
Verify schedulers respect context cancellation:
```go
func TestContextCancellation(t *testing.T) {
    scheduler := NewClipSyncScheduler(mockService, 10)
    ctx, cancel := context.WithCancel(context.Background())
    
    done := make(chan bool)
    go func() {
        scheduler.Start(ctx)
        done <- true
    }()
    
    time.Sleep(100 * time.Millisecond)
    cancel() // Cancel context
    
    select {
    case <-done:
        // Success
    case <-time.After(2 * time.Second):
        t.Fatal("Start() did not exit after context cancellation")
    }
}
```

### Test Best Practices

#### ✅ DO

1. **Use the testing framework** - Leverage `testing/` utilities for deterministic tests
2. **Use mock clock** - Replace time.Sleep with MockClock.Advance for instant tests
3. **Use mock services** - Don't rely on real databases or external APIs in unit tests
4. **Capture execution events** - Use JobExecutionHook to verify job behavior
5. **Test initialization** - Verify schedulers are configured correctly
6. **Test thread safety** - Use concurrent goroutines to test race conditions
7. **Use context timeouts** - Tests should have maximum execution times
8. **Test error handling** - Use FaultInjector to test retry logic and backoff
9. **Verify idempotency** - Multiple Stop() calls should be safe
10. **Run benchmarks** - Establish performance baselines with the testing framework

#### ❌ DON'T

1. **Don't use real time.Sleep** - Use MockClock.Advance for deterministic timing
2. **Don't use shared global state** - Each test should be isolated
3. **Don't test actual work logic** - That belongs in service tests; scheduler tests verify scheduling mechanics
4. **Don't forget cleanup** - Always defer cancel() on contexts and Stop() on schedulers
5. **Don't skip race detector** - Always run scheduler tests with `-race` flag
6. **Don't skip fault injection** - Error paths are critical for reliability

### Running Tests

```bash
# Run all scheduler tests
cd backend
go test ./internal/scheduler/

# Run with race detector (REQUIRED for scheduler tests)
go test -race ./internal/scheduler/

# Run specific test
go test -v -run TestStopWhileRunning ./internal/scheduler/

# Run tests multiple times to check for flakiness
go test -count=50 -race ./internal/scheduler/

# Run tests with coverage
go test -coverprofile=coverage.out ./internal/scheduler/
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./internal/scheduler/testing/

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./internal/scheduler/testing/
```

## Determinism and Stability

All scheduler tests are designed to be deterministic and pass consistently across multiple runs. Key strategies:

1. **Mock time-dependent operations** - Use fixed values in mocks instead of `time.Now()`
2. **Proper synchronization** - Use channels and wait groups instead of arbitrary sleeps
3. **Bounded test duration** - All tests complete in < 2 seconds with proper timeouts
4. **No external dependencies** - Tests don't depend on databases, Redis, or external APIs
5. **Clean state per test** - Each test creates fresh scheduler instances

### Verified Stability

Tests have been verified to pass:
- ✅ 50+ consecutive runs without failures
- ✅ With race detector enabled (`-race` flag)
- ✅ In parallel execution mode
- ✅ Across different Go versions (1.21, 1.22)

## Integration with CI

Scheduler tests are part of the CI pipeline:

```yaml
# .github/workflows/ci.yml
- name: Run tests with coverage
  run: |
    cd backend
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

The race detector ensures no data races exist, which is critical for production stability.

## Adding New Schedulers

When adding a new scheduler, follow this checklist:

### Implementation

- [ ] Define service interface for dependencies
- [ ] Implement scheduler struct with `stopChan` and `stopOnce`
- [ ] Use `time.Ticker` for periodic execution
- [ ] Implement thread-safe `Stop()` method using `sync.Once`
- [ ] Support context cancellation in `Start()` method
- [ ] Add Prometheus metrics for job execution
- [ ] Log important events (start, stop, errors)

### Testing

- [ ] Create mock implementation of service interface
- [ ] Test initialization with various intervals
- [ ] Test multiple Stop() calls (idempotency)
- [ ] Test concurrent Stop() calls (thread safety)
- [ ] Test Stop() while Start() is running
- [ ] Test context cancellation
- [ ] Test error handling in job execution with FaultInjector
- [ ] **Use the testing framework:**
  - [ ] Use MockClock for deterministic timing tests
  - [ ] Use JobExecutionHook to capture and verify job events
  - [ ] Use FaultInjector to test retry and backoff logic
  - [ ] Use WorkerPool if scheduler manages concurrent execution
  - [ ] Create benchmarks for throughput and latency
- [ ] Run tests with `-race` flag
- [ ] Verify tests pass 50+ consecutive runs
- [ ] Add performance benchmarks

### Documentation

- [ ] Add scheduler description to this README
- [ ] Document configuration options
- [ ] Document emitted metrics
- [ ] Add example test patterns if unique

## Troubleshooting

### Tests are flaky
1. Check for race conditions with `-race` flag
2. Verify proper synchronization (no arbitrary sleeps)
3. Ensure tests don't share global state
4. Check for proper cleanup (deferred cancel() calls)

### Scheduler doesn't stop
1. Verify `stopChan` is being closed in `Stop()`
2. Check that `Start()` select statement includes `<-s.stopChan` case
3. Ensure no blocking operations prevent loop from checking channels

### Memory leaks
1. Verify ticker is stopped with `defer ticker.Stop()`
2. Ensure contexts are cancelled properly
3. Check that goroutines exit when scheduler stops

## References

- [Go Context Package](https://pkg.go.dev/context)
- [Go Sync Package](https://pkg.go.dev/sync)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Testing Concurrent Code](https://go.dev/blog/race-detector)
