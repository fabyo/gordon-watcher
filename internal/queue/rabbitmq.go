package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/fabyo/gordon-watcher/internal/logger"
)

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL        string
	Exchange   string
	QueueName  string
	RoutingKey string
	Durable    bool

	// DLQ Configuration
	DLQEnabled  bool
	DLQExchange string
	DLQQueue    string
}

// RabbitMQQueue implements Queue interface for RabbitMQ
type RabbitMQQueue struct {
	cfg    RabbitMQConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
	logger *logger.Logger
}

// NewRabbitMQQueue creates a new RabbitMQ queue
func NewRabbitMQQueue(cfg RabbitMQConfig, log *logger.Logger) (*RabbitMQQueue, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Open channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		cfg.Durable,  // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Setup DLQ if enabled
	if cfg.DLQEnabled {
		// Declare DLQ exchange
		err = ch.ExchangeDeclare(
			cfg.DLQExchange, // name
			"topic",         // type
			cfg.Durable,     // durable
			false,           // auto-deleted
			false,           // internal
			false,           // no-wait
			nil,             // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare DLQ exchange: %w", err)
		}

		// Declare DLQ queue
		_, err = ch.QueueDeclare(
			cfg.DLQQueue, // name
			cfg.Durable,  // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare DLQ queue: %w", err)
		}

		// Bind DLQ queue to DLQ exchange
		err = ch.QueueBind(
			cfg.DLQQueue,    // queue name
			"#",             // routing key (catch all)
			cfg.DLQExchange, // exchange
			false,           // no-wait
			nil,             // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind DLQ queue: %w", err)
		}

		log.Info("DLQ configured",
			"dlqExchange", cfg.DLQExchange,
			"dlqQueue", cfg.DLQQueue,
		)
	}

	// Prepare queue arguments
	queueArgs := amqp.Table{}
	if cfg.DLQEnabled {
		// Route failed messages to DLQ
		queueArgs["x-dead-letter-exchange"] = cfg.DLQExchange
	}

	// Declare main queue
	_, err = ch.QueueDeclare(
		cfg.QueueName, // name
		cfg.Durable,   // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		queueArgs,     // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		cfg.QueueName,  // queue name
		cfg.RoutingKey, // routing key
		cfg.Exchange,   // exchange
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Info("Connected to RabbitMQ",
		"exchange", cfg.Exchange,
		"queue", cfg.QueueName,
		"routingKey", cfg.RoutingKey,
	)

	return &RabbitMQQueue{
		cfg:    cfg,
		conn:   conn,
		ch:     ch,
		logger: log,
	}, nil
}

// Publish publishes a message to RabbitMQ
func (q *RabbitMQQueue) Publish(ctx context.Context, msg *Message) error {
	tracer := otel.Tracer("gordon-watcher")
	ctx, span := tracer.Start(ctx, "rabbitmq.publish")
	defer span.End()

	span.SetAttributes(
		attribute.String("message.id", msg.ID),
		attribute.String("message.filename", msg.Filename),
		attribute.String("message.kind", msg.Kind),
	)

	// Marshal message
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish
	err = q.ch.PublishWithContext(
		ctx,
		q.cfg.Exchange,   // exchange
		q.cfg.RoutingKey, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			MessageId:    msg.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	q.logger.Debug("Message published to RabbitMQ",
		"messageId", msg.ID,
		"filename", msg.Filename,
		"path", msg.Path,
	)

	return nil
}

// Close closes the RabbitMQ connection
func (q *RabbitMQQueue) Close() error {
	if q.ch != nil {
		q.ch.Close()
	}
	if q.conn != nil {
		q.conn.Close()
	}
	q.logger.Info("RabbitMQ connection closed")
	return nil
}
