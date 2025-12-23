package testutil

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/google/uuid"
)

type FakeRequestStatus = int

const (
	StatusNone FakeRequestStatus = iota
	StatusAcked
	StatusNacked
)

type FakeMessage struct {
	identifier string
	payload    domain.Notification
	Status     FakeRequestStatus
}

func GetFakeMessage(notification domain.Notification) *FakeMessage {
	return &FakeMessage{
		identifier: uuid.New().String(),
		payload:    notification,
		Status:     StatusNone,
	}
}

func (f *FakeMessage) Payload() domain.Notification {
	return f.payload
}

func (f *FakeMessage) Ack(_ context.Context) error {
	if f.Status != StatusNone {
		return domain.ErrAlreadyReplied
	}

	f.Status = StatusAcked
	return nil
}

func (f *FakeMessage) Nack(_ context.Context, _ bool) error {
	if f.Status != StatusNone {
		return domain.ErrAlreadyReplied
	}

	f.Status = StatusNacked
	return nil
}

func (f *FakeMessage) Identifier() string {
	return f.identifier
}
