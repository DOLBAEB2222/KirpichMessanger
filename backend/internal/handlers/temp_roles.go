package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type TempRoleHandler struct {
	db *gorm.DB
}

func NewTempRoleHandler(db *gorm.DB) *TempRoleHandler {
	return &TempRoleHandler{db: db}
}

func (h *TempRoleHandler) GrantTempRole(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	grantedByID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.CreateTempRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	hasPermission := h.checkPermission(grantedByID, req.TargetID, req.TargetType, []string{"manage_roles", "admin"})
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions to grant roles",
		})
	}

	expiresAt := time.Now().Add(time.Duration(req.DurationHours) * time.Hour)

	var existingRole models.TempRole
	err = h.db.Where("target_id = ? AND target_type = ? AND user_id = ? AND is_active = ?",
		req.TargetID, req.TargetType, req.UserID, true).
		Where("expires_at > ?", time.Now()).
		First(&existingRole).Error

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User already has an active temporary role for this target",
		})
	}

	tempRole := models.TempRole{
		TargetID:    req.TargetID,
		TargetType:  req.TargetType,
		UserID:      req.UserID,
		RoleType:    req.RoleType,
		CustomName:  req.CustomName,
		Permissions: req.Permissions,
		GrantedByID: grantedByID,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	if err := h.db.Create(&tempRole).Error; err != nil {
		log.Printf("Error creating temp role: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to grant temporary role",
		})
	}

	if err := h.db.Preload("GrantedBy").Preload("User").First(&tempRole, tempRole.ID).Error; err != nil {
		log.Printf("Error loading temp role with relations: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(tempRole.ToResponse())
}

func (h *TempRoleHandler) RevokeTempRole(c fiber.Ctx) error {
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
			"error": "Invalid role ID",
		})
	}

	var tempRole models.TempRole
	if err := h.db.First(&tempRole, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Temporary role not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	hasPermission := h.checkPermission(uid, tempRole.TargetID, tempRole.TargetType, []string{"manage_roles", "admin"})
	if !hasPermission && tempRole.GrantedByID != uid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions to revoke this role",
		})
	}

	tempRole.IsActive = false
	if err := h.db.Save(&tempRole).Error; err != nil {
		log.Printf("Error revoking temp role: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke temporary role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Temporary role revoked successfully",
	})
}

func (h *TempRoleHandler) UpdateTempRole(c fiber.Ctx) error {
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
			"error": "Invalid role ID",
		})
	}

	var tempRole models.TempRole
	if err := h.db.First(&tempRole, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Temporary role not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	hasPermission := h.checkPermission(uid, tempRole.TargetID, tempRole.TargetType, []string{"manage_roles", "admin"})
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions to update this role",
		})
	}

	var req models.UpdateTempRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.IsEnabled != nil {
		tempRole.IsActive = *req.IsEnabled
	}

	if req.DurationHours != nil {
		tempRole.ExpiresAt = time.Now().Add(time.Duration(*req.DurationHours) * time.Hour)
	}

	if err := h.db.Save(&tempRole).Error; err != nil {
		log.Printf("Error updating temp role: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update temporary role",
		})
	}

	if err := h.db.Preload("GrantedBy").Preload("User").First(&tempRole, tempRole.ID).Error; err != nil {
		log.Printf("Error loading temp role with relations: %v", err)
	}

	return c.JSON(tempRole.ToResponse())
}

func (h *TempRoleHandler) GetTempRole(c fiber.Ctx) error {
	id := c.Params("id")

	sid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	var tempRole models.TempRole
	if err := h.db.Preload("GrantedBy").Preload("User").First(&tempRole, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Temporary role not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(tempRole.ToResponse())
}

func (h *TempRoleHandler) ListTargetRoles(c fiber.Ctx) error {
	targetID := c.Params("targetId")
	targetType := c.Params("targetType")

	tid, err := uuid.Parse(targetID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid target ID",
		})
	}

	var target models.TempRoleTargetType
	if targetType == "chat" {
		target = models.TempRoleTargetChat
	} else if targetType == "channel" {
		target = models.TempRoleTargetChannel
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid target type",
		})
	}

	includeExpired := c.Query("include_expired", "false") == "true"
	includeInactive := c.Query("include_inactive", "false") == "true"

	query := h.db.Model(&models.TempRole{}).
		Where("target_id = ? AND target_type = ?", tid, target)

	if !includeExpired {
		query = query.Where("expires_at > ?", time.Now())
	}

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	var tempRoles []models.TempRole
	if err := query.Preload("GrantedBy").Preload("User").
		Order("created_at DESC").Find(&tempRoles).Error; err != nil {
		log.Printf("Error listing temp roles: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list temporary roles",
		})
	}

	responses := make([]models.TempRoleResponse, len(tempRoles))
	for i, role := range tempRoles {
		responses[i] = role.ToResponse()
	}

	return c.JSON(responses)
}

func (h *TempRoleHandler) ListUserRoles(c fiber.Ctx) error {
	userID := c.Params("userId")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	includeExpired := c.Query("include_expired", "false") == "true"
	includeInactive := c.Query("include_inactive", "false") == "true"

	query := h.db.Model(&models.TempRole{}).Where("user_id = ?", uid)

	if !includeExpired {
		query = query.Where("expires_at > ?", time.Now())
	}

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	var tempRoles []models.TempRole
	if err := query.Preload("GrantedBy").Preload("User").
		Order("expires_at ASC").Find(&tempRoles).Error; err != nil {
		log.Printf("Error listing temp roles: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list temporary roles",
		})
	}

	responses := make([]models.TempRoleResponse, len(tempRoles))
	for i, role := range tempRoles {
		responses[i] = role.ToResponse()
	}

	return c.JSON(responses)
}

func (h *TempRoleHandler) CheckUserPermission(c fiber.Ctx) error {
	userID := c.Params("userId")
	targetID := c.Params("targetId")
	targetType := c.Query("target_type")
	permission := c.Query("permission")

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	tid, err := uuid.Parse(targetID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid target ID",
		})
	}

	if targetType == "" || permission == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "target_type and permission are required",
		})
	}

	var target models.TempRoleTargetType
	if targetType == "chat" {
		target = models.TempRoleTargetChat
	} else if targetType == "channel" {
		target = models.TempRoleTargetChannel
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid target type",
		})
	}

	hasPermission := h.checkPermission(uid, tid, target, []string{permission, "admin"})

	return c.JSON(fiber.Map{
		"user_id":       uid,
		"target_id":     tid,
		"target_type":   target,
		"permission":    permission,
		"has_permission": hasPermission,
	})
}

func (h *TempRoleHandler) checkPermission(userID uuid.UUID, targetID uuid.UUID, targetType models.TempRoleTargetType, permissions []string) bool {
	var tempRole models.TempRole
	err := h.db.Where("user_id = ? AND target_id = ? AND target_type = ? AND is_active = ?",
		userID, targetID, targetType, true).
		Where("expires_at > ?", time.Now()).
		First(&tempRole).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false
		}
		return false
	}

	for _, perm := range permissions {
		for _, rolePerm := range tempRole.Permissions {
			if rolePerm == perm || rolePerm == "admin" {
				return true
			}
		}
	}

	return false
}
