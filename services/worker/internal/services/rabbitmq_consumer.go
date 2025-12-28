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

var ErrConnection = errors.New("consumer: communication dropped with rabbitmq")

type RabbitMqConsumer struct {
	url             string
	maxRetryTimeout uint
	healthyTimeout  uint
	prefetch        uint
	queue           rabbitmqnotifications.QueueName
}

func NewRabbitMqConsumer(url string, queue rabbitmqnotifications.QueueName, maxRetryTimeout uint, healthyTimeout uint, prefetch uint) domain.Consumer[domain.Notification] {
	return &RabbitMqConsumer{
		url:             url,
		queue:           queue,
		maxRetryTimeout: maxRetryTimeout,
		healthyTimeout:  healthyTimeout,
		prefetch:        prefetch,
	}
}

func (r *RabbitMqConsumer) Consume(ctx context.Context) (<-chan domain.Message[domain.Notification], error) {
	outputCh := make(chan domain.Message[domain.Notification])

	go r.handleReconnect(ctx, outputCh)
	return outputCh, nil
}

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
		if runningTime <= time.Duration(r.healthyTimeout*uint(time.Second)) {
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
				d.Nack(false, false)
				slog.Error("consumer: failed to unmarshal message body", "body", d.Body)
				continue
			}

			notification := RabbitmqNotification{
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
				d.Nack(false, true)
			case outputCh <- &notification:
			}
		}
	}
}
