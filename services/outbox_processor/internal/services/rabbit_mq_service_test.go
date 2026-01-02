package services

import (
	"context"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	sharedTestutil "github.com/allanhechen/distributed-notification-system/utils/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRabbitMqService_SendNotification(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// set up container
	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	require.NoError(t, err)
	defer c.Close(context.Background())

	err = c.DeclareEntities(ctx)
	require.NoError(t, err)

	mqs := GetRabbitMqService(c.ConnString, 1, 1, 250*time.Millisecond, 1)
	mqs.Start()
	defer mqs.Stop()

	// set up test items
	expiryTime := time.Now().UTC().Add(1 * time.Second)
	n := notification.GetFakeNotification(notification.TestDeviceType, notification.StatusUndelivered, 0, expiryTime)
	updates := make(chan domain.StatusUpdate, 1)
	done, err := mqs.SendNotification(n, updates)

	// check results
	assert.NoError(t, err)
	<-done
	updateMessage := <-updates
	assert.Equal(
		t,
		domain.StatusUpdate{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusQueued,
		},
		updateMessage,
	)
}

func TestRabbitMqService_Reconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// set up container
	c, err := sharedTestutil.GetRabbitMqContainer(ctx)
	require.NoError(t, err)
	defer c.Close(context.Background())

	err = c.DeclareEntities(ctx)
	require.NoError(t, err)

	mqs := GetRabbitMqService(c.ConnString, 1, 1, 250*time.Millisecond, 1)
	mqs.Start()
	defer mqs.Stop()

	// send first message
	expiryTime := time.Now().UTC().Add(1 * time.Second)
	n := notification.GetFakeNotification(notification.TestDeviceType, notification.StatusUndelivered, 0, expiryTime)
	updates := make(chan domain.StatusUpdate, 1)
	done, err := mqs.SendNotification(n, updates)

	// check first message
	assert.NoError(t, err)
	<-done
	updateMessage := <-updates
	assert.Equal(
		t,
		domain.StatusUpdate{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusQueued,
		},
		updateMessage,
	)

	// disconnect and try again
	c.Disconnect()
	<-time.After(250 * time.Millisecond)
	c.Reconnect()

	expiryTime = time.Now().UTC().Add(1 * time.Second)
	n = notification.GetFakeNotification(notification.TestDeviceType, notification.StatusUndelivered, 0, expiryTime)
	updates = make(chan domain.StatusUpdate, 1)
	done, err = mqs.SendNotification(n, updates)

	// check second message
	assert.NoError(t, err)
	<-done
	updateMessage = <-updates
	assert.Equal(
		t,
		domain.StatusUpdate{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusQueued,
		},
		updateMessage,
	)
}
