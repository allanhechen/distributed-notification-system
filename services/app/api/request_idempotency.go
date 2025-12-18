package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/allanhechen/distributed-notification-system/services/app/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/services"
	"github.com/allanhechen/distributed-notification-system/utils"
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
		fromCtx, err := utils.GetValuesFromContext(ctx)
		if err != nil {
			slog.Error("idempotency: failed to get values from request context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger := fromCtx.Logger
		requestId := fromCtx.RequestId
		userId := fromCtx.UserId

		request, action, err := i.service.GetOrBeginRequest(ctx, requestId, userId)
		if err != nil {
			if errors.Is(err, services.ErrConflict) {
				logger.Info("request hit conflict with another request")
				w.WriteHeader(http.StatusTooEarly)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("handler: inserting request failed due to other error", "error", err)
			return
		}

		switch action {
		case services.Replay:
			logger.Info("previous successful request hit idempotency cache")
			w.Header().Add("X-Cache-Status", "Idempotency-Hit")
			w.Header().Set("Content-Type", "application/json")
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
