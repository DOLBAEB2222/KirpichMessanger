package handlers

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MessageHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewMessageHandler(db *gorm.DB, redisClient *redis.Client) *MessageHandler {
	return &MessageHandler{
		db:    db,
		redis: redisClient,
	}
}

func (h *MessageHandler) SendMessage(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.SendMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	chatID, err := uuid.Parse(req.ChatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, uid).First(&chatMember).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not a member of this chat",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	messageType := models.MessageTypeText
	if req.MessageType != "" {
		messageType = req.MessageType
	}

	var replyToID *uuid.UUID
	if req.ReplyToID != nil {
		rid, err := uuid.Parse(*req.ReplyToID)
		if err == nil {
			replyToID = &rid
		}
	}

	message := models.Message{
		SenderID:    &uid,
		ChatID:      chatID,
		Content:     req.Content,
		MessageType: messageType,
		ReplyToID:   replyToID,
	}

	if err := h.db.Create(&message).Error; err != nil {
		log.Printf("Error creating message: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send message",
		})
	}

	if err := h.db.Preload("Sender").First(&message, message.ID).Error; err != nil {
		log.Printf("Error loading message with sender: %v", err)
	}

	messageJSON, err := json.Marshal(message.ToResponse())
	if err == nil {
		channelName := "chat:" + chatID.String()
		h.redis.Publish(c.Context(), channelName, messageJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(message.ToResponse())
}

func (h *MessageHandler) GetMessage(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	messageID := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	mid, err := uuid.Parse(messageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid message ID",
		})
	}

	var message models.Message
	if err := h.db.Preload("Sender").First(&message, mid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Message not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", message.ChatID, uid).First(&chatMember).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(message.ToResponse())
}
