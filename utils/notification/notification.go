package notification

// RequestStatus are the possible states that notitications can be in
type RequestStatus int32

const (
	StatusUndelivered RequestStatus = 0
	StatusQueued      RequestStatus = 1
	StatusProcessing  RequestStatus = 2
	StatusComplete    RequestStatus = 3
	StatusFailed      RequestStatus = 4
)

type DeviceType int32

const (
	IosDeviceType     DeviceType = 0
	AndroidDeviceType DeviceType = 1
	EmailDeviceType   DeviceType = 2
)
