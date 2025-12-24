package domain

import (
	"context"
	"errors"
)

// ErrAlreadyReplied denotes that a particular message has been replied
// to multiple times.
var ErrAlreadyReplied = errors.New("consumer: already replied to message")

// Message is an abstraction around a particular message to be handled
// from the Consumer.
type Message[T any] interface {
	Payload() T
	Ack(ctx context.Context) error
	Nack(ctx context.Context, requeue bool) error
	Identifier() string
}
