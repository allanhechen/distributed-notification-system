package types

// Request status type for the idempotent_request_statuses table in the
// application server.
//
// sqlc generates code pointing to this status.
type RequestStatus int32

const (
	StatusProcessing RequestStatus = 0
	StatusComplete   RequestStatus = 1
	StatusFailed     RequestStatus = 2
)
