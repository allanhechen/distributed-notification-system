package domain

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

// MqService is an interface that abstracts communication with the message
// queue.
type MqService interface {
	// SendNotification sends a notification to the implementor's message queue.
	// It sends responses (ACK/NACK) to the given channel, intended to be used with
	// a ResponseService to update the statuses within the database.
	SendNotification(ctx context.Context, n notification.Notification, responses chan<- StatusUpdate) error
}
