package domain

import (
	"errors"

	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

var ErrAlreadyClosed = errors.New("mq service: mq connection already closed")

// MqService is an interface that abstracts communication with the message
// queue.
type MqService interface {
	// SendNotification sends a notification to the implementor's message queue.
	// It sends updated statuses to the given channel.
	// It sends a message to the returned channel when done
	//
	// Returned errors exist only for logging purposes.
	SendNotification(n notification.Notification, responses chan<- StatusUpdate) (done <-chan struct{}, err error)
}
