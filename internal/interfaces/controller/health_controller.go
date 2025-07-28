package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/pkg/logger"
)

// HealthController handles health check requests
type HealthController struct {
	logger     logger.Logger
	connection *rabbitmq.Connection
	publisher  *rabbitmq.Publisher
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
}

// NewHealthController creates a new health controller
func NewHealthController(logger logger.Logger, connection *rabbitmq.Connection, publisher *rabbitmq.Publisher) *HealthController {
	return &HealthController{
		logger:     logger,
		connection: connection,
		publisher:  publisher,
	}
}

// HealthCheck handles health check requests
func (h *HealthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check RabbitMQ connection health
	rabbitMQHealthy := h.connection.IsHealthy()
	publisherHealthy := h.publisher.IsHealthy()

	// Determine overall status
	status := "healthy"
	httpStatus := http.StatusOK

	if !rabbitMQHealthy || !publisherHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Services: map[string]interface{}{
			"rabbitmq": map[string]interface{}{
				"connection": rabbitMQHealthy,
				"publisher":  publisherHealthy,
			},
			"application": map[string]interface{}{
				"status": "running",
			},
		},
	}

	// Log health check
	h.logger.WithFields(map[string]interface{}{
		"status":            status,
		"rabbitmq_healthy":  rabbitMQHealthy,
		"publisher_healthy": publisherHealthy,
	}).Info("Health check performed")

	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)
}
