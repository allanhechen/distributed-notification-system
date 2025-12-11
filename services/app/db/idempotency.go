package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/allanhechen/distributed-notification-system/services/app/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Idempotency is the group of queries related to handling request
// idempotency.
type Idempotency struct {
	pool *pgxpool.Pool
}

// GetStoredRequest checks the database for a stored request with the
// given requestId.
//
// Returns an error that wraps ErrNoRows if no record is found for the
// given requestId. Callers should check for this error for determining
// if a request is new. Failed RequestStatusIds on the returned struct
// should also be taken into consideration for retrying requests.
func (i *Idempotency) GetStoredRequest(ctx context.Context, requestId uuid.UUID) (*generated.IdempotentRequest, error) {
	q := generated.New(i.pool)
	res, err := q.GetRequestStatus(ctx, requestId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, fmt.Errorf("db: failed to retrieve stored request %s: %w", requestId, err)
	}

	return &res, nil
}
