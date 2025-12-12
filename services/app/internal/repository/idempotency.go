package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Idempotency is the repository that handles request idempotency.
type Idempotency interface {
	GetStoredRequest(context.Context, uuid.UUID) (*db.IdempotentRequest, error)
}

// pgx implementation of the idempotency repository
type PgxIdempotency struct {
	pool *pgxpool.Pool
}

// GetStoredRequest checks the database for a stored request with the
// given requestId.
//
// Returns ErrNoRows if no record is found for the given requestId.
// Callers should check for this error for determining if a request is
// new. Failed RequestStatusIds on the returned struct should also be
// taken into consideration for retrying requests.
func (i *PgxIdempotency) GetStoredRequest(ctx context.Context, requestId uuid.UUID) (*db.IdempotentRequest, error) {
	q := db.New(i.pool)
	res, err := q.GetRequestStatus(ctx, requestId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, fmt.Errorf("db: failed to retrieve stored request %s: %w", requestId, err)
	}

	return &res, nil
}
