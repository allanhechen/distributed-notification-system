package domain

import "github.com/allanhechen/distributed-notification-system/utils/notification"

type Notification struct {
	Identifier       string
	NotificationType notification.DeviceType
	DeviceIdentifier string
	Message          string
}
