package testutil

import (
	"context"
	"sync"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

type FakeNotifier struct {
	mu                    sync.Mutex
	ReceivedMessages      []domain.Notification
	SendNotificationError error
}

func GetFakeNotifier() *FakeNotifier {
	return &FakeNotifier{
		ReceivedMessages: make([]domain.Notification, 0),
	}
}

func (f *FakeNotifier) SendNotification(_ context.Context, notification domain.Notification) error {
	if f.SendNotificationError != nil {
		return f.SendNotificationError
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.ReceivedMessages = append(f.ReceivedMessages, notification)

	return nil
}
