package testutil

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishMessages is a context-aware message publisher intended to be
// used in integration tests to publish notifications to a RabbitMQ
// instance.
func PublishMessages(
	ctx context.Context,
	url string,
	notifications []domain.Notification,
	exchange rabbitmqnotifications.ExchangeName,
	routingKey rabbitmqnotifications.RoutingKey,
) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		slog.Error("publish messages: failed to open connection")
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		slog.Error("publish messages: failed to open channel")
		return err
	}
	defer channel.Close()

	for _, n := range notifications {
		if ctx.Err() != nil {
			slog.Warn("publish messages: context cancelled while publishing messages")
			return ctx.Err()
		}

		body, err := json.Marshal(n)
		if err != nil {
			slog.Error("publish messages: failed marshal payload")
			return err
		}
		err = channel.PublishWithContext(ctx, string(exchange), string(routingKey), false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
		if err != nil {
			slog.Error("publish messages: failed to publish message", "exchange", exchange, "routingKey", routingKey, "body", body)
			return err
		}
		slog.Info("publish messages: published message", "exchange", exchange, "routingKey", routingKey, "body", body)
	}

	return nil
}
