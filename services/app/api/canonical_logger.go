package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/allanhechen/distributed-notification-system/utils"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *wrappedWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func CanonicalLogger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestId := "d4cc4732-4a8a-4044-a868-0cacb61bfc8a" // TODO: Retrieve this off the request using validation
		userId := "3c4e79fd-feec-4f37-9dbd-b7054796e24d"    // TODO: Retrieve this off the request using validation
		serviceId := "bd1f7110-1862-4b4e-9972-78b203a3cb32" // TODO: Generate and pass this on server startup
		start := time.Now()

		customWriter := &wrappedWriter{w, 0}
		requestLogger := slog.Default().With("request_id", requestId, "user_id", userId, "service_id", serviceId)

		l := &utils.LogState{}
		ctx := req.Context()
		ctx = context.WithValue(ctx, utils.Logger, requestLogger)
		ctx = context.WithValue(ctx, utils.LoggedState, l)
		utils.AddField(ctx, "method", req.Method)
		utils.AddField(ctx, "path", req.URL.Path)
		req = req.WithContext(ctx)
		next.ServeHTTP(customWriter, req)

		duration := time.Since(start)
		utils.AddField(ctx, "duration", duration)
		finalFields := l.Snapshot()

		status := customWriter.statusCode
		finalFields["status"] = status
		logAttrs := utils.FlattenMap(finalFields)

		if status >= 500 {
			requestLogger.Error("request failed (server error)", logAttrs...)
		} else if status >= 400 {
			requestLogger.Warn("request failed (client error)", logAttrs...)
		} else {
			requestLogger.Info("request succeeded", logAttrs...)
		}
	}
}
