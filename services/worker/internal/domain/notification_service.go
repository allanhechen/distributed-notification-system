package domain

import "context"

// NotificationService handles sending notifications
type NotificationService interface {
	HandleNotifications(context.Context) error
}
