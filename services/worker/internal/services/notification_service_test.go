package services

import (
	"context"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/testutil"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/stretchr/testify/assert"
)

type fakeConcreteNotificationService struct {
	ns       *ConcreteNotifcationService
	db       *testutil.FakeRepository
	consumer *testutil.FakeConsumer
	notifier *testutil.FakeNotifier
}

func GetFakeConcreteNotificationService(maxParallelism ...uint) fakeConcreteNotificationService {
	mp := uint(1)
	if len(maxParallelism) > 0 {
		mp = maxParallelism[0]
	}

	db := testutil.GetFakeRepository()
	consumer := testutil.GetFakeConsumer()
	notifier := testutil.GetFakeNotifier()

	ns := &ConcreteNotifcationService{
		db:             db,
		consumer:       consumer,
		notifier:       notifier,
		maxParallelism: mp,
	}

	return fakeConcreteNotificationService{
		ns:       ns,
		db:       db,
		consumer: consumer,
		notifier: notifier,
	}
}

func TestNotificationService_UpdateSuccessExisting(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	fakes := GetFakeConcreteNotificationService()
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusProcessing,
	}

	ns.updateStatusSuccess(ctx, identifier, fakeMessage)
	assert.Equal(t, testutil.StatusAcked, fakeMessage.Status)
	assert.Equal(t, notification.StatusComplete, fakes.db.Db[identifier].Status)
}

func TestNotificationService_UpdateSuccessNonExistent(t *testing.T) {
	ctx := context.Background()
	fakes := GetFakeConcreteNotificationService()
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()

	ns.updateStatusSuccess(ctx, identifier, fakeMessage)
	assert.Equal(t, testutil.StatusNacked, fakeMessage.Status)
}

func TestNotificationService_UpdateSuccessBadStatus(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	fakes := GetFakeConcreteNotificationService()
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusComplete,
	}

	ns.updateStatusSuccess(ctx, identifier, fakeMessage)
	assert.Equal(t, testutil.StatusNacked, fakeMessage.Status)
	assert.Equal(t, notification.StatusComplete, fakes.db.Db[identifier].Status)
}

func TestNotificationService_AcquireExisting(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	fakes := GetFakeConcreteNotificationService()
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusUndelivered,
	}

	shouldContinue := ns.acquireNotificationLock(ctx, identifier, fakeMessage)
	assert.True(t, shouldContinue)
	assert.Equal(t, testutil.StatusNone, fakeMessage.Status)
	assert.Equal(t, notification.StatusProcessing, fakes.db.Db[identifier].Status)
}

func TestNotificationService_AcquireNonExistent(t *testing.T) {
	ctx := context.Background()
	fakes := GetFakeConcreteNotificationService()
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()

	shouldContinue := ns.acquireNotificationLock(ctx, identifier, fakeMessage)
	assert.False(t, shouldContinue)
	assert.Equal(t, testutil.StatusNacked, fakeMessage.Status)
}

func TestNotificationService_AcquireProcessing(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	fakes := GetFakeConcreteNotificationService()
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusProcessing,
	}

	shouldContinue := ns.acquireNotificationLock(ctx, identifier, fakeMessage)
	assert.False(t, shouldContinue)
	assert.Equal(t, testutil.StatusNacked, fakeMessage.Status)
}

func TestNotificationService_AcquireComplete(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	fakes := GetFakeConcreteNotificationService()
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusComplete,
	}

	shouldContinue := ns.acquireNotificationLock(ctx, identifier, fakeMessage)
	assert.False(t, shouldContinue)
	assert.Equal(t, testutil.StatusAcked, fakeMessage.Status)
}

func TestNotificationService_ProcessMessageSuccess(t *testing.T) {
	ctx := context.Background()
	fakes := GetFakeConcreteNotificationService()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusUndelivered,
	}

	ns.processMessage(ctx, fakeMessage)
	assert.Equal(t, testutil.StatusAcked, fakeMessage.Status)
	assert.Equal(t, notification.StatusComplete, fakes.db.Db[identifier].Status)
}

func TestNotificationService_ProcessMessageNoAcquire(t *testing.T) {
	ctx := context.Background()
	fakes := GetFakeConcreteNotificationService()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusProcessing,
	}

	ns.processMessage(ctx, fakeMessage)
	assert.Equal(t, testutil.StatusNacked, fakeMessage.Status)
}

func TestNotificationService_ProcessMessageTimeout(t *testing.T) {
	ctx := context.Background()
	fakes := GetFakeConcreteNotificationService()
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	ns := fakes.ns

	fakeNotification := testutil.GetFakeNotification()
	fakeMessage := testutil.GetFakeMessage(fakeNotification)
	identifier := fakeMessage.Identifier()
	fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusUndelivered,
	}
	fakes.notifier.SendNotificationError = context.DeadlineExceeded

	ns.processMessage(ctx, fakeMessage)
	assert.Equal(t, testutil.StatusNacked, fakeMessage.Status)
	assert.Equal(t, notification.StatusFailed, fakes.db.Db[identifier].Status)
}

func TestNotificationService_HandleNotifications(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fakes := GetFakeConcreteNotificationService(10)
	now := time.Now().UTC()
	expiryTime := now.Add(time.Hour)
	ns := fakes.ns

	fakeMessages := make([]*testutil.FakeMessage, 0, 10)

	for range 10 {
		fakeNotification := testutil.GetFakeNotification()
		fakeMessage := testutil.GetFakeMessage(fakeNotification)
		identifier := fakeMessage.Identifier()

		fakes.db.Db[identifier] = testutil.FakeRepositoryEntry{
			ExpiresAt: expiryTime,
			Status:    notification.StatusUndelivered,
		}
		fakes.consumer.AddMessage(fakeMessage)
		fakeMessages = append(fakeMessages, fakeMessage)
	}

	err := ns.HandleNotifications(ctx)
	assert.NoError(t, err)

	for _, m := range fakeMessages {
		assert.Equal(t, testutil.StatusAcked, m.Status)
		assert.Equal(t, notification.StatusComplete, fakes.db.Db[m.Identifier()].Status)
	}
}
