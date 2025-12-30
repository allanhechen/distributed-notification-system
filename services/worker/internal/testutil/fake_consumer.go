package testutil

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

type FakeConsumer struct {
	jobs chan domain.Message[notification.Notification]
}

// GetFakeConsumer creates a FakeConsumer with its internal jobs channel buffered to a capacity of 100.
// The returned FakeConsumer is intended for tests and lets callers enqueue notification.Notification messages for consumption.
func GetFakeConsumer() *FakeConsumer {
	return &FakeConsumer{
		jobs: make(chan domain.Message[notification.Notification], 100),
	}
}

func (f *FakeConsumer) Consume(_ context.Context) (<-chan domain.Message[notification.Notification], error) {
	return f.jobs, nil
}

func (f *FakeConsumer) AddMessage(message domain.Message[notification.Notification]) {
	f.jobs <- message
}
