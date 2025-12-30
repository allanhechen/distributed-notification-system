package domain

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

// Notifier is an abstraction around a particular notification. A notifier
// performs the actions required to deliver messages.
type Notifier interface {
	SendNotification(context.Context, notification.Notification) error
}
