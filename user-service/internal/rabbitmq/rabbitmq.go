package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	amqplib "github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/config"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/logger"
)

type RabbitMQ interface {
	Health() error
	MaintainConnection(ctx context.Context)
	Publish(ctx context.Context, queue string, message *MessageType) error
}

type rabbitmq struct {
	config        *config.AMQP
	conn          *amqplib.Connection
	pubChannel    *amqplib.Channel
	reconnectLock sync.Mutex
	logger        logger.Logger
}

type Opts struct {
	Logger logger.Logger
	Config *config.AMQP
}

type MessageType struct {
	Pattern string `json:"pattern"`
	Data    any    `json:"data"`
}

func NewRabbitMQ(ctx context.Context, opts *Opts) RabbitMQ {
	b := &rabbitmq{
		reconnectLock: sync.Mutex{},
		config:        opts.Config,
		logger:        opts.Logger,
	}

	go b.MaintainConnection(ctx)

	return b
}

func (b *rabbitmq) Health() error {
	if b.conn == nil || b.conn.IsClosed() {
		b.logger.Warn("AMQP health check failed!")
		return errors.New("amqp healthcheck failed")
	}

	return nil
}

func (b *rabbitmq) MaintainConnection(ctx context.Context) {
	if err := b.connect(); err != nil {
		b.logger.Error("Initial AMQP connection attempt failed", logger.Field{Key: "error", Value: err.Error()})
	}

	attempts := b.config.ConnectionRetryAttempts
	interval := b.config.ConnectionRetryInterval

	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			if b.pubChannel != nil {
				if err := b.pubChannel.Close(); err != nil {
					b.logger.Error("AMQP failed to close channel", logger.Field{Key: "error", Value: err.Error()})
				}
			}

			if b.conn != nil {
				if err := b.conn.Close(); err != nil {
					b.logger.Error("AMQP failed to close connection", logger.Field{Key: "error", Value: err.Error()})
				}
			}

			return
		case <-t.C:
			if err := b.tryReconnect(attempts, interval); err != nil {
				return
			}
		}
	}
}

func (b *rabbitmq) Publish(ctx context.Context, queue string, message *MessageType) error {
	q, err := declareQueue(queue, b.pubChannel)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, b.config.PublishTimeout)
	defer cancel()

	messageData, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	err = b.pubChannel.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqplib.Publishing{
			ContentType: "application/json",
			Body:        messageData,
		},
	)
	if err != nil {
		return err
	}
	b.logger.Info("AMQP message sent", logger.Field{Key: "event", Value: message.Pattern})

	return nil
}

func (b *rabbitmq) connect() error {
	address := fmt.Sprintf("amqp://%s:%s@%s:%d", b.config.Username, b.config.Password, b.config.Host, b.config.Port)

	var err error

	b.conn, err = amqplib.Dial(address)
	if err != nil {
		b.logger.Error("AMQP connection error", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	b.logger.Info("AMQP connected on " + address)

	channel, err := b.newChannel()
	if err != nil {
		b.logger.Error("Unable to create listen channel", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	b.pubChannel = channel

	return nil
}

func (b *rabbitmq) tryReconnect(attempts int, interval time.Duration) error {
	b.reconnectLock.Lock()
	defer b.reconnectLock.Unlock()

	if b.Health() == nil {
		return nil
	}

	for i := range attempts {
		b.logger.Info("AMQP attempting reconnection", logger.Field{Key: "attempt", Value: i + 1}, logger.Field{Key: "interval", Value: interval * (1 << i)})

		if err := b.connect(); err == nil {
			return nil
		}

		if i+1 < attempts {
			//retry with exponential backoff
			exponent := math.Pow(2, float64(i))
			delay := time.Duration(float64(interval) * exponent)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("could not reconnect after %d retries", attempts)
}

func (b *rabbitmq) newChannel() (*amqplib.Channel, error) {
	c, err := b.conn.Channel()

	if err != nil {
		b.logger.Error("AMQP channel error", logger.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	return c, nil
}

func declareQueue(queue string, channel *amqplib.Channel) (*amqplib.Queue, error) {
	q, err := channel.QueueDeclare(
		queue,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &q, err
}
