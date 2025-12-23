package services

import (
	"context"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain/testutil"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	idempotencyTypes "github.com/allanhechen/distributed-notification-system/utils/idempotency"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeRepository struct {
	mockDatabase        map[uuid.UUID]domain.IdempotentRequest
	createErr           error
	updateProcessingErr error
	updateSuccessErr    error
	updateFailedErr     error
}

func (f *fakeRepository) GetStoredRequest(_ context.Context, requestId uuid.UUID) (*domain.IdempotentRequest, error) {
	request, ok := f.mockDatabase[requestId]
	if !ok {
		return nil, repository.ErrNoRows
	}

	return &request, nil
}

func (f *fakeRepository) CreateStoredRequest(_ context.Context, params repository.CreateRequestParams) error {
	if f.createErr != nil {
		return f.createErr
	}

	_, ok := f.mockDatabase[params.RequestID]
	if ok {
		return repository.ErrAlreadyExists
	}

	request := domain.IdempotentRequest{
		RequestID:          params.RequestID,
		UserID:             params.UserID,
		RequestStatusID:    params.RequestStatusID,
		CachedResponseCode: nil,
		CachedResponse:     nil,
		ExpiresAt:          params.ExpiresAt,
	}
	f.mockDatabase[request.RequestID] = request

	return nil
}

func (f *fakeRepository) UpdateRequestFailed(_ context.Context, requestId uuid.UUID) error {
	if f.updateFailedErr != nil {
		return f.updateFailedErr
	}

	request, ok := f.mockDatabase[requestId]
	if !ok {
		return repository.ErrNoRows
	}

	f.mockDatabase[requestId] = domain.IdempotentRequest{
		RequestID:          request.RequestID,
		UserID:             request.UserID,
		RequestStatusID:    idempotencyTypes.StatusFailed,
		CachedResponseCode: request.CachedResponseCode,
		CachedResponse:     request.CachedResponse,
		ExpiresAt:          request.ExpiresAt,
	}

	return nil
}

func (f *fakeRepository) UpdateRequestSuccess(_ context.Context, params repository.UpdateRequestSuccessParams) error {
	if f.updateSuccessErr != nil {
		return f.updateSuccessErr
	}

	request, ok := f.mockDatabase[params.RequestID]
	if !ok {
		return repository.ErrNoRows
	}

	f.mockDatabase[params.RequestID] = domain.IdempotentRequest{
		RequestID:          request.RequestID,
		UserID:             request.UserID,
		RequestStatusID:    idempotencyTypes.StatusComplete,
		CachedResponseCode: &params.CachedResponseCode,
		CachedResponse:     &params.CachedResponse,
		ExpiresAt:          params.ExpiresAt,
	}

	return nil
}

func (f *fakeRepository) UpdateRequestReprocess(_ context.Context, params repository.UpdateRequestReprocessParams) error {
	if f.updateProcessingErr != nil {
		return f.updateProcessingErr
	}

	request, ok := f.mockDatabase[params.RequestID]
	if !ok {
		return repository.ErrNoRows
	}

	now := time.Now().UTC()
	if (request.RequestStatusID == idempotencyTypes.StatusComplete) ||
		(request.RequestStatusID == idempotencyTypes.StatusProcessing && !now.After(request.ExpiresAt)) {
		return repository.ErrNoRows
	}

	f.mockDatabase[params.RequestID] = domain.IdempotentRequest{
		RequestID:          request.RequestID,
		UserID:             request.UserID,
		RequestStatusID:    idempotencyTypes.StatusProcessing,
		CachedResponseCode: request.CachedResponseCode,
		CachedResponse:     request.CachedResponse,
		ExpiresAt:          params.ExpiresAt,
	}

	return nil
}

func TestGetOrBeginRequest_NonExistent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	nonExistentRequest := uuid.New()
	userId := testutil.UserId
	_, at, err := service.GetOrBeginRequest(ctx, nonExistentRequest, userId)
	assert.NoError(t, err)
	assert.Equal(t, Proceed, at)
}

