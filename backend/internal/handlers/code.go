package handlers

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type CodeHandler struct {
	db *gorm.DB
}

func NewCodeHandler(db *gorm.DB) *CodeHandler {
	return &CodeHandler{db: db}
}

func (h *CodeHandler) CreateCodeSnippet(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.CreateCodeSnippetRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	messageID, err := uuid.Parse(req.MessageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid message ID",
		})
	}

	chatID, err := uuid.Parse(req.ChatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	var message models.Message
	if err := h.db.Where("id = ? AND chat_id = ?", messageID, chatID).First(&message).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Message not found in this chat",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	var existingSnippet models.CodeSnippet
	if err := h.db.Where("message_id = ?", messageID).First(&existingSnippet).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Code snippet already exists for this message",
		})
	}

	codeSnippet := models.CodeSnippet{
		MessageID:   messageID,
		ChatID:      chatID,
		Language:    req.Language,
		Code:        req.Code,
		FileName:    req.FileName,
		CreatedByID: uid,
	}

	if err := h.db.Create(&codeSnippet).Error; err != nil {
		log.Printf("Error creating code snippet: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create code snippet",
		})
	}

	if err := h.db.Preload("CreatedBy").First(&codeSnippet, codeSnippet.ID).Error; err != nil {
		log.Printf("Error loading code snippet with creator: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(codeSnippet.ToResponse())
}

func (h *CodeHandler) GetCodeSnippet(c fiber.Ctx) error {
	id := c.Params("id")

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid code snippet ID",
		})
	}

	var codeSnippet models.CodeSnippet
	if err := h.db.Preload("CreatedBy").First(&codeSnippet, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Code snippet not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(codeSnippet.ToResponse())
}

func (h *CodeHandler) UpdateCodeSnippet(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid code snippet ID",
		})
	}

	var codeSnippet models.CodeSnippet
	if err := h.db.First(&codeSnippet, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Code snippet not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if codeSnippet.CreatedByID != uuid.MustParse(userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the creator can edit this code snippet",
		})
	}

	var req models.UpdateCodeSnippetRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Language != nil {
		codeSnippet.Language = *req.Language
	}
	if req.Code != nil {
		codeSnippet.Code = *req.Code
	}
	if req.FileName != nil {
		codeSnippet.FileName = req.FileName
	}

	if err := h.db.Save(&codeSnippet).Error; err != nil {
		log.Printf("Error updating code snippet: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update code snippet",
		})
	}

	if err := h.db.Preload("CreatedBy").First(&codeSnippet, codeSnippet.ID).Error; err != nil {
		log.Printf("Error loading code snippet with creator: %v", err)
	}

	return c.JSON(codeSnippet.ToResponse())
}

func (h *CodeHandler) DeleteCodeSnippet(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid code snippet ID",
		})
	}

	var codeSnippet models.CodeSnippet
	if err := h.db.First(&codeSnippet, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Code snippet not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if codeSnippet.CreatedByID != uuid.MustParse(userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the creator can delete this code snippet",
		})
	}

	if err := h.db.Delete(&codeSnippet).Error; err != nil {
		log.Printf("Error deleting code snippet: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete code snippet",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Code snippet deleted successfully",
	})
}

func (h *CodeHandler) ListCodeSnippetsByChat(c fiber.Ctx) error {
	chatID := c.Params("chatId")

	cid, err := uuid.Parse(chatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	language := c.Query("language")
	query := h.db.Model(&models.CodeSnippet{}).Where("chat_id = ?", cid)

	if language != "" {
		query = query.Where("language = ?", language)
	}

	var codeSnippets []models.CodeSnippet
	if err := query.Preload("CreatedBy").Order("created_at DESC").Find(&codeSnippets).Error; err != nil {
		log.Printf("Error listing code snippets: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list code snippets",
		})
	}

	responses := make([]models.CodeSnippetResponse, len(codeSnippets))
	for i, snippet := range codeSnippets {
		responses[i] = snippet.ToResponse()
	}

	return c.JSON(responses)
}

func (h *CodeHandler) GetCodeSnippetByMessage(c fiber.Ctx) error {
	messageID := c.Params("messageId")

	mid, err := uuid.Parse(messageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid message ID",
		})
	}

	var codeSnippet models.CodeSnippet
	if err := h.db.Preload("CreatedBy").Where("message_id = ?", mid).First(&codeSnippet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Code snippet not found for this message",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(codeSnippet.ToResponse())
}
