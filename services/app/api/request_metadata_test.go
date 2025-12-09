package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestRequestDataMiddleware(t *testing.T) {
	fakeRequestIdStr := "84e7a4bf-b687-499c-bcae-86d1b8454d93"
	fakeUserIdStr := "69eb61d2-8f13-484a-881e-3577c8c7d770"

	fakeRequestId, _ := uuid.Parse(fakeRequestIdStr)
	fakeUserId, _ := uuid.Parse(fakeUserIdStr)

	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if receivedRequestId, ok := ctx.Value(requestId).(uuid.UUID); !ok || receivedRequestId != fakeRequestId {
			t.Errorf("expected requestId to be %v, got %v", fakeRequestId, receivedRequestId)
		}

		if receivedUserId, ok := ctx.Value(userId).(uuid.UUID); !ok || receivedUserId != fakeUserId {
			t.Errorf("expected userId to be %v, got %v", fakeUserId, receivedUserId)
		}
	}
	server := httptest.NewServer(RequestMetadataMiddleware(mockHandler))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"request_id": fakeRequestId,
		"user_id":    fakeUserId,
	})
	tokenString, _ := token.SignedString([]byte("mock signing secret"))
	request, _ := http.NewRequest("GET", server.URL, nil)
	request.Header.Add("Authorization", "Bearer "+tokenString)
	server.Client().Do(request)
}
