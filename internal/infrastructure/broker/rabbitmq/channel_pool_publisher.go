package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
)

// ChannelPoolPublisher represents a RabbitMQ message publisher with channel pool
type ChannelPoolPublisher struct {
	conn        *Connection
	config      config.RabbitMQConfig
	channelPool chan *PooledChannel
	poolSize    int
	deliveryTag uint64 // Atomic counter
	mu          sync.RWMutex
	closed      bool
}

// PooledChannel wraps an AMQP channel with confirmation handling
type PooledChannel struct {
	channel   *amqp.Channel
	confirms  chan amqp.Confirmation
	publisher *ChannelPoolPublisher
}

// NewChannelPoolPublisher creates a new publisher with channel pool
func NewChannelPoolPublisher(conn *Connection, config config.RabbitMQConfig, poolSize int) (*ChannelPoolPublisher, error) {
	if poolSize <= 0 {
		poolSize = 10 // Default pool size
	}

	publisher := &ChannelPoolPublisher{
		conn:        conn,
		config:      config,
		channelPool: make(chan *PooledChannel, poolSize),
		poolSize:    poolSize,
		deliveryTag: 0,
	}

	// Initialize channel pool
	for i := 0; i < poolSize; i++ {
		pooledChan, err := publisher.createPooledChannel()
		if err != nil {
			// Cleanup already created channels
			publisher.Close()
			return nil, fmt.Errorf("failed to create pooled channel %d: %w", i, err)
		}
		publisher.channelPool <- pooledChan
	}

	return publisher, nil
}

// createPooledChannel creates a new pooled channel with setup
func (p *ChannelPoolPublisher) createPooledChannel() (*PooledChannel, error) {
	channel, err := p.conn.CreateChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Enable publisher confirms
	if err := channel.Confirm(false); err != nil {
		channel.Close()
		return nil, fmt.Errorf("failed to put channel in confirm mode: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		p.config.ExchangeName,
		"direct",
		p.config.Durable,
		p.config.AutoDelete,
		false,
		p.config.NoWait,
		nil,
	)
	if err != nil {
		channel.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	queue, err := channel.QueueDeclare(
		p.config.QueueName,
		p.config.Durable,
		p.config.AutoDelete,
		p.config.Exclusive,
		p.config.NoWait,
		nil,
	)
	if err != nil {
		channel.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,
		p.config.RoutingKey,
		p.config.ExchangeName,
		p.config.NoWait,
		nil,
	)
	if err != nil {
		channel.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Set up confirmation notifications
	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	return &PooledChannel{
		channel:   channel,
		confirms:  confirms,
		publisher: p,
	}, nil
}

// getChannel gets a channel from the pool
func (p *ChannelPoolPublisher) getChannel(ctx context.Context) (*PooledChannel, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("publisher is closed")
	}
	p.mu.RUnlock()

	select {
	case pooledChan := <-p.channelPool:
		// Check if channel is still healthy
		if pooledChan.channel.IsClosed() {
			// Recreate channel
			newChan, err := p.createPooledChannel()
			if err != nil {
				return nil, fmt.Errorf("failed to recreate channel: %w", err)
			}
			return newChan, nil
		}
		return pooledChan, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("failed to get channel from pool: %w", ctx.Err())
	}
}

// returnChannel returns a channel to the pool
func (p *ChannelPoolPublisher) returnChannel(pooledChan *PooledChannel) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		// Publisher is closed, close the channel
		pooledChan.channel.Close()
		return
	}

	select {
	case p.channelPool <- pooledChan:
		// Successfully returned to pool
	default:
		// Pool is full, close the channel
		pooledChan.channel.Close()
	}
}

// getNextDeliveryTag returns the next delivery tag atomically
func (p *ChannelPoolPublisher) getNextDeliveryTag() uint64 {
	return atomic.AddUint64(&p.deliveryTag, 1)
}

// Publish publishes a message using channel pool (NO LOCKS!)
func (p *ChannelPoolPublisher) Publish(ctx context.Context, body []byte, contentType string) error {
	// Get channel from pool
	pooledChan, err := p.getChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer p.returnChannel(pooledChan)

	// Create message
	msg := amqp.Publishing{
		DeliveryMode:  amqp.Persistent,
		Timestamp:     time.Now(),
		ContentType:   contentType,
		Body:          body,
		Headers:       amqp.Table{},
		CorrelationId: fmt.Sprintf("%d", time.Now().UnixNano()),
		MessageId:     fmt.Sprintf("msg_%d", time.Now().UnixNano()),
	}

	// Publish message (NO LOCKS NEEDED!)
	err = pooledChan.channel.PublishWithContext(
		ctx,
		p.config.ExchangeName,
		p.config.RoutingKey,
		p.config.Mandatory,
		p.config.Immediate,
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation on this specific channel
	select {
	case confirmation := <-pooledChan.confirms:
		if !confirmation.Ack {
			return fmt.Errorf("message nack'd by broker")
		}
		// Increment delivery tag for statistics tracking
		p.getNextDeliveryTag()
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("confirmation timeout")
	case <-ctx.Done():
		return fmt.Errorf("confirmation cancelled: %w", ctx.Err())
	}
}

// PublishWithRetry publishes a message with retry mechanism
func (p *ChannelPoolPublisher) PublishWithRetry(ctx context.Context, body []byte, contentType string, maxRetries int) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		err := p.Publish(ctx, body, contentType)
		if err == nil {
			return nil
		}

		lastErr = err

		if i < maxRetries {
			waitTime := time.Duration(1<<uint(i)) * time.Second
			select {
			case <-time.After(waitTime):
				continue
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("failed to publish after %d retries: %w", maxRetries, lastErr)
}

// PublishBatch publishes multiple messages concurrently using channel pool
func (p *ChannelPoolPublisher) PublishBatch(ctx context.Context, messages [][]byte, contentType string) error {
	// Use worker pool pattern for batch publishing
	const maxWorkers = 10

	workers := len(messages)
	if workers > maxWorkers {
		workers = maxWorkers
	}

	jobs := make(chan int, len(messages))
	results := make(chan error, len(messages))

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for msgIndex := range jobs {
				err := p.Publish(ctx, messages[msgIndex], contentType)
				results <- err
			}
		}()
	}

	// Send jobs
	for i := range messages {
		jobs <- i
	}
	close(jobs)

	// Collect results
	var errors []error
	for i := 0; i < len(messages); i++ {
		if err := <-results; err != nil {
			errors = append(errors, fmt.Errorf("message %d: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch publish failed with %d errors: %v", len(errors), errors)
	}

	return nil
}

// IsHealthy checks if the publisher is healthy
func (p *ChannelPoolPublisher) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return !p.closed && len(p.channelPool) > 0
}

// GetPoolStats returns pool statistics
func (p *ChannelPoolPublisher) GetPoolStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"pool_size":      p.poolSize,
		"available":      len(p.channelPool),
		"total_messages": atomic.LoadUint64(&p.deliveryTag),
		"closed":         p.closed,
	}
}

// Close closes all channels in the pool
func (p *ChannelPoolPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// Close all channels in pool
	close(p.channelPool)
	for pooledChan := range p.channelPool {
		pooledChan.channel.Close()
	}

	return nil
}
