package domain

import "github.com/google/uuid"

// StatusUpdate contains the identifier of a notification along with an
// ACK or NACK status. True for ACK, false for NACK.
type StatusUpdate struct {
	Identifier uuid.UUID
	AckStatus  bool
}
