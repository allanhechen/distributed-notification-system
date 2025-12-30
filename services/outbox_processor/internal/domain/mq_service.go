package domain

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

// MqService is an interface that abstracts communication with the message
// queue.
type MqService interface {
	// SendNotification sends a notification to the implementor's message queue.
	// It sends updated statuses to the given channel.
	SendNotification(ctx context.Context, n notification.Notification, responses chan<- StatusUpdate) error
}
