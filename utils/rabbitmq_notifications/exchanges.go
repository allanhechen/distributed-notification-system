package rabbitmqnotifications

import amqp "github.com/rabbitmq/amqp091-go"

// ExchangeName is the type denoting an exchange name.
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
	// NotificationExchange is the exchange used to route notifications.
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
