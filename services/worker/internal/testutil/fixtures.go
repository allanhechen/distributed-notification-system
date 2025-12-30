package testutil

import (
	"github.com/allanhechen/distributed-notification-system/utils"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

// GetFakeNotification returns a notification.Notification with synthetic data for tests.
// The notification has a new UUID for Identifier, NotificationType set to notification.EmailDeviceType,
// a new UUID for DeviceIdentifier, and Message set to a 64-character random string.
func GetFakeNotification() notification.Notification {
	fakeIdentifier := uuid.New()
	fakeDeviceIdentifier := uuid.New()

	return notification.Notification{
		Identifier:       fakeIdentifier,
		NotificationType: notification.EmailDeviceType,
		DeviceIdentifier: fakeDeviceIdentifier,
		Message:          utils.GetRandomString(64),
	}
}
