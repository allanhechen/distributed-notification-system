package domain

import (
	"context"
	"errors"
)

var ErrAlreadyReplied = errors.New("consumer: already replied to message")

type Message[T any] interface {
	Payload() T
	Ack(ctx context.Context) error
	Nack(ctx context.Context, requeue bool) error
	Identifier() string
}
