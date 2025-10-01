package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohammedhassan/etl-pipeline/internal/api"
	"github.com/mohammedhassan/etl-pipeline/internal/config"
	"github.com/mohammedhassan/etl-pipeline/internal/database"
	"github.com/mohammedhassan/etl-pipeline/internal/etl"
	"github.com/mohammedhassan/etl-pipeline/internal/logging"
	"github.com/mohammedhassan/etl-pipeline/internal/metrics"
	"github.com/mohammedhassan/etl-pipeline/internal/server"
	"github.com/mohammedhassan/etl-pipeline/internal/storage"
	"github.com/mohammedhassan/etl-pipeline/internal/transform"
)

func main() {
	// Initialize logger
	logger, err := logging.NewLogger("logs/etl.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting ETL Pipeline Service...")

	// Load configuration
	cfg := config.LoadConfig()
	logger.Info(fmt.Sprintf("Configuration loaded: API=%s, Interval=%ds", cfg.APIURL, cfg.FetchInterval))

	// Initialize metrics
	metricsCollector := metrics.NewMetrics()
	
	// Initialize database
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL database")

	// Initialize storage
	fileStorage := storage.NewFileStorage("data", logger)

	// Initialize API client
	apiClient := api.NewClient(cfg.APIURL, logger, metricsCollector)

	// Initialize transformer
	transformer := transform.NewTransformer(logger, metricsCollector)

	// Initialize ETL service
	etlService := etl.NewETLService(
		apiClient,
		db,
		fileStorage,
		transformer,
		logger,
		metricsCollector,
	)

	// Start HTTP server for health and metrics
	srv := server.NewServer(cfg.ServerPort, db, logger, metricsCollector)
	go func() {
		logger.Info(fmt.Sprintf("Starting HTTP server on port %s", cfg.ServerPort))
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start ETL pipeline
	go etlService.Start(ctx, time.Duration(cfg.FetchInterval)*time.Second)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutdown signal received, stopping ETL pipeline...")
	cancel()

	// Graceful shutdown of HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(fmt.Sprintf("Server shutdown error: %v", err))
	}

	logger.Info("ETL Pipeline Service stopped gracefully")
}

