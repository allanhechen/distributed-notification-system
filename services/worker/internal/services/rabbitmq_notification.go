package services

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

type RabbitMqNotification struct {
	payload        domain.Notification
	identifier     string
	alreadyReplied bool
	ackFn          func(ctx context.Context) error
	nackFn         func(ctx context.Context, requeue bool) error
}

func (r *RabbitMqNotification) Payload() domain.Notification {
	return r.payload
}

func (r *RabbitMqNotification) Identifier() string {
	return r.identifier
}

func (r *RabbitMqNotification) Ack(ctx context.Context) error {
	if r.alreadyReplied {
		return domain.ErrAlreadyReplied
	}
	return r.ackFn(ctx)
}

func (r *RabbitMqNotification) Nack(ctx context.Context, requeue bool) error {
	if r.alreadyReplied {
		return domain.ErrAlreadyReplied
	}
	return r.nackFn(ctx, requeue)
}
