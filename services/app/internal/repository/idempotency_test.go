package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/testutils"
	"github.com/allanhechen/distributed-notification-system/utils/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyRepository(t *testing.T) {
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

	reqID := uuid.New()
	newRequest := CreateRequestParams{
		RequestID:       reqID,
		UserID:          uuid.New(),
		RequestStatusID: types.StatusProcessing,
		ExpiresAt:       time.Now().Add(120 * time.Second).Truncate(time.Microsecond).UTC(),
	}

	payload := map[string]any{
		"key": 1234,
	}

	b, err := json.Marshal(payload)
	require.NoError(t, err)
	cachedResponseCode := int32(200)

	t.Run("Create and Get Success", func(t *testing.T) {
		err := repo.CreateStoredRequest(ctx, newRequest)
		assert.NoError(t, err)

		result, err := repo.GetStoredRequest(ctx, reqID)
		assert.NoError(t, err)

		assert.Equal(t, newRequest.RequestID, result.RequestID)
		assert.Equal(t, newRequest.UserID, result.UserID)
		assert.Equal(t, newRequest.RequestStatusID, result.RequestStatusID)

		assert.WithinDuration(t, newRequest.ExpiresAt, result.ExpiresAt, time.Microsecond, "ExpiresAt should match within DB precision")
	})

	t.Run("Duplicate Creation (Idempotency)", func(t *testing.T) {
		err := repo.CreateStoredRequest(ctx, newRequest)

		assert.ErrorIs(t, err, ErrAlreadyExists)
	})

	t.Run("Get Non-Existent Request", func(t *testing.T) {
		randomID := uuid.New()
		_, err := repo.GetStoredRequest(ctx, randomID)

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Mark Non-Existent Request as Failed", func(t *testing.T) {
		randomID := uuid.New()
		err := repo.UpdateRequestFailed(ctx, randomID)

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Mark Non-Expired Request as Failed", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: types.StatusProcessing,
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestFailed(ctx, reqID)
		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Mark Expired Request as Failed", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: types.StatusProcessing,
			ExpiresAt:       time.Now().Add(-1 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestFailed(ctx, reqID)
		assert.NoError(t, err)
	})

	t.Run("Mark Non-Existent Request as Success", func(t *testing.T) {

		randomID := uuid.New()
		err = repo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
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
			RequestStatusID: types.StatusProcessing,
			ExpiresAt:       time.Now().Add(-1 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
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
			RequestStatusID: types.StatusFailed, // already terminal
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}
		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
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
			RequestStatusID: types.StatusProcessing,
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}

		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestSuccess(ctx, UpdateRequestSuccessParams{
			RequestID:          reqID,
			ExpiresAt:          time.Now().Add(24 * time.Hour).UTC(),
			CachedResponseCode: cachedResponseCode,
			CachedResponse:     b,
		})

		assert.NoError(t, err)
	})

	t.Run("Reprocess Non-Existent Request", func(t *testing.T) {
		randomID := uuid.New()

		err := repo.UpdateRequestReprocess(ctx, UpdateRequestReprocessParams{
			RequestID: randomID,
		})

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Reprocess Non-Expired Processing Request", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: types.StatusProcessing,
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}

		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestReprocess(ctx, UpdateRequestReprocessParams{
			RequestID: reqID,
		})

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Reprocess Succeeded Request", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: types.StatusComplete,
			ExpiresAt:       time.Now().Add(-120 * time.Second).UTC(),
		}

		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestReprocess(ctx, UpdateRequestReprocessParams{
			RequestID: reqID,
		})

		assert.ErrorIs(t, err, ErrNoRows)
	})

	t.Run("Reprocess Failed Request", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: types.StatusFailed,
			ExpiresAt:       time.Now().Add(120 * time.Second).UTC(),
		}

		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestReprocess(ctx, UpdateRequestReprocessParams{
			RequestID: reqID,
		})

		assert.NoError(t, err)
	})

	t.Run("Reprocess Expired Processing Request", func(t *testing.T) {
		reqID := uuid.New()
		newRequest := CreateRequestParams{
			RequestID:       reqID,
			UserID:          uuid.New(),
			RequestStatusID: types.StatusProcessing,
			ExpiresAt:       time.Now().Add(-1 * time.Second).UTC(),
		}

		err := repo.CreateStoredRequest(ctx, newRequest)
		require.NoError(t, err)

		err = repo.UpdateRequestReprocess(ctx, UpdateRequestReprocessParams{
			RequestID: reqID,
		})

		assert.NoError(t, err)
	})
}
