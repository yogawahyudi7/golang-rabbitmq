.PHONY: build run test clean docker-up docker-down

# Build both API and Worker
build:
	go build -o bin/api ./cmd/api/main.go
	go build -o bin/worker ./cmd/worker/main.go

# Run API server
run-api:
	go run ./cmd/api/main.go

# Run Worker
run-worker:
	go run ./cmd/worker/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Start Docker containers
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Rebuild and restart Docker containers
docker-rebuild:
	docker-compose down
	docker-compose up --build -d

# Show logs
docker-logs:
	docker-compose logs -f

# Initialize the environment
init:
	cp .env.example .env
	go mod tidy

# Install dependencies
deps:
	go mod download

# Generate mock files for testing
mocks:
	mockgen -source=internal/domain/repository/message_repository.go -destination=internal/domain/repository/mock/message_repository_mock.go

# Check code quality
lint:
	golangci-lint run
