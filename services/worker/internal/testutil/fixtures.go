package testutil

import (
	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

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
