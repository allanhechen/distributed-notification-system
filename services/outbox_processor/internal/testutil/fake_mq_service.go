package testutil

import (
	"context"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

type FakeMqService struct {
	failIdentifiers map[uuid.UUID]struct{}
}

func GetFakeMqService() domain.MqService {
	return &FakeMqService{
		failIdentifiers: make(map[uuid.UUID]struct{}),
	}
}

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

func (f *FakeMqService) AddFailedIdentifier(u uuid.UUID) {
	f.failIdentifiers[u] = struct{}{}
}

func (f *FakeMqService) RemoveFailedIdentifier(u uuid.UUID) {
	delete(f.failIdentifiers, u)
}
