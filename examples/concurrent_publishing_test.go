package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
)

// TestConcurrentPublishing tests concurrent publishing performance
func TestConcurrentPublishing(t *testing.T) {
	// Skip if RabbitMQ is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg := config.RabbitMQConfig{
		Host:         "localhost",
		Port:         "5672",
		Username:     "guest",
		Password:     "guest",
		VirtualHost:  "/",
		ExchangeName: "test_exchange",
		QueueName:    "test_queue",
		RoutingKey:   "test_key",
		Durable:      true,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
		Mandatory:    false,
		Immediate:    false,
	}

	// Create connection
	conn, err := rabbitmq.NewConnection(cfg)
	if err != nil {
		t.Skipf("Failed to create connection (RabbitMQ not available): %v", err)
	}
	defer conn.Close()

	// Create publisher with channel pool
	publisher, err := rabbitmq.NewPublisherWithPoolSize(conn, cfg, 10)
	if err != nil {
		t.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Test concurrent publishing
	t.Run("ChannelPool", func(t *testing.T) {
		testChannelPoolPerformance(t, publisher)
	})
}

func testChannelPoolPerformance(t *testing.T, publisher *rabbitmq.Publisher) {
	const (
		numGoroutines        = 20
		messagesPerGoroutine = 10
		totalMessages        = numGoroutines * messagesPerGoroutine
	)

	t.Logf("Testing with %d goroutines, %d messages each (%d total)",
		numGoroutines, messagesPerGoroutine, totalMessages)

	var wg sync.WaitGroup
	var successCount, errorCount int
	var mu sync.Mutex

	startTime := time.Now()

	// Launch concurrent publishers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				message := fmt.Sprintf("Test message from goroutine %d, message %d", goroutineID, j)

				err := publisher.Publish(ctx, []byte(message), "text/plain")

				mu.Lock()
				if err != nil {
					errorCount++
					t.Logf("❌ G%d-M%d failed: %v", goroutineID, j, err)
				} else {
					successCount++
				}
				mu.Unlock()

				cancel()
				time.Sleep(10 * time.Millisecond) // Small delay
			}
		}(i)
	}

	// Wait for completion
	wg.Wait()
	duration := time.Since(startTime)

	// Verify results
	t.Logf("✅ Results: %d successful, %d errors in %v", successCount, errorCount, duration)
	t.Logf("🚀 Throughput: %.2f messages/second", float64(successCount)/duration.Seconds())

	// Get final pool stats
	stats := publisher.GetPoolStats()
	t.Logf("📊 Final pool stats: %+v", stats)

	// Assertions
	if successCount == 0 {
		t.Fatal("No messages were published successfully")
	}

	if float64(errorCount)/float64(totalMessages) > 0.1 { // Allow max 10% error rate
		t.Fatalf("Too many errors: %d/%d (%.1f%%)", errorCount, totalMessages,
			float64(errorCount)/float64(totalMessages)*100)
	}

	expectedMinThroughput := 50.0 // messages per second
	actualThroughput := float64(successCount) / duration.Seconds()
	if actualThroughput < expectedMinThroughput {
		t.Fatalf("Throughput too low: %.2f < %.2f msg/sec", actualThroughput, expectedMinThroughput)
	}

	t.Logf("🎉 Test passed! Channel pool performing well.")
}

// BenchmarkChannelPoolPublishing benchmarks the channel pool publisher
func BenchmarkChannelPoolPublishing(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	cfg := config.RabbitMQConfig{
		Host:         "localhost",
		Port:         "5672",
		Username:     "guest",
		Password:     "guest",
		VirtualHost:  "/",
		ExchangeName: "benchmark_exchange",
		QueueName:    "benchmark_queue",
		RoutingKey:   "benchmark_key",
		Durable:      true,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
		Mandatory:    false,
		Immediate:    false,
	}

	conn, err := rabbitmq.NewConnection(cfg)
	if err != nil {
		b.Skipf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	publisher, err := rabbitmq.NewPublisherWithPoolSize(conn, cfg, 15)
	if err != nil {
		b.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	message := []byte("benchmark message")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := publisher.Publish(ctx, message, "text/plain")
			if err != nil {
				b.Errorf("Publish failed: %v", err)
			}
			cancel()
		}
	})
}
