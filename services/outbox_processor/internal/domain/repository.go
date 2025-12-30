package domain

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

// Repository handles communication with the database.
type Repository interface {
	// GetUnprocessedNotifications returns messages in unprocessed or locked
	// but expired status
	GetUnprocessedNotifications(ctx context.Context, count uint) ([]notification.Notification, error)
	// UpdateNotificationStatuses updates the notifications specified within
	// updates to the provided statuses.
	UpdateNotificationStatuses(ctx context.Context, updates []StatusUpdate) error
}
