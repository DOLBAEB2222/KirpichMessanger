package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RSSFeed struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ChannelID   uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_rss_channel" json:"channel_id"`
	URL         string     `gorm:"type:text;not null;index:idx_rss_url" json:"url"`
	Title       string     `gorm:"type:varchar(500);not null" json:"title"`
	Description *string    `gorm:"type:text" json:"description,omitempty"`
	IconURL     *string    `gorm:"type:text" json:"icon_url,omitempty"`
	AddedByID   uuid.UUID  `gorm:"type:uuid;not null" json:"added_by_id"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	LastFetched *time.Time `json:"last_fetched,omitempty"`
	FetchError  *string    `gorm:"type:text" json:"fetch_error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	AddedBy      *User       `gorm:"foreignKey:AddedByID" json:"added_by,omitempty"`
	Channel      *Channel    `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Items        []RSSItem   `gorm:"foreignKey:FeedID" json:"items,omitempty"`
}

type RSSItem struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	FeedID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_rss_item_feed" json:"feed_id"`
	GUID        string     `gorm:"type:varchar(500);not null;index:idx_rss_item_guid" json:"guid"`
	Title       string     `gorm:"type:varchar(500);not null" json:"title"`
	Description string     `gorm:"type:text;not null" json:"description"`
	Content     *string    `gorm:"type:text" json:"content,omitempty"`
	Link        string     `gorm:"type:text;not null" json:"link"`
	Author      *string    `gorm:"type:varchar(255)" json:"author,omitempty"`
	Category    *string    `gorm:"type:varchar(255)" json:"category,omitempty"`
	PublishedAt time.Time  `gorm:"not null;index:idx_rss_item_published" json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`

	Feed      *RSSFeed  `gorm:"foreignKey:FeedID" json:"feed,omitempty"`
	MessageID *uuid.UUID `gorm:"type:uuid" json:"message_id,omitempty"`
}

type RSSFeedStatus struct {
	TotalFeeds    int       `json:"total_feeds"`
	ActiveFeeds   int       `json:"active_feeds"`
	TotalItems    int       `json:"total_items"`
	LastFetchTime time.Time `json:"last_fetch_time"`
}

type CreateRSSFeedRequest struct {
	ChannelID string `json:"channel_id" validate:"required,uuid"`
	URL       string `json:"url" validate:"required,url"`
}

type UpdateRSSFeedRequest struct {
	IsActive *bool  `json:"is_active"`
	URL      *string `json:"url" validate:"omitempty,url"`
}

type RSSFeedResponse struct {
	ID          uuid.UUID     `json:"id"`
	ChannelID   uuid.UUID     `json:"channel_id"`
	URL         string        `json:"url"`
	Title       string        `json:"title"`
	Description *string       `json:"description,omitempty"`
	IconURL     *string       `json:"icon_url,omitempty"`
	AddedByID   uuid.UUID     `json:"added_by_id"`
	IsActive    bool          `json:"is_active"`
	LastFetched *time.Time    `json:"last_fetched,omitempty"`
	FetchError  *string       `json:"fetch_error,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	AddedBy     *UserResponse `json:"added_by,omitempty"`
	ItemCount   int           `json:"item_count"`
}

type RSSItemResponse struct {
	ID          uuid.UUID  `json:"id"`
	FeedID      uuid.UUID  `json:"feed_id"`
	GUID        string     `json:"guid"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Content     *string    `json:"content,omitempty"`
	Link        string     `json:"link"`
	Author      *string    `json:"author,omitempty"`
	Category    *string    `json:"category,omitempty"`
	PublishedAt time.Time  `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	MessageID   *uuid.UUID `json:"message_id,omitempty"`
}

func (r *RSSFeed) ToResponse() RSSFeedResponse {
	resp := RSSFeedResponse{
		ID:          r.ID,
		ChannelID:   r.ChannelID,
		URL:         r.URL,
		Title:       r.Title,
		Description: r.Description,
		IconURL:     r.IconURL,
		AddedByID:   r.AddedByID,
		IsActive:    r.IsActive,
		LastFetched: r.LastFetched,
		FetchError:  r.FetchError,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		ItemCount:   len(r.Items),
	}

	if r.AddedBy != nil {
		userResp := r.AddedBy.ToResponse()
		resp.AddedBy = &userResp
	}

	return resp
}

func (r *RSSItem) ToResponse() RSSItemResponse {
	return RSSItemResponse{
		ID:          r.ID,
		FeedID:      r.FeedID,
		GUID:        r.GUID,
		Title:       r.Title,
		Description: r.Description,
		Content:     r.Content,
		Link:        r.Link,
		Author:      r.Author,
		Category:    r.Category,
		PublishedAt: r.PublishedAt,
		CreatedAt:   r.CreatedAt,
		MessageID:   r.MessageID,
	}
}

func (r *RSSFeed) BeforeCreate(tx *gorm.DB) error {
	if r.AddedByID == uuid.Nil {
		return nil
	}
	return nil
}
