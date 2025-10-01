package transform

import (
	"fmt"
	"strings"
	"time"

	"github.com/mohammedhassan/etl-pipeline/internal/database"
	"github.com/mohammedhassan/etl-pipeline/internal/logging"
	"github.com/mohammedhassan/etl-pipeline/internal/metrics"
)

// Transformer handles data transformation operations
type Transformer struct {
	logger  *logging.Logger
	metrics *metrics.Metrics
}

// NewTransformer creates a new transformer instance
func NewTransformer(logger *logging.Logger, metrics *metrics.Metrics) *Transformer {
	return &Transformer{
		logger:  logger,
		metrics: metrics,
	}
}

// TransformedData represents the output of transformation
type TransformedData struct {
	Records       []database.ProcessedRecord `json:"records"`
	ProcessedAt   string                     `json:"processed_at"`
	TotalRecords  int                        `json:"total_records"`
	ProcessedByUTC string                    `json:"processed_by_utc"`
}

// Transform processes raw data and returns structured data
func (t *Transformer) Transform(rawData []map[string]interface{}) (*TransformedData, error) {
	t.logger.Info(fmt.Sprintf("Starting transformation of %d records", len(rawData)))

	var processedRecords []database.ProcessedRecord
	errorCount := 0

	for i, record := range rawData {
		transformed, err := t.transformRecord(record)
		if err != nil {
			t.metrics.TransformationErrorTotal.Inc()
			t.logger.Warn(fmt.Sprintf("Failed to transform record %d: %v", i, err))
			errorCount++
			continue
		}

		processedRecords = append(processedRecords, transformed)
		t.metrics.RecordsProcessedTotal.Inc()
	}

	if errorCount > 0 {
		t.logger.Warn(fmt.Sprintf("Transformation completed with %d errors", errorCount))
	} else {
		t.logger.Info(fmt.Sprintf("Transformation successful: %d records processed", len(processedRecords)))
	}

	return &TransformedData{
		Records:        processedRecords,
		ProcessedAt:    time.Now().UTC().Format(time.RFC3339),
		TotalRecords:   len(processedRecords),
		ProcessedByUTC: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// transformRecord transforms a single record
func (t *Transformer) transformRecord(record map[string]interface{}) (database.ProcessedRecord, error) {
	// Extract fields with type checking
	userID, ok := record["userId"].(float64)
	if !ok {
		return database.ProcessedRecord{}, fmt.Errorf("invalid or missing userId")
	}

	title, ok := record["title"].(string)
	if !ok {
		title = ""
	}

	body, ok := record["body"].(string)
	if !ok {
		body = ""
	}

	// Normalize data
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	// Validate required fields
	if title == "" {
		return database.ProcessedRecord{}, fmt.Errorf("title cannot be empty")
	}

	return database.ProcessedRecord{
		UserID: int(userID),
		Title:  title,
		Body:   body,
	}, nil
}
