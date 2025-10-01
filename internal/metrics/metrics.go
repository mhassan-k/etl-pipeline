package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the ETL pipeline
type Metrics struct {
	APIRequestsTotal         prometheus.Counter
	APIRequestsFailedTotal   prometheus.Counter
	APIRequestDuration       prometheus.Histogram
	RecordsProcessedTotal    prometheus.Counter
	TransformationErrorTotal prometheus.Counter
	DataSavedTotal           prometheus.Counter
	DatabaseWritesTotal      prometheus.Counter
	DatabaseWriteErrorsTotal prometheus.Counter
}

// NewMetrics creates and registers all metrics
func NewMetrics() *Metrics {
	return &Metrics{
		APIRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "etl_api_requests_total",
			Help: "Total number of API requests made",
		}),
		APIRequestsFailedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "etl_api_requests_failed_total",
			Help: "Total number of failed API requests",
		}),
		APIRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "etl_api_request_duration_seconds",
			Help:    "Duration of API requests in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		RecordsProcessedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "etl_records_processed_total",
			Help: "Total number of records processed",
		}),
		TransformationErrorTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "etl_transformation_errors_total",
			Help: "Total number of transformation errors",
		}),
		DataSavedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "etl_data_saved_total",
			Help: "Total number of successful data saves",
		}),
		DatabaseWritesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "etl_database_writes_total",
			Help: "Total number of database write operations",
		}),
		DatabaseWriteErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "etl_database_write_errors_total",
			Help: "Total number of database write errors",
		}),
	}
}
