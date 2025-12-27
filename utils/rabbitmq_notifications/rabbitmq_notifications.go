package rabbitmqnotifications

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareEntities(ctx context.Context, url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		slog.Error("rabbitmq notifications: failed to open connection")
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		slog.Error("rabbitmq notifications: failed to open channel")
		return err
	}
	defer channel.Close()

	if ctx.Err() != nil {
		slog.Warn("rabbitmq notifications: context closed before declaring entities, exiting")
		return err
	}

	slog.Info("rabbitmq notifications: begin declaring exchanges")
	for _, e := range exchanges {
		err = channel.ExchangeDeclare(string(e.name), e.kind, e.durable, e.autoDelete, e.internal, e.noWait, e.args)
		slog.Info("rabbitmq notifications: declaring exchange", "name", e.name)
		if err != nil {
			slog.Error("rabbitmq notifications: failed to declare exchange", "error", err)
			return err
		}
	}

	slog.Info("rabbitmq notifications: begin declaring queues")
	for _, q := range queues {
		_, err = channel.QueueDeclare(string(q.name), q.durable, q.autoDelete, q.exclusive, q.noWait, q.args)
		slog.Info("rabbitmq notifications: declaring queue", "name", q.name)
		if err != nil {
			slog.Error("rabbitmq notifications: failed to declare queue", "error", err)
			return err
		}
	}

	slog.Info("rabbitmq notifications: begin declaring bindings")
	for _, b := range bindings {
		err = channel.QueueBind(string(b.name), b.key, string(b.exchange), b.noWait, b.args)
		slog.Info("rabbitmq notifications: declaring binding", "name", b.name)
		if err != nil {
			slog.Error("rabbitmq notifications: failed to declare binding", "error", err)
			return err
		}
	}

	slog.Info("rabbitmq notifications: successful initialization")
	return nil
}
