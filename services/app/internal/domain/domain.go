package domain

import (
	"time"

	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
	"github.com/google/uuid"
)

// IdempotentRequest is the complete representation of a cached request.
// Optional fields are represented with a pointer, nil means not present.
type IdempotentRequest struct {
	RequestID          uuid.UUID                      `json:"request_id"`
	UserID             uuid.UUID                      `json:"user_id"`
	RequestStatusID    idempotencyTypes.RequestStatus `json:"request_status_id"`
	CachedResponseCode *int32                         `json:"cached_response_code"`
	CachedResponse     *[]byte                        `json:"cached_response"`
	ExpiresAt          time.Time                      `json:"expires_at"`
}

// ShortRequestTtl is the TTL associated with processing requests
var ShortRequestTtl = 120 * time.Second

// ProcessingTtl is the TTL allocated to process requests. Gives some
// freedom for slower requests compared to ShortRequestTtl
var ProcessingTtl = 100 * time.Second

// LongRequestTtl is the TTL associated with storing processed requests
var LongRequestTtl = 24 * time.Hour
