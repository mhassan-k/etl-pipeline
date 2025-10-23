---
title: "Prevent overlapping ETL cycles with mutex or atomic flag"
labels: bug, concurrency, race-condition, reliability
---

## Description
If an ETL pipeline cycle takes longer than the fetch interval (default 30s), the next cycle will start before the previous one completes. This can cause database connection contention, duplicate data insertion, and unpredictable behavior.

## File Location
`internal/etl/service.go:55-63`

## Current Code
```go
func (e *ETLService) Start(ctx context.Context, interval time.Duration) {
    e.logger.Info(fmt.Sprintf("ETL pipeline started with interval: %v", interval))

    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    e.runPipeline()  // Run immediately

    for {
        select {
        case <-ctx.Done():
            e.logger.Info("ETL pipeline stopped")
            return
        case <-ticker.C:
            e.runPipeline()  // ❌ Can run while previous cycle executing
        }
    }
}
```

## Problem Scenario

```
Time    Event
00:00   Cycle 1 starts
00:15   API call completes (slow network)
00:20   Database writes in progress
00:30   ⚠️ Cycle 2 starts (Cycle 1 still running!)
00:35   Cycle 1 completes
00:45   Database deadlock - both cycles writing
```

## Real-World Triggers
1. **Slow API responses** (network issues, API throttling)
2. **Database contention** (locks, slow queries)
3. **Large datasets** (processing takes longer than expected)
4. **Resource constraints** (CPU/memory pressure)
5. **Short interval** (configured < actual cycle time)

## Impact
- Database deadlocks and connection pool exhaustion
- Duplicate data insertion
- Race conditions in file writes
- Resource exhaustion (goroutines, memory)
- Unpredictable metric values
- Difficult-to-reproduce bugs

## Proposed Fix

### Solution 1: Atomic Boolean Flag (Recommended)
Simple and efficient for single-instance deployments:

```go
package etl

import (
    "context"
    "fmt"
    "sync/atomic"
    "time"
    // ... other imports
)

type ETLService struct {
    apiClient   *api.Client
    db          *database.PostgresDB
    storage     *storage.FileStorage
    transformer *transform.Transformer
    logger      *logging.Logger
    metrics     *metrics.Metrics
    isRunning   atomic.Bool  // Add atomic flag
}

func (e *ETLService) Start(ctx context.Context, interval time.Duration) {
    e.logger.Info(fmt.Sprintf("ETL pipeline started with interval: %v", interval))

    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    // Run immediately
    e.runPipelineWithLock()

    for {
        select {
        case <-ctx.Done():
            e.logger.Info("ETL pipeline stopped")
            return
        case <-ticker.C:
            e.runPipelineWithLock()
        }
    }
}

func (e *ETLService) runPipelineWithLock() {
    // Try to acquire lock
    if !e.isRunning.CompareAndSwap(false, true) {
        e.logger.Warn("Skipping ETL cycle - previous cycle still running")
        e.metrics.SkippedCyclesTotal.Inc()  // Track skipped cycles
        return
    }

    // Ensure lock is released even if panic occurs
    defer e.isRunning.Store(false)

    // Run the actual pipeline
    e.runPipeline()
}
```

### Solution 2: Mutex (Alternative)
Use if you need more complex locking logic:

```go
import "sync"

type ETLService struct {
    // ... existing fields
    mu sync.Mutex
}

func (e *ETLService) runPipelineWithLock() {
    if !e.mu.TryLock() {
        e.logger.Warn("Skipping ETL cycle - previous cycle still running")
        return
    }
    defer e.mu.Unlock()

    e.runPipeline()
}
```

### Solution 3: Channel-Based Semaphore
More Go-idiomatic:

