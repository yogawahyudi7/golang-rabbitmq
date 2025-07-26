package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	RabbitMQ RabbitMQConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	Host         string
	Port         string
	Username     string
	Password     string
	VirtualHost  string
	ExchangeName string
	QueueName    string
	RoutingKey   string
	Durable      bool
	AutoDelete   bool
	Exclusive    bool
	NoWait       bool
	Mandatory    bool
	Immediate    bool
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Initialize RabbitMQ configuration
	durable, _ := strconv.ParseBool(getEnv("RABBITMQ_QUEUE_DURABLE", "true"))
	autoDelete, _ := strconv.ParseBool(getEnv("RABBITMQ_QUEUE_AUTODELETE", "false"))
	exclusive, _ := strconv.ParseBool(getEnv("RABBITMQ_QUEUE_EXCLUSIVE", "false"))
	noWait, _ := strconv.ParseBool(getEnv("RABBITMQ_QUEUE_NOWAIT", "false"))
	mandatory, _ := strconv.ParseBool(getEnv("RABBITMQ_PUBLISH_MANDATORY", "false"))
	immediate, _ := strconv.ParseBool(getEnv("RABBITMQ_PUBLISH_IMMEDIATE", "false"))

	rabbitConfig := RabbitMQConfig{
		Host:         getEnv("RABBITMQ_HOST", "localhost"),
		Port:         getEnv("RABBITMQ_PORT", "5672"),
		Username:     getEnv("RABBITMQ_USERNAME", "guest"),
		Password:     getEnv("RABBITMQ_PASSWORD", "guest"),
		VirtualHost:  getEnv("RABBITMQ_VHOST", "/"),
		ExchangeName: getEnv("RABBITMQ_EXCHANGE", "message_exchange"),
		QueueName:    getEnv("RABBITMQ_QUEUE", "message_queue"),
		RoutingKey:   getEnv("RABBITMQ_ROUTING_KEY", "message_key"),
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		NoWait:       noWait,
		Mandatory:    mandatory,
		Immediate:    immediate,
	}

	// Initialize server configuration
	serverConfig := ServerConfig{
		Port: getEnv("SERVER_PORT", "8080"),
	}

	return &Config{
		Server:   serverConfig,
		RabbitMQ: rabbitConfig,
	}, nil
}

// getEnv gets environment variable or returns default value if not exists
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
