package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mohammedhassan/etl-pipeline/internal/logging"
	"github.com/mohammedhassan/etl-pipeline/internal/metrics"
)

// Client represents an API client for data extraction
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *logging.Logger
	metrics    *metrics.Metrics
}

// NewClient creates a new API client
func NewClient(baseURL string, logger *logging.Logger, metrics *metrics.Metrics) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:  logger,
		metrics: metrics,
	}
}

// FetchData fetches data from the API
func (c *Client) FetchData() ([]map[string]interface{}, error) {
	start := time.Now()
	c.metrics.APIRequestsTotal.Inc()

	c.logger.Info(fmt.Sprintf("Fetching data from API: %s", c.baseURL))

	resp, err := c.httpClient.Get(c.baseURL)
	if err != nil {
		c.metrics.APIRequestsFailedTotal.Inc()
		c.logger.Error(fmt.Sprintf("API request failed: %v", err))
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start).Seconds()
	c.metrics.APIRequestDuration.Observe(duration)

	if resp.StatusCode != http.StatusOK {
		c.metrics.APIRequestsFailedTotal.Inc()
		c.logger.Error(fmt.Sprintf("API returned non-200 status: %d", resp.StatusCode))
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.metrics.APIRequestsFailedTotal.Inc()
		c.logger.Error(fmt.Sprintf("Failed to read response body: %v", err))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.metrics.APIRequestsFailedTotal.Inc()
		c.logger.Error(fmt.Sprintf("Failed to parse JSON response: %v", err))
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	c.logger.Info(fmt.Sprintf("API request successful: fetched %d records in %.2fs", len(data), duration))
	return data, nil
}
