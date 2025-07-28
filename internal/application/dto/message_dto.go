package dto

import "time"

// MessageDTO represents the message data transfer object
type MessageDTO struct {
	ID        string `json:"id,omitempty"`
	Content   string `json:"content"`
	Priority  int    `json:"priority"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// NewMessageRequestDTO creates a new message request DTO
type NewMessageRequestDTO struct {
	Content  string `json:"content" validate:"required"`
	Priority int    `json:"priority" validate:"min=0,max=10"`
}

// MessageResponseDTO represents the message response DTO
type MessageResponseDTO struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Priority  int    `json:"priority"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PublishMessageRequest represents the request to publish a message
type PublishMessageRequest struct {
	Message string            `json:"message" validate:"required"`
	Type    string            `json:"type,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// PublishMessageResponse represents the response after publishing a message
type PublishMessageResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	MessageID string `json:"message_id,omitempty"`
}

// MessageStatsResponse represents message statistics
type MessageStatsResponse struct {
	TotalPublished int64     `json:"total_published"`
	TotalConsumed  int64     `json:"total_consumed"`
	LastActivity   time.Time `json:"last_activity"`
	QueueStatus    struct {
		Name         string `json:"name"`
		Messages     int    `json:"messages"`
		Consumers    int    `json:"consumers"`
		MessageRate  int    `json:"message_rate"`
		DeliveryRate int    `json:"delivery_rate"`
	} `json:"queue_status"`
}
