package testutil

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

type FakeNotifier struct {
	ReceivedMessages      []domain.Notification
	SendNotificationError error
}

// GetFakeNotifier returns a new FakeNotifier with ReceivedMessages initialized to an empty slice.
// The returned FakeNotifier can be used in tests to capture notifications and optionally simulate send failures via SendNotificationError.
func GetFakeNotifier() *FakeNotifier {
	return &FakeNotifier{
		ReceivedMessages: make([]domain.Notification, 0),
	}
}

func (f *FakeNotifier) SendNotification(_ context.Context, notification domain.Notification) error {
	if f.SendNotificationError != nil {
		return f.SendNotificationError
	}

	f.ReceivedMessages = append(f.ReceivedMessages, notification)

	return nil
}