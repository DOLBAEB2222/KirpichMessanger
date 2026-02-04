package handlers

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type WikiHandler struct {
	db *gorm.DB
}

func NewWikiHandler(db *gorm.DB) *WikiHandler {
	return &WikiHandler{db: db}
}

func slugify(text string) string {
	reg := regexp.MustCompile(`[^\p{L}\p{N}]+`)
	slug := reg.ReplaceAllString(text, "-")
	slug = strings.Trim(slug, "-")
	slug = strings.ToLower(slug)
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

func (h *WikiHandler) CreateWikiPage(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.CreateWikiPageRequest
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

	if req.Slug == "" {
		req.Slug = slugify(req.Title)
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

	var existingPage models.WikiPage
	if err := h.db.Where("channel_id = ? AND slug = ?", channelID, req.Slug).First(&existingPage).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Wiki page with this slug already exists",
		})
	}

	isPublished := true
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}

	order := 0
	if req.Order != nil {
		order = *req.Order
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid parent ID",
			})
		}
		parentID = &pid
	}

	wikiPage := models.WikiPage{
		ChannelID:   channelID,
		Slug:        req.Slug,
		Title:       req.Title,
		Content:     req.Content,
		CreatedByID: uid,
		ParentID:    parentID,
		IsPublished: isPublished,
		Order:       order,
	}

	if err := h.db.Create(&wikiPage).Error; err != nil {
		log.Printf("Error creating wiki page: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create wiki page",
		})
	}

	if err := h.db.Preload("CreatedBy").Preload("Children").First(&wikiPage, wikiPage.ID).Error; err != nil {
		log.Printf("Error loading wiki page with relations: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(wikiPage.ToResponse())
}

func (h *WikiHandler) GetWikiPage(c fiber.Ctx) error {
	channelID := c.Params("channelId")
	slug := c.Params("slug")

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	var wikiPage models.WikiPage
	if err := h.db.Preload("CreatedBy").Preload("Children").Where("channel_id = ? AND slug = ?", cid, slug).First(&wikiPage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Wiki page not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(wikiPage.ToResponse())
}

func (h *WikiHandler) UpdateWikiPage(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	channelID := c.Params("channelId")
	slug := c.Params("slug")

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	var wikiPage models.WikiPage
	if err := h.db.Where("channel_id = ? AND slug = ?", cid, slug).First(&wikiPage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Wiki page not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if wikiPage.CreatedByID != uuid.MustParse(userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the creator can edit this page",
		})
	}

	var req models.UpdateWikiPageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title != nil {
		wikiPage.Title = *req.Title
	}
	if req.Content != nil {
		wikiPage.Content = *req.Content
	}
	if req.IsPublished != nil {
		wikiPage.IsPublished = *req.IsPublished
	}
	if req.Order != nil {
		wikiPage.Order = *req.Order
	}
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid parent ID",
			})
		}
		wikiPage.ParentID = &pid
	}

	if err := h.db.Save(&wikiPage).Error; err != nil {
		log.Printf("Error updating wiki page: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update wiki page",
		})
	}

	if err := h.db.Preload("CreatedBy").Preload("Children").First(&wikiPage, wikiPage.ID).Error; err != nil {
		log.Printf("Error loading wiki page with relations: %v", err)
	}

	return c.JSON(wikiPage.ToResponse())
}

func (h *WikiHandler) DeleteWikiPage(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	channelID := c.Params("channelId")
	slug := c.Params("slug")

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	var wikiPage models.WikiPage
	if err := h.db.Where("channel_id = ? AND slug = ?", cid, slug).First(&wikiPage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Wiki page not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if wikiPage.CreatedByID != uuid.MustParse(userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the creator can delete this page",
		})
	}

	if err := h.db.Delete(&wikiPage).Error; err != nil {
		log.Printf("Error deleting wiki page: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete wiki page",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Wiki page deleted successfully",
	})
}

func (h *WikiHandler) ListWikiPages(c fiber.Ctx) error {
	channelID := c.Params("channelId")

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	includeUnpublished := c.Query("include_unpublished", "false") == "true"

	query := h.db.Model(&models.WikiPage{}).Where("channel_id = ?", cid)

	if !includeUnpublished {
		query = query.Where("is_published = ?", true)
	}

	var wikiPages []models.WikiPage
	if err := query.Preload("CreatedBy").Order("order ASC, created_at DESC").Find(&wikiPages).Error; err != nil {
		log.Printf("Error listing wiki pages: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list wiki pages",
		})
	}

	responses := make([]models.WikiPageResponse, len(wikiPages))
	for i, page := range wikiPages {
		responses[i] = page.ToResponse()
	}

	return c.JSON(responses)
}

func (h *WikiHandler) GetWikiTree(c fiber.Ctx) error {
	channelID := c.Params("channelId")

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	var wikiPages []models.WikiPage
	if err := h.db.Where("channel_id = ? AND is_published = ?", cid, true).
		Order("order ASC, created_at ASC").
		Find(&wikiPages).Error; err != nil {
		log.Printf("Error listing wiki pages: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list wiki pages",
		})
	}

	tree := buildWikiTree(wikiPages, nil)

	return c.JSON(tree)
}

func buildWikiTree(pages []models.WikiPage, parentID *uuid.UUID) []models.WikiPageResponse {
	var result []models.WikiPageResponse

	for _, page := range pages {
		if (parentID == nil && page.ParentID == nil) || (parentID != nil && page.ParentID != nil && *page.ParentID == *parentID) {
			resp := page.ToResponse()
			resp.Children = buildWikiTree(pages, &page.ID)
			result = append(result, resp)
		}
	}

	return result
}
