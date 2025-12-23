package internal

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	domainTestutil "github.com/allanhechen/distributed-notification-system/services/app/internal/domain/testutil"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository/testutil"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/testutils"
	"github.com/allanhechen/distributed-notification-system/utils"
	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyLayer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = context.WithValue(ctx, utils.RequestIdKey, domainTestutil.RequestId)
	ctx = context.WithValue(ctx, utils.UserIdKey, domainTestutil.UserId)
	l := &utils.LogState{}
	ctx = context.WithValue(ctx, utils.Logger, slog.Default())
	ctx = context.WithValue(ctx, utils.LoggedState, l)

	dbContainer, err := testutils.GetCrdbDatabaseContainer(ctx)
	require.NoError(t, err)
	defer dbContainer.Container.Terminate(ctx)

	err = testutils.Migrate(ctx, dbContainer)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dbContainer.ConnString)
	require.NoError(t, err)
	defer pool.Close()

	q := db.New(pool)
	fakeGetRepo := func(q db.Querier) repository.IdempotencyLayerRepo {
		return repository.NewIdempotencyLayerRepo(q)
	}

	insertRepo := repository.NewIdempotencyRepo(pool)
	idempotency := NewConcreteIdempotencyLayer(pool, fakeGetRepo)

	t.Run("it should commit the values when the callback succeeds", func(t *testing.T) {
		createRequestParams := testutil.GetCreateRequestParams()
		createRequestParams.ExpiresAt = time.Now().Add(time.Hour).UTC()
		err := insertRepo.CreateStoredRequest(ctx, *createRequestParams)
		assert.NoError(t, err)

		cb := func(_ context.Context, _ db.Querier) (int, []byte, error) {
			return 200, nil, nil
		}
		status, resp, err := idempotency.Handle(ctx, createRequestParams.RequestID, cb)
		assert.NoError(t, err)
		assert.Equal(t, 200, status)
		assert.Nil(t, resp)

		req, err := q.GetRequest(ctx, createRequestParams.RequestID)
		assert.NoError(t, err)
		assert.Equal(t, idempotencyTypes.StatusComplete, req.RequestStatusID)
		assert.Equal(t, pgtype.Int4(pgtype.Int4{Int32: 200, Valid: true}), req.CachedResponseCode)
	})

	t.Run("it should rollback the values when the callback fails", func(t *testing.T) {
		requestId := uuid.New()
		createRequestParams := testutil.GetCreateRequestParams()
		createRequestParams.ExpiresAt = time.Now().Add(time.Hour).UTC()
		createRequestParams.RequestID = requestId

		err := insertRepo.CreateStoredRequest(ctx, *createRequestParams)
		assert.NoError(t, err)

		cb := func(_ context.Context, _ db.Querier) (int, []byte, error) {
			return 500, nil, errors.New("test error")
		}
		status, resp, err := idempotency.Handle(ctx, createRequestParams.RequestID, cb)
		assert.Error(t, err)
		assert.Equal(t, 500, status)
		assert.Nil(t, resp)

		req, err := q.GetRequest(ctx, createRequestParams.RequestID)
		assert.NoError(t, err)
		assert.Equal(t, idempotencyTypes.StatusProcessing, req.RequestStatusID)
		assert.Equal(t, pgtype.Int4(pgtype.Int4{Int32: 0, Valid: false}), req.CachedResponseCode)
	})

	t.Run("it should handle failure correctly", func(t *testing.T) {
		requestId := uuid.New()
		createRequestParams := testutil.GetCreateRequestParams()
		createRequestParams.ExpiresAt = time.Now().Add(-1 * time.Hour).UTC()
		createRequestParams.RequestID = requestId

		err := insertRepo.CreateStoredRequest(ctx, *createRequestParams)
		assert.NoError(t, err)

		idempotency.HandleFailure(requestId)

		req, err := q.GetRequest(ctx, createRequestParams.RequestID)
		assert.NoError(t, err)
		assert.Equal(t, idempotencyTypes.StatusFailed, req.RequestStatusID)
	})
}
