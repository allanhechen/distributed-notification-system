package testutil

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishMessages(ctx context.Context, url string, notifications []domain.Notification, exchange rabbitmqnotifications.ExchangeName, routingKey rabbitmqnotifications.RoutingKey) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	for _, n := range notifications {
		body, err := json.Marshal(n)
		if err != nil {
			return err
		}
		slog.Info("publish messages: publishing message", "exchange", exchange, "routingKey", routingKey)
		err = channel.PublishWithContext(ctx, string(exchange), string(routingKey), false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
