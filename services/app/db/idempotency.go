package db

import (
	"context"

	"github.com/allanhechen/distributed-notification-system/services/app/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Idempotency struct {
	pool *pgxpool.Pool
}

func (i *Idempotency) GetStoredRequest(requestId uuid.UUID) (*generated.IdempotentRequest, error) {
	ctx := context.Background()
	tx, err := i.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	q := generated.New(tx)
	res, err := q.GetRequestStatus(ctx, requestId)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
