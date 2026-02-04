package handlers

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/messenger/backend/pkg/media"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MessageHandler struct {
	db             *gorm.DB
	redis          *redis.Client
	mediaUploader  *media.MediaUploader
}

func NewMessageHandler(db *gorm.DB, redisClient *redis.Client) *MessageHandler {
	return &MessageHandler{
		db:            db,
		redis:         redisClient,
		mediaUploader: media.NewMediaUploader(),
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

func (h *MessageHandler) SendMediaMessage(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	chatIDStr := c.FormValue("chat_id")
	if chatIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "chat_id is required",
		})
	}

	chatID, err := uuid.Parse(chatIDStr)
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

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	allowedTypes := mergeMaps(media.AllowedImageTypes, media.AllowedVideoTypes, media.AllowedAudioTypes)
	
	if err := h.mediaUploader.ValidateFile(file, allowedTypes); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	openedFile, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer openedFile.Close()

	result, err := h.mediaUploader.SaveFile(openedFile, file, uid)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	messageType := determineMessageType(result.MimeType)
	content := c.FormValue("content")
	if content == "" {
		content = file.Filename
	}

	message := models.Message{
		SenderID:    &uid,
		ChatID:      chatID,
		Content:     content,
		MessageType: messageType,
		MediaURL:    &result.FilePath,
		MediaSize:   &result.FileSize,
	}

	if err := h.db.Create(&message).Error; err != nil {
		h.mediaUploader.DeleteFile(result.FilePath)
		if result.Thumbnail != nil {
			h.mediaUploader.DeleteFile(*result.Thumbnail)
		}
		log.Printf("Error creating message: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send message",
		})
	}

	mediaFile := models.MediaFile{
		UserID:    &uid,
		FilePath:  result.FilePath,
		FileName:  result.FileName,
		FileSize:  result.FileSize,
		MimeType:  result.MimeType,
		MessageID: &message.ID,
	}

	if err := h.db.Create(&mediaFile).Error; err != nil {
		log.Printf("Error saving media file record: %v", err)
	}

	if err := h.db.Preload("Sender").First(&message, message.ID).Error; err != nil {
		log.Printf("Error loading message with sender: %v", err)
	}

	messageJSON, err := json.Marshal(message.ToResponse())
	if err == nil {
		channelName := "chat:" + chatID.String()
		h.redis.Publish(c.Context(), channelName, messageJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": message.ToResponse(),
		"media": fiber.Map{
			"file_path":  result.FilePath,
			"file_size":  result.FileSize,
			"mime_type":  result.MimeType,
			"width":      result.Width,
			"height":     result.Height,
			"thumbnail":  result.Thumbnail,
			"compressed": result.Compressed,
		},
	})
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

func (h *MessageHandler) DeleteMessage(c fiber.Ctx) error {
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
	if err := h.db.First(&message, mid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Message not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if message.SenderID == nil || *message.SenderID != uid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only delete your own messages",
		})
	}

	if err := h.db.Model(&message).Update("is_deleted", true).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete message",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Message deleted successfully",
	})
}

func (h *MessageHandler) EditMessage(c fiber.Ctx) error {
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

	var req struct {
		Content string `json:"content" validate:"required"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var message models.Message
	if err := h.db.First(&message, mid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Message not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if message.SenderID == nil || *message.SenderID != uid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only edit your own messages",
		})
	}

	if message.IsDeleted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot edit deleted messages",
		})
	}

	updates := map[string]interface{}{
		"content":  req.Content,
		"is_edited": true,
	}

	if err := h.db.Model(&message).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to edit message",
		})
	}

	h.db.Preload("Sender").First(&message, message.ID)

	messageJSON, err := json.Marshal(message.ToResponse())
	if err == nil {
		channelName := "chat:" + message.ChatID.String()
		h.redis.Publish(c.Context(), channelName, messageJSON)
	}

	return c.JSON(message.ToResponse())
}

func determineMessageType(mimeType string) models.MessageType {
	if strings.HasPrefix(mimeType, "image/") {
		return models.MessageTypeImage
	}
	if strings.HasPrefix(mimeType, "video/") {
		return models.MessageTypeVideo
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return models.MessageTypeAudio
	}
	return models.MessageTypeFile
}

func mergeMaps(maps ...map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func (h *MessageHandler) GetMediaFile(c fiber.Ctx) error {
	filePath := c.Params("*")
	if filePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File path required",
		})
	}

	timestamp := c.QueryInt("t", 0)
	if timestamp == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	fullPath := h.mediaUploader.GetRelativePath(filePath)
	fullPath = "uploads/" + filePath

	if _, err := media.GetImageDimensions(fullPath); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	return c.SendFile(fullPath)
}
