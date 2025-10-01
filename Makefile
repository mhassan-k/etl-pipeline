.PHONY: build run test clean docker-build docker-up docker-down help

# Build the Go binary
build:
	go build -o etl-pipeline .

# Run the application locally
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f etl-pipeline
	rm -rf data/ logs/

# Build Docker image
docker-build:
	docker-compose build

# Start all services with Docker Compose
docker-up:
	docker-compose up -d

# Stop all services
docker-down:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f etl-pipeline

# Restart services
docker-restart:
	docker-compose restart

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Help command
help:
	@echo "Available commands:"
	@echo "  make build         - Build the Go binary"
	@echo "  make run           - Run the application locally"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts and data"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start services with Docker Compose"
	@echo "  make docker-down   - Stop services"
	@echo "  make docker-logs   - View logs"
	@echo "  make docker-restart- Restart services"
	@echo "  make deps          - Install dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
