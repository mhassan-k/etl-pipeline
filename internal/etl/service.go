package etl

import (
	"context"
	"fmt"
	"time"

	"github.com/mohammedhassan/etl-pipeline/internal/api"
	"github.com/mohammedhassan/etl-pipeline/internal/database"
	"github.com/mohammedhassan/etl-pipeline/internal/logging"
	"github.com/mohammedhassan/etl-pipeline/internal/metrics"
	"github.com/mohammedhassan/etl-pipeline/internal/storage"
	"github.com/mohammedhassan/etl-pipeline/internal/transform"
)

// ETLService orchestrates the ETL pipeline
type ETLService struct {
	apiClient   *api.Client
	db          *database.PostgresDB
	storage     *storage.FileStorage
	transformer *transform.Transformer
	logger      *logging.Logger
	metrics     *metrics.Metrics
}

// NewETLService creates a new ETL service
func NewETLService(
	apiClient *api.Client,
	db *database.PostgresDB,
	storage *storage.FileStorage,
	transformer *transform.Transformer,
	logger *logging.Logger,
	metrics *metrics.Metrics,
) *ETLService {
	return &ETLService{
		apiClient:   apiClient,
		db:          db,
		storage:     storage,
		transformer: transformer,
		logger:      logger,
		metrics:     metrics,
	}
}

// Start begins the ETL pipeline with the specified interval
func (e *ETLService) Start(ctx context.Context, interval time.Duration) {
	e.logger.Info(fmt.Sprintf("ETL pipeline started with interval: %v", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	e.runPipeline()

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("ETL pipeline stopped")
			return
		case <-ticker.C:
			e.runPipeline()
		}
	}
}

// runPipeline executes one iteration of the ETL pipeline
func (e *ETLService) runPipeline() {
	e.logger.Info("========== Starting ETL Pipeline Cycle ==========")
	startTime := time.Now()

	// 1. Extract: Fetch data from API
	rawData, err := e.apiClient.FetchData()
	if err != nil {
		e.logger.Error(fmt.Sprintf("Extraction failed: %v", err))
		return
	}

	// 2. Store raw data in database
	e.metrics.DatabaseWritesTotal.Inc()
	if err := e.db.InsertRawData(rawData); err != nil {
		e.metrics.DatabaseWriteErrorsTotal.Inc()
		e.logger.Error(fmt.Sprintf("Failed to insert raw data into database: %v", err))
		return
	}
	e.logger.Info(fmt.Sprintf("Raw data inserted into database: %d records", len(rawData)))

	// 3. Save raw data to file system
	if err := e.storage.SaveRawData(rawData); err != nil {
		e.logger.Error(fmt.Sprintf("Failed to save raw data to file: %v", err))
		// Continue even if file save fails
	} else {
		e.metrics.DataSavedTotal.Inc()
	}

	// 4. Transform: Process the data
	transformedData, err := e.transformer.Transform(rawData)
	if err != nil {
		e.logger.Error(fmt.Sprintf("Transformation failed: %v", err))
		return
	}

	// 5. Store processed data in database
	e.metrics.DatabaseWritesTotal.Inc()
	if err := e.db.InsertProcessedData(transformedData.Records); err != nil {
		e.metrics.DatabaseWriteErrorsTotal.Inc()
		e.logger.Error(fmt.Sprintf("Failed to insert processed data into database: %v", err))
		return
	}
	e.logger.Info(fmt.Sprintf("Processed data inserted into database: %d records", len(transformedData.Records)))

	// 6. Save processed data to file system
	if err := e.storage.SaveProcessedData(transformedData); err != nil {
		e.logger.Error(fmt.Sprintf("Failed to save processed data to file: %v", err))
		// Continue even if file save fails
	} else {
		e.metrics.DataSavedTotal.Inc()
	}

	duration := time.Since(startTime)
	e.logger.Info(fmt.Sprintf("========== ETL Pipeline Cycle Completed in %.2fs ==========", duration.Seconds()))
}
