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

// CanonicalLogger is an HTTP middleware that adds standardized logging
// and panic recovery to the request handling chain. Intended to be placed
// in the beginning of request processing chain, after the metadata has
// been extracted.
//
// CanonicalLogger creates a slog with the requestId and userId embedded.
// All request handlers are expected to utilize this logger to report
// logs relevant to the request.
//
// All logged attrs are expected to be added to LogState embedded within
// the context with `utils.AddField`, all of which will be repeated at the
// end of the request.
func CanonicalLogger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		start := time.Now()
		requestId := ctx.Value(requestIdKey)
		userId := ctx.Value(userIdKey)

		customWriter := &wrappedWriter{w, 0}
		requestLogger := slog.Default().With("request_id", requestId, "user_id", userId)

		l := &utils.LogState{}
		ctx = context.WithValue(ctx, utils.Logger, requestLogger)
		ctx = context.WithValue(ctx, utils.LoggedState, l)
		utils.AddField(ctx, "method", req.Method)
		utils.AddField(ctx, "path", req.URL.Path)
		req = req.WithContext(ctx)

		defer func() {
			if r := recover(); r != nil {
				duration := time.Since(start)
				utils.AddField(ctx, "duration", duration)
				utils.AddField(ctx, "panic", r)
				finalFields := l.Snapshot()
				finalFields["status"] = 500
				logAttrs := utils.FlattenMap(finalFields)
				requestLogger.Error("request panicked", logAttrs...)
				panic(r)
			}
		}()

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
