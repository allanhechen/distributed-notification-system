package domain

import (
	"context"
	"errors"
	"time"

	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

// ErrNonExistent is returned when a specified notification does not exist
// in the database.
var ErrNonExistent = errors.New("repository: notification does not exist")

// MessageLockDuration is the duration to lock notifications during an
// attempt to send them to the queue.
const MessageLockDuration = time.Duration(5 * time.Second)

// Repository handles communication with the database.
type Repository interface {
	// GetUnprocessedNotifications returns messages in unprocessed or locked
	// but expired status
	GetUnprocessedNotifications(ctx context.Context, count int) ([]notification.Notification, error)
	// UpdateNotificationStatuses updates the notifications specified within
	// updates to the provided statuses.
	UpdateNotificationStatuses(ctx context.Context, updates []StatusUpdate) error
}
