package domain

import (
	"context"
)

type Consumer[T any] interface {
	Consume(context.Context) (<-chan Message[T], error)
}
