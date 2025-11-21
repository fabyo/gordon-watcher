package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Message represents the structure of messages from gordon-watcher
type Message struct {
	Path      string    `json:"path"`
	Hash      string    `json:"hash"`
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
	Queue     string    `json:"queue"`
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *slog.Logger
}

func NewConsumer(rabbitURL string) (*Consumer, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS to process one message at a time
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, queueName string) error {
	// Declare queue (idempotent)
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (we'll ack manually)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info("Consumer started", "queue", queueName)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer stopping...")
			return nil

		case msg, ok := <-msgs:
			if !ok {
				c.logger.Warn("Channel closed")
				return fmt.Errorf("channel closed")
			}

			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.Error("Failed to process message",
					"error", err,
					"message_id", msg.MessageId)

				// Reject and requeue (or send to DLQ if configured)
				msg.Nack(false, false)
			} else {
				// Acknowledge successful processing
				msg.Ack(false)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, delivery amqp.Delivery) error {
	var msg Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	c.logger.Info("Processing file",
		"path", msg.Path,
		"hash", msg.Hash,
		"size", msg.Size,
		"queue", msg.Queue)

	// TODO: Implement your business logic here
	// Examples:
	// - Parse XML/JSON file
	// - Store in database
	// - Send to external API
	// - Transform and forward to another queue
	// - Generate reports
	// - Trigger workflows

	// Simulate processing
	time.Sleep(100 * time.Millisecond)

	c.logger.Info("File processed successfully",
		"path", msg.Path,
		"hash", msg.Hash)

	return nil
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func main() {
	// Get configuration from environment
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "xml"
	}

	// Create consumer
	consumer, err := NewConsumer(rabbitURL)
	if err != nil {
		slog.Error("Failed to create consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
	}()

	// Start consuming
	if err := consumer.Start(ctx, queueName); err != nil {
		slog.Error("Consumer error", "error", err)
		os.Exit(1)
	}

	slog.Info("Consumer stopped gracefully")
}
