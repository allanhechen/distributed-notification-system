package domain

import "context"

// OutboxService contains the core business logic of the outbox processor.
// It is intended to utilize the MqService, the Repository, and the
// StatusService.
type OutboxService interface {
	// HandleMessages handles a continuous stream of messages with a
	// shared-nothing architecture.
	HandleMessages(ctx context.Context) error
}
