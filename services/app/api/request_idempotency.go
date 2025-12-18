package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/services"
	"github.com/allanhechen/distributed-notification-system/utils"
	"github.com/google/uuid"
)

type IdempotentRequestHandler struct {
	service services.IdempotencyService
}

// NewIdempotentRequestHandler creates a new IdempotencyRequestHandler.
func NewIdempotentRequestHandler(service services.IdempotencyService) *IdempotentRequestHandler {
	return &IdempotentRequestHandler{
		service: service,
	}
}

func (i *IdempotentRequestHandler) HandleRequest(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		logger, ok := ctx.Value(utils.Logger).(*slog.Logger)
		if !ok {
			slog.Error("logger from context is the incorrect type")
			return
		}
		requestId, ok := ctx.Value(utils.RequestIdKey).(uuid.UUID)
		if !ok {
			logger.Error("requestId from context is the incorrect type")
			return
		}
		userId, ok := ctx.Value(utils.UserIdKey).(uuid.UUID)
		if !ok {
			logger.Error("userId from context is the incorrect type")
			return
		}

		request, action, err := i.service.GetOrBeginRequest(ctx, requestId, userId)
		if err != nil {
			if errors.Is(err, services.ErrConflict) {
				logger.Info("request hit conflict with another request")
				w.WriteHeader(http.StatusTooEarly)
				return
			}

			logger.Error("handler: inserting request failed due to other error", "error", err)
			return
		}

		switch action {
		case services.Replay:
			logger.Info("previous successful request hit idempotency cache")
			w.Header().Add("X-Cache-Status", "Idempotency-Hit")
			if request.CachedResponseCode != nil {
				w.WriteHeader(int(*request.CachedResponseCode))
			}
			if request.CachedResponse != nil {
				w.Write(*request.CachedResponse)
			}
			return
		case services.Reprocess:
			logger.Info("reprocessing previous request")
		case services.Proceed:
			logger.Info("processing new request")
		}

		reqCtx, cancel := context.WithTimeout(ctx, domain.ProcessingTtl)
		defer cancel()
		req = req.WithContext(reqCtx)

		next.ServeHTTP(w, req)
	}
}
