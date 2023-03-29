package storage

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/practice-sem-2/notification-service/internal/pb"
	"github.com/sirupsen/logrus"
	"github.com/zyedidia/generic/multimap"
	"google.golang.org/protobuf/proto"
	"sync"
)

// TODO IDK how to correctly choose channel buffer size
const readerBufferSize = 16

type Worker func()

type NotificationListener struct {
	UserID   string
	store    *NotificationStore
	listener chan *pb.Notification
}

func (l *NotificationListener) Notifications() <-chan *pb.Notification {
	return l.listener
}

// Detach cancels listening and closes the listener channel
func (l *NotificationListener) Detach() {
	l.store.detach(l)
}

type NotificationStore struct {
	rm        sync.RWMutex
	consumer  sarama.ConsumerGroup
	listeners multimap.MultiMap[string, chan *pb.Notification]
	logger    *logrus.Logger
}

func NewNotificationStorage(consumer sarama.ConsumerGroup, logger *logrus.Logger) *NotificationStore {
	store := &NotificationStore{
		consumer:  consumer,
		listeners: multimap.NewMapSlice[string, chan *pb.Notification](),
		logger:    logger,
	}
	return store
}

func (s *NotificationStore) Notify(userID string, msg *pb.Notification) {
	s.rm.RLock()
	for _, reader := range s.listeners.Get(userID) {
		reader <- msg
	}
	defer s.rm.RUnlock()
}

func (s *NotificationStore) detach(listener *NotificationListener) {
	s.rm.Lock()
	defer s.rm.Unlock()
	s.listeners.Remove(listener.UserID, listener.listener)
	close(listener.listener)
}

func (s *NotificationStore) Run(ctx context.Context) error {
	c := Consumer{
		store: s,
	}
	for {
		err := s.consumer.Consume(ctx, []string{"notifications"}, &c)

		if errors.Is(err, context.Canceled) {
			return err
		}

		s.logger.
			WithField("error", err.Error()).
			Errorf("consuming error occured")
	}
}

// Listen returns a channel contains all notification connected to userID.
// TODO: replace pb.Notification with domain specific model
func (s *NotificationStore) Listen(userID string) NotificationListener {
	s.rm.Lock()
	defer s.rm.Unlock()
	listener := make(chan *pb.Notification, readerBufferSize)
	s.listeners.Put(userID, listener)
	return NotificationListener{
		UserID:   userID,
		store:    s,
		listener: listener,
	}
}

type Consumer struct {
	store *NotificationStore
}

func (c *Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	// Nothing to do...
	return nil
}

func (c *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	// Nothing to do...
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for {
		select {
		case msg, ok := <-claim.Messages():

			if !ok {
				return nil
			}

			notification := &pb.Notification{}
			err := proto.Unmarshal(msg.Value, notification)
			if err == nil {
				c.store.Notify(string(msg.Key), notification)
			} else {
				c.store.logger.
					WithField("error", err.Error()).
					Errorf("notification has invalid format")
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}

	}
}
