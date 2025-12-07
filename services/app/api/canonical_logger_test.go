package api_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/allanhechen/distributed-notification-system/services/app/api"
	"github.com/allanhechen/distributed-notification-system/utils"
)

type mockLogHandler struct {
	Records []slog.Record
}

func (h *mockLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *mockLogHandler) Handle(_ context.Context, r slog.Record) error {
	h.Records = append(h.Records, r)
	return nil
}

func (h *mockLogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *mockLogHandler) WithGroup(_ string) slog.Handler {
	return h
}

func TestCanonicalLogger(t *testing.T) {
	mockHandler := &mockLogHandler{}
	testLogger := slog.New(mockHandler)
	slog.SetDefault(testLogger)

	tests := []struct {
		name            string
		mockStatus      int
		expectedLevel   slog.Level
		expectedMessage string
	}{
		{"Success", http.StatusOK, slog.LevelInfo, "request succeeded"},
		{"ClientError", http.StatusNotFound, slog.LevelWarn, "request failed (client error)"},
		{"ServerError", http.StatusInternalServerError, slog.LevelError, "request failed (server error)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler.Records = nil

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				utils.AddField(r.Context(), "key", "value")
			})

			req := httptest.NewRequest("GET", "/test/path", nil)
			rec := httptest.NewRecorder()

			api.CanonicalLogger(nextHandler).ServeHTTP(rec, req)

			if len(mockHandler.Records) != 1 {
				t.Fatalf("Expected 1 log record, got %d", len(mockHandler.Records))
			}

			record := mockHandler.Records[0]

			if record.Level != tt.expectedLevel {
				t.Errorf("Expected level %v, got %v", tt.expectedLevel, record.Level)
			}
			if record.Message != tt.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, record.Message)
			}

			statusFound := false
			pathFound := false
			keyFound := false
			record.Attrs(func(attr slog.Attr) bool {
				if attr.Key == "status" && attr.Value.Kind() == slog.KindInt64 && attr.Value.Int64() == int64(tt.mockStatus) {
					statusFound = true
				}
				if attr.Key == "path" && attr.Value.Kind() == slog.KindString && attr.Value.String() == "/test/path" {
					pathFound = true
				}
				if attr.Key == "key" && attr.Value.Kind() == slog.KindString && attr.Value.String() == "value" {
					keyFound = true
				}
				return true
			})

			if !statusFound {
				t.Error("Log record did not contain the correct 'status' field.")
			}
			if !pathFound {
				t.Error("Log record did not contain the correct 'path' field.")
			}
			if !keyFound {
				t.Error("Log record did not contain the correct 'keyFound' field.")
			}
		})
	}
}
