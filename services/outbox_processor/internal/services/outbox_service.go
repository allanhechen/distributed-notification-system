package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
)

// ConcreteOutboxService is a concrete implementation of the OutboxService
// interface.
type ConcreteOutboxService struct {
	repo      domain.Repository
	ss        domain.StatusService
	mqs       domain.MqService
	batchSize int
	interval  time.Duration
}

// GetConcreteOutboxService returns an instance of ConcreteOutboxService.
func GetConcreteOutboxService(
	repo domain.Repository,
	ss domain.StatusService,
	mqs domain.MqService,
	batchSize int,
	interval time.Duration,
) domain.OutboxService {
	return &ConcreteOutboxService{
		repo:      repo,
		ss:        ss,
		mqs:       mqs,
		batchSize: batchSize,
		interval:  interval,
	}
}

// HandleMessages handles new notifications retrieved by the repository.
// If the handled batch size is max capacity, it runs another iteration
// immediately.
func (c *ConcreteOutboxService) HandleMessages(ctx context.Context) error {
	var listenerWg sync.WaitGroup
	var updateWg sync.WaitGroup
	batchSize := c.batchSize
	updates := make(chan domain.StatusUpdate, 2*batchSize) // closed in ctx.Done()
	iterate := make(chan struct{}, 1)
	defer close(iterate)
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	listenerWg.Go(func() {
		c.ss.Listen(updates)
	})

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox service: context closed, waiting for children and exiting gracefully")
			updateWg.Wait()
			close(updates)
			listenerWg.Wait()
			return ctx.Err()
		case <-ticker.C:
			select {
			case iterate <- struct{}{}:
			default:
			}
		case <-iterate:
			batch, err := c.repo.GetUnprocessedNotifications(ctx, batchSize)
			if err != nil {
				slog.Error("outbox service: repository failed to retrieve information", "error", err)
				continue
			}
			if len(batch) == 0 {
				continue
			}
			for _, n := range batch {
				updateWg.Go(func() {
					done, err := c.mqs.SendNotification(ctx, n, updates)
					if err != nil {
						slog.Error("outbox service: message queue service failed to send notification", "error", err, "notification", n)
						return
					}
					<-done
				})
			}

			if len(batch) == batchSize {
				iterate <- struct{}{}
			}
		}
	}
}
