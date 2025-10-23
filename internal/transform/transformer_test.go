package transform

import (
	"testing"

	"github.com/mohammedhassan/etl-pipeline/internal/logging"
	"github.com/mohammedhassan/etl-pipeline/internal/metrics"
)

func TestTransformRecord(t *testing.T) {
	logger, _ := logging.NewLogger("test.log")
	defer logger.Close()
	
	metricsCollector := metrics.NewMetrics()
	transformer := NewTransformer(logger, metricsCollector)

	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
		expectedID  int
	}{
		{
			name: "Valid record",
			input: map[string]interface{}{
				"userId": float64(1),
				"title":  "Test Title",
				"body":   "Test Body",
			},
			expectError: false,
			expectedID:  1,
		},
		{
			name: "Missing userId",
			input: map[string]interface{}{
				"title": "Test Title",
				"body":  "Test Body",
			},
			expectError: true,
		},
		{
			name: "Empty title",
			input: map[string]interface{}{
				"userId": float64(1),
				"title":  "",
				"body":   "Test Body",
			},
			expectError: true,
		},
		{
			name: "Whitespace trimming",
			input: map[string]interface{}{
				"userId": float64(2),
				"title":  "  Title with spaces  ",
				"body":   "  Body with spaces  ",
			},
			expectError: false,
			expectedID:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.transformRecord(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.UserID != tt.expectedID {
					t.Errorf("Expected UserID %d, got %d", tt.expectedID, result.UserID)
				}
			}
		})
	}
}

func TestTransform(t *testing.T) {
	logger, _ := logging.NewLogger("test.log")
	defer logger.Close()
	
	metricsCollector := metrics.NewMetrics()
	transformer := NewTransformer(logger, metricsCollector)

	rawData := []map[string]interface{}{
		{
			"userId": float64(1),
			"title":  "First Post",
			"body":   "First Body",
		},
		{
			"userId": float64(2),
			"title":  "Second Post",
			"body":   "Second Body",
		},
		{
			// Invalid record - should be skipped
			"title": "No UserID",
			"body":  "Invalid",
		},
	}

	result, err := transformer.Transform(rawData)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should process 2 valid records
	if len(result.Records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(result.Records))
	}

	if result.TotalRecords != 2 {
		t.Errorf("Expected TotalRecords to be 2, got %d", result.TotalRecords)
	}

	// Verify first record
	if result.Records[0].UserID != 1 {
		t.Errorf("Expected first record UserID to be 1, got %d", result.Records[0].UserID)
	}
	if result.Records[0].Title != "First Post" {
		t.Errorf("Expected title 'First Post', got '%s'", result.Records[0].Title)
	}
}

func TestTransformEmptyData(t *testing.T) {
	logger, _ := logging.NewLogger("test.log")
	defer logger.Close()
	
	metricsCollector := metrics.NewMetrics()
	transformer := NewTransformer(logger, metricsCollector)

	rawData := []map[string]interface{}{}
	result, err := transformer.Transform(rawData)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result.Records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(result.Records))
	}
}
