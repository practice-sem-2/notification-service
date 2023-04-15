package models

import "time"

type Update interface {
	GetTime() time.Time
	GetAudience() []string
}

type FileAttachment struct {
	MimeType string `validate:"required" db:"mime_type"`
	FileID   string `validate:"required,uuid" db:"file_id"`
}

type UpdateMeta struct {
	Timestamp time.Time
	Audience  []string
}

func (m *UpdateMeta) GetTime() time.Time {
	return m.Timestamp
}

func (m *UpdateMeta) GetAudience() []string {
	return m.Audience
}

type MessageSent struct {
	UpdateMeta
	MessageID   string  `validate:"required,uuid"`
	FromUser    string  `validate:"required"`
	ChatID      string  `validate:"required,uuid"`
	Text        string  `validate:"required_without=Attachments"`
	ReplyTo     *string `validate:"uuid"`
	Attachments []FileAttachment
}

type ChatCreated struct {
	UpdateMeta
	ChatID   string `validate:"required,uuid"`
	IsDirect bool   `validate:"required"`
	Members  []string
}
