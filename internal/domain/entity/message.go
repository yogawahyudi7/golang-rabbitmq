package entity

import (
	"time"
)

// Message represents a message entity in the domain
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Priority  int       `json:"priority"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewMessage creates a new message instance
func NewMessage(id string, content string, priority int) *Message {
	now := time.Now()
	return &Message{
		ID:        id,
		Content:   content,
		Priority:  priority,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Valid validates the message
func (m *Message) Valid() bool {
	return m.ID != "" && m.Content != ""
}

// SetStatus updates the message status
func (m *Message) SetStatus(status string) {
	m.Status = status
	m.UpdatedAt = time.Now()
}
