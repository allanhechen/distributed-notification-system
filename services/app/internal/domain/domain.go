package domain

import (
	"time"

	"github.com/allanhechen/distributed-notification-system/utils/types"
	"github.com/google/uuid"
)

// IdempotentRequest is the complete representation of a cached request.
// Optional fields are represented with a pointer, nil means not present.
type IdempotentRequest struct {
	RequestID          uuid.UUID           `json:"request_id"`
	UserID             uuid.UUID           `json:"user_id"`
	RequestStatusID    types.RequestStatus `json:"request_status_id"`
	CachedResponseCode *int32              `json:"cached_response_code"`
	CachedResponse     *[]byte             `json:"cached_response"`
	ExpiresAt          time.Time           `json:"expires_at"`
}
