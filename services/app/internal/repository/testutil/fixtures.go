package testutil

import (
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain/testutil"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	"github.com/allanhechen/distributed-notification-system/utils/types"
)

func GetCreateRequestParams() *repository.CreateRequestParams {
	return &repository.CreateRequestParams{
		RequestID:       testutil.RequestId,
		UserID:          testutil.UserId,
		RequestStatusID: types.StatusProcessing,
		ExpiresAt: time.Date(
			2025, 1, 1, 12, 0, 0, 0, time.UTC,
		),
	}

}
