package testutil

import (
	"encoding/json"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain/testutil"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/services"
)

var payload = map[string]any{
	"key": 1234,
}
var b, _ = json.Marshal(payload)

// GetUpdateRequestSuccessParams returns an UpdateRequestSuccessParams to
// be used in tests
func GetUpdateRequestSuccessParams() *services.UpdateRequestSuccessParams {
	return &services.UpdateRequestSuccessParams{
		RequestID:          testutil.RequestId,
		CachedResponseCode: int32(200),
		CachedResponse:     b,
	}
}
