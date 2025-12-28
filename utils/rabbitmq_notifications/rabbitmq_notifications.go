// rabbitmqnotifications contains definitions for the persistent RabbitMQ
// entities, and a utility function to declare them.
package rabbitmqnotifications

import (
	"context"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DeclareEntities is a context-aware utility function used to declare
// RabbitMQ entities during service initialization. This utility does not
// attempt to handle any errors raised during initialization, and instead
// returns them.
//
// The url parameter denotes the RabbitMQ instance to have entities
// declared against. The entities are hard-coded within the other files
// within this package, which is acceptable because they will not change
// frequently.
func DeclareEntities(ctx context.Context, url string) error {
	config := amqp.Config{
		Dial: amqp.DefaultDial(5 * time.Second),
	}
	conn, err := amqp.DialConfig(url, config)
	if err != nil {
		slog.Error("rabbitmq notifications: failed to open connection", "error", err)
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		slog.Error("rabbitmq notifications: failed to open channel", "error", err)
		return err
	}
	defer channel.Close()

	if ctx.Err() != nil {
		slog.Warn("rabbitmq notifications: context closed before declaring entities, exiting")
		return ctx.Err()
	}

	slog.Info("rabbitmq notifications: begin declaring exchanges")
	for _, e := range exchanges {
		err = channel.ExchangeDeclare(string(e.name), e.kind, e.durable, e.autoDelete, e.internal, e.noWait, e.args)
		if err != nil {
			slog.Error("rabbitmq notifications: failed to declare exchange", "error", err)
			return err
		}
		slog.Info("rabbitmq notifications: declared exchange", "name", e.name)
	}

	slog.Info("rabbitmq notifications: begin declaring queues")
	for _, q := range queues {
		_, err = channel.QueueDeclare(string(q.name), q.durable, q.autoDelete, q.exclusive, q.noWait, q.args)
		if err != nil {
			slog.Error("rabbitmq notifications: failed to declare queue", "error", err)
			return err
		}
		slog.Info("rabbitmq notifications: declared queue", "name", q.name)
	}

	slog.Info("rabbitmq notifications: begin declaring bindings")
	for _, b := range bindings {
		err = channel.QueueBind(string(b.name), b.key, string(b.exchange), b.noWait, b.args)
		if err != nil {
			slog.Error("rabbitmq notifications: failed to declare binding", "error", err)
			return err
		}
		slog.Info("rabbitmq notifications: declared binding", "name", b.name)
	}

	slog.Info("rabbitmq notifications: successful initialization")
	return nil
}
