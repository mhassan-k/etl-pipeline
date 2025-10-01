package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mohammedhassan/etl-pipeline/internal/database"
	"github.com/mohammedhassan/etl-pipeline/internal/logging"
	"github.com/mohammedhassan/etl-pipeline/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	port    string
	db      *database.PostgresDB
	logger  *logging.Logger
	metrics *metrics.Metrics
	server  *http.Server
}

// NewServer creates a new HTTP server
func NewServer(port string, db *database.PostgresDB, logger *logging.Logger, metrics *metrics.Metrics) *Server {
	return &Server{
		port:    port,
		db:      db,
		logger:  logger,
		metrics: metrics,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// Readiness check endpoint
	mux.HandleFunc("/ready", s.readyHandler)

	// Metrics endpoint (Prometheus)
	mux.Handle("/metrics", promhttp.Handler())

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "etl-pipeline",
	}

	// Check database health
	if err := s.db.HealthCheck(); err != nil {
		s.logger.Error(fmt.Sprintf("Health check failed: database unhealthy: %v", err))
		response["status"] = "unhealthy"
		response["database"] = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		response["database"] = "healthy"
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// readyHandler handles readiness check requests
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "ready",
		"service": "etl-pipeline",
	}

	// Check if database is accessible
	if err := s.db.HealthCheck(); err != nil {
		response["status"] = "not ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
