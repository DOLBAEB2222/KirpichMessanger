package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ChatHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewChatHandler(db *gorm.DB, redisClient *redis.Client) *ChatHandler {
	return &ChatHandler{
		db:    db,
		redis: redisClient,
	}
}

func (h *ChatHandler) CreateChat(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.CreateChatRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.MemberIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one member is required",
		})
	}

	chat := models.Chat{
		Name:    req.Name,
		Type:    req.Type,
		OwnerID: &uid,
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&chat).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating chat: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create chat",
		})
	}

	members := []models.ChatMember{
		{
			ChatID: chat.ID,
			UserID: uid,
			Role:   models.MemberRoleAdmin,
		},
	}

	for _, memberIDStr := range req.MemberIDs {
		memberID, err := uuid.Parse(memberIDStr)
		if err != nil {
			continue
		}
		if memberID == uid {
			continue
		}
		
		members = append(members, models.ChatMember{
			ChatID: chat.ID,
			UserID: memberID,
			Role:   models.MemberRoleMember,
		})
	}

	if err := tx.Create(&members).Error; err != nil {
		tx.Rollback()
		log.Printf("Error adding chat members: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add members",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	if err := h.db.Preload("Members.User").First(&chat, chat.ID).Error; err != nil {
		log.Printf("Error loading chat with members: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(chat.ToResponse())
}

func (h *ChatHandler) GetUserChats(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var chatMembers []models.ChatMember
	if err := h.db.Where("user_id = ?", uid).Find(&chatMembers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	chatIDs := make([]uuid.UUID, len(chatMembers))
	for i, cm := range chatMembers {
		chatIDs[i] = cm.ChatID
	}

	var chats []models.Chat
	if err := h.db.Where("id IN ?", chatIDs).Order("last_message_at DESC").Find(&chats).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	responses := make([]models.ChatResponse, len(chats))
	for i, chat := range chats {
		responses[i] = chat.ToResponse()
	}

	return c.JSON(responses)
}

func (h *ChatHandler) GetChat(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	chatID := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	cid, err := uuid.Parse(chatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", cid, uid).First(&chatMember).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var chat models.Chat
	if err := h.db.Preload("Members.User").First(&chat, cid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Chat not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(chat.ToResponse())
}

func (h *ChatHandler) GetChatMessages(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	chatID := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	cid, err := uuid.Parse(chatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", cid, uid).First(&chatMember).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var messages []models.Message
	query := h.db.Where("chat_id = ? AND is_deleted = ?", cid, false).
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&messages).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	var total int64
	h.db.Model(&models.Message{}).Where("chat_id = ? AND is_deleted = ?", cid, false).Count(&total)

	responses := make([]models.MessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = msg.ToResponse()
	}

	return c.JSON(fiber.Map{
		"messages": responses,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": int64(offset+len(messages)) < total,
	})
}

func (h *ChatHandler) AddMember(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	chatID := c.Params("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	cid, err := uuid.Parse(chatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", cid, uid).First(&chatMember).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	if chatMember.Role != models.MemberRoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only admins can add members",
		})
	}

	var req models.AddMemberRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	newUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	role := models.MemberRoleMember
	if req.Role != "" {
		role = req.Role
	}

	newMember := models.ChatMember{
		ChatID: cid,
		UserID: newUserID,
		Role:   role,
	}

	if err := h.db.Create(&newMember).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add member",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Member added successfully",
	})
}

func (h *ChatHandler) RemoveMember(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	chatID := c.Params("id")
	targetUserID := c.Params("userId")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	cid, err := uuid.Parse(chatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	tuid, err := uuid.Parse(targetUserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid target user ID",
		})
	}

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", cid, uid).First(&chatMember).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	if chatMember.Role != models.MemberRoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only admins can remove members",
		})
	}

	if err := h.db.Where("chat_id = ? AND user_id = ?", cid, tuid).Delete(&models.ChatMember{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove member",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Member removed successfully",
	})
}
