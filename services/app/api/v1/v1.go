// The v1 API handlers for the application
// Serves as the jumping-off point for other modules (ex. users, devices, groups)
package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/internal"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/allanhechen/distributed-notification-system/utils"
	"github.com/google/uuid"
)

// Connectivity Check
//
//	@Summary		Test server connectivity
//	@Description	Endpoint to quickly verify that the server is reachable
//	@Tags			health
//	@Produce		application/json
//	@Success		200	{object}	map[string]string	"pong response"
//	@Router			/v1/ping [get]
func ping(idempotencyLayer internal.IdempotencyLayer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		logger, ok := ctx.Value(utils.Logger).(*slog.Logger)
		if !ok {
			slog.Error("logger from context is the incorrect type")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		requestId, ok := ctx.Value(utils.RequestIdKey).(uuid.UUID)
		if !ok {
			logger.Error("requestId from context is the incorrect type")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status, resp, err := idempotencyLayer.Handle(ctx, requestId, func(_ context.Context, _ db.Querier) (int, []byte, error) {
			time.Sleep(2 * time.Second)
			response := map[string]string{"message": "pong!"}
			b, _ := json.Marshal(response)
			status := http.StatusOK

			return status, b, nil
		})
		if err != nil {
			if !errors.Is(err, internal.ErrUser) {
				logger.Error("handler: error caused by user", "error", err)
				idempotencyLayer.HandleFailure(requestId)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(resp)
	}

}

// Creates a v1 Routes handler, expected to be used once during the initialization of the application
func Routes(idempotencyLayer internal.IdempotencyLayer) *http.ServeMux {
	v1 := http.NewServeMux()

	v1.HandleFunc("/ping", ping(idempotencyLayer))

	return v1
}
