# ETL Pipeline - Data Engineering Project

A production-ready ETL (Extract, Transform, Load) pipeline built in Go that fetches data from a REST API, transforms it, and stores it in PostgreSQL and local file system with comprehensive observability.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)
- [Metrics & Monitoring](#metrics--monitoring)
- [Testing](#testing)
- [Productionization Strategy](#productionization-strategy)

---

## Features

### Core Requirements

1. **Data Extraction**
   - Fetches data from JSONPlaceholder API every 30 seconds (configurable)
   - Stores raw data in PostgreSQL database
   - Comprehensive error handling and retry logic
   - HTTP timeout and connection pooling

2. **Data Transformation**
   - Removes unnecessary fields
   - Normalizes data structure
   - Validates required fields (userId, title, body)
   - Converts timestamps to UTC ISO 8601 format
   - Trims whitespace and cleans data

3. **Data Storage**
   - **PostgreSQL**: Stores both raw and processed data in separate tables
   - **Local Files**: Saves data to `data/raw/` and `data/processed/`
   - Append-only writes (no data overwriting)
   - Timestamped files for easy tracking

4. **Logging & Observability**
   - Structured logging to both file (`logs/etl.log`) and console
   - Logs API request success/failure
   - Logs transformation errors
   - Logs successful data saves
   - Three log levels: INFO, WARN, ERROR

5. **Metrics Monitoring**
   - Prometheus metrics endpoint (`/metrics`)
   - Health check endpoint (`/health`)
   - Readiness endpoint (`/ready`)
   - 8 comprehensive metrics tracking pipeline health

6. **Containerization**
   - Multi-stage Dockerfile for optimized image size (~50MB)
   - docker-compose.yml for easy orchestration
   - Volume mounting for data and log persistence
   - Health checks configured
   - Automatic restart on failure

---

## Quick Start

### Prerequisites

- **Docker Desktop** installed and running
- **Go 1.21+** (optional, only needed for local development)

### Option 1: One-Command Start (Recommended)

```bash
./quickstart.sh
```

### Option 2: Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f etl-pipeline

# Stop services
docker-compose down
```


### Verify It's Working

```bash
# Check health
curl http://localhost:8080/health

# View metrics
curl http://localhost:8080/metrics | grep etl_

# Check logs
tail -f logs/etl.log

# View data files
ls -lh data/raw/
ls -lh data/processed/

# Check database
docker exec -it etl-postgres psql -U etl_user -d etl_db
SELECT COUNT(*) FROM raw_data;
SELECT COUNT(*) FROM processed_data;
```

---

## 🏗 Architecture

### System Flow

```mermaid
flowchart TB
    API[External API<br/>JSONPlaceholder]
    Client[API Client Go<br/>HTTP Request Handler]
    RawDB[(PostgreSQL<br/>raw_data table)]
    RawFS[/File System<br/>data/raw//]
    Transform[Transformer Go<br/>Data Processing]
    ProcDB[(PostgreSQL<br/>processed_data table)]
    ProcFS[/File System<br/>data/processed//]
    Metrics[Prometheus Metrics<br/>/metrics endpoint]
    Logs[Structured Logs<br/>logs/etl.log]
    
    API -->|Fetch every 30s| Client
    Client -->|Store Raw JSON| RawDB
    Client -->|Save Timestamped Files| RawFS
    RawDB -->|Read for Processing| Transform
    Transform -->|Validate & Normalize| ProcDB
    Transform -->|Save Processed JSON| ProcFS
    Client -.->|Track Operations| Metrics
    Transform -.->|Track Operations| Metrics
    Client -.->|Log Events| Logs
    Transform -.->|Log Events| Logs
    
    style API fill:#e1f5ff
    style Client fill:#b3e5fc
    style RawDB fill:#ffecb3
    style RawFS fill:#fff9c4
    style Transform fill:#c8e6c9
    style ProcDB fill:#ffecb3
    style ProcFS fill:#fff9c4
    style Metrics fill:#f8bbd0
    style Logs fill:#f8bbd0
```

### Component Overview

1. **main.go** - Application entry point, dependency injection, graceful shutdown
2. **internal/api/client.go** - HTTP client for data extraction with retry logic
3. **internal/database/postgres.go** - PostgreSQL operations with connection pooling
4. **internal/transform/transformer.go** - Data transformation and validation
5. **internal/storage/storage.go** - File system operations
6. **internal/etl/service.go** - Pipeline orchestration
7. **internal/server/server.go** - HTTP server with health and metrics endpoints
8. **internal/logging/logger.go** - Structured logging
9. **internal/metrics/metrics.go** - Prometheus metrics collection
10. **internal/config/config.go** - Configuration management

### Database Schema

**raw_data table:**
```sql
CREATE TABLE raw_data (
    id SERIAL PRIMARY KEY,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_raw_data_created_at ON raw_data(created_at);
```

**processed_data table:**
```sql
CREATE TABLE processed_data (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    title TEXT,
    body TEXT,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_processed_data_processed_at ON processed_data(processed_at);
CREATE INDEX idx_processed_data_user_id ON processed_data(user_id);
```

---

## 📁 Project Structure

```
etl-pipeline/
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── Dockerfile                   # Container build instructions
├── docker-compose.yml           # Multi-container orchestration
├── Makefile                     # Build automation commands
├── quickstart.sh                # Quick start script
├── .gitignore                   # Git exclusions
├── README.md                    # This file
│
├── internal/                    # Internal packages
│   ├── api/
│   │   └── client.go            # HTTP client for API calls
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── postgres.go          # PostgreSQL operations
│   ├── etl/
│   │   └── service.go           # ETL pipeline orchestration
│   ├── logging/
│   │   └── logger.go            # Structured logging
│   ├── metrics/
│   │   └── metrics.go           # Prometheus metrics
│   ├── server/
│   │   └── server.go            # HTTP server
│   ├── storage/
│   │   └── storage.go           # File storage operations
│   └── transform/
│       ├── transformer.go       # Data transformation logic
│       └── transformer_test.go  # Unit tests
│
├── .github/
│   └── workflows/
│       └── ci.yml               # CI/CD pipeline
│
├── data/                        # Data storage (created at runtime)
│   ├── raw/                     # Raw API responses
│   └── processed/               # Transformed data
│
└── logs/                        # Application logs (created at runtime)
    └── etl.log                  # Main log file
```

---

## ⚙️ Configuration

Configuration is managed through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `API_URL` | `https://jsonplaceholder.typicode.com/posts` | API endpoint to fetch data from |
| `DATABASE_URL` | `postgres://etl_user:etl_password@localhost:5432/etl_db?sslmode=disable` | PostgreSQL connection string |
| `FETCH_INTERVAL` | `30` | Seconds between API fetches |
| `SERVER_PORT` | `8080` | HTTP server port |

### Changing Configuration

**In docker-compose.yml:**
```yaml
environment:
  API_URL: https://api.example.com/data
  FETCH_INTERVAL: 60
  SERVER_PORT: 8080
```

**For local development:**
```bash
export API_URL=https://api.example.com/data
export FETCH_INTERVAL=60
go run main.go
```

---

## 🔌 API Endpoints

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "service": "etl-pipeline",
  "database": "healthy"
}
```

**Status Codes:**
- `200 OK` - All systems healthy
- `503 Service Unavailable` - Database unhealthy

### Readiness Check

**Endpoint:** `GET /ready`

**Response:**
```json
{
  "status": "ready",
  "service": "etl-pipeline"
}
```

**Use Case:** Kubernetes readiness probes

### Prometheus Metrics

**Endpoint:** `GET /metrics`

**Format:** Prometheus exposition format

**Example Output:**
```
# HELP etl_api_requests_total Total number of API requests made
# TYPE etl_api_requests_total counter
etl_api_requests_total 150

# HELP etl_records_processed_total Total number of records processed
# TYPE etl_records_processed_total counter
etl_records_processed_total 15000
```

---

## 📊 Metrics & Monitoring

### Available Metrics

| Metric Name | Type | Description | Use Case |
|------------|------|-------------|----------|
| `etl_api_requests_total` | Counter | Total API requests made | Track overall API usage |
| `etl_api_requests_failed_total` | Counter | Failed API requests | Alert on API issues |
| `etl_api_request_duration_seconds` | Histogram | API request latency | Monitor performance |
| `etl_records_processed_total` | Counter | Records processed | Track throughput |
| `etl_transformation_errors_total` | Counter | Transformation errors | Data quality monitoring |
| `etl_data_saved_total` | Counter | Successful data saves | Success rate tracking |
| `etl_database_writes_total` | Counter | Database write operations | Database load monitoring |
| `etl_database_write_errors_total` | Counter | Database write errors | Database health alerts |

### Monitoring Use Cases

1. **API Reliability**: Set alert if `etl_api_requests_failed_total` > 5% of total requests
2. **Performance**: Monitor P95 latency via `etl_api_request_duration_seconds`
3. **Data Quality**: Alert if `etl_transformation_errors_total` spikes
4. **System Health**: Alert if `etl_database_write_errors_total` > 0
5. **Capacity Planning**: Track `etl_records_processed_total` for throughput analysis

### Log Levels

- **INFO**: Normal operations (API requests, data saved, cycle completed)
- **WARN**: Non-critical issues (individual record transformation errors)
- **ERROR**: Critical failures (API failures, database errors, file write errors)

---

## 🧪 Testing

### Run Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./internal/transform/...
```

### Manual Testing

```bash
# 1. Start services
docker-compose up -d

# 2. Wait for first ETL cycle (30 seconds)
sleep 35

# 3. Test health endpoint
curl http://localhost:8080/health

# 4. Test metrics
curl http://localhost:8080/metrics | grep etl_

# 5. Verify data files exist
ls -lh data/raw/
ls -lh data/processed/

# 6. Check database
docker exec -it etl-postgres psql -U etl_user -d etl_db -c "SELECT COUNT(*) FROM raw_data;"
docker exec -it etl-postgres psql -U etl_user -d etl_db -c "SELECT COUNT(*) FROM processed_data;"

# 7. View logs
docker-compose logs etl-pipeline | tail -50
```

---

## 🚀 Productionization Strategy

### Cloud Architecture - Two AWS Deployment Options

#### **Option 1: AWS Managed Services** (Recommended for Quick Setup)

| Component | AWS Service | Cost/Month | Benefits |
|-----------|-------------|------------|----------|
| **Compute** | ECS Fargate | ~$50 | Serverless, no server management |
| **Database** | RDS PostgreSQL | ~$150 | Automated backups, Multi-AZ |
| **Storage** | S3 | ~$25 | Unlimited, lifecycle policies |
| **Monitoring** | CloudWatch | ~$30 | Native integration |
| **Total** | | **~$255** | Easy setup, fully managed |

---

#### **Option 2: Open Source Stack on AWS** ⭐ (Recommended for Cost & Flexibility)

| Component | Open Source Tool | Cost/Month | Benefits |
|-----------|------------------|------------|----------|
| **Compute** | EKS  | ~$70 | Kubernetes flexibility |
| **Database** | Self-hosted PostgreSQL | ~$30 | Full control, no limits |
| **Storage** | S3 + MinIO (optional) | ~$25 | S3-compatible, hybrid |
| **Monitoring** | Prometheus + Grafana + Loki | ~$20 | Complete observability |
| **Total** | | **~$145** | 43% cost savings |


#### Compute Layer (Managed Services Details)

**AWS ECS Fargate**
- Fully managed, serverless containers
- Auto-scaling: 2-10 tasks based on CPU/memory
- Task definition: 0.5 vCPU, 1GB memory
- Blue-green deployment for zero downtime

#### Data Storage

**Database: Amazon RDS PostgreSQL**
- Instance: `db.t3.medium` (2 vCPU, 4GB RAM)
- Multi-AZ deployment for high availability
- Automated backups (7-day retention)
- Read replicas for analytics workloads
- Storage: 100GB with auto-scaling enabled

**Data Lake: Amazon S3**
```
s3://etl-data-lake/
├── raw/year=2025/month=10/day=01/
│   └── raw_data_20251001_120000.json
└── processed/year=2025/month=10/day=01/
    └── processed_data_20251001_120000.json
```
- Lifecycle policies (move to Glacier after 90 days)
- Versioning enabled for audit trail
- Server-side encryption (KMS)
- S3 Intelligent-Tiering for cost optimization

#### Observability Stack

**Logging: Amazon CloudWatch Logs**
- Centralized log aggregation from all ECS tasks
- Log retention: 30 days
- CloudWatch Insights for log queries
- Integration with CloudWatch Alarms

**Metrics: Amazon Managed Prometheus + Grafana**
- Scrape metrics from `/metrics` endpoint
- Pre-built dashboards for ETL pipeline
- Custom alerts based on metric thresholds

**Tracing: AWS X-Ray**
- End-to-end request tracing
- Latency analysis and bottleneck identification
- Service map visualization

**Alerting: Amazon SNS + CloudWatch Alarms**
- API failure rate > 5% → Page on-call engineer
- Database connection errors → Alert Slack
- Transformation errors spike → Warning notification
- High memory usage → Auto-scaling trigger


#### CI/CD Pipeline

```
GitHub → AWS CodeBuild → Amazon ECR → ECS Fargate
```

**Pipeline Stages:**
1. **Source**: Trigger on push to `main` branch
2. **Build**: 
   - Run tests (`go test ./...`)
   - Build Docker image
   - Security scan (Trivy, Snyk)
3. **Push**: Push to Amazon ECR
4. **Deploy**:
   - Blue-green deployment to ECS
   - Health check validation
   - Automatic rollback on failure

#### Scalability Strategy

**Horizontal Scaling:**
- ECS Service Auto Scaling
- Scale-out triggers:
  - CPU utilization > 70%
  - Memory utilization > 80%
- Min: 2 tasks (high availability)
- Max: 10 tasks (cost control)

**Database Scaling:**
- Read replicas for analytics queries
- Connection pooling (25 connections configured)
- Table partitioning by date for large tables
- Vertical scaling (upgrade instance size) when needed

**Performance Optimization:**
- Connection pooling for database
- Batch processing for large datasets
- Caching layer (Redis/ElastiCache) if needed
- Compression for S3 storage

#### Reliability & Disaster Recovery

**High Availability:**
- Multi-AZ deployment for all components
- Minimum 2 ECS tasks across availability zones
- RDS Multi-AZ for automatic database failover
- Application Load Balancer for health checks


### Alternative Cloud Providers

#### Google Cloud Platform (GCP)

**Services:**
- Compute: Cloud Run (serverless) or GKE
- Database: Cloud SQL PostgreSQL
- Storage: Cloud Storage
- Monitoring: Cloud Logging + Cloud Monitoring

**Benefits:**
- Better BigQuery integration for analytics
- Cheaper egress costs
- Simpler pricing model

### Production Best Practices

1. **Data Quality**: Implement schema validation and data profiling
2. **Testing**: Unit tests (80%+ coverage), integration tests, load testing
3. **Documentation**: API docs, runbooks, architecture decision records
4. **Compliance**: GDPR/CCPA compliance, audit logging, data retention policies
5. **Performance**: Batch processing, connection pooling, caching strategies

---

## Development

### Prerequisites

```bash
# Install Go 1.21+
brew install go

# Install Docker
brew install --cask docker
```

### Local Development

```bash
# Clone repository
git clone <repo-url>
cd etl-pipeline

# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
go build -o etl-pipeline .

# Run locally
./etl-pipeline
```

### Make Commands

```bash
make build          # Build the Go binary
make run            # Run locally
make test           # Run tests
make docker-build   # Build Docker image
make docker-up      # Start with Docker Compose
make docker-down    # Stop services
make docker-logs    # View logs
make clean          # Clean artifacts
```

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

MIT License

---

## 👤 Author

Mohammed Hassan


**Built with Go, Docker, PostgreSQL** 🚀