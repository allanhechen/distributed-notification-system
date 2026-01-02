package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	amqp "github.com/rabbitmq/amqp091-go"
)

// request represents a single notification request to be sent.
type request struct {
	n       notification.Notification
	done    chan struct{}
	ctx     context.Context
	success chan bool
}

// RabbitMqService is a concrete implementation of MqService to be used
// with RabbitMQ. It handles reconnection on the connection and channel
// level, with a worker pool to maximize throughput.
//
// Expected lifecycle:
// 1. Obtain struct with GetRabbitMqService
// 2. Call Start
// 3. Pass messages with SendNotification
// 4. Call Stop
//
// RabbitMqService respects the deadline imposed by repository and will
// not deliberately send any messages to RabbitMQ after the deadline has
// expired. After calling Stop, no new messages may be queued. All
// remaining messages in jobs will be handled, after which channels and
// connections will be closed.
type RabbitMqService struct {
	mu          sync.Mutex
	connString  string
	numWorkers  int
	cancelled   chan struct{}
	jobs        chan request
	backoff     time.Duration
	maxBackoff  time.Duration
	maxAttempts int
}

// GetRabbitMqService returns an instance of RabbitMqService.
func GetRabbitMqService(
	connString string,
	numWorkers int,
	jobQueueCount int,
	backoff time.Duration,
	maxAttempts int,
) domain.MqService {
	return &RabbitMqService{
		connString:  connString,
		numWorkers:  numWorkers,
		cancelled:   make(chan struct{}),
		jobs:        make(chan request, jobQueueCount),
		backoff:     backoff,
		maxAttempts: maxAttempts,
	}
}

// SendNotification is a function that queues a message to be sent to
// RabbitMQ. Responses will be sent to the responses channel, respecting
// the notification's given timeout.
func (r *RabbitMqService) SendNotification(n notification.Notification, responses chan<- domain.StatusUpdate) (<-chan struct{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-r.cancelled:
		responses <- domain.StatusUpdate{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusUndelivered,
		}
		return nil, domain.ErrAlreadyClosed
	default:
	}

	done := make(chan struct{})
	ctx, cancel := context.WithDeadline(context.Background(), n.LockExpiryTime)
	req := request{
		n:       n,
		done:    done,
		success: make(chan bool, 1),
		ctx:     ctx,
	}
	r.jobs <- req

	go func() {
		defer cancel()
		defer func() {
			done <- struct{}{}
		}()

		update := domain.StatusUpdate{
			Identifier: n.Identifier,
		}

		var success bool
		select {
		case <-ctx.Done():
			success = false
		case success = <-req.success:
		}

		if success {
			slog.Info("mq service: notification was delivered successfully", "identifier", n.Identifier)
			update.FinalStatus = notification.StatusQueued
			responses <- update
		} else {
			slog.Warn("mq service: notification was not delivered successfully", "identifier", n.Identifier)
			update.FinalStatus = notification.StatusUndelivered
			responses <- update
		}
	}()
	return done, nil
}

// Start connects to RabbitMQ and initializes workers.
func (r *RabbitMqService) Start() {
	slog.Info("mq service: connecting to RabbitMQ")
	go r.handleReconnect()
}

// Stop signals cancellation and denies new jobs from being queued.
func (r *RabbitMqService) Stop() {
	slog.Info("mq service: halting")
	select {
	case <-r.cancelled:
		return
	default:
		r.mu.Lock()
		close(r.cancelled)
		close(r.jobs)
		r.mu.Unlock()
	}
}

// handleReconnect manages connections to RabbitMQ along with spawned
// workers.
func (r *RabbitMqService) handleReconnect() {
	var wg sync.WaitGroup
	var conn *amqp.Connection
	var err error
	currentBackoff := r.backoff

	for {
		conn, err = amqp.Dial(r.connString)
		if err != nil {
			slog.Warn("failed to connect", "error", err)
			<-time.After(currentBackoff)
			currentBackoff = min(currentBackoff*2, r.maxBackoff)
			continue
		}

		slog.Info("mq service: connection established")
		currentBackoff = r.backoff
		connCloseCh := make(chan *amqp.Error, 1)
		conn.NotifyClose(connCloseCh)

		for range r.numWorkers {
			wg.Go(func() {
				r.handleNotifications(conn, connCloseCh)
			})
		}

		select {
		case <-r.cancelled:
			wg.Wait()
			conn.Close()
			return
		case <-connCloseCh:
			slog.Warn("mq service: connection lost")
		}
	}
}

// handleNotifications belongs to a single connection. Returns when the
// connection is closed, or when jobs is closed.
func (r *RabbitMqService) handleNotifications(conn *amqp.Connection, connCloseCh <-chan *amqp.Error) {
	currentBackoff := r.backoff
	var channel *amqp.Channel
	var err error
	var chanCloseCh chan *amqp.Error

	for {
		chanCloseCh = make(chan *amqp.Error)
		channel, err = conn.Channel()
		if err != nil {
			currentBackoff = min(currentBackoff*2, r.maxBackoff)
			<-time.After(currentBackoff)
			continue
		}
		slog.Info("mq service: worker connected to channel")
		currentBackoff = r.backoff
		channel.NotifyClose(chanCloseCh)

		if shouldReturn := r.handleJobs(channel, chanCloseCh); shouldReturn {
			return
		}

		select {
		case <-connCloseCh:
			return
		case <-chanCloseCh:
			slog.Info("mq service: worker disconnected from channel")
		}
	}
}

// handleNotifications belongs to a single connection. Returns when the
// connection is closed, or when jobs is closed.
func (r *RabbitMqService) handleJobs(channel *amqp.Channel, chanCloseCh <-chan *amqp.Error) (shouldReturn bool) {
	for {
		select {
		case <-chanCloseCh:
			return false
		case req, ok := <-r.jobs:
			if !ok {
				return true
			}
			r.sendSingleNotification(req, channel, chanCloseCh)
		}
	}
}

// sendSingleNotification sends a single notification with the key
// corresponding to the notification's type.
func (r *RabbitMqService) sendSingleNotification(req request, channel *amqp.Channel, chanCloseCh <-chan *amqp.Error) {
	ctx := req.ctx
	n := req.n
	routingKey := rabbitmqnotifications.DeviceTypeToRoutingKey[n.NotificationType]
	currentBackoff := r.backoff

	body, err := json.Marshal(n)
	if err != nil {
		slog.Error("mq_service: could not marshal body")
		req.success <- false
		return
	}

	select {
	case <-ctx.Done():
		slog.Warn("mq service: context closed before delivering notification")
		req.success <- false
		return
	case <-chanCloseCh:
		req.success <- false
		return
	default:
	}

	for attempts := 1; attempts <= r.maxAttempts; attempts++ {
		err = channel.Publish(rabbitmqnotifications.NotificationExchange, routingKey, false, false, amqp.Publishing{
			Body: body,
		})
		if err != nil {
			slog.Warn("mq service: failed to deliver notification", "error", err, "identifier", n.Identifier, "attempt", attempts)
			select {
			case <-chanCloseCh:
				slog.Warn("mq service: channel closed on message", "identifier", n.Identifier)
				req.success <- false
				return
			case <-ctx.Done():
				slog.Warn("mq service: context closed on message", "identifier", n.Identifier)
				req.success <- false
				return
			case <-time.After(currentBackoff * time.Second):
				currentBackoff = min(currentBackoff*2, r.maxBackoff)
			}
		} else {
			req.success <- true
			return
		}
	}

	slog.Error("mq service: failed to deliver notification", "identifier", n.Identifier)
	req.success <- false
}
