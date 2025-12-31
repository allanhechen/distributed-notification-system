package services

import (
	"testing"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/domain"
	"github.com/allanhechen/distributed-notification-system/services/outbox_processor/internal/testutil"
	"github.com/allanhechen/distributed-notification-system/utils/notification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessBatch_SingleExisting(t *testing.T) {
	r := testutil.GetFakeRepository()
	ss := ConcreteStatusService{
		repo:          r,
		sem:           make(chan struct{}, 1),
		buf:           make([]domain.StatusUpdate, 0, 1),
		maxLen:        1,
		tickerTimeout: 1 * time.Second,
		jobTimeout:    1 * time.Second,
	}
	n := notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{})
	r.Entries[n.Identifier] = n

	statusUpdate := []domain.StatusUpdate{
		{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusQueued,
		},
	}
	// send message for this worker, normally handled by Listen
	ss.sem <- struct{}{}
	err := ss.processBatch(statusUpdate)

	// update notification to match
	n.Status = notification.StatusQueued
	assert.NoError(t, err)
	assert.Equal(t, n, r.Entries[n.Identifier])
}

func TestProcessBatch_MixExisting(t *testing.T) {
	r := testutil.GetFakeRepository()
	ss := ConcreteStatusService{
		repo:          r,
		sem:           make(chan struct{}, 1),
		buf:           make([]domain.StatusUpdate, 0, 1),
		maxLen:        1,
		tickerTimeout: 1 * time.Second,
		jobTimeout:    1 * time.Second,
	}
	n := notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{})
	r.Entries[n.Identifier] = n

	statusUpdate := []domain.StatusUpdate{
		{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusQueued,
		},
		{
			Identifier:  uuid.New(), // one that doesn't exist
			FinalStatus: notification.StatusQueued,
		},
	}
	// send message for this worker, normally handled by Listen
	ss.sem <- struct{}{}
	err := ss.processBatch(statusUpdate)

	expected := n
	expected.Status = notification.StatusQueued
	// update notification to match
	assert.ErrorIs(t, err, domain.ErrNonExistent)
	assert.Equal(t, expected, r.Entries[n.Identifier])
}

func TestProcessBatch_TickerActivation(t *testing.T) {
	r := testutil.GetFakeRepository()
	ss := ConcreteStatusService{
		repo:          r,
		sem:           make(chan struct{}, 1),
		buf:           make([]domain.StatusUpdate, 0, 1),
		maxLen:        1,
		tickerTimeout: 250 * time.Millisecond,
		jobTimeout:    1 * time.Second,
	}
	notifications := []notification.Notification{
		notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{}),
		notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{}),
	}
	var statusUpdates []domain.StatusUpdate
	for _, n := range notifications {
		r.Entries[n.Identifier] = n
		statusUpdates = append(statusUpdates, domain.StatusUpdate{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusQueued,
		})
		n.Status = notification.StatusQueued
	}
	updates := make(chan domain.StatusUpdate)
	go ss.Listen(updates)

	// send first update
	updates <- statusUpdates[0]
	<-time.After(500 * time.Millisecond)
	expected := notifications[0]
	expected.Status = notification.StatusQueued
	assert.Equal(t, expected, r.Entries[expected.Identifier])

	// send second update
	updates <- statusUpdates[1]
	<-time.After(500 * time.Millisecond)
	expected = notifications[1]
	expected.Status = notification.StatusQueued
	assert.Equal(t, expected, r.Entries[expected.Identifier])
}

func TestProcessBatch_ChannelClose(t *testing.T) {
	r := testutil.GetFakeRepository()
	ss := ConcreteStatusService{
		repo:          r,
		sem:           make(chan struct{}, 1),
		buf:           make([]domain.StatusUpdate, 0, 1),
		maxLen:        1,
		tickerTimeout: 250 * time.Millisecond,
		jobTimeout:    1 * time.Second,
	}
	notifications := []notification.Notification{
		notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{}),
		notification.GetFakeNotification(notification.EmailDeviceType, notification.StatusUndelivered, 0, time.Time{}),
	}
	var statusUpdates []domain.StatusUpdate
	for _, n := range notifications {
		r.Entries[n.Identifier] = n
		statusUpdates = append(statusUpdates, domain.StatusUpdate{
			Identifier:  n.Identifier,
			FinalStatus: notification.StatusQueued,
		})
		n.Status = notification.StatusQueued
	}
	updates := make(chan domain.StatusUpdate)
	go ss.Listen(updates)

	// send first update
	updates <- statusUpdates[0]

	// close updates before ticker timeout
	<-time.After(200 * time.Millisecond)
	close(updates)

	expected := notifications[0]
	expected.Status = notification.StatusQueued
	assert.Equal(t, expected, r.Entries[expected.Identifier])
}
