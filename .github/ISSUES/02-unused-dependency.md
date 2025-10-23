---
title: "Remove unused dependency: github.com/robfig/cron/v3"
labels: maintenance, dependencies, cleanup
---

## Description
The `github.com/robfig/cron/v3` package is declared as a dependency but is never used in the codebase. The application uses `time.Ticker` instead for scheduling.

## File Location
`go.mod:8`

## Current Code
```go
require (
    github.com/lib/pq v1.10.9
    github.com/prometheus/client_golang v1.17.0
    github.com/robfig/cron/v3 v3.0.1  // ❌ NOT USED
)
```

## Impact
- Increases dependency attack surface
- Bloats vendor directory
- Confuses contributors about scheduling approach
- Unnecessary security scanning overhead

## Fix
```bash
go mod edit -droprequire github.com/robfig/cron/v3
go mod tidy
```

## Verification
After fix, verify the dependency is removed:
```bash
go mod graph | grep cron
# Should return nothing
```

## Priority
**High** - Should be cleaned up before next release
