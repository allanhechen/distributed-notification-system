package domain

import "context"

// Notifier is an abstraction around a particular notification. A notifier
// performs the actions required to deliver messages.
type Notifier interface {
	SendNotification(context.Context, Notification) error
}
