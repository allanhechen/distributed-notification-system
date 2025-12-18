package db

import (
	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

// IdempotentRequestFromDomain converts the domain representation of a
// request into the specialized database type.
func IdempotentRequestFromDomain(domainRequest *domain.IdempotentRequest) *IdempotentRequest {
	if domainRequest == nil {
		return nil
	}

	var cachedResponseCode pgtype.Int4
	if domainRequest.CachedResponseCode != nil {
		cachedResponseCode = pgtype.Int4{
			Int32: *domainRequest.CachedResponseCode,
			Valid: true,
		}
	} else {
		cachedResponseCode = pgtype.Int4{
			Int32: 0,
			Valid: false,
		}
	}

	// TODO: test that this is valid JSON
	var cachedResponse []byte
	if domainRequest.CachedResponse != nil {
		cachedResponse = *domainRequest.CachedResponse
	}

	return &IdempotentRequest{
		RequestID:          domainRequest.RequestID,
		UserID:             domainRequest.UserID,
		RequestStatusID:    domainRequest.RequestStatusID,
		CachedResponseCode: cachedResponseCode,
		CachedResponse:     cachedResponse,
		ExpiresAt:          domainRequest.ExpiresAt,
	}
}

// IdempotentRequestToDomain converts the specialized database
// representation of a request into the domain type.
func IdempotentRequestToDomain(dbRequest *IdempotentRequest) *domain.IdempotentRequest {
	if dbRequest == nil {
		return nil
	}

	var cachedResponseCode *int32
	if dbRequest.CachedResponseCode.Valid {
		code := dbRequest.CachedResponseCode.Int32
		cachedResponseCode = &code
	}

	var cachedResponse *[]byte
	if len(dbRequest.CachedResponse) > 0 {
		cachedResponse = &dbRequest.CachedResponse
	}

	return &domain.IdempotentRequest{
		RequestID:          dbRequest.RequestID,
		UserID:             dbRequest.UserID,
		RequestStatusID:    dbRequest.RequestStatusID,
		CachedResponseCode: cachedResponseCode,
		CachedResponse:     cachedResponse,
		ExpiresAt:          dbRequest.ExpiresAt,
	}
}
