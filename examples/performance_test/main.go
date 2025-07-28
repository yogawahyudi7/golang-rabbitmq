package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
)

func main() {
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
		log.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	// Create publisher
	publisher, err := rabbitmq.NewPublisher(conn, cfg)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Test concurrent publishing with smart locking
	testConcurrentPublishing(publisher)
}

func testConcurrentPublishing(publisher *rabbitmq.Publisher) {
	fmt.Println("🚀 Testing Concurrent Publishing with Smart Locking...")

	const numGoroutines = 10
	const messagesPerGoroutine = 5

	var wg sync.WaitGroup
	startTime := time.Now()

	// Start multiple goroutines publishing concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				message := fmt.Sprintf("Message from goroutine %d, message %d", goroutineID, j)

				// Publish with smart locking - no blocking between goroutines!
				err := publisher.Publish(ctx, []byte(message), "text/plain")
				if err != nil {
					fmt.Printf("❌ Goroutine %d failed to publish message %d: %v\n", goroutineID, j, err)
				} else {
					fmt.Printf("✅ Goroutine %d published message %d successfully\n", goroutineID, j)
				}

				cancel()

				// Small delay to demonstrate concurrency
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	// Monitor active confirmations while publishing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cleanup

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return // ✅ Graceful exit
			case <-ticker.C:
				active := publisher.GetActiveConfirmations()
				deliveryTag := publisher.GetDeliveryTag()
				fmt.Printf("📊 Active confirmations: %d, Current delivery tag: %d\n", active, deliveryTag)
			}
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()
	cancel() // Stop monitoring goroutine

	duration := time.Since(startTime)
	totalMessages := numGoroutines * messagesPerGoroutine

	fmt.Printf("\n🎉 Completed publishing %d messages in %v\n", totalMessages, duration)
	fmt.Printf("📈 Average time per message: %v\n", duration/time.Duration(totalMessages))
	fmt.Printf("🔄 Final active confirmations: %d\n", publisher.GetActiveConfirmations())

	// Wait a bit to let remaining confirmations complete
	time.Sleep(2 * time.Second)
	fmt.Printf("🏁 Final active confirmations after waiting: %d\n", publisher.GetActiveConfirmations())
}
