package services

import (
	"context"
	"log/slog"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
)

// LogNotifier is a notifier that sends notifications to the slog.
// Intended to be only used for development.
type LogNotifier struct{}

// GetLogNotifier returns an instance of LogNotifier.
func GetLogNotifier() domain.Notifier {
	return &LogNotifier{}
}

// SendNotification sends a notification to the slog.
func (c *LogNotifier) SendNotification(ctx context.Context, notification notification.Notification) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	slog.Info("notifier: received notification", "message", notification.Message)
	return nil
}
