package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

type FakeRepositoryEntry struct {
	ExpiresAt time.Time
	Status    notification.RequestStatus
}

type FakeRepository struct {
	mu sync.Mutex
	Db map[uuid.UUID]FakeRepositoryEntry
}

// GetFakeRepository creates a new FakeRepository with Db initialized as an empty map.
func GetFakeRepository() *FakeRepository {
	return &FakeRepository{
		Db: make(map[uuid.UUID]FakeRepositoryEntry),
	}
}

func (f *FakeRepository) Acquire(_ context.Context, identifier uuid.UUID, expiryTime time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.Db[identifier]
	if !ok {
		return domain.ErrNoRows
	}
	if entry.Status == notification.StatusComplete {
		return domain.ErrAlreadyComplete
	}

	now := time.Now().UTC()
	if entry.Status == notification.StatusProcessing && entry.ExpiresAt.After(now) {
		return domain.ErrAlreadyProcessing
	}

	f.Db[identifier] = FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusProcessing,
	}
	return nil
}

func (f *FakeRepository) MarkSuccess(_ context.Context, identifier uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.Db[identifier]
	if !ok {
		return domain.ErrNoRows
	} else if existing.Status == notification.StatusComplete {
		return domain.ErrAlreadyComplete
	}

	now := time.Now().UTC()
	f.Db[identifier] = FakeRepositoryEntry{
		ExpiresAt: now.Add(time.Hour), // for testing only
		Status:    notification.StatusComplete,
	}
	return nil
}

func (f *FakeRepository) MarkFailure(_ context.Context, identifier uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.Db[identifier]
	if !ok {
		return domain.ErrNoRows
	}

	now := time.Now().UTC()
	f.Db[identifier] = FakeRepositoryEntry{
		ExpiresAt: now.Add(time.Hour), // for testing only
		Status:    notification.StatusFailed,
	}
	return nil
}
