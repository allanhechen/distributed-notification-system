package rabbitmqnotifications

import amqp "github.com/rabbitmq/amqp091-go"

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
	TestNotificationQueue    = "test"
	AppleNotificationQueue   = "apple"
	AndroidNotificationQueue = "android"
	EmailNotificationQueue   = "email"
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
