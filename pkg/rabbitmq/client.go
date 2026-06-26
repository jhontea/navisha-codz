package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Queue names.
const (
	QueueCodeExecution  = "code.execution.pending"
	QueueNotifications  = "notifications"
	QueueDLX            = "dead.letter"
	ExchangeCodeExec    = "code.execution"
	ExchangeEvents     = "events"
)

// Config holds RabbitMQ connection configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

// Client wraps amqp.Connection with additional functionality.
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
}

// New creates a new RabbitMQ client with retry logic.
func New(cfg Config) (*Client, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)

	var conn *amqp.Connection
	var err error

	// Retry connection with exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		backoff := time.Duration(1<<i) * time.Second
		log.Printf("RabbitMQ connection attempt %d/%d failed: %v. Retrying in %v...", i+1, maxRetries, err, backoff)
		time.Sleep(backoff)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &Client{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}

	// Declare exchanges and queues
	if err := client.declareTopology(); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare topology: %w", err)
	}

	log.Printf("RabbitMQ connected: %s@%s:%d", cfg.User, cfg.Host, cfg.Port)
	return client, nil
}

// NewFromEnv creates a RabbitMQ client from environment variables.
func NewFromEnv() (*Client, error) {
	cfg := Config{
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnvInt("RABBITMQ_PORT", 5672),
		User:     getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		VHost:    getEnv("RABBITMQ_VHOST", "/"),
	}
	return New(cfg)
}

// declareTopology sets up exchanges, queues, and bindings.
func (c *Client) declareTopology() error {
	// Dead letter exchange
	if err := c.channel.ExchangeDeclare(
		"dlx", "direct", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	// Main code execution exchange
	if err := c.channel.ExchangeDeclare(
		ExchangeCodeExec, "topic", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare code execution exchange: %w", err)
	}

	// Events exchange (for pub/sub patterns)
	if err := c.channel.ExchangeDeclare(
		ExchangeEvents, "fanout", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare events exchange: %w", err)
	}

	// Dead letter queue
	if _, err := c.channel.QueueDeclare(
		QueueDLX, true, false, false, false,
		amqp.Table{"x-message-ttl": 30000},
	); err != nil {
		return fmt.Errorf("failed to declare DLX queue: %w", err)
	}

	// Code execution queue
	if _, err := c.channel.QueueDeclare(
		QueueCodeExecution, true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    "dlx",
			"x-message-ttl":            300000, // 5 min TTL
			"x-max-priority":           10,
		},
	); err != nil {
		return fmt.Errorf("failed to declare code execution queue: %w", err)
	}

	// Notifications queue
	if _, err := c.channel.QueueDeclare(
		QueueNotifications, true, false, false, false,
		amqp.Table{"x-message-ttl": 86400000}, // 24h TTL
	); err != nil {
		return fmt.Errorf("failed to declare notifications queue: %w", err)
	}

	// Bindings
	if err := c.channel.QueueBind(QueueDLX, "dead.letter", "dlx", false, nil); err != nil {
		return fmt.Errorf("failed to bind DLX queue: %w", err)
	}
	if err := c.channel.QueueBind(QueueCodeExecution, "execution.*", ExchangeCodeExec, false, nil); err != nil {
		return fmt.Errorf("failed to bind code execution queue: %w", err)
	}
	if err := c.channel.QueueBind(QueueNotifications, "notification.*", ExchangeCodeExec, false, nil); err != nil {
		return fmt.Errorf("failed to bind notifications queue: %w", err)
	}

	return nil
}

// Publish sends a message to the specified exchange with routing key.
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = c.channel.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			MessageId:    generateMessageID(),
			Headers: amqp.Table{
				"source": "execution-service",
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// PublishToQueue sends a message directly to a named queue.
func (c *Client) PublishToQueue(ctx context.Context, queueName string, message interface{}) error {
	return c.Publish(ctx, ExchangeCodeExec, queueName, message)
}

// Consume starts consuming messages from a queue.
func (c *Client) Consume(ctx context.Context, queueName string, handler func(amqp.Delivery) error) error {
	msgs, err := c.channel.Consume(
		queueName,   // queue
		"",          // consumer tag (auto-generated)
		false,       // auto-ack (manual ack)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming from %s: %w", queueName, err)
	}

	log.Printf("Started consuming from queue: %s", queueName)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping consumer for queue: %s", queueName)
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("Channel closed for queue: %s, reconnecting...", queueName)
				return fmt.Errorf("channel closed for queue %s", queueName)
			}

			if err := handler(msg); err != nil {
				log.Printf("Message handler error: %v", err)
				// Negative acknowledge, requeue
				if nackErr := msg.Nack(false, true); nackErr != nil {
					log.Printf("Failed to nack message: %v", nackErr)
				}
			} else {
				// Acknowledge
				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("Failed to ack message: %v", ackErr)
				}
			}
		}
	}
}

// QueueInfo returns information about a queue.
func (c *Client) QueueInfo(queueName string) (amqp.Queue, error) {
	return c.channel.QueueInspect(queueName)
}

// HealthCheck verifies RabbitMQ connection.
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	return nil
}

// Close gracefully closes the channel and connection.
func (c *Client) Close() error {
	if err := c.channel.Close(); err != nil {
		log.Printf("Failed to close channel: %v", err)
	}
	if err := c.conn.Close(); err != nil {
		log.Printf("Failed to close connection: %v", err)
	}
	return nil
}

func generateMessageID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}
