package domain

import (
	"context"
)

// Consumer is the abstraction around a message queue. It is intended to
// be used with dependency injection, and returns a channel representing
// messages relevant to the current worker.
type Consumer[T any] interface {
	Consume(context.Context) (<-chan Message[T], error)
}