```go
type ETLService struct {
    // ... existing fields
    sem chan struct{}
}

func NewETLService(...) *ETLService {
    return &ETLService{
        // ... existing initialization
        sem: make(chan struct{}, 1),  // Buffer of 1 = binary semaphore
    }
}

func (e *ETLService) runPipelineWithLock() {
    select {
    case e.sem <- struct{}{}:  // Try to acquire
        defer func() { <-e.sem }()  // Release
        e.runPipeline()
    default:  // Couldn't acquire
        e.logger.Warn("Skipping ETL cycle - previous cycle still running")
    }
}
```

## Additional Improvements

### Add Metrics for Skipped Cycles
Update `internal/metrics/metrics.go`:

```go
type Metrics struct {
    // ... existing metrics
    SkippedCyclesTotal prometheus.Counter
    CycleDuration     prometheus.Histogram
}

func NewMetrics() *Metrics {
    return &Metrics{
        // ... existing metrics
        SkippedCyclesTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "etl_skipped_cycles_total",
            Help: "Total number of ETL cycles skipped due to previous cycle still running",
        }),
        CycleDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "etl_cycle_duration_seconds",
            Help:    "Time taken for each ETL cycle to complete",
            Buckets: prometheus.DefBuckets,
        }),
    }
}
```

### Track Cycle Duration
In `runPipeline()`:

```go
func (e *ETLService) runPipeline() {
    e.logger.Info("========== Starting ETL Pipeline Cycle ==========")
    startTime := time.Now()

    // ... existing pipeline code ...

    duration := time.Since(startTime)
    e.metrics.CycleDuration.Observe(duration.Seconds())  // Track duration
    e.logger.Info(fmt.Sprintf("========== ETL Pipeline Cycle Completed in %.2fs ==========", duration.Seconds()))

    // Warn if cycle took longer than interval
    if duration > 30*time.Second {  // Or use configured interval
        e.logger.Warn(fmt.Sprintf("ETL cycle took %.2fs, longer than interval!", duration.Seconds()))
    }
}
```

## Testing

Add unit test to verify concurrency protection:

```go
// internal/etl/service_test.go
func TestNoConcurrentCycles(t *testing.T) {
    // Setup mock dependencies
    service := NewETLService(...)

    // Track concurrent executions
    var concurrentRuns atomic.Int32
    var maxConcurrent atomic.Int32

    // Wrap runPipeline to track concurrency
    originalRun := service.runPipeline
    service.runPipeline = func() {
        current := concurrentRuns.Add(1)
        defer concurrentRuns.Add(-1)

        // Track max
        for {
            max := maxConcurrent.Load()
            if current <= max || maxConcurrent.CompareAndSwap(max, current) {
                break
            }
        }

        time.Sleep(100 * time.Millisecond)  // Simulate work
        originalRun()
    }

    // Start service with short interval
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    go service.Start(ctx, 10*time.Millisecond)

    <-ctx.Done()

    // Verify no more than 1 concurrent execution
    if maxConcurrent.Load() > 1 {
        t.Errorf("Expected max 1 concurrent cycle, got %d", maxConcurrent.Load())
    }
}
```

## Monitoring

After deploying, monitor these metrics:
- `etl_skipped_cycles_total` - Should be 0 in normal operation
- `etl_cycle_duration_seconds` - Should be < fetch interval
- If skipped cycles > 0: Increase interval or optimize pipeline

## Configuration Recommendations

Update README to recommend:
```yaml
# Good: Interval > expected cycle time
FETCH_INTERVAL=30  # If cycles take ~10-15s

# Bad: Interval < expected cycle time
FETCH_INTERVAL=5   # If cycles take 10s = overlaps!
```

## Priority
**High** - Can cause data corruption and database issues

## Acceptance Criteria
- [ ] ETL service uses atomic flag or mutex to prevent overlaps
- [ ] Skipped cycles are logged with WARN level
- [ ] Metric added to track skipped cycles
- [ ] Metric added to track cycle duration
- [ ] Warning logged if cycle exceeds interval
- [ ] Unit test verifies no concurrent cycles
- [ ] README updated with interval configuration guidance
- [ ] Tested with interval < cycle duration
- [ ] No database deadlocks under load testing
