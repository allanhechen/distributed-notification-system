package repository

import (
	"context"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
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
	newRequest := db.CreateRequestStatusParams{
		RequestID:       reqID,
		UserID:          uuid.New(),
		RequestStatusID: types.StatusProcessing,
		ExpiresAt:       time.Now().Add(120 * time.Second).Truncate(time.Microsecond).UTC(),
	}

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
}
