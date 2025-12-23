package testutil

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

type FakeConsumer struct {
	jobs chan domain.Message[domain.Notification]
}

func GetFakeConsumer() *FakeConsumer {
	return &FakeConsumer{
		jobs: make(chan domain.Message[domain.Notification]),
	}
}

func (f *FakeConsumer) Consume(_ context.Context) (<-chan domain.Message[domain.Notification], error) {
	return f.jobs, nil
}

func (f *FakeConsumer) AddMessage(message domain.Message[domain.Notification]) {
	f.jobs <- message
}
