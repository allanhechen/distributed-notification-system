package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/testutils"
	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyLayerRepository(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dbContainer, err := testutils.GetCrdbDatabaseContainer(ctx)
	require.NoError(t, err)
	defer dbContainer.Container.Terminate(ctx)

	err = testutils.Migrate(ctx, dbContainer)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dbContainer.ConnString)
	require.NoError(t, err)
	defer pool.Close()

	repo := PgxIdempotency{
		pool: pool,
	}

	payload := map[string]any{
		"key": 1234,
	}

	b, err := json.Marshal(payload)
	require.NoError(t, err)
	cachedResponseCode := int32(200)

	t.Run("Mark Non-Existent Request as Failed", func(t *testing.T) {
		layerRepo, trx := getLayerRepo(ctx, t, pool)
		defer trx.Rollback(ctx)

		randomID := uuid.New()
		err := layerRepo.UpdateRequestFailed(ctx, randomID)

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Mark Non-Expired Request as Failed", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: idempotencyTypes.StatusProcessing,
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		layerRepo, trx := getLayerRepo(ctx, t, pool)
		defer trx.Rollback(ctx)

		err = layerRepo.UpdateRequestFailed(ctx, reqID)
		assert.NoError(t, err)
	})

	t.Run("Mark Expired Request as Failed", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: idempotencyTypes.StatusProcessing,
			ExpiresAt:       time.Now().Add(-1 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		layerRepo, trx := getLayerRepo(ctx, t, pool)
		defer trx.Rollback(ctx)

		err = layerRepo.UpdateRequestFailed(ctx, reqID)
		assert.NoError(t, err)
	})

	t.Run("Mark Non-Existent Request as Success", func(t *testing.T) {
		layerRepo, trx := getLayerRepo(ctx, t, pool)
		defer trx.Rollback(ctx)

		randomID := uuid.New()
		err = layerRepo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
			RequestID:          randomID,
			ExpiresAt:          time.Now().Add(24 * time.Hour).UTC(),
			CachedResponseCode: cachedResponseCode,
			CachedResponse:     b,
		})

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Mark Expired Request as Success", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: idempotencyTypes.StatusProcessing,
			ExpiresAt:       time.Now().Add(-1 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		layerRepo, trx := getLayerRepo(ctx, t, pool)
		defer trx.Rollback(ctx)

		err = layerRepo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
			RequestID:          reqID,
			ExpiresAt:          time.Now().Add(24 * time.Hour).UTC(),
			CachedResponseCode: cachedResponseCode,
			CachedResponse:     b,
		})

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Mark Non-Processing Request as Success", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: idempotencyTypes.StatusFailed, // already terminal
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		layerRepo, trx := getLayerRepo(ctx, t, pool)
		defer trx.Rollback(ctx)

		err = layerRepo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
			RequestID:          reqID,
			ExpiresAt:          time.Now().Add(24 * time.Hour).UTC(),
			CachedResponseCode: cachedResponseCode,
			CachedResponse:     b,
		})

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Mark Processing Request as Success", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: idempotencyTypes.StatusProcessing,
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}

		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		layerRepo, trx := getLayerRepo(ctx, t, pool)
		defer trx.Rollback(ctx)

		err = layerRepo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
			RequestID:          reqID,
			ExpiresAt:          time.Now().Add(24 * time.Hour).UTC(),
			CachedResponseCode: cachedResponseCode,
			CachedResponse:     b,
		})
		assert.NoError(t, err)
	})
}

func getLayerRepo(ctx context.Context, t *testing.T, pool *pgxpool.Pool) (QuerierIdempotency, pgx.Tx) {
	trx, err := pool.Begin(ctx)
	require.NoError(t, err)
	q := db.New(pool)
	qtx := q.WithTx(trx)
	layerRepo := QuerierIdempotency{
		querier: qtx,
	}
	return layerRepo, trx
}
