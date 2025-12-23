package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdempotencyRepo is the repository that handles request idempotency.
type IdempotencyRepo interface {
	GetStoredRequest(ctx context.Context, requestId uuid.UUID) (*domain.IdempotentRequest, error)
	CreateStoredRequest(context.Context, CreateRequestParams) error
	UpdateRequestReprocess(context.Context, UpdateRequestReprocessParams) error
}

// pgx implementation of the idempotency repository
type PgxIdempotency struct {
	pool *pgxpool.Pool
}

// NewIdempotencyRepo creates a new IdempotencyRepo.
func NewIdempotencyRepo(pool *pgxpool.Pool) IdempotencyRepo {
	return &PgxIdempotency{
		pool: pool,
	}
}

// GetStoredRequest checks the database for a stored request with the
// given requestId.
//
// Returns ErrNoRows if no record is found for the given requestId.
// Callers should check for this error for determining if a request is
// new. Failed RequestStatusIds on the returned struct should also be
// taken into consideration for retrying requests.
func (p *PgxIdempotency) GetStoredRequest(ctx context.Context, requestId uuid.UUID) (*domain.IdempotentRequest, error) {
	q := db.New(p.pool)
	res, err := q.GetRequest(ctx, requestId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, fmt.Errorf("db: failed to retrieve stored request %s: %w", requestId, err)
	}

	domainRequest := db.IdempotentRequestToDomain(&res)
	return domainRequest, nil
}

type CreateRequestParams struct {
	RequestID       uuid.UUID
	UserID          uuid.UUID
	RequestStatusID idempotencyTypes.RequestStatus
	ExpiresAt       time.Time
}

// CreateStoredRequest creates a record for request passed in
// idempotentRequest.
//
// Returns ErrAlreadyExists if the request already exists. Callers should
// check for this error to ensure the request was correctly inserted.
func (p *PgxIdempotency) CreateStoredRequest(ctx context.Context, params CreateRequestParams) error {
	dbParams := db.CreateRequestParams{
		RequestID:       params.RequestID,
		UserID:          params.UserID,
		RequestStatusID: params.RequestStatusID,
		ExpiresAt:       params.ExpiresAt,
	}

	q := db.New(p.pool)
	err := q.CreateRequest(ctx, dbParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return ErrAlreadyExists
			}
		}

		return fmt.Errorf("db: failed to create request with id %s: %w", params.RequestID, err)
	}

	return nil
}

type UpdateRequestReprocessParams struct {
	ExpiresAt time.Time
	RequestID uuid.UUID
}

// UpdateFailedRequestProcessed marks an existing request as processing.
// This request must already exist, and be in either failed status or
// processing status with an expired timestamp.
//
// If either condition is not met, ErrNoRows is returned.
func (p *PgxIdempotency) UpdateRequestReprocess(ctx context.Context, params UpdateRequestReprocessParams) error {
	dbParams := db.UpdateRequestReprocessParams{
		ExpiresAt: params.ExpiresAt,
		RequestID: params.RequestID,
	}
	q := db.New(p.pool)
	count, err := q.UpdateRequestReprocess(ctx, dbParams)
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoRows
	}
	return nil
}
