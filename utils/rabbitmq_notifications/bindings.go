package rabbitmqnotifications

import amqp "github.com/rabbitmq/amqp091-go"

type amqpBinding struct {
	name     QueueName
	key      string
	exchange ExchangeName
	noWait   bool
	args     amqp.Table
}

// RoutingKey is the type for RabbitMQ routing keys.
type RoutingKey = string

const (
	// TestNotificationKey sends messages to the test queue.
	TestNotificationKey RoutingKey = "notifications.test"

	// AppleNotificationKey sends messages to the Apple device queue (IOS, MacOS).
	AppleNotificationKey RoutingKey = "notifications.apple"

	// AndroidNotificationKey sends messages to the Android device queue.
	AndroidNotificationKey RoutingKey = "notifications.android"

	// EmailNotificationKey sends messages to the email device queue.
	EmailNotificationKey RoutingKey = "notifications.email"
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
