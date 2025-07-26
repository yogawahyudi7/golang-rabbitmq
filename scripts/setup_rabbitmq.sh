#!/bin/bash

# This script helps set up RabbitMQ for local development

# Check if RabbitMQ is running via Docker
docker ps | grep rabbitmq > /dev/null
if [ $? -ne 0 ]; then
    echo "RabbitMQ container is not running. Starting it..."
    docker-compose up -d rabbitmq
    
    # Wait for RabbitMQ to start
    echo "Waiting for RabbitMQ to start..."
    sleep 10
else
    echo "RabbitMQ is already running."
fi

# Check if management plugin is enabled
PLUGINS=$(docker exec rabbitmq rabbitmq-plugins list -e)
if [[ ! $PLUGINS == *"rabbitmq_management"* ]]; then
    echo "Enabling management plugin..."
    docker exec rabbitmq rabbitmq-plugins enable rabbitmq_management
    
    # Restart RabbitMQ to apply changes
    docker restart rabbitmq
    
    # Wait for RabbitMQ to restart
    echo "Waiting for RabbitMQ to restart..."
    sleep 10
fi

# Create exchange and queue
echo "Setting up exchange and queue..."
docker exec rabbitmq rabbitmqadmin declare exchange name=message_exchange type=direct
docker exec rabbitmq rabbitmqadmin declare queue name=message_queue durable=true
docker exec rabbitmq rabbitmqadmin declare binding source=message_exchange destination=message_queue routing_key=message_key

echo "RabbitMQ setup completed."
echo "Management UI is available at http://localhost:15672"
echo "Username: guest"
echo "Password: guest"
