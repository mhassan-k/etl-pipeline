---
title: "Expand test coverage from 15% to 80%+"
labels: testing, quality, technical-debt, good-first-issue
---

## Description
Currently only the `transformer` package has tests (~15% coverage). All other 8 packages have 0% test coverage, making the codebase fragile and difficult to refactor safely.

## Current Test Status

```
?   github.com/mohammedhassan/etl-pipeline                    [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/api       [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/config    [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/database  [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/etl       [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/logging   [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/metrics   [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/server    [no test files]
?   github.com/mohammedhassan/etl-pipeline/internal/storage   [no test files]
ok    github.com/mohammedhassan/etl-pipeline/internal/transform  0.123s  coverage: 85.7% of statements
```

**Coverage:** 1/10 packages = ~15%

## Impact
- **No regression detection** - Changes can break existing functionality
- **Difficult refactoring** - Fear of breaking things prevents improvements
- **Low confidence** - Unclear if code works correctly
- **Slow debugging** - Manual testing required for every change
- **Production bugs** - Issues discovered by users instead of tests

## Testing Priority by Package

### 🔴 Critical (Must Test)

#### 1. Database Package (`internal/database/`)
**Priority:** Highest - handles data persistence

Tests needed:
- [ ] `TestNewPostgresDB` - Connection success/failure
- [ ] `TestInsertRawData` - Successful batch insert
- [ ] `TestInsertRawDataTransactionRollback` - Error handling
- [ ] `TestInsertProcessedData` - Record insertion
- [ ] `TestHealthCheck` - Connection health
- [ ] `TestInitSchema` - Schema creation (idempotent)
- [ ] `TestConnectionPool` - Pool limits respected
- [ ] `TestConcurrentWrites` - Thread safety

Example test:
```go
// internal/database/postgres_test.go
func TestInsertRawData(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    pgDB := &PostgresDB{db: db}

    mock.ExpectBegin()
    mock.ExpectPrepare("INSERT INTO raw_data")
    mock.ExpectExec("INSERT INTO raw_data").
        WithArgs(sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    data := []map[string]interface{}{{"test": "data"}}
    err = pgDB.InsertRawData(data)

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %v", err)
    }
}
```

#### 2. API Client Package (`internal/api/`)
**Priority:** High - external dependency, failure-prone

Tests needed:
- [ ] `TestFetchDataSuccess` - Successful API call
- [ ] `TestFetchDataTimeout` - Request timeout handling
- [ ] `TestFetchDataInvalidJSON` - Malformed response
- [ ] `TestFetchDataHTTPError` - 4xx/5xx status codes
- [ ] `TestFetchDataNetworkError` - Connection failures
- [ ] `TestMetricsIncremented` - Metrics updated correctly

Example test:
```go
// internal/api/client_test.go
func TestFetchDataSuccess(t *testing.T) {
    // Mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode([]map[string]interface{}{
            {"userId": 1, "title": "Test", "body": "Body"},
        })
    }))
    defer server.Close()

    logger, _ := logging.NewLogger("test.log")
    metrics := metrics.NewMetrics()
    client := NewClient(server.URL, logger, metrics)

    data, err := client.FetchData()

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if len(data) != 1 {
        t.Errorf("Expected 1 record, got %d", len(data))
    }
}
```

#### 3. ETL Service Package (`internal/etl/`)
**Priority:** High - core business logic

Tests needed:
- [ ] `TestRunPipelineSuccess` - Full pipeline flow
- [ ] `TestRunPipelineAPIFailure` - Graceful failure on API error
- [ ] `TestRunPipelineDBFailure` - Graceful failure on DB error
- [ ] `TestStartStop` - Context cancellation
- [ ] `TestMetricsUpdated` - All metrics incremented correctly

### 🟡 Important (Should Test)

#### 4. Storage Package (`internal/storage/`)
Tests needed:
- [ ] `TestSaveRawDataSuccess`
- [ ] `TestSaveProcessedDataSuccess`
- [ ] `TestSaveDataDirectoryCreation`
- [ ] `TestSaveDataPermissionError`
- [ ] `TestSaveDataDiskFull`
- [ ] `TestFilenamingNoCollision`

#### 5. Server Package (`internal/server/`)
Tests needed:
- [ ] `TestHealthEndpoint` - Returns healthy status
- [ ] `TestHealthEndpointDBDown` - Returns unhealthy when DB down
- [ ] `TestReadyEndpoint` - Readiness check
- [ ] `TestMetricsEndpoint` - Prometheus metrics served
- [ ] `TestGracefulShutdown` - Server stops cleanly

Example:
```go
func TestHealthEndpoint(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    server := NewServer(mockDB, mockLogger, mockMetrics, "8080")
    server.healthHandler(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }

    var response map[string]string
    json.NewDecoder(w.Body).Decode(&response)

    if response["status"] != "healthy" {
        t.Errorf("Expected healthy status, got %s", response["status"])
    }
}
```

### 🟢 Nice to Have

#### 6-8. Config, Logging, Metrics Packages
Lower priority as they're thin wrappers, but still valuable:
- [ ] `TestLoadConfig` - Environment variable parsing
- [ ] `TestLogger` - Log output formatting
- [ ] `TestMetricsRegistration` - Prometheus registration

## Test Implementation Plan

### Phase 1: Critical Path (Week 1)
1. Set up testing infrastructure
   - Install `github.com/DATA-DOG/go-sqlmock` for DB mocking
   - Install `github.com/stretchr/testify` for assertions
2. Add database tests (highest risk)
3. Add API client tests (external dependency)
4. Target: 40% coverage

### Phase 2: Core Logic (Week 2)
1. Add ETL service tests
2. Add storage tests
3. Add server tests
4. Target: 70% coverage

### Phase 3: Completeness (Week 3)
1. Add config tests
2. Add integration tests
3. Add edge case tests
4. Target: 80%+ coverage

## Testing Tools & Dependencies

Add to `go.mod`:
```go
require (
    // ... existing dependencies
    github.com/DATA-DOG/go-sqlmock v1.5.0
    github.com/stretchr/testify v1.8.4
)
```

## CI/CD Integration

Update `.github/workflows/ci.yml`:
```yaml
- name: Run tests with coverage
  run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 80.0" | bc -l) )); then
      echo "Coverage $coverage% is below 80% threshold"
      exit 1
    fi
```

## Testing Best Practices

1. **Table-Driven Tests** - Test multiple scenarios efficiently
2. **Mocking** - Use interfaces and mocks for external dependencies
3. **Test Isolation** - Each test should be independent
4. **Clear Names** - `TestFunctionName_Scenario_ExpectedBehavior`
5. **Fast Tests** - Unit tests should run in milliseconds
6. **Integration Tests** - Separate from unit tests

## Success Metrics

- [ ] Coverage badge in README: [![Coverage](https://codecov.io/gh/...)](...)
- [ ] CI fails if coverage drops below 80%
- [ ] All critical paths have tests
- [ ] Tests run in < 10 seconds
- [ ] No flaky tests (run 10x without failures)

## Priority
**Medium** - Critical for long-term maintainability

## Labels
- `testing` - Testing infrastructure
- `quality` - Code quality improvement
- `technical-debt` - Paying down debt
- `good-first-issue` - Some test files can be starter tasks

## Acceptance Criteria
- [ ] Test coverage >= 80%
- [ ] All packages have test files
- [ ] Critical paths fully tested (database, API, ETL)
- [ ] CI enforces coverage threshold
- [ ] Test documentation added to CONTRIBUTING.md
- [ ] Tests pass reliably (no flakiness)
- [ ] README updated with testing instructions
