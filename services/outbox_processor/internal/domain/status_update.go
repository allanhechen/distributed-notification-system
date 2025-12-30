package domain

import (
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

// StatusUpdate contains the updated state of a notification after a
// delivery attempt.
type StatusUpdate struct {
	Identifier  uuid.UUID
	FinalStatus notification.RequestStatus
}
