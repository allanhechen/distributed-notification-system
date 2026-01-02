package testutil

import (
	"sync"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

// FakeMqService is a fake implementation of MqService.
type FakeMqService struct {
	mu              sync.Mutex
	failIdentifiers map[uuid.UUID]struct{}
}

// GetFakeMqService returns an instance of FakeMqService.
func GetFakeMqService() *FakeMqService {
	return &FakeMqService{
		failIdentifiers: make(map[uuid.UUID]struct{}),
	}
}

// SendNotification pretends to send a notification with the message
// queue, and returns responses to the provided channel.
//
// A notification is "failed" to be sent if it exists within the failed
// set contained within memory.
func (f *FakeMqService) SendNotification(n notification.Notification, responses chan<- domain.StatusUpdate) (<-chan struct{}, error) {
	resp := domain.StatusUpdate{
		Identifier: n.Identifier,
	}
	f.mu.Lock()
	_, ok := f.failIdentifiers[n.Identifier]
	if !ok {
		resp.FinalStatus = notification.StatusQueued
	} else {
		resp.FinalStatus = notification.StatusUndelivered
	}
	f.mu.Unlock()

	responses <- resp
	done := make(chan struct{})
	go func() {
		defer close(done)
		done <- struct{}{}
	}()

	return done, nil
}

func (f *FakeMqService) Start() {}
func (f *FakeMqService) Stop()  {}

// AddFailedIdentifier adds additional notifications that will fail to
// deliver.
func (f *FakeMqService) AddFailedIdentifier(u uuid.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failIdentifiers[u] = struct{}{}
}

// RemoveFailedIdentifier removes notification failure markers, allowing
// once-failing notifications to succeed again.
func (f *FakeMqService) RemoveFailedIdentifier(u uuid.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.failIdentifiers, u)
}
