package server

import (
	"encoding/json"
	"net/http"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/pkg/logger"
)

// MessageRequest represents the structure of the message request
type MessageRequest struct {
	Content  string `json:"content"`
	Priority int    `json:"priority"`
}

// NewRouter creates a new HTTP router
func NewRouter(logger logger.Logger, publisher *rabbitmq.Publisher) http.Handler {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Send message to RabbitMQ
	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req MessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.WithError(err).Error("Failed to decode request body")
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Convert message to JSON
		msgJSON, err := json.Marshal(req)
		if err != nil {
			logger.WithError(err).Error("Failed to marshal message")
			http.Error(w, "Failed to process message", http.StatusInternalServerError)
			return
		}

		// Publish message to RabbitMQ
		if err := publisher.Publish(r.Context(), msgJSON, "application/json"); err != nil {
			logger.WithError(err).Error("Failed to publish message to RabbitMQ")
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Message sent to queue",
		})
	})

	// Wrap with logging middleware
	return loggingMiddleware(logger, mux)
}

// loggingMiddleware adds request logging
func loggingMiddleware(logger logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}).Info("Request received")

		next.ServeHTTP(w, r)
	})
}
