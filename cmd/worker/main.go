package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
	"github.com/yogawahyudi7/golang-rabbitmq/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()

	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.WithError(err).Fatal("Could not load config")
	}

	// Initialize RabbitMQ connection
	rabbitMQConn, err := rabbitmq.NewConnection(cfg.RabbitMQ)
	if err != nil {
		log.WithError(err).Fatal("Could not connect to RabbitMQ")
	}
	defer rabbitMQConn.Close()

	// Initialize RabbitMQ consumer
	consumer, err := rabbitmq.NewConsumer(rabbitMQConn, cfg.RabbitMQ, log)
	if err != nil {
		log.WithError(err).Fatal("Could not create RabbitMQ consumer")
	}
	defer consumer.Close()

	// Create a channel to listen for OS signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.WithError(err).Fatal("Failed to start consumer")
		}
	}()

	log.Info("Worker started successfully, waiting for messages")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	cancel()
	log.Info("Worker exited gracefully")
}
