package utils

import (
	"context"
	"maps"
	"sync"
)

// Unique keys for the LogState context
type contextKey string

const Logger contextKey = "logger"
const LoggedState contextKey = "loggedState"

type LogState struct {
	mu     sync.Mutex
	fields map[string]any
}

func (l *LogState) setField(key string, value any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fields == nil {
		l.fields = make(map[string]any)
	}
	l.fields[key] = value
}

func (l *LogState) Snapshot() map[string]any {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fields == nil {
		return make(map[string]any)
	}
	return maps.Clone(l.fields)
}

func AddField(ctx context.Context, key string, value any) {
	if state, ok := ctx.Value(LoggedState).(*LogState); ok {
		state.setField(key, value)
	}
}
