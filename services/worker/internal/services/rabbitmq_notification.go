package services

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
)

type RabbitmqNotification struct {
	payload        domain.Notification
	identifier     string
	alreadyReplied bool
	ackFn          func(ctx context.Context) error
	nackFn         func(ctx context.Context, requeue bool) error
}

func (r *RabbitmqNotification) Payload() domain.Notification {
	return r.payload
}

func (r *RabbitmqNotification) Identifier() string {
	return r.Identifier()
}

func (r *RabbitmqNotification) Ack(ctx context.Context) error {
	if r.alreadyReplied {
		return domain.ErrAlreadyReplied
	}
	return r.ackFn(ctx)
}

func (r *RabbitmqNotification) Nack(ctx context.Context, requeue bool) error {
	if r.alreadyReplied {
		return domain.ErrAlreadyReplied
	}
	return r.nackFn(ctx, requeue)
}
