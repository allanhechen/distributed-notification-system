package testutil

import (
	"context"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

// FakeMqService is a fake implementation of MqService.
type FakeMqService struct {
	failIdentifiers map[uuid.UUID]struct{}
}

// GetFakeMqService returns an instance of FakeMqService.
func GetFakeMqService() *FakeMqService {
	return &FakeMqService{
		failIdentifiers: make(map[uuid.UUID]struct{}),
	}
}

// SendNotification pretends to send a notification with the message
// queue, and returns responses to the provided channel. SendNotification
// also handles business logic to determine the final state of the given
// notifications.
//
// A notification is "failed" to be sent if it exists within the failed
// set contained within memory.
func (f *FakeMqService) SendNotification(_ context.Context, n notification.Notification, maxQueueAttempts uint, responses chan<- domain.StatusUpdate) error {
	resp := domain.StatusUpdate{
		Identifier:     n.Identifier,
		LockExpiryTime: time.Time{},
	}
	_, ok := f.failIdentifiers[n.Identifier]
	if !ok {
		resp.FinalStatus = notification.StatusQueued
		resp.FailedQueueAttempts = n.FailedQueueAttempts
	} else if n.FailedQueueAttempts+1 >= maxQueueAttempts {
		resp.FinalStatus = notification.StatusFailed
		resp.FailedQueueAttempts = n.FailedQueueAttempts + 1
	} else {
		resp.FinalStatus = notification.StatusUndelivered
		resp.FailedQueueAttempts = n.FailedQueueAttempts + 1
	}

	responses <- resp
	return nil
}

// AddFailedIdentifier adds additional notifications that will fail to
// deliver.
func (f *FakeMqService) AddFailedIdentifier(u uuid.UUID) {
	f.failIdentifiers[u] = struct{}{}
}

// AddFailedIdentifier removes notification failure markers, allowing
// once-failing notifications to succeed again.
func (f *FakeMqService) RemoveFailedIdentifier(u uuid.UUID) {
	delete(f.failIdentifiers, u)
}
