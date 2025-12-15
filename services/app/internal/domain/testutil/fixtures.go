package testutil

import (
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/types"
	"github.com/google/uuid"
)

var RequestId = uuid.MustParse("7cdc4e71-73b0-4053-9ce6-43e319eda7d7")
var UserId = uuid.MustParse("d4e9c895-df17-4e6c-b9d8-97f2f79a9285")

func GetIdempotentRequest() *domain.IdempotentRequest {
	return &domain.IdempotentRequest{
		RequestID:          RequestId,
		UserID:             UserId,
		RequestStatusID:    types.StatusProcessing,
		CachedResponseCode: nil,
		CachedResponse:     nil,
		ExpiresAt: time.Date(
			2025, 1, 1, 12, 0, 0, 0, time.UTC,
		),
	}

}
