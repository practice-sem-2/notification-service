package storage

import (
	"github.com/practice-sem-2/notification-service/internal/models"
	"github.com/practice-sem-2/notification-service/internal/pb/chats/updates"
	"time"
)

func MessageSentUpdateToDomain(meta *updates.UpdateMeta, msg *updates.MessageSent) *models.MessageSent {
	attachLen := 0
	if msg.Attachments != nil {
		attachLen = len(msg.Attachments)
	}
	upd := &models.MessageSent{
		UpdateMeta: models.UpdateMeta{
			Timestamp: time.Unix(meta.Timestamp, 0).UTC(),
			Audience:  meta.Audience,
		},
		MessageID:   msg.MessageId,
		FromUser:    msg.FromUser,
		ChatID:      msg.ChatId,
		Text:        msg.Text,
		ReplyTo:     msg.ReplyTo,
		Attachments: make([]models.FileAttachment, 0, attachLen),
	}

	for _, file := range msg.Attachments {
		upd.Attachments = append(upd.Attachments, models.FileAttachment{
			FileID:   file.FileId,
			MimeType: file.MimeType,
		})
	}
	return upd
}

func ChatCreatedToDomain(meta *updates.UpdateMeta, msg *updates.ChatCreated) *models.ChatCreated {
	return &models.ChatCreated{
		UpdateMeta: models.UpdateMeta{
			Timestamp: time.Unix(meta.Timestamp, 0).UTC(),
			Audience:  meta.Audience,
		},
		ChatID:   msg.ChatId,
		IsDirect: msg.IsDirect,
		Members:  msg.Members,
	}
}
