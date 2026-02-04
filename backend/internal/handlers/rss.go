package handlers

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type RSSHandler struct {
	db        *gorm.DB
	client    *http.Client
}

func NewRSSHandler(db *gorm.DB) *RSSHandler {
	return &RSSHandler{
		db:     db,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Language    string `xml:"language"`
		Image       struct {
			URL string `xml:"url"`
		} `xml:"image"`
		Items []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Content     string `xml:"encoded"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Author      string `xml:"author"`
	Category    string `xml:"category"`
}

func (h *RSSHandler) AddRSSFeed(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.CreateRSSFeedRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	var channel models.Channel
	if err := h.db.First(&channel, channelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Channel not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	var existingFeed models.RSSFeed
	if err := h.db.Where("channel_id = ?", channelID).First(&existingFeed).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "RSS feed already exists for this channel",
		})
	}

	parsedFeed, err := h.fetchRSSFeed(req.URL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch RSS feed: %v", err),
		})
	}

	iconURL := ""
	if parsedFeed.Channel.Image.URL != "" {
		iconURL = parsedFeed.Channel.Image.URL
	}

	rssFeed := models.RSSFeed{
		ChannelID:   channelID,
		URL:         req.URL,
		Title:       parsedFeed.Channel.Title,
		Description: &parsedFeed.Channel.Description,
		IconURL:     &iconURL,
		AddedByID:   uid,
		IsActive:    true,
	}

	if err := h.db.Create(&rssFeed).Error; err != nil {
		log.Printf("Error creating RSS feed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add RSS feed",
		})
	}

	for _, item := range parsedFeed.Channel.Items {
		publishedAt := time.Now()
		if item.PubDate != "" {
			if t, err := parseRSSDate(item.PubDate); err == nil {
				publishedAt = t
			}
		}

		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}

		description := stripHTML(item.Description)
		content := ""
		if item.Content != "" {
			content = stripHTML(item.Content)
		}

		rssItem := models.RSSItem{
			FeedID:      rssFeed.ID,
			GUID:        guid,
			Title:       item.Title,
			Description: description,
			Content:     &content,
			Link:        item.Link,
			PublishedAt: publishedAt,
		}

		if item.Author != "" {
			rssItem.Author = &item.Author
		}

		if item.Category != "" {
			rssItem.Category = &item.Category
		}

		h.db.Create(&rssItem)
	}

	now := time.Now()
	rssFeed.LastFetched = &now
	h.db.Save(&rssFeed)

	if err := h.db.Preload("AddedBy").First(&rssFeed, rssFeed.ID).Error; err != nil {
		log.Printf("Error loading RSS feed with added by: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(rssFeed.ToResponse())
}

func (h *RSSHandler) GetRSSFeed(c fiber.Ctx) error {
	id := c.Params("id")

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid RSS feed ID",
		})
	}

	var rssFeed models.RSSFeed
	if err := h.db.Preload("AddedBy").First(&rssFeed, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "RSS feed not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(rssFeed.ToResponse())
}

func (h *RSSHandler) UpdateRSSFeed(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid RSS feed ID",
		})
	}

	var rssFeed models.RSSFeed
	if err := h.db.First(&rssFeed, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "RSS feed not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if rssFeed.AddedByID != uid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the creator can update this RSS feed",
		})
	}

	var req models.UpdateRSSFeedRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.IsActive != nil {
		rssFeed.IsActive = *req.IsActive
	}

	if req.URL != nil && *req.URL != rssFeed.URL {
		parsedFeed, err := h.fetchRSSFeed(*req.URL)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to fetch RSS feed: %v", err),
			})
		}

		rssFeed.URL = *req.URL
		rssFeed.Title = parsedFeed.Channel.Title
		rssFeed.Description = &parsedFeed.Channel.Description

		iconURL := ""
		if parsedFeed.Channel.Image.URL != "" {
			iconURL = parsedFeed.Channel.Image.URL
		}
		rssFeed.IconURL = &iconURL
	}

	if err := h.db.Save(&rssFeed).Error; err != nil {
		log.Printf("Error updating RSS feed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update RSS feed",
		})
	}

	if err := h.db.Preload("AddedBy").First(&rssFeed, rssFeed.ID).Error; err != nil {
		log.Printf("Error loading RSS feed with added by: %v", err)
	}

	return c.JSON(rssFeed.ToResponse())
}

func (h *RSSHandler) DeleteRSSFeed(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid RSS feed ID",
		})
	}

	var rssFeed models.RSSFeed
	if err := h.db.First(&rssFeed, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "RSS feed not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if rssFeed.AddedByID != uid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the creator can delete this RSS feed",
		})
	}

	if err := h.db.Delete(&rssFeed).Error; err != nil {
		log.Printf("Error deleting RSS feed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete RSS feed",
		})
	}

	return c.JSON(fiber.Map{
		"message": "RSS feed deleted successfully",
	})
}

func (h *RSSHandler) ListRSSFeeds(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var feeds []models.RSSFeed
	if err := h.db.Joins("JOIN channel_subscribers ON channel_subscribers.channel_id = rss_feeds.channel_id").
		Where("channel_subscribers.user_id = ?", uid).
		Preload("AddedBy").
		Find(&feeds).Error; err != nil {
		log.Printf("Error listing RSS feeds: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list RSS feeds",
		})
	}

	responses := make([]models.RSSFeedResponse, len(feeds))
	for i, feed := range feeds {
		responses[i] = feed.ToResponse()
	}

	return c.JSON(responses)
}

func (h *RSSHandler) GetRSSFeedItems(c fiber.Ctx) error {
	id := c.Params("id")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	if limit > 100 {
		limit = 100
	}

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid RSS feed ID",
		})
	}

	var items []models.RSSItem
	if err := h.db.Where("feed_id = ?", sid).
		Order("published_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		log.Printf("Error listing RSS items: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list RSS items",
		})
	}

	responses := make([]models.RSSItemResponse, len(items))
	for i, item := range items {
		responses[i] = item.ToResponse()
	}

	return c.JSON(responses)
}

func (h *RSSHandler) RefreshRSSFeed(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid RSS feed ID",
		})
	}

	var rssFeed models.RSSFeed
	if err := h.db.First(&rssFeed, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "RSS feed not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if rssFeed.AddedByID != uid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the creator can refresh this RSS feed",
		})
	}

	parsedFeed, err := h.fetchRSSFeed(rssFeed.URL)
	if err != nil {
		now := time.Now()
		errorMsg := err.Error()
		rssFeed.FetchError = &errorMsg
		rssFeed.LastFetched = &now
		h.db.Save(&rssFeed)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch RSS feed: %v", err),
		})
	}

	rssFeed.Title = parsedFeed.Channel.Title
	rssFeed.Description = &parsedFeed.Channel.Description
	rssFeed.FetchError = nil

	iconURL := ""
	if parsedFeed.Channel.Image.URL != "" {
		iconURL = parsedFeed.Channel.Image.URL
	}
	rssFeed.IconURL = &iconURL

	newItems := 0
	for _, item := range parsedFeed.Channel.Items {
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}

		var existingItem models.RSSItem
		if err := h.db.Where("feed_id = ? AND guid = ?", rssFeed.ID, guid).First(&existingItem).Error; err == nil {
			continue
		}

		publishedAt := time.Now()
		if item.PubDate != "" {
			if t, err := parseRSSDate(item.PubDate); err == nil {
				publishedAt = t
			}
		}

		description := stripHTML(item.Description)
		content := ""
		if item.Content != "" {
			content = stripHTML(item.Content)
		}

		rssItem := models.RSSItem{
			FeedID:      rssFeed.ID,
			GUID:        guid,
			Title:       item.Title,
			Description: description,
			Content:     &content,
			Link:        item.Link,
			PublishedAt: publishedAt,
		}

		if item.Author != "" {
			rssItem.Author = &item.Author
		}

		if item.Category != "" {
			rssItem.Category = &item.Category
		}

		if err := h.db.Create(&rssItem).Error; err == nil {
			newItems++
		}
	}

	now := time.Now()
	rssFeed.LastFetched = &now
	h.db.Save(&rssFeed)

	if err := h.db.Preload("AddedBy").First(&rssFeed, rssFeed.ID).Error; err != nil {
		log.Printf("Error loading RSS feed with added by: %v", err)
	}

	return c.JSON(fiber.Map{
		"feed":      rssFeed.ToResponse(),
		"new_items": newItems,
	})
}

func (h *RSSHandler) fetchRSSFeed(url string) (*RSSFeed, error) {
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	return &feed, nil
}

func parseRSSDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Now(), nil
}

func stripHTML(s string) string {
	var result strings.Builder
	var inTag bool

	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	str := result.String()
	str = strings.Join(strings.Fields(str), " ")
	return str
}
