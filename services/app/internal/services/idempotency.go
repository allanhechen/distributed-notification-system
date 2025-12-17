package services

import (
	"context"
	"errors"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	"github.com/allanhechen/distributed-notification-system/utils/types"
	"github.com/google/uuid"
)

// IdempotencyService is the service that handles request idempotency.
type IdempotencyService interface {
	getExistingRequest(ctx context.Context, requestId uuid.UUID) (*domain.IdempotentRequest, error)
	beginProcessingRequest(ctx context.Context, requestId uuid.UUID, userId uuid.UUID) error
	UpdateRequestSuccess(context.Context, UpdateRequestSuccessParams) error
	UpdateRequestFailed(ctx context.Context, requestId uuid.UUID) error
}

// IdempotencyServiceImplementation is the concrete implementation of
// IdempotencyService.
type IdempotencyServiceImplementation struct {
	repo repository.IdempotencyRepo
}

// NewIdempotencyService creates a new IdempotencyService.
func NewIdempotencyService(repo repository.IdempotencyRepo) IdempotencyService {
	return &IdempotencyServiceImplementation{
		repo: repo,
	}
}

// getExistingRequest finds the request with the given requestId. Simple
// wrapper around the repository method.
//
// Returns ErrNotFound if no rows were reported by the repository.
// Returns ErrExpired if the found row was marked expired.
// Returns ErrFailed if the found row was marked failed.
func (i *IdempotencyServiceImplementation) getExistingRequest(ctx context.Context, requestId uuid.UUID) (*domain.IdempotentRequest, error) {
	request, err := i.repo.GetStoredRequest(ctx, requestId)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if request.RequestStatusID == types.StatusFailed {
		return nil, ErrFailed
	}

	if time.Now().UTC().After(request.ExpiresAt) {
		return nil, ErrExpired
	}

	return request, nil
}

// beginProcessingRequest begins processing of the given requestId. If
// inserting a stored request returns repository.ErrAlreadyExists, it
// tries to update the same row. If this this fails, one of the following
// has occurred:
// 1. Previous request has not yet failed
// 2. Previous request is not expired
// 3. Previous request is already marked complete
//
// In all situations, ErrConflict is returned.
func (i *IdempotencyServiceImplementation) beginProcessingRequest(ctx context.Context, requestId uuid.UUID, userId uuid.UUID) error {
	newExpiryTime := time.Now().Add(domain.ShortRequestTtl).UTC()
	newRequest := repository.CreateRequestParams{
		RequestID:       requestId,
		UserID:          userId,
		RequestStatusID: types.StatusProcessing,
		ExpiresAt:       newExpiryTime,
	}
	err := i.repo.CreateStoredRequest(ctx, newRequest)

	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			updateRequestReprocess := repository.UpdateRequestReprocessParams{
				ExpiresAt: newExpiryTime,
				RequestID: requestId,
			}

			err = i.repo.UpdateRequestReprocess(ctx, updateRequestReprocess)
			if err != nil {
				if errors.Is(err, repository.ErrNoRows) {
					return ErrConflict
				}
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

type UpdateRequestSuccessParams struct {
	RequestID          uuid.UUID
	CachedResponseCode int32
	CachedResponse     []byte
}

// UpdateRequestSuccess marks the status of the given requestId to
// success.
//
// Returns ErrNotFound no rows were updated.
func (i *IdempotencyServiceImplementation) UpdateRequestSuccess(ctx context.Context, params UpdateRequestSuccessParams) error {
	newExpiryTime := time.Now().Add(domain.LongRequestTtl).UTC()
	UpdateRequestSuccess := repository.UpdateRequestSuccessParams{
		RequestID:          params.RequestID,
		CachedResponseCode: params.CachedResponseCode,
		CachedResponse:     params.CachedResponse,
		ExpiresAt:          newExpiryTime,
	}
	err := i.repo.UpdateRequestSuccess(ctx, UpdateRequestSuccess)

	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

// UpdateRequestFailed marks the status of the given requestId to
// failed.
//
// Returns ErrNotFound no rows were updated.
func (i *IdempotencyServiceImplementation) UpdateRequestFailed(ctx context.Context, requestId uuid.UUID) error {
	err := i.repo.UpdateRequestFailed(ctx, requestId)

	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
