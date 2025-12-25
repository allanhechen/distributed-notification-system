package rabbitmqnotifications

import amqp "github.com/rabbitmq/amqp091-go"

type amqpBinding struct {
	name     QueueName
	key      string
	exchange ExchangeName
	noWait   bool
	args     amqp.Table
}

type RoutingKey = string

const (
	TestNotificationKey    RoutingKey = "notifications.test"
	AppleNotificationKey   RoutingKey = "notifications.apple"
	AndroidNotificationKey RoutingKey = "notifications.android"
	EmailNotificationKey   RoutingKey = "notifications.email"
)

var bindings = []amqpBinding{
	{
		name:     TestNotificationQueue,
		key:      TestNotificationKey,
		exchange: NotificationExchange,
		noWait:   false,
		args:     amqp.Table{},
	},
	{
		name:     AppleNotificationQueue,
		key:      AppleNotificationKey,
		exchange: NotificationExchange,
		noWait:   false,
		args:     amqp.Table{},
	},
	{
		name:     AndroidNotificationQueue,
		key:      AndroidNotificationKey,
		exchange: NotificationExchange,
		noWait:   false,
		args:     amqp.Table{},
	},
	{
		name:     EmailNotificationQueue,
		key:      EmailNotificationKey,
		exchange: NotificationExchange,
		noWait:   false,
		args:     amqp.Table{},
	},
}
