package rabbitmqnotifications

import amqp "github.com/rabbitmq/amqp091-go"

// QueueName is the type denoting queue names.
type QueueName string

type amqpQueue struct {
	name       QueueName
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
	args       amqp.Table
}

var quorumQueueArgs = amqp.Table{
	amqp.QueueTypeArg: amqp.QueueTypeQuorum,
}

const (
	// TestNotificationQueue is the queue dedicated for test devices.
	TestNotificationQueue = "test"

	// AppleNotificationQueue is the queue dedicated for Apple devices (IOS, MacOS).
	AppleNotificationQueue = "apple"

	// AndroidNotificationQueue is the queue dedicated for Android devices.
	AndroidNotificationQueue = "android"

	// EmailNotificationQueue is the queue dedicated for email devices.
	EmailNotificationQueue = "email"
)

var queues = []amqpQueue{
	{
		name:       TestNotificationQueue,
		durable:    true,
		autoDelete: false,
		exclusive:  false,
		noWait:     false,
		args:       quorumQueueArgs,
	},
	{
		name:       AppleNotificationQueue,
		durable:    true,
		autoDelete: false,
		exclusive:  false,
		noWait:     false,
		args:       quorumQueueArgs,
	},
	{
		name:       AndroidNotificationQueue,
		durable:    true,
		autoDelete: false,
		exclusive:  false,
		noWait:     false,
		args:       quorumQueueArgs,
	},
	{
		name:       EmailNotificationQueue,
		durable:    true,
		autoDelete: false,
		exclusive:  false,
		noWait:     false,
		args:       quorumQueueArgs,
	},
}
