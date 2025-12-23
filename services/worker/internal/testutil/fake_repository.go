package testutil

import (
	"context"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

type FakeRepositoryEntry struct {
	ExpiresAt time.Time
	Status    notification.RequestStatus
}

type FakeRepository struct {
	Db map[string]FakeRepositoryEntry
}

func GetFakeRepository() *FakeRepository {
	return &FakeRepository{
		Db: make(map[string]FakeRepositoryEntry),
	}
}

func (f *FakeRepository) Acquire(_ context.Context, identifier string, expiryTime time.Time) error {
	entry, ok := f.Db[identifier]
	if !ok {
		return domain.ErrNoRows
	}
	if entry.Status == notification.StatusComplete {
		return domain.ErrAlreadyComplete
	}

	now := time.Now().UTC()
	if entry.Status == notification.StatusProcessing && expiryTime.After(now) {
		return domain.ErrAlreadyProcessing
	}

	f.Db[identifier] = FakeRepositoryEntry{
		ExpiresAt: expiryTime,
		Status:    notification.StatusProcessing,
	}
	return nil
}

func (f *FakeRepository) MarkSuccess(_ context.Context, identifier string) error {
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

func (f *FakeRepository) MarkFailure(_ context.Context, identifier string) error {
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
