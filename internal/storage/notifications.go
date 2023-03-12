package storage

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/practice-sem-2/notification-service/internal/pb"
	"github.com/sirupsen/logrus"
	"github.com/zyedidia/generic/multimap"
	"sync"
)

// TODO IDK how to correctly choose channel buffer size
const readerBufferSize = 16

type Worker func()

type NotificationReader struct {
	id     int64
	store  *NotificationStore
	reader chan pb.Notification
}

func (r *NotificationReader) Iter() <-chan pb.Notification {
	return r.reader
}

func (r *NotificationReader) Close() {

}

type NotificationStore struct {
	consumer sarama.ConsumerGroup
	readers  multimap.MultiMap[string, chan pb.Notification]
	wg       *sync.WaitGroup
	logger   *logrus.Logger
	ctx      context.Context
}

func NewNotificationStorage(consumer sarama.ConsumerGroup, logger *logrus.Logger) (*NotificationStore, error) {
	wg := &sync.WaitGroup{}

	store := &NotificationStore{
		consumer: consumer,
		readers:  multimap.NewMapSlice[string, chan pb.Notification](),
		wg:       wg,
		logger:   logger,
	}

	c := Consumer{
		store: store,
	}
	err := consumer.Consume(store.ctx, []string{"notifications"}, &c)
	return store
}

func (s *NotificationStore) Run() {

}

// Listen TODO replace pb.Notification with domain specific model
func (s *NotificationStore) Listen(userID string) <-chan pb.Notification {
	reader := make(chan pb.Notification, readerBufferSize)
	s.readers.Put(userID, reader)
	return reader
}

func (s *NotificationStore) Close() {
	s.done <- struct{}{}
	s.wg.Wait()
}

type Consumer struct {
	store *NotificationStore
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	//TODO implement me
	panic("implement me")
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	//TODO implement me
	panic("implement me")
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	//TODO implement me
	panic("implement me")
}
