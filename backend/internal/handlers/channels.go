package handlers

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type ChannelHandler struct {
	db *gorm.DB
}

func NewChannelHandler(db *gorm.DB) *ChannelHandler {
	return &ChannelHandler{db: db}
}

func (h *ChannelHandler) CreateChannel(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.CreateChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	channel := models.Channel{
		Name:        req.Name,
		OwnerID:     uid,
		Description: req.Description,
		IsPublic:    isPublic,
	}

	if err := h.db.Create(&channel).Error; err != nil {
		log.Printf("Error creating channel: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create channel",
		})
	}

	subscriber := models.ChannelSubscriber{
		ChannelID: channel.ID,
		UserID:    uid,
	}
	h.db.Create(&subscriber)

	if err := h.db.Preload("Owner").First(&channel, channel.ID).Error; err != nil {
		log.Printf("Error loading channel with owner: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(channel.ToResponse())
}

func (h *ChannelHandler) GetChannel(c fiber.Ctx) error {
	channelID := c.Params("id")

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	var channel models.Channel
	if err := h.db.Preload("Owner").First(&channel, cid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Channel not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(channel.ToResponse())
}

func (h *ChannelHandler) Subscribe(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	channelID := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	var channel models.Channel
	if err := h.db.First(&channel, cid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Channel not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	var existingSubscriber models.ChannelSubscriber
	if err := h.db.Where("channel_id = ? AND user_id = ?", cid, uid).First(&existingSubscriber).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Already subscribed",
		})
	}

	subscriber := models.ChannelSubscriber{
		ChannelID: cid,
		UserID:    uid,
	}

	if err := h.db.Create(&subscriber).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to subscribe",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Subscribed successfully",
	})
}

func (h *ChannelHandler) Unsubscribe(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	channelID := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel ID",
		})
	}

	if err := h.db.Where("channel_id = ? AND user_id = ?", cid, uid).Delete(&models.ChannelSubscriber{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to unsubscribe",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Unsubscribed successfully",
	})
}
