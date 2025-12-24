package testutil

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

type FakeConsumer struct {
	jobs chan domain.Message[domain.Notification]
}

// GetFakeConsumer creates a FakeConsumer with its internal jobs channel buffered to a capacity of 100.
// The returned FakeConsumer is intended for tests and lets callers enqueue domain.Notification messages for consumption.
func GetFakeConsumer() *FakeConsumer {
	return &FakeConsumer{
		jobs: make(chan domain.Message[domain.Notification], 100),
	}
}

func (f *FakeConsumer) Consume(_ context.Context) (<-chan domain.Message[domain.Notification], error) {
	return f.jobs, nil
}

func (f *FakeConsumer) AddMessage(message domain.Message[domain.Notification]) {
	f.jobs <- message
}
