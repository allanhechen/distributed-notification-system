package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	"github.com/allanhechen/distributed-notification-system/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyLayer interface {
	Handle(ctx context.Context, requestId uuid.UUID, cb func(context.Context, db.Querier) (int, []byte, error)) (int, []byte, error)
	HandleFailure(requestId uuid.UUID)
}

type ConcreteIdempotencyLayer struct {
	pool    *pgxpool.Pool
	getRepo func(db.Querier) repository.IdempotencyLayerRepo
}

func NewConcreteIdempotencyLayer(pool *pgxpool.Pool, getRepo func(db.Querier) repository.IdempotencyLayerRepo) IdempotencyLayer {
	return &ConcreteIdempotencyLayer{
		pool:    pool,
		getRepo: getRepo,
	}
}

func isValidHTTPStatus(code int) bool {
	return code >= 100 && code <= 599
}

var ErrNotJson = errors.New("idempotency: non-JSON response returned from callback")
var ErrInvalidStatus = errors.New("idempotency: invalid repsonse status returned from callback")
var ErrTransaction = errors.New("idempotency: could not interact with database transaction")
var ErrUser = errors.New("idempotency: error caused by the user")

const internalFailureStatus = http.StatusInternalServerError

// Handle wraps the entire handler (passed in cb) in a transaction. This
// function allows for the request's status and response to be captured
// and inserted atomically before responding to the client, enabling
// idempotency.
//
// The callback must return the status to be sent by the handler, and the
// JSON response to be sent in the body.
// All errors returned are expected to be handled by HandleFailure.
func (c *ConcreteIdempotencyLayer) Handle(ctx context.Context, requestId uuid.UUID, cb func(context.Context, db.Querier) (int, []byte, error)) (int, []byte, error) {
	fromCtx, err := utils.GetValuesFromContext(ctx)
	if err != nil {
		slog.Error("idempotency: failed to get values from request context")
		return 500, nil, err
	}
	logger := fromCtx.Logger

	trx, err := c.pool.Begin(ctx)
	if err != nil {
		logger.Error("idempotency: could not acquire transaction", "error", err)
		return internalFailureStatus, nil, ErrTransaction
	}
	defer trx.Rollback(ctx)

	q := db.New(trx)
	qtx := q.WithTx(trx)
	repo := c.getRepo(qtx)

	status, response, err := cb(ctx, qtx)
	if err != nil {
		logger.Error("idempotency: callback failed", "error", err)
		rollbackErr := trx.Rollback(ctx)
		if rollbackErr != nil {
			logger.Error("idempotency: error rolling back the transaction", "error", rollbackErr)
		}
		return 500, response, err
	}
	if response != nil && !json.Valid(response) {
		return 500, nil, ErrNotJson
	}
	if !isValidHTTPStatus(status) {
		return 500, nil, ErrInvalidStatus
	}
	if status >= 400 && status < 500 {
		return status, response, ErrUser
	}

	successParams := repository.UpdateRequestSuccessParams{
		RequestID:          requestId,
		CachedResponseCode: int32(status),
		CachedResponse:     response,
		ExpiresAt:          time.Now().Add(domain.LongRequestTtl).UTC(),
	}
	err = repo.UpdateRequestSuccess(ctx, successParams)
	if err != nil {
		logger.Error("idempotency: updating success", "error", err)
		rollbackErr := trx.Rollback(ctx)
		if rollbackErr != nil {
			logger.Error("idempotency: error rolling back the transaction", "error", rollbackErr)
		}
		return internalFailureStatus, nil, err
	}

	err = trx.Commit(ctx)
	if err != nil {
		logger.Error("idempotency: transaction failed", "error", err)

		return internalFailureStatus, nil, err
	}

	logger.Info("idempotency: transaction success")
	return status, response, nil
}

// HandleFailure handles failure cases returned by Handle by calling the
// repository's UpdateRequestFailed method. This method is separate from
// the request context because that may already be expired. It does not
// return any errors because any failures at this state cannot be handled.
func (c *ConcreteIdempotencyLayer) HandleFailure(requestId uuid.UUID) {
	slog.Info("idempotency: setting failed request status", "requestId", requestId)
	ctx, cancel := context.WithTimeout(context.Background(), domain.DatabaseTimeout)
	defer cancel()

	q := db.New(c.pool)
	repo := c.getRepo(q)
	err := repo.UpdateRequestFailed(ctx, requestId)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("idempotency: setting failed request status timed out", "requestId", requestId)
		}

		slog.Error("idempotency: setting request status failed for other reason", "requestId", requestId, "error", err)
	}
}
