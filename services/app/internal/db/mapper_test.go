package db_test

import (
	"testing"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/types"
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
				RequestStatusID:    types.StatusComplete,
				CachedResponseCode: nil,
				CachedResponse:     nil,
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: types.StatusComplete,
				CachedResponseCode: pgtype.Int4{
					Int32: 0,
					Valid: false,
				},
				CachedResponse: make([]byte, 0),
			},
		},
		{
			name: "Convert non-nil values",
			domainRequest: domain.IdempotentRequest{
				RequestID:          requestId,
				UserID:             userId,
				RequestStatusID:    types.StatusFailed,
				CachedResponseCode: &cachedResponseCode,
				CachedResponse:     &cachedResponse,
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: types.StatusFailed,
				CachedResponseCode: pgtype.Int4{
					Int32: 200,
					Valid: true,
				},
				CachedResponse: cachedResponse,
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
				RequestStatusID:    types.StatusComplete,
				CachedResponseCode: nil,
				CachedResponse:     nil,
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: types.StatusComplete,
				CachedResponseCode: pgtype.Int4{
					Int32: 0,
					Valid: false,
				},
				CachedResponse: make([]byte, 0),
			},
		},
		{
			name: "Convert non-nil values",
			domainRequest: domain.IdempotentRequest{
				RequestID:          requestId,
				UserID:             userId,
				RequestStatusID:    types.StatusFailed,
				CachedResponseCode: &cachedResponseCode,
				CachedResponse:     &cachedResponse,
			},
			dbRequest: db.IdempotentRequest{
				RequestID:       requestId,
				UserID:          userId,
				RequestStatusID: types.StatusFailed,
				CachedResponseCode: pgtype.Int4{
					Int32: 200,
					Valid: true,
				},
				CachedResponse: cachedResponse,
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
