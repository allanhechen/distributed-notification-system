package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	amqp "github.com/rabbitmq/amqp091-go"
)

func generateRandomSuffix() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type RabbitMqConsumer struct {
	conn *amqp.Connection
}

func (r *RabbitMqConsumer) Consume(ctx context.Context) (<-chan domain.Message[domain.Notification], error) {
	channel, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.Qos(10, 0, false); err != nil {
		channel.Close()
		return nil, err
	}

	consumer := fmt.Sprintf("%s-%d-%s", rabbitmqnotifications.TestNotificationQueue, time.Now().UTC().UnixMilli(), generateRandomSuffix())
	inputCh, err := channel.Consume(string(rabbitmqnotifications.TestNotificationQueue), consumer, false, false, false, false, nil)
	if err != nil {
		channel.Close()
		return nil, err
	}
	outputCh := make(chan domain.Message[domain.Notification])

	go func() {
		defer close(outputCh)
		defer channel.Close()
		defer channel.Cancel(consumer, false)

		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-inputCh:
				if !ok {
					slog.Warn("consumer: channel closed before context closed")
					return
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
				case <-ctx.Done():
					d.Nack(false, true)
				case outputCh <- &notification:
				}
			}
		}
	}()
	return outputCh, nil
}
