package testutil

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

type FakeNotifier struct {
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

	f.ReceivedMessages = append(f.ReceivedMessages, notification)

	return nil
}
