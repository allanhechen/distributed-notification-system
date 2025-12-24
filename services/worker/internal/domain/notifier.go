package domain

import "context"

type Notifier interface {
	SendNotification(context.Context, Notification) error
}
