package services

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

// RabbitMqNotification is a concrete implementation of the Message
// interface for RabbitMQ messages and domain notifications.
type RabbitMqNotification struct {
	payload        domain.Notification
	identifier     string
	alreadyReplied bool
	ackFn          func(ctx context.Context) error
	nackFn         func(ctx context.Context, requeue bool) error
}

// Payload retrieves the payload of this message.
func (r *RabbitMqNotification) Payload() domain.Notification {
	return r.payload
}

// Identifier retrieves the identifier of this message.
func (r *RabbitMqNotification) Identifier() string {
	return r.identifier
}

// Ack sends an ACK to RabbitMQ by calling the ackFn. Does not requeue.
func (r *RabbitMqNotification) Ack(ctx context.Context) error {
	if r.alreadyReplied {
		return domain.ErrAlreadyReplied
	}
	return r.ackFn(ctx)
}

// Nack sends a NACK to RabbitMQ by cakking the nackFn. Requeue optional.
func (r *RabbitMqNotification) Nack(ctx context.Context, requeue bool) error {
	if r.alreadyReplied {
		return domain.ErrAlreadyReplied
	}
	return r.nackFn(ctx, requeue)
}
