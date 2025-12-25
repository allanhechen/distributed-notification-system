package rabbitmqnotifications

import amqp "github.com/rabbitmq/amqp091-go"

type ExchangeName string

type amqpExchange struct {
	name       ExchangeName
	kind       string
	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
	args       amqp.Table
}

const (
	NotificationExchange = "notifications"
)

var exchanges = []amqpExchange{
	{
		name:       NotificationExchange,
		kind:       "topic",
		durable:    true,
		autoDelete: false,
		internal:   false,
		noWait:     false,
		args:       amqp.Table{},
	},
}
