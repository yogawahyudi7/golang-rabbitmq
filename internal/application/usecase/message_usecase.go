package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/application/dto"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/domain/entity"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/domain/service"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/pkg/logger"
)

type MessageUseCase interface {
	SendMessage(ctx context.Context, request dto.NewMessageRequestDTO) (*dto.MessageResponseDTO, error)
	GetMessage(ctx context.Context, id string) (*dto.MessageResponseDTO, error)
	GetAllMessages(ctx context.Context) ([]*dto.MessageResponseDTO, error)
	HandleIncomingMessage(ctx context.Context, messageData []byte) error
	PublishMessage(ctx context.Context, req dto.PublishMessageRequest) error
	GetMessageStats(ctx context.Context) (*dto.MessageStatsResponse, error)
}

type messageUseCase struct {
	service   service.MessageService
	publisher *rabbitmq.Publisher
	logger    logger.Logger
}

// NewMessageUseCase creates a new message use case
func NewMessageUseCase(service service.MessageService, publisher *rabbitmq.Publisher, logger logger.Logger) MessageUseCase {
	return &messageUseCase{
		service:   service,
		publisher: publisher,
		logger:    logger,
	}
}

// SendMessage sends a message to RabbitMQ
func (uc *messageUseCase) SendMessage(ctx context.Context, request dto.NewMessageRequestDTO) (*dto.MessageResponseDTO, error) {
	// Create message entity
	id := generateID()
	message := entity.NewMessage(id, request.Content, request.Priority)

	// Save message to repository
	if err := uc.service.CreateMessage(ctx, message); err != nil {
		return nil, err
	}

	// Convert to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	// Publish message to RabbitMQ
	if err := uc.publisher.Publish(ctx, messageJSON, "application/json"); err != nil {
		return nil, err
	}

	// Return response
	return mapMessageToDTO(message), nil
}

// GetMessage gets a message by ID
func (uc *messageUseCase) GetMessage(ctx context.Context, id string) (*dto.MessageResponseDTO, error) {
	message, err := uc.service.GetMessageByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapMessageToDTO(message), nil
}

// GetAllMessages gets all messages
func (uc *messageUseCase) GetAllMessages(ctx context.Context) ([]*dto.MessageResponseDTO, error) {
	messages, err := uc.service.GetAllMessages(ctx)
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	var result []*dto.MessageResponseDTO
	for _, message := range messages {
		result = append(result, mapMessageToDTO(message))
	}
	return result, nil
}

// HandleIncomingMessage handles incoming message from RabbitMQ
func (uc *messageUseCase) HandleIncomingMessage(ctx context.Context, messageData []byte) error {
	// Parse message
	var message entity.Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		return err
	}

	// Log received message
	uc.logger.WithFields(map[string]interface{}{
		"messageId": message.ID,
		"content":   message.Content,
		"priority":  message.Priority,
	}).Info("Received message for processing")

	// Process message
	if err := uc.service.ProcessMessage(ctx, &message); err != nil {
		return err
	}

	return nil
}

// PublishMessage publishes a message to RabbitMQ
func (uc *messageUseCase) PublishMessage(ctx context.Context, req dto.PublishMessageRequest) error {
	// Create message data
	messageData := map[string]interface{}{
		"message":    req.Message,
		"type":       req.Type,
		"headers":    req.Headers,
		"timestamp":  time.Now(),
		"message_id": generateID(),
	}

	// Convert to JSON
	messageJSON, err := json.Marshal(messageData)
	if err != nil {
		return err
	}

	// Publish message to RabbitMQ
	if err := uc.publisher.Publish(ctx, messageJSON, "application/json"); err != nil {
		return err
	}

	uc.logger.WithFields(map[string]interface{}{
		"message": req.Message,
		"type":    req.Type,
	}).Info("Message published successfully")

	return nil
}

// GetMessageStats returns message statistics
func (uc *messageUseCase) GetMessageStats(ctx context.Context) (*dto.MessageStatsResponse, error) {
	// Get pool stats from publisher
	poolStats := uc.publisher.GetPoolStats()

	// Create response
	response := &dto.MessageStatsResponse{
		TotalPublished: int64(poolStats["total_messages"].(int)),
		TotalConsumed:  0, // This would need to be tracked separately
		LastActivity:   time.Now(),
	}

	// Set queue status
	response.QueueStatus.Name = "message_queue"
	response.QueueStatus.Messages = poolStats["in_use"].(int)
	response.QueueStatus.Consumers = 1
	response.QueueStatus.MessageRate = 0  // This would need rate calculation
	response.QueueStatus.DeliveryRate = 0 // This would need rate calculation

	return response, nil
}

// Helper functions

// mapMessageToDTO maps a message entity to a message DTO
func mapMessageToDTO(message *entity.Message) *dto.MessageResponseDTO {
	return &dto.MessageResponseDTO{
		ID:        message.ID,
		Content:   message.Content,
		Priority:  message.Priority,
		Status:    message.Status,
		CreatedAt: message.CreatedAt.Format(time.RFC3339),
		UpdatedAt: message.UpdatedAt.Format(time.RFC3339),
	}
}

// generateID generates a unique ID for a message
func generateID() string {
	return "msg_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}
