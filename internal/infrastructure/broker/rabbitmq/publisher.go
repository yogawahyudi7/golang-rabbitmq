package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
)

// Publisher represents a RabbitMQ message publisher
type Publisher struct {
	conn       *Connection
	channel    *amqp.Channel
	config     config.RabbitMQConfig
	exchange   string
	routingKey string
	mandatory  bool
	immediate  bool
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(conn *Connection, config config.RabbitMQConfig) (*Publisher, error) {
	channel, err := conn.CreateChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
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

	return &Publisher{
		conn:       conn,
		channel:    channel,
		config:     config,
		exchange:   config.ExchangeName,
		routingKey: config.RoutingKey,
		mandatory:  config.Mandatory,
		immediate:  config.Immediate,
	}, nil
}

// Publish publishes a message to RabbitMQ
func (p *Publisher) Publish(ctx context.Context, body []byte, contentType string) error {
	// Create publishing message
	msg := amqp.Publishing{
		DeliveryMode:  amqp.Persistent,
		Timestamp:     time.Now(),
		ContentType:   contentType,
		Body:          body,
		Headers:       amqp.Table{},
		CorrelationId: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// Publish message
	err := p.channel.PublishWithContext(
		ctx,
		p.exchange,   // exchange
		p.routingKey, // routing key
		p.mandatory,  // mandatory
		p.immediate,  // immediate
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close closes the channel
func (p *Publisher) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}
