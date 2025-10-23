---
title: "Add authentication to /metrics endpoint"
labels: security, monitoring, enhancement
---

## Description
The Prometheus `/metrics` endpoint is publicly accessible without authentication, potentially leaking sensitive system information that could be used for reconnaissance by attackers.

## File Location
`internal/server/server.go` - metrics handler

## Exposed Information
The endpoint currently exposes:
- Request counts and traffic patterns
- Error rates and failure patterns
- Database operation metrics
- API call patterns and latency
- System performance characteristics
- Internal service architecture

## Security Risk
- **Information Disclosure**: Attackers can study system behavior
- **Reconnaissance**: Learn about dependencies and architecture
- **Business Intelligence Leakage**: Traffic patterns reveal business metrics
- **Attack Surface**: Understanding error patterns aids exploit development

## Proposed Solutions

### Option 1: Basic Authentication (Recommended)
```go
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
    username, password, ok := r.BasicAuth()
    if !ok || !s.validateMetricsCredentials(username, password) {
        w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    promhttp.Handler().ServeHTTP(w, r)
}

func (s *Server) validateMetricsCredentials(username, password string) bool {
    expectedUsername := os.Getenv("METRICS_USERNAME")
    expectedPassword := os.Getenv("METRICS_PASSWORD")

    // Use constant-time comparison to prevent timing attacks
    usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(expectedUsername)) == 1
    passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1

    return usernameMatch && passwordMatch
}
```

### Option 2: Token-Based Authentication
```go
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")
    expectedToken := "Bearer " + os.Getenv("METRICS_TOKEN")

    if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    promhttp.Handler().ServeHTTP(w, r)
}
```

### Option 3: IP Allowlist (Network-Level)
If metrics are only scraped from internal network:
```go
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
    clientIP := r.RemoteAddr
    if !s.isAllowedIP(clientIP) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    promhttp.Handler().ServeHTTP(w, r)
}
```

### Option 4: Separate Internal Port
Expose metrics on different port that's not publicly accessible:
```go
// Main server on :8080
go s.server.ListenAndServe()

// Metrics server on :9090 (internal only)
metricsServer := &http.Server{
    Addr:    ":9090",
    Handler: promhttp.Handler(),
}
go metricsServer.ListenAndServe()
```

## Recommended Implementation
Use **Option 1 (Basic Authentication)** because:
- Prometheus supports basic auth natively
- Simple to implement and configure
- Works with existing Prometheus configurations
- Provides adequate security for internal metrics

## Configuration Changes

Add to `internal/config/config.go`:
```go
type Config struct {
    // ... existing fields
    MetricsUsername string
    MetricsPassword string
}

func LoadConfig() *Config {
    return &Config{
        // ... existing config
        MetricsUsername: getEnv("METRICS_USERNAME", ""),
        MetricsPassword: getEnv("METRICS_PASSWORD", ""),
    }
}
```

Add to `.env.example`:
```env
# Metrics Authentication
# Leave empty to disable authentication (not recommended for production)
METRICS_USERNAME=prometheus
METRICS_PASSWORD=secure_password_here
```

Update Prometheus configuration:
```yaml
scrape_configs:
  - job_name: 'etl-pipeline'
    basic_auth:
      username: prometheus
      password: secure_password_here
    static_configs:
      - targets: ['localhost:8080']
```

## Priority
**High** - Security vulnerability, but mitigated if running in private network

## Acceptance Criteria
- [ ] Metrics endpoint requires authentication
- [ ] Authentication credentials configurable via environment variables
- [ ] Uses constant-time comparison to prevent timing attacks
- [ ] Backwards compatible (optional, can be disabled)
- [ ] README updated with Prometheus configuration example
- [ ] `.env.example` includes metrics auth variables
- [ ] Tested with and without authentication enabled
- [ ] Prometheus can successfully scrape with auth configured
