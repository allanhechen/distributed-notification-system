package utils

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

type contextKey string

const Logger contextKey = "logger"
const LoggedState contextKey = "loggedState"
const RequestIdKey contextKey = "requestId"
const UserIdKey contextKey = "userId"

type RequestUtils struct {
	Logger      *slog.Logger
	LoggedState *LogState
	RequestId   uuid.UUID
	UserId      uuid.UUID
}

var ErrRequestUtils = errors.New("request does not contain expected values in context")

func GetValuesFromContext(ctx context.Context) (RequestUtils, error) {
	logger, loggerOk := ctx.Value(Logger).(*slog.Logger)
	loggedState, loggedStateOk := ctx.Value(LoggedState).(*LogState)
	requestId, requestIdOk := ctx.Value(RequestIdKey).(uuid.UUID)
	userId, userIdOk := ctx.Value(UserIdKey).(uuid.UUID)

	if !loggerOk || !loggedStateOk || !requestIdOk || !userIdOk {
		return RequestUtils{}, ErrRequestUtils
	}

	return RequestUtils{
		Logger:      logger,
		LoggedState: loggedState,
		RequestId:   requestId,
		UserId:      userId,
	}, nil
}
