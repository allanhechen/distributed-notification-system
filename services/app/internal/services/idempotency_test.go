package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain/testutil"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	"github.com/allanhechen/distributed-notification-system/utils/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var payload = map[string]any{
	"key": 1234,
}
var b, _ = json.Marshal(payload)

// GetUpdateRequestSuccessParams returns an UpdateRequestSuccessParams to
// be used in tests
func GetUpdateRequestSuccessParams() *UpdateRequestSuccessParams {
	return &UpdateRequestSuccessParams{
		RequestID:          testutil.RequestId,
		CachedResponseCode: int32(200),
		CachedResponse:     b,
	}
}

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
		RequestStatusID:    types.StatusFailed,
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
		RequestStatusID:    types.StatusComplete,
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
	if (request.RequestStatusID == types.StatusComplete) ||
		(request.RequestStatusID == types.StatusProcessing && !now.After(request.ExpiresAt)) {
		return repository.ErrNoRows
	}

	f.mockDatabase[params.RequestID] = domain.IdempotentRequest{
		RequestID:          request.RequestID,
		UserID:             request.UserID,
		RequestStatusID:    types.StatusProcessing,
		CachedResponseCode: request.CachedResponseCode,
		CachedResponse:     request.CachedResponse,
		ExpiresAt:          params.ExpiresAt,
	}

	return nil
}

func TestGetExistingRequest_NonExistent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()

	nonExistentRequest := uuid.New()
	_, err := service.GetExistingRequest(ctx, nonExistentRequest)
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
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	// retrieve request with service
	request, err := service.GetExistingRequest(ctx, fakeRequest.RequestID)
	assert.NoError(t, err)
	assert.Equal(t, fakeRequest, request)
}

func TestBeginProcessingRequest_NewRequest(t *testing.T) {
	repo := fakeRepository{
		mockDatabase: make(map[uuid.UUID]domain.IdempotentRequest),
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId
	userId := testutil.UserId

	err := service.BeginProcessingRequest(ctx, requestId, userId)
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
	fakeRequest.RequestStatusID = types.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	err := service.BeginProcessingRequest(ctx, requestId, userId)
	assert.NoError(t, err)

	request, ok := repo.mockDatabase[requestId]
	assert.True(t, ok, "expected request to remain in the database")
	assert.Equal(t, userId, request.UserID, "expected updated userId to match given userId")
	assert.Equal(t, types.StatusProcessing, request.RequestStatusID)
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
	fakeRequest.RequestStatusID = types.StatusComplete
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	err := service.BeginProcessingRequest(ctx, requestId, userId)
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
	fakeRequest.RequestStatusID = types.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	err := service.BeginProcessingRequest(ctx, requestId, userId)
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

	err := service.BeginProcessingRequest(ctx, requestId, userId)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestUpdateRequestSuccess_Existent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		createErr:           repository.ErrAlreadyExists,
		updateProcessingErr: repository.ErrNoRows,
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId

	// insert a processing request
	initialExpiry := time.Now().Add(1 * time.Second).UTC()
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = initialExpiry
	fakeRequest.RequestStatusID = types.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	fakeRequestUpdate := GetUpdateRequestSuccessParams()

	err := service.UpdateRequestSuccess(ctx, *fakeRequestUpdate)
	assert.NoError(t, err)

	request, ok := repo.mockDatabase[requestId]
	assert.True(t, ok, "expected request to remain in the database")
	assert.Equal(t, types.StatusComplete, request.RequestStatusID)
	assert.True(t, request.ExpiresAt.After(initialExpiry), "ExpiresAt should be updated to a later time")
}

func TestUpdateRequestSuccess_NonExistent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		createErr:           repository.ErrAlreadyExists,
		updateProcessingErr: repository.ErrNoRows,
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	fakeRequestUpdate := GetUpdateRequestSuccessParams()

	err := service.UpdateRequestSuccess(ctx, *fakeRequestUpdate)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestUpdateRequestFailed_Existent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		createErr:           repository.ErrAlreadyExists,
		updateProcessingErr: repository.ErrNoRows,
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId

	// insert a processing request
	initialExpiry := time.Now().Add(1 * time.Second).UTC()
	fakeRequest := testutil.GetIdempotentRequest()
	fakeRequest.ExpiresAt = initialExpiry
	fakeRequest.RequestStatusID = types.StatusProcessing
	repo.mockDatabase[fakeRequest.RequestID] = *fakeRequest

	err := service.UpdateRequestFailed(ctx, requestId)
	assert.NoError(t, err)

	request, ok := repo.mockDatabase[requestId]
	assert.True(t, ok, "expected request to remain in the database")
	assert.Equal(t, types.StatusFailed, request.RequestStatusID)
}

func TestUpdateRequestFailed_NonExistent(t *testing.T) {
	repo := fakeRepository{
		mockDatabase:        make(map[uuid.UUID]domain.IdempotentRequest),
		createErr:           repository.ErrAlreadyExists,
		updateProcessingErr: repository.ErrNoRows,
	}
	service := NewIdempotencyService(&repo)
	ctx := context.Background()
	requestId := testutil.RequestId

	err := service.UpdateRequestFailed(ctx, requestId)
	assert.ErrorIs(t, err, ErrNotFound)
}
