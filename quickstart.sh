#!/bin/bash

# ETL Pipeline Quick Start Script
# This script helps you get started with the ETL pipeline

set -e

echo "=========================================="
echo "ETL Pipeline Quick Start"
echo "=========================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://www.docker.com/products/docker-desktop/"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo "❌ Docker daemon is not running. Please start Docker Desktop."
    exit 1
fi

echo "✅ Docker is installed and running"
echo ""

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "⚠️  docker-compose not found, trying docker compose..."
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "Starting ETL Pipeline services..."
echo ""

# Start services
$DOCKER_COMPOSE up -d

echo ""
echo "=========================================="
echo "Services Started Successfully! 🚀"
echo "=========================================="
echo ""
echo "Access the following endpoints:"
echo "  Health Check: http://localhost:8080/health"
echo "  Readiness:    http://localhost:8080/ready"
echo "  Metrics:      http://localhost:8080/metrics"
echo ""
echo "View logs:"
echo "  docker-compose logs -f etl-pipeline"
echo ""
echo "Check database:"
echo "  docker exec -it etl-postgres psql -U etl_user -d etl_db"
echo ""
echo "Stop services:"
echo "  docker-compose down"
echo ""
echo "=========================================="

# Wait for services to be healthy
echo "Waiting for services to be healthy..."
sleep 5

# Test health endpoint
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Health check passed!"
    echo ""
    echo "Testing health endpoint:"
    curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/health
    echo ""
else
    echo "⚠️  Services are starting... Check logs with: docker-compose logs -f"
fi

echo ""
echo "Data will be stored in:"
echo "  - ./data/raw/       (raw API responses)"
echo "  - ./data/processed/ (transformed data)"
echo "  - ./logs/           (application logs)"
echo ""
