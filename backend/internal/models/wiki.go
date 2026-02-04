package models

import (
	"time"

	"github.com/google/uuid"
)

type WikiPage struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ChannelID   uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_wiki_channel_slug" json:"channel_id"`
	Slug        string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_wiki_channel_slug" json:"slug"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	CreatedByID uuid.UUID  `gorm:"type:uuid;not null" json:"created_by_id"`
	ParentID    *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	IsPublished bool       `gorm:"default:true" json:"is_published"`
	Order       int        `gorm:"default:0" json:"order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	CreatedBy *User    `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Channel   *Channel `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Parent    *WikiPage `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []WikiPage `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

type CreateWikiPageRequest struct {
	ChannelID   string     `json:"channel_id" validate:"required,uuid"`
	Slug        string     `json:"slug" validate:"required,min=1,max=255,alphanum"`
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Content     string     `json:"content" validate:"required"`
	ParentID    *string    `json:"parent_id" validate:"omitempty,uuid"`
	IsPublished *bool      `json:"is_published"`
	Order       *int       `json:"order"`
}

type UpdateWikiPageRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=1,max=255"`
	Content     *string `json:"content" validate:"omitempty,min=1"`
	IsPublished *bool   `json:"is_published"`
	Order       *int    `json:"order"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid"`
}

type WikiPageResponse struct {
	ID          uuid.UUID        `json:"id"`
	ChannelID   uuid.UUID        `json:"channel_id"`
	Slug        string           `json:"slug"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	CreatedByID uuid.UUID        `json:"created_by_id"`
	ParentID    *uuid.UUID       `json:"parent_id,omitempty"`
	IsPublished bool             `json:"is_published"`
	Order       int              `json:"order"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	CreatedBy   *UserResponse    `json:"created_by,omitempty"`
	Children    []WikiPageResponse `json:"children,omitempty"`
}

func (w *WikiPage) ToResponse() WikiPageResponse {
	resp := WikiPageResponse{
		ID:          w.ID,
		ChannelID:   w.ChannelID,
		Slug:        w.Slug,
		Title:       w.Title,
		Content:     w.Content,
		CreatedByID: w.CreatedByID,
		ParentID:    w.ParentID,
		IsPublished: w.IsPublished,
		Order:       w.Order,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}

	if w.CreatedBy != nil {
		userResp := w.CreatedBy.ToResponse()
		resp.CreatedBy = &userResp
	}

	if len(w.Children) > 0 {
		resp.Children = make([]WikiPageResponse, len(w.Children))
		for i, child := range w.Children {
			resp.Children[i] = child.ToResponse()
		}
	}

	return resp
}
