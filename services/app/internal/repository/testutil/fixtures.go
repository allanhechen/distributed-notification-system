package testutil

import (
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain/testutil"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
)

// GetCreateRequestParams returns args to create an IdempotentRequest to
// be used in unit tests
func GetCreateRequestParams() *repository.CreateRequestParams {
	return &repository.CreateRequestParams{
		RequestID:       testutil.RequestId,
		UserID:          testutil.UserId,
		RequestStatusID: idempotencyTypes.StatusProcessing,
		ExpiresAt: time.Date(
			2025, 1, 1, 12, 0, 0, 0, time.UTC,
		),
	}

}
