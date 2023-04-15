package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/practice-sem-2/notification-service/internal/models"
	"github.com/practice-sem-2/notification-service/internal/pb/chats/updates"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"sync"
)

var (
	ErrParseMessage = errors.New("error while parsing message")
)

type UpdatesConsumer struct {
	consumer sarama.Consumer
	topic    string
	logger   *logrus.Logger
}

func NewUpdatesConsumer(c sarama.Consumer, topic string, l *logrus.Logger) *UpdatesConsumer {
	return &UpdatesConsumer{
		consumer: c,
		topic:    topic,
		logger:   l,
	}
}

func (c *UpdatesConsumer) Run(ctx context.Context, updates chan<- models.Update) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	partitions, err := c.consumer.Partitions(c.topic)

	if err != nil {
		return err
	}

	for _, part := range partitions {
		cons, err := c.consumer.ConsumePartition(c.topic, part, sarama.OffsetNewest)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(cons sarama.PartitionConsumer, c *UpdatesConsumer) {
			defer wg.Done()
			for {
				select {
				case _ = <-ctx.Done():
					cons.AsyncClose()
					break
				case msg, ok := <-cons.Messages():
					if !ok {
						break
					}
					upd, err := parseUpdate(msg)
					if err != nil {
						c.logger.Errorf("error occurred while parsing message %v:", err)
					}
					updates <- upd
				}
			}
		}(cons, c)
	}
	return nil
}

func parseUpdate(msg *sarama.ConsumerMessage) (models.Update, error) {
	u := &updates.Update{}
	err := proto.Unmarshal(msg.Value, u)

	if err != nil {
		return nil, err
	}

	meta := u.Meta
	switch u.Update.(type) {
	case *updates.Update_Message:
		upd := u.Update.(*updates.Update_Message).Message
		return MessageSentUpdateToDomain(meta, upd), nil
	case *updates.Update_CreatedChat:
		upd := u.Update.(*updates.Update_CreatedChat).CreatedChat
		ChatCreatedToDomain(meta, upd)
	case *updates.Update_DeletedChat:
	case *updates.Update_MemberAdded:
	case *updates.Update_MemberRemoved:
	}
	return nil, fmt.Errorf("%v: unsupported body type", ErrParseMessage)
}
