package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/application/dto"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/application/usecase"
	"github.com/yogawahyudi7/golang-rabbitmq/pkg/logger"
)

// MessageController handles message-related HTTP requests
type MessageController struct {
	messageUsecase usecase.MessageUseCase
	logger         logger.Logger
}

// NewMessageController creates a new message controller
func NewMessageController(messageUsecase usecase.MessageUseCase, logger logger.Logger) *MessageController {
	return &MessageController{
		messageUsecase: messageUsecase,
		logger:         logger,
	}
}

// PublishMessage handles POST /api/messages
func (c *MessageController) PublishMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req dto.PublishMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to decode request body")

		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Message == "" {
		c.logger.Error("Message is required")
		http.Error(w, `{"error": "Message is required"}`, http.StatusBadRequest)
		return
	}

	// Publish message
	if err := c.messageUsecase.PublishMessage(r.Context(), req); err != nil {
		c.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"message": req.Message,
		}).Error("Failed to publish message")

		http.Error(w, `{"error": "Failed to publish message"}`, http.StatusInternalServerError)
		return
	}

	// Return success response
	response := dto.PublishMessageResponse{
		Success: true,
		Message: "Message published successfully",
	}

	c.logger.WithFields(map[string]interface{}{
		"message": req.Message,
		"type":    req.Type,
	}).Info("Message published successfully")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetMessageStats handles GET /api/messages/stats
func (c *MessageController) GetMessageStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get message statistics
	stats, err := c.messageUsecase.GetMessageStats(r.Context())
	if err != nil {
		c.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to get message stats")

		http.Error(w, `{"error": "Failed to get message stats"}`, http.StatusInternalServerError)
		return
	}

	c.logger.Info("Message stats retrieved successfully")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}
