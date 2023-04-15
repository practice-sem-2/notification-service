package storage

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/google/uuid"
	"github.com/practice-sem-2/notification-service/internal/models"
	"github.com/practice-sem-2/notification-service/internal/pb/chats/updates"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"testing"
	"time"
)

func TestUpdatesConsumer_Run(t *testing.T) {
	topic := "chat.updates"
	userId := uuid.New().String()
	chatId := uuid.New().String()
	expectedMsg := &models.MessageSent{
		UpdateMeta: models.UpdateMeta{
			Timestamp: time.Now().UTC().Truncate(time.Second),
			Audience:  []string{userId},
		},
		MessageID:   chatId,
		FromUser:    uuid.New().String(),
		ChatID:      uuid.New().String(),
		Text:        "Hello, world!",
		Attachments: make([]models.FileAttachment, 0),
	}

	protoMsg := &updates.Update{
		Meta: &updates.UpdateMeta{
			Timestamp: expectedMsg.Timestamp.Unix(),
			Audience:  expectedMsg.Audience,
		},
		Update: &updates.Update_Message{
			Message: &updates.MessageSent{
				MessageId: expectedMsg.MessageID,
				FromUser:  expectedMsg.FromUser,
				ChatId:    expectedMsg.ChatID,
				Text:      expectedMsg.Text,
			},
		},
	}
	value, _ := proto.Marshal(protoMsg)

	cfg := sarama.NewConfig()
	c := mocks.NewConsumer(t, cfg)
	c.SetTopicMetadata(map[string][]int32{
		"chat.updates": {1},
	})
	p1 := c.ExpectConsumePartition(topic, 1, sarama.OffsetNewest)
	key, _ := sarama.StringEncoder(chatId).Encode()
	p1.YieldMessage(&sarama.ConsumerMessage{
		Headers:        nil,
		Timestamp:      time.Now().UTC(),
		BlockTimestamp: time.Now().UTC(),
		Key:            key,
		Value:          value,
		Topic:          topic,
		Partition:      0,
		Offset:         0,
	})
	consumer := NewUpdatesConsumer(c, topic, logrus.New())
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	result := make(chan models.Update)
	defer cancel()
	go consumer.Run(ctx, result)
	actualMsg, ok := <-result
	assert.True(t, ok, "should correctly consume from channel")
	assert.Equal(t, expectedMsg, actualMsg)
}
