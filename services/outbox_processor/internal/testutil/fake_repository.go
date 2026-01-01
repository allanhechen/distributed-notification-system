package testutil

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

// FakeRepository is an in-memory implementation of the repository.
type FakeRepository struct {
	mu              sync.Mutex
	Entries         map[uuid.UUID]notification.Notification
	QueueRetryLimit int
}

// GetFakeRepository returns an instance of the fake repository.
func GetFakeRepository() *FakeRepository {
	return &FakeRepository{
		Entries:         make(map[uuid.UUID]notification.Notification),
		QueueRetryLimit: notification.DefaultQueueRetryLimit,
	}
}

// GetUnprocessedNotifications returns all the notifications eligible to
// be queued from memory.
func (f *FakeRepository) GetUnprocessedNotifications(_ context.Context, count int) ([]notification.Notification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// get result container
	r := make([]notification.Notification, 0, count)
	now := time.Now().UTC()

	// loop through entries until full
	for k, v := range f.Entries {
		if len(r) == int(count) {
			break
		}

		// fetch undelivered notifications with queue attempts remaining
		if v.Status == notification.StatusUndelivered &&
			now.After(v.LockExpiryTime) &&
			v.FailedQueueAttempts < f.QueueRetryLimit {

			v.LockExpiryTime = now.Add(domain.MessageLockDuration)
			r = append(r, v)
			f.Entries[k] = v
		}
	}

	return r, nil
}

// UpdateNotificationStatuses updates the notifications to the provided
// statuses.
func (f *FakeRepository) UpdateNotificationStatuses(_ context.Context, updates []domain.StatusUpdate) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var errs []error

	for _, u := range updates {
		n, ok := f.Entries[u.Identifier]
		if !ok {
			errs = append(errs, fmt.Errorf("notification %s: %w", u.Identifier, domain.ErrNonExistent))
			continue
		}

		// if it previously succeeded we don't care
		if n.Status == notification.StatusQueued {
			continue
		}

		n.LockExpiryTime = time.Time{}
		n.Status = u.FinalStatus
		if u.FinalStatus == notification.StatusUndelivered {
			n.FailedQueueAttempts = n.FailedQueueAttempts + 1
		}
		f.Entries[u.Identifier] = n
	}

	return errors.Join(errs...)
}
