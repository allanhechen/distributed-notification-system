package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RequestIdentifiers struct {
	jwt.RegisteredClaims
	RequestId uuid.UUID `json:"request_id"`
	UserId    uuid.UUID `json:"user_id"`
}

// Unique keys for the metadata context
type metadataContextKey string

const requestId metadataContextKey = "requestId"
const userId metadataContextKey = "userId"

// RequestMetadataMiddleware retrieves the identifying information for the
// current request from its associated JWT, and immediately rejects any
// requests with invalid identifiers. Additional information about the
// identification fields can be found in the [logging documentation].
//
// The request has already been verified at this point, so we can skip
// checking the JWT signature. We also know that it exists, and is valid.
//
// [logging documentation]: https://github.com/allanhechen/distributed-notification-system/blob/main/docs/observability/logging.md
func RequestMetadataMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			slog.Error("request received without authorization header")
			panic("request received without authorization header")
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			slog.Error("request received invalid authorization header")
			panic("request received invalid authorization header")
		}

		headerToken := strings.TrimPrefix(auth, prefix)
		token, _, err := new(jwt.Parser).ParseUnverified(headerToken, &RequestIdentifiers{})
		if err != nil {
			http.Error(w, "invalid JWT", http.StatusUnauthorized)
			slog.Error("request received with invalid JWT", "error", err)
			panic("request received with invalid JWT")
		}

		claims, ok := token.Claims.(*RequestIdentifiers)
		if !ok {
			http.Error(w, "invalid JWT", http.StatusUnauthorized)
			slog.Error("request received with invalid JWT")
			panic("request received with invalid JWT")
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, requestId, claims.RequestId)
		ctx = context.WithValue(ctx, userId, claims.UserId)
		req = req.WithContext(ctx)

		next.ServeHTTP(w, req)
	}
}
