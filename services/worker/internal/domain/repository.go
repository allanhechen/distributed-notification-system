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

var ErrNoRows = errors.New("repository: no rows were found")
var ErrAlreadyProcessing = errors.New("repository: notification is already processing")
var ErrAlreadyComplete = errors.New("repository: notification is already complete")

type Repository interface {
	Acquire(ctx context.Context, identifier string, expiryTime time.Time) error
	MarkSuccess(ctx context.Context, identifier string) error
	MarkFailure(ctx context.Context, identifier string) error
}
