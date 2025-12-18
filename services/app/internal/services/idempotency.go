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

type IdempotencyActionType int

const (
	Proceed IdempotencyActionType = iota
	Reprocess
	Replay
)

// IdempotencyService is the service that handles request idempotency.
type IdempotencyService interface {
	GetOrBeginRequest(ctx context.Context, requestId uuid.UUID, userId uuid.UUID) (*domain.IdempotentRequest, IdempotencyActionType, error)
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

// GetOrBeginRequest either retrieves or begins the request with the
// given information.
//
// 1. If the request already exists, the request is returned with a Replay
// command and a nil error.
//
// 2. If the request is expired or failed, the request is marked as
// processing, a Reprocess command with a nil error is returned.
//
// 3.If the request is not found, the request is marked as processing, a
// Proceed command with a nil error is returned.
//
// 4. If the request is found but still processing and not expired, an
// ErrConflict is returned.
//
// If ErrConflict is returned while marking the request as processing, it
// is returned with no other values.
func (i *IdempotencyServiceImplementation) GetOrBeginRequest(ctx context.Context, requestId uuid.UUID, userId uuid.UUID) (*domain.IdempotentRequest, IdempotencyActionType, error) {
	request, err := i.getExistingRequest(ctx, requestId)
	if err == nil {
		return request, Replay, nil
	}

	var at IdempotencyActionType

	switch {
	case errors.Is(err, ErrExpired), errors.Is(err, ErrFailed):
		at = Reprocess
	case errors.Is(err, ErrNotFound):
		at = Proceed
	case errors.Is(err, ErrConflict):
		return nil, at, err
	default:
		return nil, at, err
	}

	err = i.beginProcessingRequest(ctx, requestId, userId)

	if err != nil {
		if errors.Is(err, ErrConflict) {
			return nil, at, ErrConflict
		}

		return nil, at, err
	}

	return nil, at, nil
}

// getExistingRequest finds the request with the given requestId. Simple
// wrapper around the repository method.
//
// Returns ErrNotFound if no rows were reported by the repository.
// Returns ErrConflict if the retrieved request is still being processed.
// Returns ErrExpired if the found row was marked expired.
// Returns ErrFailed if the found row was marked failed.
func (i *IdempotencyServiceImplementation) getExistingRequest(ctx context.Context, requestId uuid.UUID) (*domain.IdempotentRequest, error) {
	request, err := i.repo.GetStoredRequest(ctx, requestId)
	now := time.Now()
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if request.RequestStatusID == types.StatusFailed {
		return nil, ErrFailed
	}

	if request.ExpiresAt.After(now.UTC()) && request.RequestStatusID == types.StatusProcessing {
		return nil, ErrConflict
	}

	if now.UTC().After(request.ExpiresAt) {
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
