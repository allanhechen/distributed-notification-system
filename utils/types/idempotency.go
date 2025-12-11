package types

// Request status type for the idempotent_request_statuses table in the
// application server.
//
// sqlc generates code pointing to this status.
type RequestStatus int

const (
	StatusProcessing RequestStatus = iota
	StatusComplete
	StatusFailed
)
