package service

import (
	"context"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/domain/entity"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/domain/repository"
)

// MessageService defines the message service interface
type MessageService interface {
	CreateMessage(ctx context.Context, message *entity.Message) error
	GetMessageByID(ctx context.Context, id string) (*entity.Message, error)
	GetAllMessages(ctx context.Context) ([]*entity.Message, error)
	UpdateMessage(ctx context.Context, message *entity.Message) error
	DeleteMessage(ctx context.Context, id string) error
	ProcessMessage(ctx context.Context, message *entity.Message) error
}

type messageService struct {
	repo repository.MessageRepository
}

// NewMessageService creates a new message service
func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{
		repo: repo,
	}
}

// CreateMessage creates a new message
func (s *messageService) CreateMessage(ctx context.Context, message *entity.Message) error {
	if !message.Valid() {
		return ErrInvalidMessage
	}
	return s.repo.Save(ctx, message)
}

// GetMessageByID gets a message by ID
func (s *messageService) GetMessageByID(ctx context.Context, id string) (*entity.Message, error) {
	return s.repo.FindByID(ctx, id)
}

// GetAllMessages gets all messages
func (s *messageService) GetAllMessages(ctx context.Context) ([]*entity.Message, error) {
	return s.repo.FindAll(ctx)
}

// UpdateMessage updates a message
func (s *messageService) UpdateMessage(ctx context.Context, message *entity.Message) error {
	if !message.Valid() {
		return ErrInvalidMessage
	}
	return s.repo.Update(ctx, message)
}

// DeleteMessage deletes a message
func (s *messageService) DeleteMessage(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ProcessMessage processes a message (business logic)
func (s *messageService) ProcessMessage(ctx context.Context, message *entity.Message) error {
	// Set message as processing
	message.SetStatus("processing")
	if err := s.repo.Update(ctx, message); err != nil {
		return err
	}

	// Simulate processing
	// In a real-world application, you would implement your business logic here

	// Set message as completed
	message.SetStatus("completed")
	return s.repo.Update(ctx, message)
}

// Common errors
var (
	ErrInvalidMessage = NewError("invalid message")
)

// Error represents a domain error
type Error struct {
	message string
}

// NewError creates a new domain error
func NewError(message string) *Error {
	return &Error{
		message: message,
	}
}

// Error returns the error message
func (e *Error) Error() string {
	return e.message
}
