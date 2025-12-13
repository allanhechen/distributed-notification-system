package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Idempotency is the repository that handles request idempotency.
type Idempotency interface {
	GetStoredRequest(context.Context, uuid.UUID) (*db.IdempotentRequest, error)
	CreateStoredRequest(context.Context, db.CreateRequestStatusParams) error
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
func (p *PgxIdempotency) GetStoredRequest(ctx context.Context, requestId uuid.UUID) (*db.IdempotentRequest, error) {
	q := db.New(p.pool)
	res, err := q.GetRequestStatus(ctx, requestId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, fmt.Errorf("db: failed to retrieve stored request %s: %w", requestId, err)
	}

	return &res, nil
}

// CreateStoredRequest creates a record for request passed in
// idempotentRequest.
//
// Returns ErrAlreadyExists if the request already exists. Callers should
// check for this error to ensure the request was correctly inserted.
func (p *PgxIdempotency) CreateStoredRequest(ctx context.Context, idempotentRequest db.CreateRequestStatusParams) error {
	q := db.New(p.pool)
	err := q.CreateRequestStatus(ctx, idempotentRequest)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return ErrAlreadyExists
			}
		}

		return fmt.Errorf("db: failed to create request with id %s: %w", idempotentRequest.RequestID, err)
	}

	return nil
}
