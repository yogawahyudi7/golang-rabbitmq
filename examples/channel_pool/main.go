package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
)

func main() {
	fmt.Println("🚀 Testing NEW Channel Pool Publisher Implementation")
	fmt.Println(strings.Repeat("=", 60))

	cfg := config.RabbitMQConfig{
		Host:         "localhost",
		Port:         "5672",
		Username:     "guest",
		Password:     "guest",
		VirtualHost:  "/",
		ExchangeName: "pool_test_exchange",
		QueueName:    "pool_test_queue",
		RoutingKey:   "pool_test_key",
		Durable:      true,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
		Mandatory:    false,
		Immediate:    false,
	}

	conn, err := rabbitmq.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	publisher, err := rabbitmq.NewPublisherWithPoolSize(conn, cfg, 15)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	fmt.Printf("✅ Channel Pool Publisher created with 15 channels\n\n")

	// Test concurrent publishing
	const numGoroutines = 20
	const messagesPerGoroutine = 5

	fmt.Printf("Testing %d concurrent goroutines, %d messages each...\n\n", numGoroutines, messagesPerGoroutine)

	startTime := time.Now()
	var wg sync.WaitGroup
	var successCount int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				message := fmt.Sprintf("Pool message G%d-M%d", goroutineID, j)

				err := publisher.Publish(ctx, []byte(message), "text/plain")
				if err != nil {
					fmt.Printf("❌ G%d-M%d: %v\n", goroutineID, j, err)
				} else {
					successCount++
					fmt.Printf("✅ G%d-M%d published\n", goroutineID, j)
				}
				cancel()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Printf("\n🎉 RESULTS:\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Success: %d messages\n", successCount)
	fmt.Printf("Throughput: %.2f msg/sec\n", float64(successCount)/duration.Seconds())

	stats := publisher.GetPoolStats()
	fmt.Printf("Pool stats: %+v\n", stats)
}
