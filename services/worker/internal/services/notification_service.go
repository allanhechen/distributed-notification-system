package services

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/worker/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
)

// TODO: extend this with actual errors
// isTransient reports whether the provided error should be considered transient and therefore eligible for retry.
// It returns true for context.DeadlineExceeded and context.Canceled, and for errors implementing net.Error whose Temporary method returns true.
func isTransient(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}

type NotificationService interface {
	HandleNotifications(context.Context) error
}

type ConcreteNotificationService struct {
	db             domain.Repository
	consumer       domain.Consumer[notification.Notification]
	notifier       domain.Notifier
	maxParallelism uint
}

// GetConcreteNotificationService creates a NotificationService backed by a ConcreteNotificationService
// using the provided repository, consumer, notifier, and worker pool size.
// The maxParallelism parameter controls the number of worker goroutines that will process messages concurrently.
func GetConcreteNotificationService(db domain.Repository, consumer domain.Consumer[notification.Notification], notifier domain.Notifier, maxParallelism uint) NotificationService {
	return &ConcreteNotificationService{
		db:             db,
		consumer:       consumer,
		notifier:       notifier,
		maxParallelism: maxParallelism,
	}
}

func (c *ConcreteNotificationService) HandleNotifications(ctx context.Context) error {
	jobs, err := c.consumer.Consume(ctx)
	if err != nil {
		slog.Error("notification service: failed to start consumer", "error", err)
		return err
	}

	var wg sync.WaitGroup
	for range c.maxParallelism {
		wg.Go(func() {
			c.startWorker(ctx, jobs)
		})
	}

	wg.Wait()
	return nil
}

func (c *ConcreteNotificationService) startWorker(ctx context.Context, jobs <-chan domain.Message[notification.Notification]) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("notification service: worker received cancellation signal")
			return
		case msg, ok := <-jobs:
			if !ok {
				slog.Info("notification service: worker tried to pull from closed channel")
				return
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *ConcreteNotificationService) processMessage(ctx context.Context, msg domain.Message[notification.Notification]) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("notification service: worker recovered from panic", "panic", r)
			msg.Nack(ctx, false)
		}
	}()

	jobCtx, cancel := context.WithTimeout(ctx, domain.JobProcessingTime)
	defer cancel()

	payload := msg.Payload()
	identifier := msg.Identifier()

	if shouldContinue := c.acquireNotificationLock(jobCtx, identifier, msg); !shouldContinue {
		return
	}

	slog.Info("notification service: begin processing notification", "identifier", identifier)
	if err := c.notifier.SendNotification(jobCtx, payload); err != nil {
		requeue := isTransient(err)

		slog.Info("notification service: failed to process notification", "identifier", identifier, "error", err)
		errCtx, errCancel := context.WithTimeout(context.WithoutCancel(jobCtx), domain.ErrorProcessingTime)
		defer errCancel()
		msg.Nack(errCtx, requeue)
		c.db.MarkFailure(errCtx, msg.Identifier())
		return
	}

	slog.Info("notification service: updating request with success status", "identifier", identifier)
	c.updateStatusSuccess(jobCtx, identifier, msg)
}

func (c *ConcreteNotificationService) updateStatusSuccess(jobCtx context.Context, identifier uuid.UUID, msg domain.Message[notification.Notification]) {
	updateCtx, updateCancel := context.WithTimeout(context.WithoutCancel(jobCtx), domain.SuccessProcessingTime)
	defer updateCancel()

	err := c.db.MarkSuccess(updateCtx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrNoRows) {
			slog.Error("notification service: message not found while marking success", "identifier", identifier)
		} else {
			slog.Error("notification service: failed to mark notification success", "error", err, "identifier", identifier)
		}
		msg.Nack(updateCtx, true)
		return
	}

	err = msg.Ack(updateCtx)
	if err != nil {
		slog.Error("notification service: failed to ack notification", "error", err, "identifier", identifier)
		return
	}
}

func (c *ConcreteNotificationService) acquireNotificationLock(ctx context.Context, identifier uuid.UUID, msg domain.Message[notification.Notification]) bool {
	now := time.Now().UTC()
	expiryTime := now.Add(domain.ProcessingLockTime)
	err := c.db.Acquire(ctx, identifier, expiryTime)
	if err != nil {
		if errors.Is(err, domain.ErrNoRows) {
			slog.Error("notification service: message not found", "identifier", identifier)
			msg.Nack(ctx, false)
		} else if errors.Is(err, domain.ErrAlreadyProcessing) {
			slog.Info("notification service: message is already processing", "identifier", identifier)
			msg.Nack(ctx, true)
		} else if errors.Is(err, domain.ErrAlreadyComplete) {
			slog.Info("notification service: message is already complete", "identifier", identifier)
			msg.Ack(ctx)
		} else {
			slog.Error("notification service: failed to handle message for other reason", "error", err)
			msg.Nack(ctx, true)
		}

		return false
	}

	return true
}
