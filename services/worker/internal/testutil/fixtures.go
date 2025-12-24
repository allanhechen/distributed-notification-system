package testutil

import (
	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

// GetFakeNotification returns a domain.Notification with synthetic data for tests.
// The notification has a new UUID for Identifier, NotificationType set to notification.EmailDeviceType,
// a new UUID for DeviceIdentifier, and Message set to a 64-character random string.
func GetFakeNotification() domain.Notification {
	fakeIdentifier := uuid.New().String()
	fakeDeviceIdentifier := uuid.New().String()

	return domain.Notification{
		Identifier:       fakeIdentifier,
		NotificationType: notification.EmailDeviceType,
		DeviceIdentifier: fakeDeviceIdentifier,
		Message:          utils.GetRandomString(64),
	}
}
