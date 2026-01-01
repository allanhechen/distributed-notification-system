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
	// It sends a message to the returned channel when done
	//
	// Returned errors exist only for logging purposes.
	SendNotification(ctx context.Context, n notification.Notification, responses chan<- StatusUpdate) (done <-chan struct{}, err error)
}
