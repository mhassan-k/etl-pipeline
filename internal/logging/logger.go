package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger handles application logging
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	file        *os.File
}

// NewLogger creates a new logger instance
func NewLogger(logPath string) (*Logger, error) {
	// Create logs directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		infoLogger:  log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(file, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		file:        file,
	}, nil
}

// Info logs an informational message
func (l *Logger) Info(message string) {
	l.infoLogger.Output(2, message)
	fmt.Printf("[%s] INFO: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.errorLogger.Output(2, message)
	fmt.Printf("[%s] ERROR: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.warnLogger.Output(2, message)
	fmt.Printf("[%s] WARN: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
