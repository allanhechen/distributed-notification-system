package domain

import (
	"context"
	"errors"
	"time"
)

const (
	ProcessingLockTime    = 15 * time.Second
	JobProcessingTime     = 5 * time.Second
	ErrorProcessingTime   = 5 * time.Second
	SuccessProcessingTime = 5 * time.Second
)

// ErrNoRows signifies that no rows were found or updated matching the
// criteria
var ErrNoRows = errors.New("repository: no rows were found")

// ErrAlreadyProcessing signifies that the message is currently being
// processed and is not yet expired.
var ErrAlreadyProcessing = errors.New("repository: notification is already processing")

// ErrAlreadyComplete signifies that a message has already finished being
// processed
var ErrAlreadyComplete = errors.New("repository: notification is already complete")

// Repository is an abstraction around the database.
type Repository interface {
	Acquire(ctx context.Context, identifier string, expiryTime time.Time) error
	MarkSuccess(ctx context.Context, identifier string) error
	MarkFailure(ctx context.Context, identifier string) error
}
