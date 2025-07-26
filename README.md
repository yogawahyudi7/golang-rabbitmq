# Golang RabbitMQ Clean Architecture

This project demonstrates a Go application with RabbitMQ message broker, following clean architecture principles.

## Project Structure

```
golang-rabbitmq/
├── cmd/
│   ├── api/           # API server entry point
│   └── worker/        # Worker service entry point
├── internal/
│   ├── domain/        # Business domain layer (entities, repositories interfaces)
│   │   ├── entity/
│   │   ├── repository/
│   │   └── service/
│   ├── application/   # Application use cases and DTOs
│   │   ├── dto/
│   │   └── usecase/
│   ├── interfaces/    # Interface adapters (controllers, presenters)
│   │   ├── controller/
│   │   └── presenter/
│   └── infrastructure/# External frameworks & tools
│       ├── broker/
│       │   └── rabbitmq/ # RabbitMQ implementation
│       ├── config/       # App configuration
│       ├── persistence/  # Database implementations
│       └── server/       # HTTP server setup
├── pkg/                 # Shared packages
│   └── logger/          # Logrus logger implementation
└── scripts/             # Helper scripts
```

## Features

- Clean Architecture design
- RabbitMQ message broker with publisher/consumer pattern
- Structured logging using logrus
- Graceful shutdown
- Dockerized development environment
- RabbitMQ Management UI for monitoring

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- RabbitMQ

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/yogawahyudi7/golang-rabbitmq.git
   cd golang-rabbitmq
   ```
   
2. Initialize the environment:
   ```
   make init
   ```
   
3. Start RabbitMQ with Docker:
   ```
   docker-compose up -d rabbitmq
   ```
   
4. Run the API server:
   ```
   go run cmd/api/main.go
   ```
   
5. Run the worker in a separate terminal:
   ```
   go run cmd/worker/main.go
   ```
   
6. Access the RabbitMQ Management UI at http://localhost:15672
   - Username: guest
   - Password: guest

## Development

### Running the API Server

```bash
go run cmd/api/main.go
```

### Running the Worker

```bash
go run cmd/worker/main.go
```

### Building the Application

```bash
# Build API server
go build -o bin/api ./cmd/api

# Build worker
go build -o bin/worker ./cmd/worker
```

### Testing

```bash
go test -v ./...
```

### Docker Development Environment

You can also run the complete stack with Docker Compose:

```bash
docker-compose up -d
```

## API Endpoints

- `GET /health` - Health check endpoint
- `POST /messages` - Publish a message to RabbitMQ

Example message request:
```json
{
  "content": "Test message",
  "priority": 1
}
```

## Monitoring

Access the RabbitMQ Management UI at http://localhost:15672 to monitor:
- Queue status
- Message rates
- Consumer status
- Exchange bindings

## Architecture Overview

The application follows the Clean Architecture pattern with the following layers:

1. **Domain Layer** - Contains business entities and business rules
   - Entities represent core business objects
   - Repository interfaces define data access contracts
   - Domain services implement business logic

2. **Application Layer** - Implements use cases
   - Use cases orchestrate the flow of data to and from entities
   - DTOs define data transfer objects for external communication

3. **Interface Layer** - Adapters for UI and external interfaces
   - Controllers handle HTTP requests
   - Presenters format data for responses

4. **Infrastructure Layer** - Technical implementations
   - RabbitMQ for message broker
   - Configuration management
   - Database implementations

## License

MIT