func TestGetOrBeginRequest_AlreadyProcessing(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	// insert a new request
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = time.Now().Add(24 * time.Hour).UTC()
	fakeRequest.RequestStatusID = idempotencyTypes.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest
	userId := testutil.UserId

	_, _, err := service.GetOrBeginRequest(ctx, fakeRequest.RequestID, userId)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestGetOrBeginRequest_CachedInvalid(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	// insert a new request
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = time.Now().Add(-24 * time.Hour).UTC()
	fakeRequest.RequestStatusID = idempotencyTypes.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest
	userId := testutil.UserId

	_, at, err := service.GetOrBeginRequest(ctx, fakeRequest.RequestID, userId)
	assert.NoError(t, err)
	assert.Equal(t, Reprocess, at)
}

func TestGetOrBeginRequest_UpdateConflict(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		updateProcessingErr: repository.ErrNoRows, // leads to ErrConflict from beginProcessingRequest
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	// insert a new request
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = time.Now().Add(-24 * time.Hour).UTC()
	fakeRequest.RequestStatusID = idempotencyTypes.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest
	userId := testutil.UserId

	_, _, err := service.GetOrBeginRequest(ctx, fakeRequest.RequestID, userId)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestGetExistingRequest_NonExistent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	nonExistentRequest := uuid.New()
	_, err := service.getExistingRequest(ctx, nonExistentRequest)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetExistingRequest_Existent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	// insert a new request
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = time.Now().Add(24 * time.Hour).UTC()
	fakeRequest.RequestStatusID = idempotencyTypes.StatusComplete
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	// retrieve request with service
	request, err := service.getExistingRequest(ctx, fakeRequest.RequestID)
	assert.NoError(t, err)
	assert.Equal(t, fakeRequest, request)
}

func TestGetExistingRequest_Failed(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	// insert a failed request
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = time.Now().Add(24 * time.Hour).UTC()
	fakeRequest.RequestStatusID = idempotencyTypes.StatusFailed
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	// retrieve request with service
	_, err := service.getExistingRequest(ctx, fakeRequest.RequestID)
	assert.ErrorIs(t, err, ErrFailed)
}

func TestGetExistingRequest_Expired(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	// insert a new request
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = time.Now().Add(-24 * time.Hour).UTC()
	fakeRequest.RequestStatusID = idempotencyTypes.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	// retrieve request with service
	_, err := service.getExistingRequest(ctx, fakeRequest.RequestID)
	assert.ErrorIs(t, err, ErrExpired)
}

func TestBeginProcessingRequest_NewRequest(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId
	userId := testutil.UserId

	err := service.beginProcessingRequest(ctx, requestId, userId)
	assert.NoError(t, err)

	request, ok := repo.mockDatabase[requestId]
	assert.True(t, ok, "expected request to have been inserted into the database")
	assert.Equal(t, userId, request.UserID, "expected inserted userId to match given userId")
}

func TestBeginProcessingRequest_ExpiredRequest(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId
	userId := testutil.UserId

	// insert an expired request
	initialExpiry := time.Now().Add(-1 * time.Second).UTC()
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = initialExpiry
	fakeRequest.RequestStatusID = idempotencyTypes.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	err := service.beginProcessingRequest(ctx, requestId, userId)
	assert.NoError(t, err)

	request, ok := repo.mockDatabase[requestId]
	assert.True(t, ok, "expected request to remain in the database")
	assert.Equal(t, userId, request.UserID, "expected updated userId to match given userId")
	assert.Equal(t, idempotencyTypes.StatusProcessing, request.RequestStatusID)
	assert.True(t, request.ExpiresAt.After(initialExpiry), "ExpiresAt should be updated to a later time")
}

func TestBeginProcessingRequest_AlreadyComplete(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		createErr:           repository.ErrAlreadyExists,
		updateProcessingErr: repository.ErrNoRows,
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId
	userId := testutil.UserId

	// insert a complete request
	expiry := time.Now().Add(24 * time.Hour).UTC()
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = expiry
	fakeRequest.RequestStatusID = idempotencyTypes.StatusComplete
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	err := service.beginProcessingRequest(ctx, requestId, userId)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestBeginProcessingRequest_NotExpired(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		createErr:           repository.ErrAlreadyExists,
		updateProcessingErr: repository.ErrNoRows,
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId
	userId := testutil.UserId

	// insert a not-expired request
	expiry := time.Now().Add(24 * time.Hour).UTC()
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = expiry
	fakeRequest.RequestStatusID = idempotencyTypes.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	err := service.beginProcessingRequest(ctx, requestId, userId)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestBeginProcessingRequest_OtherConflict(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		createErr:           repository.ErrAlreadyExists,
		updateProcessingErr: repository.ErrNoRows,
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId
	userId := testutil.UserId

	err := service.beginProcessingRequest(ctx, requestId, userId)
	assert.ErrorIs(t, err, ErrConflict)
}
