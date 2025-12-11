package types

type RequestStatus int

const (
	StatusProcessing RequestStatus = iota
	StatusComplete
	StatusFailed
)
