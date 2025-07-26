package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/broker/rabbitmq"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/config"
	"github.com/yogawahyudi7/golang-rabbitmq/internal/infrastructure/server"
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

	// Initialize RabbitMQ publisher
	publisher, err := rabbitmq.NewPublisher(rabbitMQConn, cfg.RabbitMQ)
	if err != nil {
		log.WithError(err).Fatal("Could not create RabbitMQ publisher")
	}
	defer publisher.Close()

	// Set up HTTP server with router
	router := server.NewRouter(log, publisher)
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.WithField("port", cfg.Server.Port).Info("Starting API server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Could not start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}

	log.Info("Server exited gracefully")
}
