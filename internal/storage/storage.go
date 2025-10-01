package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mohammedhassan/etl-pipeline/internal/logging"
)

// FileStorage handles file-based storage operations
type FileStorage struct {
	basePath string
	logger   *logging.Logger
}

// NewFileStorage creates a new file storage instance
func NewFileStorage(basePath string, logger *logging.Logger) *FileStorage {
	return &FileStorage{
		basePath: basePath,
		logger:   logger,
	}
}

// SaveRawData saves raw data to the file system
func (fs *FileStorage) SaveRawData(data []map[string]interface{}) error {
	rawPath := filepath.Join(fs.basePath, "raw")
	if err := os.MkdirAll(rawPath, 0755); err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to create raw data directory: %v", err))
		return fmt.Errorf("failed to create directory: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102_150405")
	filename := filepath.Join(rawPath, fmt.Sprintf("raw_data_%s.json", timestamp))

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to marshal raw data: %v", err))
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Append to file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to open raw data file: %v", err))
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(jsonData); err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to write raw data: %v", err))
		return fmt.Errorf("failed to write data: %w", err)
	}

	fs.logger.Info(fmt.Sprintf("Raw data saved successfully: %s", filename))
	return nil
}

// SaveProcessedData saves processed data to the file system
func (fs *FileStorage) SaveProcessedData(data interface{}) error {
	processedPath := filepath.Join(fs.basePath, "processed")
	if err := os.MkdirAll(processedPath, 0755); err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to create processed data directory: %v", err))
		return fmt.Errorf("failed to create directory: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102_150405")
	filename := filepath.Join(processedPath, fmt.Sprintf("processed_data_%s.json", timestamp))

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to marshal processed data: %v", err))
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Append to file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to open processed data file: %v", err))
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(jsonData); err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to write processed data: %v", err))
		return fmt.Errorf("failed to write data: %w", err)
	}

	fs.logger.Info(fmt.Sprintf("Processed data saved successfully: %s", filename))
	return nil
}
