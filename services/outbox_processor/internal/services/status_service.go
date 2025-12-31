package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
)

type ConcreteStatusService struct {
	repo          domain.Repository
	sem           chan struct{}
	buf           []domain.StatusUpdate
	maxLen        int
	tickerTimeout time.Duration
	jobTimeout    time.Duration
}

func GetConcreteStatusService(
	repo domain.Repository,
	numWorkers int,
	maxLen int,
	tickerTimeout time.Duration,
	jobTimeout time.Duration,
) domain.StatusService {
	return &ConcreteStatusService{
		repo:          repo,
		sem:           make(chan struct{}, numWorkers),
		buf:           make([]domain.StatusUpdate, 0, maxLen),
		maxLen:        maxLen,
		tickerTimeout: tickerTimeout,
		jobTimeout:    jobTimeout,
	}
}
func (c *ConcreteStatusService) Listen(updates <-chan domain.StatusUpdate) {
	var wg sync.WaitGroup
	ticker := time.NewTicker(c.tickerTimeout)
	defer ticker.Stop()
	for {
		select {
		case u, ok := <-updates:
			if !ok {
				slog.Info("status service: channel closed, clearing buffers")
				if len(c.buf) > 0 {
					old := c.buf
					c.buf = nil
					c.sem <- struct{}{}
					wg.Go(func() {
						c.processBatch(old)
					})
				}
				wg.Wait()
				slog.Info("status service: workers returned, exiting")
				return
			}
			c.buf = append(c.buf, u)
			if len(c.buf) == c.maxLen {
				old := c.buf
				c.buf = make([]domain.StatusUpdate, 0, c.maxLen)
				c.sem <- struct{}{}
				wg.Go(func() {
					c.processBatch(old)
				})
			}
		case <-ticker.C:
			if len(c.buf) > 0 {
				old := c.buf
				c.buf = make([]domain.StatusUpdate, 0, c.maxLen)
				c.sem <- struct{}{}
				wg.Go(func() {
					c.processBatch(old)
				})
			}
		}
	}
}
func (c *ConcreteStatusService) processBatch(buf []domain.StatusUpdate) error {
	if len(buf) == 0 {
		slog.Info("status service: no updates to be sent to the repository")
		<-c.sem
		return nil
	}
	slog.Info("status service: updating repository statuses", "entries", len(buf))
	jobCtx, cancel := context.WithTimeout(context.Background(), c.jobTimeout)
	defer cancel()
	err := c.repo.UpdateNotificationStatuses(jobCtx, buf)
	if err != nil {
		if errors.Is(err, domain.ErrNonExistent) {
			slog.Warn("status service: some notifications specified no longer exist", "error", err)
		} else {
			slog.Warn("status service: failed to mark notifications", "error", err)
		}
	}
	<-c.sem
	return err
}
