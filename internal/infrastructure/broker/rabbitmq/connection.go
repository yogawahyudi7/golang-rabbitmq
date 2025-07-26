package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
)

// Connection represents a RabbitMQ connection
type Connection struct {
	conn *amqp.Connection
}

// NewConnection creates a new RabbitMQ connection
func NewConnection(config config.RabbitMQConfig) (*Connection, error) {
	// Create connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.VirtualHost)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &Connection{conn: conn}, nil
}

// GetConnection returns the underlying AMQP connection
func (c *Connection) GetConnection() *amqp.Connection {
	return c.conn
}

// CreateChannel creates a new channel
func (c *Connection) CreateChannel() (*amqp.Channel, error) {
	return c.conn.Channel()
}

// Close closes the RabbitMQ connection
func (c *Connection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
