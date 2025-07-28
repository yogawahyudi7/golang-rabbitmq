package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
	"github.com/yogawahyudi7/golang-rabbitmq/pkg/logger"
)

// Consumer represents a RabbitMQ message consumer
type Consumer struct {
	conn              *Connection
	channel           *amqp.Channel
	config            config.RabbitMQConfig
	logger            logger.Logger
	queue             string
	mu                sync.RWMutex
	messageHandler    func([]byte) error
	errorHandler      func(error)
	reconnectInterval time.Duration
}

// NewConsumer creates a new RabbitMQ consumer
func NewConsumer(conn *Connection, config config.RabbitMQConfig, logger logger.Logger) (*Consumer, error) {
	channel, err := conn.CreateChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Set prefetch count
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		config.ExchangeName, // exchange name
		"direct",            // exchange type
		config.Durable,      // durable
		config.AutoDelete,   // auto-deleted
		false,               // internal
		config.NoWait,       // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	queue, err := channel.QueueDeclare(
		config.QueueName,  // queue name
		config.Durable,    // durable
		config.AutoDelete, // delete when unused
		config.Exclusive,  // exclusive
		config.NoWait,     // no-wait
		nil,               // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,          // queue name
		config.RoutingKey,   // routing key
		config.ExchangeName, // exchange
		config.NoWait,       // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &Consumer{
		conn:              conn,
		channel:           channel,
		config:            config,
		logger:            logger,
		queue:             config.QueueName,
		reconnectInterval: 5 * time.Second,
		errorHandler: func(err error) {
			logger.WithError(err).Error("Consumer error occurred")
		},
	}, nil
}

// SetMessageHandler sets the message processing function
func (c *Consumer) SetMessageHandler(handler func([]byte) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageHandler = handler
}

// SetErrorHandler sets the error handling function
func (c *Consumer) SetErrorHandler(handler func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorHandler = handler
}

// Start starts consuming messages from RabbitMQ with reconnection logic
func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer stopping due to context cancellation")
			return nil
		default:
			if err := c.consume(ctx); err != nil {
				c.logger.WithError(err).Error("Consumer connection failed, attempting to reconnect")

				select {
				case <-time.After(c.reconnectInterval):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

// consume handles the actual message consumption
func (c *Consumer) consume(ctx context.Context) error {
	// Create message delivery channel
	deliveries, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Process messages
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("consumer channel closed")
			}
			c.processMessage(delivery)
		}
	}
}

// processMessage handles a delivered message with proper error handling and retry logic
func (c *Consumer) processMessage(delivery amqp.Delivery) {
	startTime := time.Now()

	// Log received message
	c.logger.WithFields(map[string]interface{}{
		"messageId":     delivery.MessageId,
		"correlationId": delivery.CorrelationId,
		"contentType":   delivery.ContentType,
		"routingKey":    delivery.RoutingKey,
		"deliveryTag":   delivery.DeliveryTag,
		"redelivered":   delivery.Redelivered,
	}).Info("Processing message")

	// Use custom message handler if set, otherwise use default processing
	var err error
	c.mu.RLock()
	handler := c.messageHandler
	c.mu.RUnlock()

	if handler != nil {
		err = handler(delivery.Body)
	} else {
		err = c.defaultMessageHandler(delivery.Body)
	}

	// Handle processing result
	if err != nil {
		c.logger.WithError(err).WithFields(map[string]interface{}{
			"messageId":   delivery.MessageId,
			"redelivered": delivery.Redelivered,
		}).Error("Failed to process message")

		// Check if message has been redelivered too many times
		if delivery.Redelivered {
			// Send to dead letter queue or log for manual inspection
			c.logger.WithField("messageId", delivery.MessageId).Warn("Message redelivered, sending negative ack")
			if nackErr := delivery.Nack(false, false); nackErr != nil {
				c.logger.WithError(nackErr).Error("Failed to nack message")
			}
		} else {
			// Requeue for retry
			if nackErr := delivery.Nack(false, true); nackErr != nil {
				c.logger.WithError(nackErr).Error("Failed to nack and requeue message")
			}
		}

		// Call error handler if set
		c.mu.RLock()
		errorHandler := c.errorHandler
		c.mu.RUnlock()

		if errorHandler != nil {
			errorHandler(err)
		}
		return
	}

	// Acknowledge successful processing
	if err := delivery.Ack(false); err != nil {
		c.logger.WithError(err).Error("Failed to acknowledge message")
		return
	}

	// Log successful processing
	c.logger.WithFields(map[string]interface{}{
		"messageId":      delivery.MessageId,
		"processingTime": time.Since(startTime).String(),
	}).Info("Message processed successfully")
}

// defaultMessageHandler provides default message processing
func (c *Consumer) defaultMessageHandler(body []byte) error {
	// Default implementation: just log the message
	c.logger.WithField("body", string(body)).Info("Message received (default handler)")

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	return nil
}

// IsHealthy checks if the consumer connection is healthy
func (c *Consumer) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.channel != nil && !c.channel.IsClosed()
}

// Close closes the channel
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil && !c.channel.IsClosed() {
		return c.channel.Close()
	}
	return nil
}
