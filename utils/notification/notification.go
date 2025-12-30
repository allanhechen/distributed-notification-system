package notification

// RequestStatus are the possible states that notifications can be in
type RequestStatus int32

const (
	StatusUndelivered RequestStatus = 0
	StatusQueued      RequestStatus = 1
	StatusProcessing  RequestStatus = 2
	StatusComplete    RequestStatus = 3
	StatusFailed      RequestStatus = 4
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
	Identifier       string
	NotificationType DeviceType
	DeviceIdentifier string
	Message          string
}
