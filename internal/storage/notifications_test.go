package storage

import (
	"github.com/magiconair/properties/assert"
	"github.com/practice-sem-2/notification-service/internal/pb"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNotificationStore_Notify(t *testing.T) {
	store := NewNotificationStorage(nil, logrus.New())
	l := store.Listen("def")

	msg1 := &pb.Notification{
		Type: pb.NotificationType_NewMessageNotification,
		Notification: &pb.Notification_Message{
			Message: &pb.NewMessage{
				MessageId:   "aslkdf",
				CreatedAt:   "sdfjskldaf",
				Text:        "salfksaldkjf",
				FromUser:    "aslkfdjsdlaf",
				ChatId:      "def",
				Attachments: nil,
			},
		},
	}
	store.Notify("def", msg1)

	msg := <-l.Notifications()
	assert.Equal(t, msg, msg1, "should be equal")
}
