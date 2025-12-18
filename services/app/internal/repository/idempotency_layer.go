package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// IdempotencyLayerRepo handles ending requests.
type IdempotencyLayerRepo interface {
	UpdateRequestFailed(ctx context.Context, requestid uuid.UUID) error
	UpdateRequestSuccess(context.Context, UpdateRequestSuccessParams) error
}

// Querier implementation of the idempotency repository
type QuerierIdempotency struct {
	querier db.Querier
}

// NewIdempotencyRepo creates a new IdempotencyLayerRepo.
func NewIdempotencyLayerRepo(q db.Querier) IdempotencyLayerRepo {
	return &QuerierIdempotency{
		querier: q,
	}
}

// UpdateRequestFailed marks the selected record as failed. This record
// must already be expired (current time > record expiry time), and also
// exist in the database.
//
// If either condition is not met, ErrNoRows is returned.
func (q *QuerierIdempotency) UpdateRequestFailed(ctx context.Context, requestId uuid.UUID) error {
	count, err := q.querier.UpdateRequestFailed(ctx, requestId)
	if err != nil {
		return fmt.Errorf("db: failed to mark request with requestId %s as failed: %w", requestId, err)
	}

	if count == 0 {
		return ErrNoRows
	}
	return nil
}

type UpdateRequestSuccessParams struct {
	ExpiresAt          time.Time
	CachedResponseCode int32
	CachedResponse     []byte
	RequestID          uuid.UUID
}

// UpdateRequestSuccess marks an existing request as successful, and
// sets the row expiry time to the time provided in the params. This
// record must already exist in the database, and currently be in
// processing status.
//
// If either condition is not met, ErrNoRows is returned.
func (q *QuerierIdempotency) UpdateRequestSuccess(ctx context.Context, params UpdateRequestSuccessParams) error {
	dbParams := db.UpdateRequestSuccessParams{
		ExpiresAt: params.ExpiresAt,
		CachedResponseCode: pgtype.Int4{
			Int32: params.CachedResponseCode,
			Valid: true,
		},
		CachedResponse: params.CachedResponse,
		RequestID:      params.RequestID,
	}
	count, err := q.querier.UpdateRequestSuccess(ctx, dbParams)
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoRows
	}
	return nil
}
