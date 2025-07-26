package dto

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
