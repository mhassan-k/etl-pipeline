---
title: "Fix file timestamp collision risk in storage layer"
labels: bug, data-integrity, storage
---

## Description
File timestamps use second-level precision (`20060102_150405`), which can cause filename collisions if multiple ETL cycles run within the same second. This would result in data being appended to the same file instead of creating separate files, potentially corrupting the JSON structure.

## File Location
`internal/storage/storage.go:35, 69`

## Current Code
```go
// Only second precision
timestamp := time.Now().UTC().Format("20060102_150405")
filename := filepath.Join(rawPath, fmt.Sprintf("raw_data_%s.json", timestamp))

// File opened with O_APPEND flag
file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
```

## Problem Scenarios

### Scenario 1: Fast Fetch Interval
```bash
# If FETCH_INTERVAL < 1 second
FETCH_INTERVAL=0.5

# Both runs get same timestamp
# 2025-10-23 14:30:45.123 -> raw_data_20251023_143045.json
# 2025-10-23 14:30:45.789 -> raw_data_20251023_143045.json  # COLLISION!
```

### Scenario 2: Manual Triggers
```bash
# Quick successive manual triggers
curl http://localhost:8080/trigger  # 14:30:45.100
curl http://localhost:8080/trigger  # 14:30:45.500
# Both write to same file!
```

### Scenario 3: Clock Skew After NTP Sync
```bash
# NTP adjusts clock backward
# Writes could collide with previous files
```

## Data Corruption Example
If two writes happen to the same file:
```json
[{"id":1,"data":"first"}][{"id":2,"data":"second"}]
```
This is **invalid JSON** - should be two separate files.

## Impact
- **Data Corruption**: Invalid JSON files
- **Data Loss**: Overwritten or malformed data
- **Processing Errors**: Downstream systems can't parse files
- **Debugging Difficulty**: Hard to identify which records failed

## Proposed Fixes

### Fix 1: Millisecond Precision (Recommended)
```go
// Use millisecond precision (good for intervals >= 1ms)
timestamp := time.Now().UTC().Format("20060102_150405.000")
filename := filepath.Join(rawPath, fmt.Sprintf("raw_data_%s.json", timestamp))
```

### Fix 2: Nanosecond Unix Timestamp (Most Reliable)
```go
// Guaranteed unique per execution
timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
filename := filepath.Join(rawPath, fmt.Sprintf("raw_data_%s.json", timestamp))
```

### Fix 3: UUID/ULID (Best for Distributed Systems)
```go
import "github.com/google/uuid"

// Globally unique, sortable
id := uuid.New().String()
timestamp := time.Now().UTC().Format("20060102_150405")
filename := filepath.Join(rawPath, fmt.Sprintf("raw_data_%s_%s.json", timestamp, id))
```

### Fix 4: Sequential Counter (If Order Matters)
```go
type FileStorage struct {
    basePath string
    logger   *logging.Logger
    counter  atomic.Uint64  // Add counter
}

func (fs *FileStorage) SaveRawData(data []map[string]interface{}) error {
    count := fs.counter.Add(1)
    timestamp := time.Now().UTC().Format("20060102_150405")
    filename := filepath.Join(rawPath, fmt.Sprintf("raw_data_%s_%06d.json", timestamp, count))
    // ...
}
```

## Recommended Solution
Use **Fix 1 (Millisecond Precision)** because:
- Simple, minimal code change
- Human-readable filenames
- Sufficient for current use case (30s interval)
- No external dependencies
- Maintains chronological sorting

If you later add manual triggers or sub-second intervals, upgrade to Fix 2 or Fix 3.

## Implementation

Update both functions in `internal/storage/storage.go`:

```go
func (fs *FileStorage) SaveRawData(data []map[string]interface{}) error {
    rawPath := filepath.Join(fs.basePath, "raw")
    if err := os.MkdirAll(rawPath, 0755); err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to create raw data directory: %v", err))
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // Use millisecond precision to prevent collisions
    timestamp := time.Now().UTC().Format("20060102_150405.000")
    filename := filepath.Join(rawPath, fmt.Sprintf("raw_data_%s.json", timestamp))

    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to marshal raw data: %v", err))
        return fmt.Errorf("failed to marshal data: %w", err)
    }

    // Use O_CREATE | O_WRONLY | O_EXCL to fail if file exists
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
    if err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to create raw data file: %v", err))
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    if _, err := file.Write(jsonData); err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to write raw data: %v", err))
        return fmt.Errorf("failed to write data: %w", err)
    }

    fs.logger.Info(fmt.Sprintf("Raw data saved successfully: %s", filename))
    return nil
}

func (fs *FileStorage) SaveProcessedData(data interface{}) error {
    processedPath := filepath.Join(fs.basePath, "processed")
    if err := os.MkdirAll(processedPath, 0755); err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to create processed data directory: %v", err))
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // Use millisecond precision to prevent collisions
    timestamp := time.Now().UTC().Format("20060102_150405.000")
    filename := filepath.Join(processedPath, fmt.Sprintf("processed_data_%s.json", timestamp))

    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to marshal processed data: %v", err))
        return fmt.Errorf("failed to marshal data: %w", err)
    }

    // Use O_CREATE | O_WRONLY | O_EXCL to fail if file exists
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
    if err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to create processed data file: %v", err))
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    if _, err := file.Write(jsonData); err != nil {
        fs.logger.Error(fmt.Sprintf("Failed to write processed data: %v", err))
        return fmt.Errorf("failed to write data: %w", err)
    }

    fs.logger.Info(fmt.Sprintf("Processed data saved successfully: %s", filename))
    return nil
}
```

**Key Changes:**
1. Timestamp format: `"20060102_150405.000"` (added `.000` for milliseconds)
2. File flags: Changed from `O_APPEND` to `O_EXCL` (fails if file exists)
3. Operation: Write instead of Append (each file is new)

## Testing

Test collision detection:
```go
func TestSaveDataNoCollision(t *testing.T) {
    storage := NewFileStorage("/tmp/test", logger)

    // Rapid successive saves
    data1 := []map[string]interface{}{{"id": 1}}
    data2 := []map[string]interface{}{{"id": 2}}

    err1 := storage.SaveRawData(data1)
    err2 := storage.SaveRawData(data2)

    // Both should succeed
    if err1 != nil || err2 != nil {
        t.Error("Expected both saves to succeed")
    }

    // Should create two different files
    files, _ := filepath.Glob("/tmp/test/raw/*.json")
    if len(files) != 2 {
        t.Errorf("Expected 2 files, got %d", len(files))
    }
}
```

## Priority
**High** - Data integrity risk, though mitigated by current 30s interval

## Acceptance Criteria
- [ ] Timestamp format includes milliseconds
- [ ] File open flags use O_EXCL to prevent overwrites
- [ ] Unit tests verify no collisions on rapid saves
- [ ] Existing data files remain readable
- [ ] Tested with intervals < 1 second
- [ ] Documentation updated if file naming changes
