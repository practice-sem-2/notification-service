package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/practice-sem-2/notification-service/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/zyedidia/generic/queue"
	"strconv"
	"testing"
	"time"
)

type FakeConsumer struct {
	q *queue.Queue[models.Update]
}

func NewFakeConsumer() *FakeConsumer {
	return &FakeConsumer{q: queue.New[models.Update]()}
}

func (f *FakeConsumer) Fit(upds ...models.Update) {
	for _, u := range upds {
		f.q.Enqueue(u)
	}
}

func (f *FakeConsumer) Run(ctx context.Context, upds chan<- models.Update) error {
	f.q.Each(func(u models.Update) {
		upds <- u
	})
	return nil
}

func ReadWithTimeout[T any](t *testing.T, ch <-chan T, timeout time.Duration, msg string) *T {
	select {
	case data, ok := <-ch:
		assert.True(t, ok, "data must be read without errors")
		return &data
	case <-time.After(timeout):
		assert.Failf(t, "timeout error", msg)
	}
	return nil
}

func TestNotificationStore_Notify(t *testing.T) {

	const userId = "burenotti"
	var chatId = uuid.New().String()
	var fromUser = uuid.New().String()
	var messageId = uuid.New().String()
	var timeSent = time.Date(2023, 04, 15, 20, 0, 0, 0, time.UTC)

	store := NewNotificationStorage(logrus.New())
	l := store.Listen(userId)
	defer l.Detach()
	expectedMsg := models.MessageSent{
		UpdateMeta: models.UpdateMeta{
			Timestamp: timeSent,
			Audience:  []string{userId},
		},
		MessageID:   messageId,
		FromUser:    fromUser,
		ChatID:      chatId,
		Text:        "Hello, world!",
		ReplyTo:     nil,
		Attachments: make([]models.FileAttachment, 0),
	}
	store.Notify(userId, &expectedMsg)

	msg := *ReadWithTimeout[models.Update](t, l.Notifications(), 1*time.Second, "should correctly read msg")
	actualMsg := msg.(*models.MessageSent)
	assert.Equal(t, chatId, actualMsg.ChatID)
	assert.Equal(t, "Hello, world!", actualMsg.Text)
}

func TestFanoutUpdates(t *testing.T) {
	userId1 := uuid.New().String()
	userId2 := uuid.New().String()
	userId3 := uuid.New().String()
	upds := make(chan models.Update, 2)
	upds <- &models.ChatCreated{
		UpdateMeta: models.UpdateMeta{
			Timestamp: time.Now().UTC(),
			Audience:  []string{userId1, userId2},
		},
		ChatID:   uuid.New().String(),
		IsDirect: true,
		Members:  []string{uuid.New().String(), uuid.New().String()},
	}
	upds <- &models.ChatCreated{
		UpdateMeta: models.UpdateMeta{
			Timestamp: time.Now().UTC(),
			Audience:  []string{userId2, userId3},
		},
		ChatID:   uuid.New().String(),
		IsDirect: true,
		Members:  []string{uuid.New().String(), uuid.New().String()},
	}
	store := NewNotificationStorage(logrus.New())
	l1 := store.Listen(userId1)
	l2 := store.Listen(userId2)
	defer l1.Detach()
	defer l2.Detach()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go store.fanOutUpdates(ctx, upds)

	msg1 := *ReadWithTimeout(t, l1.Notifications(), 1*time.Second, "should correctly read msg")
	msg2 := *ReadWithTimeout(t, l2.Notifications(), 1*time.Second, "should correctly read msg")
	msg3 := *ReadWithTimeout(t, l2.Notifications(), 1*time.Second, "should correctly read msg")
	assert.Contains(t, msg1.GetAudience(), userId1)
	assert.Contains(t, msg1.GetAudience(), userId2)
	assert.Contains(t, msg2.GetAudience(), userId1)
	assert.Contains(t, msg2.GetAudience(), userId2)
	assert.Contains(t, msg3.GetAudience(), userId2)
	assert.Contains(t, msg3.GetAudience(), userId3)
}

func TestNotificationStorage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c1 := NewFakeConsumer()
	c2 := NewFakeConsumer()
	for i := 1; i <= 2; i++ {
		msg := &models.ChatCreated{
			UpdateMeta: models.UpdateMeta{
				Timestamp: time.Now().UTC(),
				Audience:  []string{strconv.Itoa(i)},
			},
			ChatID:   uuid.New().String(),
			IsDirect: true,
			Members:  []string{uuid.New().String(), uuid.New().String()},
		}
		c1.Fit(msg)
		c2.Fit(msg)
	}

	s := NewNotificationStorage(logrus.New(), c1, c2)
	l1 := s.Listen("1")
	l2 := s.Listen("2")
	go func(t *testing.T) {
		err := s.Run(ctx)
		assert.NoError(t, err)
	}(t)

	msg1 := *ReadWithTimeout(t, l1.Notifications(), 1*time.Second, "")
	msg2 := *ReadWithTimeout(t, l1.Notifications(), 1*time.Second, "")
	msg3 := *ReadWithTimeout(t, l2.Notifications(), 1*time.Second, "")
	msg4 := *ReadWithTimeout(t, l2.Notifications(), 1*time.Second, "")
	assert.Equal(t, "1", msg1.GetAudience()[0])
	assert.Equal(t, "1", msg2.GetAudience()[0])
	assert.Equal(t, "2", msg3.GetAudience()[0])
	assert.Equal(t, "2", msg4.GetAudience()[0])
}
