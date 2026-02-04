package models

import (
    "time"

    "github.com/google/uuid"
)

type ChatType string
type MemberRole string

const (
    ChatTypeDM    ChatType = "dm"
    ChatTypeGroup ChatType = "group"

    MemberRoleAdmin  MemberRole = "admin"
    MemberRoleMember MemberRole = "member"
)

type Chat struct {
    ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    Name          *string    `gorm:"type:varchar(255)" json:"name,omitempty"`
    Type          ChatType   `gorm:"type:varchar(20);not null" json:"type"`
    OwnerID       *uuid.UUID `gorm:"type:uuid" json:"owner_id"`
    AvatarURL     *string    `gorm:"type:text" json:"avatar_url,omitempty"`
    Description   *string    `gorm:"type:text" json:"description,omitempty"`
    MemberCount   int        `gorm:"default:0" json:"member_count"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
    LastMessageAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"last_message_at"`
    
    Owner         *User        `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
    Members       []ChatMember `gorm:"foreignKey:ChatID" json:"members,omitempty"`
}

type ChatMember struct {
    ChatID     uuid.UUID  `gorm:"primaryKey" json:"chat_id"`
    UserID     uuid.UUID  `gorm:"primaryKey" json:"user_id"`
    Role       MemberRole `gorm:"type:varchar(20);default:'member'" json:"role"`
    JoinedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
    LastReadAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"last_read_at"`
    
    User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Chat       *Chat      `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
}

type CreateChatRequest struct {
    Name      *string   `json:"name" validate:"omitempty,max=255"`
    Type      ChatType  `json:"type" validate:"required,oneof=dm group"`
    MemberIDs []string  `json:"member_ids" validate:"required,min=1"`
}

type AddMemberRequest struct {
    UserID string     `json:"user_id" validate:"required,uuid"`
    Role   MemberRole `json:"role" validate:"omitempty,oneof=admin member"`
}

type ChatResponse struct {
    ID            uuid.UUID       `json:"id"`
    Name          *string         `json:"name,omitempty"`
    Type          ChatType        `json:"type"`
    OwnerID       *uuid.UUID      `json:"owner_id"`
    AvatarURL     *string         `json:"avatar_url,omitempty"`
    MemberCount   int             `json:"member_count"`
    CreatedAt     time.Time       `json:"created_at"`
    LastMessageAt time.Time       `json:"last_message_at"`
    Members       []MemberResponse `json:"members,omitempty"`
}

type ChatWithLastMessageResponse struct {
    ChatResponse
    LastMessage *MessageResponse `json:"last_message,omitempty"`
    UnreadCount int64            `json:"unread_count"`
}

type MemberResponse struct {
    UserID   uuid.UUID    `json:"user_id"`
    Role     MemberRole   `json:"role"`
    JoinedAt time.Time    `json:"joined_at"`
    User     *UserResponse `json:"user,omitempty"`
}

func (c *Chat) ToResponse() ChatResponse {
    resp := ChatResponse{
        ID:            c.ID,
        Name:          c.Name,
        Type:          c.Type,
        OwnerID:       c.OwnerID,
        AvatarURL:     c.AvatarURL,
        MemberCount:   c.MemberCount,
        CreatedAt:     c.CreatedAt,
        LastMessageAt: c.LastMessageAt,
    }
    
    if len(c.Members) > 0 {
        resp.Members = make([]MemberResponse, len(c.Members))
        for i, m := range c.Members {
            memberResp := MemberResponse{
                UserID:   m.UserID,
                Role:     m.Role,
                JoinedAt: m.JoinedAt,
            }
            if m.User != nil {
                userResp := m.User.ToResponse()
                memberResp.User = &userResp
            }
            resp.Members[i] = memberResp
        }
    }
    
    return resp
}

type Channel struct {
    ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    Name            string    `gorm:"type:varchar(255);not null" json:"name"`
    OwnerID         uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
    Description     *string   `gorm:"type:text" json:"description,omitempty"`
    AvatarURL       *string   `gorm:"type:text" json:"avatar_url,omitempty"`
    SubscriberCount int       `gorm:"default:0" json:"subscriber_count"`
    IsPublic        bool      `gorm:"default:true" json:"is_public"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    
    Owner           *User               `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}

type ChannelSubscriber struct {
    ChannelID    uuid.UUID `gorm:"primaryKey" json:"channel_id"`
    UserID       uuid.UUID `gorm:"primaryKey" json:"user_id"`
    SubscribedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"subscribed_at"`
    
    User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Channel      *Channel  `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
}

type CreateChannelRequest struct {
    Name        string  `json:"name" validate:"required,max=255"`
    Description *string `json:"description" validate:"omitempty,max=1000"`
    IsPublic    *bool   `json:"is_public"`
}

type ChannelResponse struct {
    ID              uuid.UUID    `json:"id"`
    Name            string       `json:"name"`
    OwnerID         uuid.UUID    `json:"owner_id"`
    Description     *string      `json:"description,omitempty"`
    AvatarURL       *string      `json:"avatar_url,omitempty"`
    SubscriberCount int          `json:"subscriber_count"`
    IsPublic        bool         `json:"is_public"`
    CreatedAt       time.Time    `json:"created_at"`
    Owner           *UserResponse `json:"owner,omitempty"`
}

func (c *Channel) ToResponse() ChannelResponse {
    resp := ChannelResponse{
        ID:              c.ID,
        Name:            c.Name,
        OwnerID:         c.OwnerID,
        Description:     c.Description,
        AvatarURL:       c.AvatarURL,
        SubscriberCount: c.SubscriberCount,
        IsPublic:        c.IsPublic,
        CreatedAt:       c.CreatedAt,
    }
    
    if c.Owner != nil {
        ownerResp := c.Owner.ToResponse()
        resp.Owner = &ownerResp
    }
    
    return resp
}
