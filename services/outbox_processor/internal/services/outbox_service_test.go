package services

import (
	"context"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/testutil"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOutboxService_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	repo := testutil.GetFakeRepository()
	ss := GetConcreteStatusService(repo, 1, 2, 10*time.Millisecond, 10*time.Millisecond)
	mqs := testutil.GetFakeMqService()
	os := GetConcreteOutboxService(repo, ss, mqs, 1, 10*time.Millisecond)

	notifications := []notification.Notification{
		notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{}),
		notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{}),
		notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{}),
	}
	expectedMap := make(map[uuid.UUID]notification.Notification)
	for _, n := range notifications {
		repo.Entries[n.Identifier] = n
		expected := n
		expected.Status = notification.StatusQueued
		expectedMap[n.Identifier] = expected
	}

	done := make(chan struct{})
	go func() {
		os.HandleMessages(ctx)
		close(done)
	}()

	<-time.After(100 * time.Millisecond)
	cancel()
	<-done

	for _, n := range notifications {
		assert.Equal(t, expectedMap[n.Identifier], repo.Entries[n.Identifier])
	}
}

func TestOutboxService_Failure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	repo := testutil.GetFakeRepository()
	ss := GetConcreteStatusService(repo, 1, 2, 10*time.Millisecond, 10*time.Millisecond)
	mqs := testutil.GetFakeMqService()
	os := GetConcreteOutboxService(repo, ss, mqs, 1, 10*time.Millisecond)

	n := notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{})
	repo.Entries[n.Identifier] = n
	mqs.AddFailedIdentifier(n.Identifier)

	expected := n
	expected.FailedQueueAttempts = notification.DefaultQueueRetryLimit

	done := make(chan struct{})
	go func() {
		os.HandleMessages(ctx)
		close(done)
	}()

	<-time.After(100 * time.Millisecond)
	cancel()
	<-done

	assert.Equal(t, expected, repo.Entries[n.Identifier])
}
