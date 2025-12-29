package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ErrConnection represents a dropped connection from RabbitMQ.
var ErrConnection = errors.New("consumer: communication dropped with rabbitmq")

// RabbitMqConsumer is a concrete implementation of the Consumer interface
// to be used with RabbitMQ.
//
// It handles reconnects with exponential-backoff while simultaneously
// being context-aware.
type RabbitMqConsumer struct {
	url             string
	maxRetryTimeout uint
	healthyTimeout  uint
	prefetch        uint
	queue           rabbitmqnotifications.QueueName
}

// NewRabbitMqConsumer returns an instance of RabbitMqConsumer with the given parameters.
func NewRabbitMqConsumer(url string, queue rabbitmqnotifications.QueueName, maxRetryTimeout uint, healthyTimeout uint, prefetch uint) domain.Consumer[domain.Notification] {
	return &RabbitMqConsumer{
		url:             url,
		queue:           queue,
		maxRetryTimeout: maxRetryTimeout,
		healthyTimeout:  healthyTimeout,
		prefetch:        prefetch,
	}
}

// Consume returns a channel satisfying Consumer.Consume representing a
// series of notification messages from RabbitMQ.
func (r *RabbitMqConsumer) Consume(ctx context.Context) (<-chan domain.Message[domain.Notification], error) {
	outputCh := make(chan domain.Message[domain.Notification])

	go r.handleReconnect(ctx, outputCh)
	return outputCh, nil
}

// handleReconnect is the method handling reconnection logic. It runs
// individual iterations using the handleIteration method, using
// exponential backoff during failures. The backoff is reset if the
// connection is "healthy" for healthyTimeout seconds.
//
// handleIteration must be completed before closing this channel. The
// current implementation runs handleIteration synchronously.
func (r *RabbitMqConsumer) handleReconnect(ctx context.Context, outputCh chan domain.Message[domain.Notification]) {
	defer close(outputCh)

	backoff := uint(1)
	for {
		// ensure that closing context returns us
		if ctx.Err() != nil {
			slog.Info("consumer: received context close signal, exiting")
			return
		}

		begin := time.Now()
		slog.Info("consumer: (re)connecting to rabbitmq")
		r.handleIteration(ctx, outputCh)

		if ctx.Err() != nil {
			slog.Info("consumer: received context close signal, exiting")
			return
		}

		runningTime := time.Since(begin)
		if runningTime <= time.Duration(r.healthyTimeout)*time.Second {
			backoff = min(backoff*2, r.maxRetryTimeout)
			slog.Info("consumer: previous iteration failed, waiting for backoff", "backoff", backoff)

			select {
			case <-ctx.Done():
				slog.Info("consumer: context closed while waiting for timeout")
				return
			case <-time.After(time.Duration(backoff) * time.Second):
			}
		} else {
			backoff = uint(1)
			slog.Info("consumer: previous healthy connection dropped, resetting backoff")
		}
	}
}

// handleIteration connects to RabbitMQ with its own connection and
// channel. It streams the received messages to the output channel.
func (r *RabbitMqConsumer) handleIteration(ctx context.Context, outputCh chan domain.Message[domain.Notification]) error {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.Qos(int(r.prefetch), 0, false); err != nil {
		return err
	}

	consumer := uuid.New().String()
	inputCh, err := channel.Consume(string(r.queue), consumer, false, false, false, false, nil)
	if err != nil {
		return err
	}
	defer channel.Cancel(consumer, false)

	chanCh := make(chan *amqp.Error, 1)
	connCh := make(chan *amqp.Error, 1)
	channel.NotifyClose(chanCh)
	conn.NotifyClose(connCh)

	shutdownCh := make(chan struct{}, 1)

	slog.Info("consumer: connected successfully with rabbitmq")
	for {
		select {
		case <-ctx.Done():
			slog.Info("consumer: context cancelled, dropping in-flight messages")
			close(shutdownCh)
			return nil
		case <-chanCh:
			slog.Warn("consumer: channel dropped")
			close(shutdownCh)
			return ErrConnection
		case <-connCh:
			slog.Warn("consumer: connection dropped")
			close(shutdownCh)
			return ErrConnection
		case d, ok := <-inputCh:
			if !ok {
				slog.Warn("consumer: channel closed before context closed")
				return nil
			}

			var payload domain.Notification
			err := json.Unmarshal(d.Body, &payload)
			if err != nil {
				if nackErr := d.Nack(false, false); nackErr != nil {
					slog.Error("consumer: failed to nack malformed message", "error", nackErr)
				} else {
					slog.Error("consumer: failed to unmarshal message body", "body", d.Body)
				}
				continue
			}

			notification := RabbitMqNotification{
				payload:    payload,
				identifier: d.MessageId,
				ackFn: func(ctx context.Context) error {
					return d.Ack(false)
				},
				nackFn: func(ctx context.Context, requeue bool) error {
					return d.Nack(false, requeue)
				},
			}

			select {
			case <-shutdownCh:
				if nackErr := d.Nack(false, false); nackErr != nil {
					slog.Error("consumer: failed to nack message during shutdown", "error", nackErr)
				}
			case outputCh <- &notification:
			}
		}
	}
}
