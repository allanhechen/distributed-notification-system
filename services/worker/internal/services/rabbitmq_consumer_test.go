package services

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/testutil"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	sharedTestutil "github.com/allanhechen/distributed-notification-system/utils/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testContainer *sharedTestutil.RabbitMqContainer

func TestMain(m *testing.M) {
	ctx := context.Background()

	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	testContainer = c

	if err := c.DeclareEntities(ctx); err != nil {
		c.Close(ctx)
		log.Fatalf("Failed to declare entities: %v", err)
	}

	code := m.Run()
	testContainer.Close(ctx)
	os.Exit(code)
}

func TestRabbitMqConsumer_ReceiveMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// insert messages
	notifications := []notification.Notification{
		{
			Identifier:       uuid.New(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
		{
			Identifier:       uuid.New(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
		{
			Identifier:       uuid.Nil,
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
	}
	err := testutil.PublishMessages(ctx, testContainer.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(testContainer.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	result := make([]notification.Notification, 0, 3)
	for msg := range out {
		payload := msg.Payload()
		err = msg.Ack(ctx)
		assert.NoError(t, err)
		result = append(result, payload)
		if payload.Identifier == uuid.Nil {
			break
		}
	}

	assert.Equal(t, notifications, result)
	cancel()
}

func TestRabbitMqConsumer_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// listen to messages
	consumer := NewRabbitMqConsumer(testContainer.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	var wg sync.WaitGroup

	wg.Go(func() {
		<-out
	})

	cancel()
	wg.Wait()
}

func TestRabbitMqConsumer_Reconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// insert messages
	notifications := []notification.Notification{
		{
			Identifier:       uuid.New(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
		{
			Identifier:       uuid.New(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
		{
			Identifier:       uuid.Nil,
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
	}
	err := testutil.PublishMessages(ctx, testContainer.ConnString, notifications[:1], rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(testContainer.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
	out, err := consumer.Consume(ctx)
	assert.NoError(t, err)

	// receive first message
	result := make([]notification.Notification, 0, 3)
	msg := <-out
	payload := msg.Payload()
	err = msg.Ack(ctx)
	assert.NoError(t, err)
	result = append(result, payload)

	// simulate network dropout
	err = testContainer.Disconnect()
	require.NoError(t, err)
	err = testContainer.Reconnect()
	require.NoError(t, err)

	// publish additional messages
	err = testutil.PublishMessages(ctx, testContainer.ConnString, notifications[1:], rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to all remaining messages
	for msg := range out {
		payload := msg.Payload()
		err = msg.Ack(ctx)
		assert.NoError(t, err)
		result = append(result, payload)
		if payload.Identifier == uuid.Nil {
			break
		}
	}

	assert.Equal(t, notifications, result)
	cancel()
}

func TestRabbitMqConsumer_Ack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// insert messages
	notifications := []notification.Notification{
		{
			Identifier:       uuid.New(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
	}
	err := testutil.PublishMessages(ctx, testContainer.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(testContainer.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
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
	case <-time.After(1 * time.Second):
	}

	cancel()
}

func TestRabbitMqConsumer_NackNoRequeue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// insert messages
	notifications := []notification.Notification{
		{
			Identifier:       uuid.New(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
	}
	err := testutil.PublishMessages(ctx, testContainer.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(testContainer.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
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
	case <-time.After(1 * time.Second):
	}

	cancel()
}

func TestRabbitMqConsumer_NackRequeue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// insert messages
	notifications := []notification.Notification{
		{
			Identifier:       uuid.New(),
			NotificationType: notification.IosDeviceType,
			DeviceIdentifier: uuid.New(),
		},
	}
	err := testutil.PublishMessages(ctx, testContainer.ConnString, notifications, rabbitmqnotifications.NotificationExchange, rabbitmqnotifications.TestNotificationKey)
	require.NoError(t, err)

	// listen to messages
	consumer := NewRabbitMqConsumer(testContainer.ConnString, rabbitmqnotifications.TestNotificationQueue, 30, 30, 1)
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
	case <-time.After(1 * time.Second):
		assert.Fail(t, "expected channel to receive the same message, received nothing")
	}

	cancel()
}
