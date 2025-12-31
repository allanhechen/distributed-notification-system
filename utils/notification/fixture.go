package notification

import (
	"time"

	"github.com/allanhechen/distributed-notification-system/utils"
	"github.com/google/uuid"
)

// GetFakeNotification returns a notification.Notification with synthetic data for tests.
// The notification has a new UUID for Identifier, NotificationType set to notification.EmailDeviceType,
// a new UUID for DeviceIdentifier, and Message set to a 64-character random string.
//
// Some additional information can be passed by params.
func GetFakeNotification(
	notificationType DeviceType,
	status RequestStatus,
	failedQueueAttempts int,
	lockExpiryTime time.Time,
) Notification {
	fakeIdentifier := uuid.New()
	fakeDeviceIdentifier := uuid.New()

	return Notification{
		Identifier:          fakeIdentifier,
		NotificationType:    notificationType,
		Status:              status,
		DeviceIdentifier:    fakeDeviceIdentifier,
		Message:             utils.GetRandomString(64),
		FailedQueueAttempts: failedQueueAttempts,
		LockExpiryTime:      lockExpiryTime,
	}
}
