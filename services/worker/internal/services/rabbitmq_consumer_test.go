package services

import (
	"context"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/worker/internal/testutil"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	sharedTestutil "github.com/allanhechen/distributed-notification-system/utils/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRabbitMqConsumer_ReceiveMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// set up container
	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	require.NoError(t, err)
	err = c.DeclareEntities(ctx)
	require.NoError(t, err)

	// insert messages
	notifications := []domain.Notification{
		{
			Identifier:       uuid.NewString(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
		{
			Identifier:       uuid.NewString(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
		{
			Identifier:       "stop",
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
	}
	err = testutil.PublishMessages(ctx, c.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(c.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	result := make([]domain.Notification, 0, 3)
	for msg := range out {
		payload := msg.Payload()
		err = msg.Ack(ctx)
		assert.NoError(t, err)
		result = append(result, payload)
		if payload.Identifier == "stop" {
			break
		}
	}

	assert.Equal(t, notifications, result)
	cancel()

	c.Close(ctx)
}

func TestRabbitMqConsumer_Reconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// set up container
	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	require.NoError(t, err)
	err = c.DeclareEntities(ctx)
	require.NoError(t, err)

	// insert messages
	notifications := []domain.Notification{
		{
			Identifier:       uuid.NewString(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
		{
			Identifier:       uuid.NewString(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
		{
			Identifier:       "stop",
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
	}
	err = testutil.PublishMessages(ctx, c.ConnString, notifications[:1], rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(c.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	// receive first message
	result := make([]domain.Notification, 0, 3)
	msg := <-out
	payload := msg.Payload()
	err = msg.Ack(ctx)
	assert.NoError(t, err)
	result = append(result, payload)

	// simulate network dropout
	err = c.Disconnect()
	require.NoError(t, err)
	err = c.Reconnect()
	require.NoError(t, err)

	// publish additional messages
	err = testutil.PublishMessages(ctx, c.ConnString, notifications[1:], rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to all remaining messages
	for msg := range out {
		payload := msg.Payload()
		err = msg.Ack(ctx)
		assert.NoError(t, err)
		result = append(result, payload)
		if payload.Identifier == "stop" {
			break
		}
	}

	assert.Equal(t, notifications, result)
	cancel()

	c.Close(ctx)
}

func TestRabbitMqConsumer_Ack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// set up container
	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	require.NoError(t, err)
	err = c.DeclareEntities(ctx)
	require.NoError(t, err)

	// insert messages
	notifications := []domain.Notification{
		{
			Identifier:       uuid.NewString(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
	}
	err = testutil.PublishMessages(ctx, c.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(c.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	// retrieve single message
	msg := <-out
	err = msg.Ack(ctx)
	assert.NoError(t, err)

	// ensure that no more messages appear
	select {
	case msg := <-out:
		assert.Fail(t, "expected channel to be empty, received %v", msg)
	case <-time.After(2 * time.Second):
	}

	cancel()
	c.Close(ctx)
}

func TestRabbitMqConsumer_NackNoRequeue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// set up container
	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	require.NoError(t, err)
	err = c.DeclareEntities(ctx)
	require.NoError(t, err)

	// insert messages
	notifications := []domain.Notification{
		{
			Identifier:       uuid.NewString(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
	}
	err = testutil.PublishMessages(ctx, c.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(c.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	// retrieve single message
	msg := <-out
	err = msg.Nack(ctx, false)
	assert.NoError(t, err)

	// ensure that no more messages appear
	select {
	case msg := <-out:
		assert.Fail(t, "expected channel to be empty, received %v", msg)
	case <-time.After(2 * time.Second):
	}

	cancel()
	c.Close(ctx)
}

func TestRabbitMqConsumer_NackRequeue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// set up container
	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	require.NoError(t, err)
	err = c.DeclareEntities(ctx)
	require.NoError(t, err)

	// insert messages
	notifications := []domain.Notification{
		{
			Identifier:       uuid.NewString(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.NewString(),
		},
	}
	err = testutil.PublishMessages(ctx, c.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(c.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	// retrieve single message
	msg := <-out
	err = msg.Nack(ctx, true)
	assert.NoError(t, err)

	// ensure the same message appears again
	select {
	case msg := <-out:
		assert.Equal(t, notifications[0].Identifier, msg.Payload().Identifier)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "expected channel to receive the same message, received nothing")
	}

	cancel()
	c.Close(ctx)
}
