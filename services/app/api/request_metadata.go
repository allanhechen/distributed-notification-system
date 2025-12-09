package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type requestIdentifiers struct {
	jwt.RegisteredClaims
	UserId uuid.UUID `json:"user_id"`
}

// Unique keys for the metadata context
type metadataContextKey string

const requestIdKey metadataContextKey = "requestId"
const userIdKey metadataContextKey = "userId"

// RequestMetadataMiddleware retrieves the identifying information for the
// current request from its associated JWT, and immediately rejects any
// requests with invalid identifiers. Additional information about the
// identification fields can be found in the [logging documentation].
//
// The request has already been verified at this point, so we can skip
// checking the JWT signature. We also know that it exists, and is valid.
// We call os.Exit(1) to kill the program as quickly as possible in case
// this invariant is ever violated
//
// [logging documentation]: https://github.com/allanhechen/distributed-notification-system/blob/main/docs/observability/logging.md
func RequestMetadataMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")
		requestIdStr := req.Header.Get("X-REQUEST-ID")
		requestId, err := uuid.Parse(requestIdStr)
		if err != nil {
			http.Error(w, "invalid request id", http.StatusUnauthorized)
			slog.Error("request received with invalid request id")
			os.Exit(1)
		}

		if auth == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			slog.Error("request received without authorization header")
			os.Exit(1)
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			slog.Error("request received invalid authorization header")
			os.Exit(1)
		}

		headerToken := strings.TrimPrefix(auth, prefix)
		token, _, err := new(jwt.Parser).ParseUnverified(headerToken, &requestIdentifiers{})
		if err != nil {
			http.Error(w, "invalid JWT", http.StatusUnauthorized)
			slog.Error("request received with invalid JWT", "error", err)
			os.Exit(1)
		}

		claims, ok := token.Claims.(*requestIdentifiers)
		if !ok {
			http.Error(w, "invalid JWT", http.StatusUnauthorized)
			slog.Error("request received with invalid JWT")
			os.Exit(1)
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, requestIdKey, requestId)
		ctx = context.WithValue(ctx, userIdKey, claims.UserId)
		req = req.WithContext(ctx)

		next.ServeHTTP(w, req)
	}
}
