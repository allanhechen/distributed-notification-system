package testutil

import (
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
	"github.com/google/uuid"
)

// Fake RequestId to be used in idempotency fixtures
var RequestId = uuid.MustParse("7cdc4e71-73b0-4053-9ce6-43e319eda7d7")

// Fake UserId to be used in idempotency fixtures
var UserId = uuid.MustParse("d4e9c895-df17-4e6c-b9d8-97f2f79a9285")

// GetIdempotentRequest returns an IdempotentRequest to be used in tests
func GetIdempotentRequest() *domain.IdempotentRequest {
	return &domain.IdempotentRequest{
		RequestID:          RequestId,
		UserID:             UserId,
		RequestStatusID:    idempotencyTypes.StatusProcessing,
		CachedResponseCode: nil,
		CachedResponse:     nil,
		ExpiresAt: time.Date(
			2025, 1, 1, 12, 0, 0, 0, time.UTC,
		),
	}
}
