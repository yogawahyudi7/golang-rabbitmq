package repository

import (
	"context"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/domain/entity"
)

// MessageRepository defines the message repository interface
type MessageRepository interface {
	Save(ctx context.Context, message *entity.Message) error
	FindByID(ctx context.Context, id string) (*entity.Message, error)
	FindAll(ctx context.Context) ([]*entity.Message, error)
	Update(ctx context.Context, message *entity.Message) error
	Delete(ctx context.Context, id string) error
}
