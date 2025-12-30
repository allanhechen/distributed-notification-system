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
	mu      sync.Mutex
	entries map[uuid.UUID]notification.Notification
}

// GetFakeRepository returns an instance of the fake repository.
func GetFakeRepository() *FakeRepository {
	return &FakeRepository{
		entries: make(map[uuid.UUID]notification.Notification),
	}
}

// GetUnprocessedNotifications returns all the notifications eligible to
// be queued from memory.
func (f *FakeRepository) GetUnprocessedNotifications(_ context.Context, count uint) ([]notification.Notification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// get result container
	r := make([]notification.Notification, 0, count)
	now := time.Now().UTC()

	// loop through entries until full
	for _, v := range f.entries {
		if len(r) == int(count) {
			break
		}

		// fetch undelivered or locked and expired entries
		if v.Status == notification.StatusUndelivered || (v.Status == notification.StatusLocked && now.After(v.LockExpiryTime)) {
			r = append(r, v)
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
		n, ok := f.entries[u.Identifier]
		if !ok {
			errs = append(errs, fmt.Errorf("notification %s: %w", u.Identifier, domain.ErrNonExistent))
			continue
		}

		// if it previously succeeded we don't care
		if n.Status == notification.StatusQueued {
			continue
		}

		// if another iteration took place after this one we don't care
		if n.LockExpiryTime != u.LockExpiryTime {
			continue
		}

		n.FailedQueueAttempts = u.FailedQueueAttempts
		n.LockExpiryTime = u.LockExpiryTime
		n.Status = u.FinalStatus
		f.entries[u.Identifier] = n
	}

	return errors.Join(errs...)
}
