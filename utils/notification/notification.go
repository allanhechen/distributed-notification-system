package notification

import (
	"time"

	"github.com/google/uuid"
)

const DefaultQueueRetryLimit = 3

// RequestStatus are the possible states that notifications can be in
type RequestStatus int32

const (
	// StatusUndelivered is a notification never sent to the queue.
	StatusUndelivered RequestStatus = 0

	// StatusQueued is a notification waiting in the message queue.
	StatusQueued RequestStatus = 1

	// StatusProcessing is a notification being processed by a worker.
	StatusProcessing RequestStatus = 2

	// StatusComplete is a notification successfully delivered by a
	// worker.
	StatusComplete RequestStatus = 3

	// StatusFailed is a notification that could not be delivered.
	StatusFailed RequestStatus = 4
)

// DeviceType represents the notification targets we support
type DeviceType int32

const (
	IosDeviceType     DeviceType = 0
	AndroidDeviceType DeviceType = 1
	EmailDeviceType   DeviceType = 2
)

// TODO: update this type when the database schema is merged
// Notification is an abstraction around a notification sent through a
// message queue. Intended to be the payload of a Message.
type Notification struct {
	Identifier          uuid.UUID
	NotificationType    DeviceType
	Status              RequestStatus
	DeviceIdentifier    uuid.UUID
	Message             string
	FailedQueueAttempts int
	LockExpiryTime      time.Time
}
