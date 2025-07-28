package rabbitmq

import (
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
)

// Connection represents a RabbitMQ connection with health monitoring
type Connection struct {
	conn     *amqp.Connection
	config   config.RabbitMQConfig
	url      string
	mu       sync.RWMutex
	isClosed bool
}

// NewConnection creates a new RabbitMQ connection with retry logic
func NewConnection(config config.RabbitMQConfig) (*Connection, error) {
	// Create connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.VirtualHost)

	// Connect to RabbitMQ with retry
	conn, err := connectWithRetry(url, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	return &Connection{
		conn:   conn,
		config: config,
		url:    url,
	}, nil
}

// connectWithRetry attempts to connect with retry logic
func connectWithRetry(url string, maxRetries int) (*amqp.Connection, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		if i < maxRetries {
			waitTime := time.Duration(1<<uint(i)) * time.Second
			time.Sleep(waitTime)
		}
	}

	return nil, lastErr
}

// GetConnection returns the underlying AMQP connection
func (c *Connection) GetConnection() *amqp.Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// CreateChannel creates a new channel with error handling
func (c *Connection) CreateChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed || c.conn.IsClosed() {
		return nil, fmt.Errorf("connection is closed")
	}

	return c.conn.Channel()
}

// IsHealthy checks if the connection is healthy
func (c *Connection) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return !c.isClosed && c.conn != nil && !c.conn.IsClosed()
}

// Reconnect attempts to reconnect to RabbitMQ
func (c *Connection) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		c.conn.Close()
	}

	conn, err := connectWithRetry(c.url, 3)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	c.conn = conn
	c.isClosed = false
	return nil
}

// Close closes the RabbitMQ connection
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.isClosed {
		c.isClosed = true
		return c.conn.Close()
	}
	return nil
}
