package domain

import "github.com/allanhechen/distributed-notification-system/utils/notification"

// TODO: update this type when the database schema is merged
// Notification is an abstraction around a notification sent through a
// message queue. Intended to be the payload of a Message.
type Notification struct {
	Identifier       string
	NotificationType notification.DeviceType
	DeviceIdentifier string
	Message          string
}
