package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
	"github.com/yogawahyudi7/golang-rabbitmq/pkg/logger"
)

// Consumer represents a RabbitMQ message consumer
type Consumer struct {
	conn    *Connection
	channel *amqp.Channel
	config  config.RabbitMQConfig
	logger  logger.Logger
	queue   string
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
		conn:    conn,
		channel: channel,
		config:  config,
		logger:  logger,
		queue:   config.QueueName,
	}, nil
}

// Start starts consuming messages from RabbitMQ
func (c *Consumer) Start(ctx context.Context) error {
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
			c.logger.Info("Consumer stopping due to context cancellation")
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.Error("Consumer channel closed unexpectedly")
				return fmt.Errorf("consumer channel closed")
			}
			c.processMessage(delivery)
		}
	}
}

// processMessage handles a delivered message
func (c *Consumer) processMessage(delivery amqp.Delivery) {
	startTime := time.Now()

	// Log received message
	c.logger.WithFields(map[string]interface{}{
		"messageId":     delivery.MessageId,
		"correlationId": delivery.CorrelationId,
		"contentType":   delivery.ContentType,
		"routingKey":    delivery.RoutingKey,
	}).Info("Processing message")

	// Process message (in a real application, you would implement your business logic here)
	// For this example, we're just logging the message
	c.logger.WithField("body", string(delivery.Body)).Info("Message received")

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Acknowledge message
	if err := delivery.Ack(false); err != nil {
		c.logger.WithError(err).Error("Failed to acknowledge message")
		return
	}

	// Log processing time
	c.logger.WithField("processingTime", time.Since(startTime).String()).Info("Message processed successfully")
}

// Close closes the channel
func (c *Consumer) Close() error {
	if c.channel != nil {
		return c.channel.Close()
	}
	return nil
}
