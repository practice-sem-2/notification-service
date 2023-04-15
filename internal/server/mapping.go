package server

import (
	"github.com/practice-sem-2/notification-service/internal/models"
	"github.com/practice-sem-2/notification-service/internal/pb/notify"
)

func NotificationFromUpdate(upd models.Update) *notify.Notification {
	switch upd.(type) {
	case *models.MessageSent:
		return makeMessageSentNotification(upd.(*models.MessageSent))
	}
	return nil
}

func makeMessageSentNotification(upd *models.MessageSent) *notify.Notification {
	attachments := make([]*notify.Attachment, 0, len(upd.Attachments))
	for _, att := range upd.Attachments {
		attachments = append(attachments, &notify.Attachment{
			FileId:   att.FileID,
			MimeType: att.MimeType,
		})
	}
	return &notify.Notification{
		Notification: &notify.Notification_Message{
			Message: &notify.NewMessage{
				MessageId:   upd.MessageID,
				CreatedAt:   upd.Timestamp.UTC().Unix(),
				Text:        upd.Text,
				FromUser:    upd.FromUser,
				ChatId:      upd.ChatID,
				ReplyTo:     upd.ReplyTo,
				Attachments: attachments,
			},
		},
	}
}
