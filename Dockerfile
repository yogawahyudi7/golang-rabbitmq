FROM golang:1.21-alpine AS builder

ARG APP_NAME=api

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/${APP_NAME} ./cmd/${APP_NAME}/main.go

# Final stage
FROM alpine:latest

ARG APP_NAME=api

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/${APP_NAME} .

# Expose port
EXPOSE 8080

# Command to run
CMD ["/app/api"]
