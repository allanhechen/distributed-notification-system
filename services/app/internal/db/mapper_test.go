package db_test

import (
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestIdempotentRequestFromDomain(t *testing.T) {
	requestId := uuid.New()
	userId := uuid.New()

	// define types to be used in conversion
	cachedResponseCode := int32(200)
	cachedResponse := []byte("request response")

	tt := []struct {
		name          string
		domainRequest domain.IdempotentRequest
		dbRequest     db.IdempotentRequest
	}{
		{
			name: "Convert nil values",
			domainRequest: domain.IdempotentRequest{
				RequestID:          requestId,
				UserID:             userId,
				RequestStatusID:    idempotencyTypes.StatusComplete,
				CachedResponseCode: nil,
				CachedResponse:     nil,
				ExpiresAt:          time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: idempotencyTypes.StatusComplete,
				CachedResponseCode: pgtype.Int4{
					Int32: 0,
					Valid: false,
				},
				CachedResponse: nil,
				ExpiresAt:      time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Convert non-nil values",
			domainRequest: domain.IdempotentRequest{
				RequestID:          requestId,
				UserID:             userId,
				RequestStatusID:    idempotencyTypes.StatusFailed,
				CachedResponseCode: &cachedResponseCode,
				CachedResponse:     &cachedResponse,
				ExpiresAt:          time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: idempotencyTypes.StatusFailed,
				CachedResponseCode: pgtype.Int4{
					Int32: 200,
					Valid: true,
				},
				CachedResponse: cachedResponse,
				ExpiresAt:      time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			result := db.IdempotentRequestFromDomain(&test.domainRequest)
			assert.Equal(t, test.dbRequest, *result)
		})
	}
}

func TestIdempotentRequestToDomain(t *testing.T) {
	requestId := uuid.New()
	userId := uuid.New()

	// define types to be used in conversion
	cachedResponseCode := int32(200)
	cachedResponse := []byte("request response")

	tt := []struct {
		name          string
		domainRequest domain.IdempotentRequest
		dbRequest     db.IdempotentRequest
	}{
		{
			name: "Convert nil values",
			domainRequest: domain.IdempotentRequest{
				RequestID:          requestId,
				UserID:             userId,
				RequestStatusID:    idempotencyTypes.StatusComplete,
				CachedResponseCode: nil,
				CachedResponse:     nil,
				ExpiresAt:          time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: idempotencyTypes.StatusComplete,
				CachedResponseCode: pgtype.Int4{
					Int32: 0,
					Valid: false,
				},
				CachedResponse: make([]byte, 0),
				ExpiresAt:      time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Convert non-nil values",
			domainRequest: domain.IdempotentRequest{
				RequestID:          requestId,
				UserID:             userId,
				RequestStatusID:    idempotencyTypes.StatusFailed,
				CachedResponseCode: &cachedResponseCode,
				CachedResponse:     &cachedResponse,
				ExpiresAt:          time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: idempotencyTypes.StatusFailed,
				CachedResponseCode: pgtype.Int4{
					Int32: 200,
					Valid: true,
				},
				CachedResponse: cachedResponse,
				ExpiresAt:      time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			result := db.IdempotentRequestToDomain(&test.dbRequest)
			assert.Equal(t, test.domainRequest, *result)
		})
	}
}
