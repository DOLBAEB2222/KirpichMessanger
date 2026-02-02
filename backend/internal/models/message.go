package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
	MessageTypeFile  MessageType = "file"
)

type Message struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SenderID    *uuid.UUID   `gorm:"type:uuid" json:"sender_id"`
	ChatID      uuid.UUID    `gorm:"type:uuid;not null;index:idx_messages_chat_created" json:"chat_id"`
	Content     string       `gorm:"type:text;not null" json:"content"`
	MessageType MessageType  `gorm:"type:varchar(20);default:'text'" json:"message_type"`
	MediaURL    *string      `gorm:"type:text" json:"media_url,omitempty"`
	MediaSize   *int64       `json:"media_size,omitempty"`
	ReplyToID   *uuid.UUID   `gorm:"type:uuid" json:"reply_to_id,omitempty"`
	IsEdited    bool         `gorm:"default:false" json:"is_edited"`
	IsDeleted   bool         `gorm:"default:false" json:"is_deleted"`
	CreatedAt   time.Time    `gorm:"index:idx_messages_chat_created" json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	
	Sender      *User        `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Chat        *Chat        `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	ReplyTo     *Message     `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
}

type SendMessageRequest struct {
	ChatID      string      `json:"chat_id" validate:"required,uuid"`
	Content     string      `json:"content" validate:"required"`
	MessageType MessageType `json:"message_type" validate:"omitempty,oneof=text image video audio file"`
	ReplyToID   *string     `json:"reply_to_id" validate:"omitempty,uuid"`
}

type MessageResponse struct {
	ID          uuid.UUID    `json:"id"`
	SenderID    *uuid.UUID   `json:"sender_id"`
	ChatID      uuid.UUID    `json:"chat_id"`
	Content     string       `json:"content"`
	MessageType MessageType  `json:"message_type"`
	MediaURL    *string      `json:"media_url,omitempty"`
	ReplyToID   *uuid.UUID   `json:"reply_to_id,omitempty"`
	IsEdited    bool         `json:"is_edited"`
	CreatedAt   time.Time    `json:"created_at"`
	Sender      *UserResponse `json:"sender,omitempty"`
}

func (m *Message) ToResponse() MessageResponse {
	resp := MessageResponse{
		ID:          m.ID,
		SenderID:    m.SenderID,
		ChatID:      m.ChatID,
		Content:     m.Content,
		MessageType: m.MessageType,
		MediaURL:    m.MediaURL,
		ReplyToID:   m.ReplyToID,
		IsEdited:    m.IsEdited,
		CreatedAt:   m.CreatedAt,
	}
	
	if m.Sender != nil {
		senderResp := m.Sender.ToResponse()
		resp.Sender = &senderResp
	}
	
	return resp
}

type MediaFile struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid" json:"user_id"`
	FilePath  string     `gorm:"type:text;not null" json:"file_path"`
	FileName  string     `gorm:"type:varchar(255);not null" json:"file_name"`
	FileSize  int64      `gorm:"not null" json:"file_size"`
	MimeType  string     `gorm:"type:varchar(100);not null" json:"mime_type"`
	MessageID *uuid.UUID `gorm:"type:uuid" json:"message_id"`
	CreatedAt time.Time  `json:"created_at"`
}
