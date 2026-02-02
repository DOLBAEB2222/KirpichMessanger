package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type SubscriptionHandler struct {
	db *gorm.DB
}

func NewSubscriptionHandler(db *gorm.DB) *SubscriptionHandler {
	return &SubscriptionHandler{db: db}
}

func (h *SubscriptionHandler) PurchaseSubscription(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.PurchaseSubscriptionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var user models.User
	if err := h.db.First(&user, uid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var activeSubscription models.Subscription
	if err := h.db.Where("user_id = ? AND status = ? AND end_date >= ?", 
		uid, models.SubscriptionStatusActive, time.Now()).First(&activeSubscription).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Active subscription already exists",
		})
	}

	amount := models.GetSubscriptionPrice(req.SubscriptionType)
	if amount == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subscription type",
		})
	}

	paymentMethod := "stub"
	if req.PaymentMethod != nil {
		paymentMethod = *req.PaymentMethod
	}

	startDate := time.Now()
	endDate := startDate.Add(models.GetSubscriptionDuration(req.SubscriptionType))

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	subscription := models.Subscription{
		UserID:    uid,
		Type:      req.SubscriptionType,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    models.SubscriptionStatusActive,
		AutoRenew: false,
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating subscription: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create subscription",
		})
	}

	stubNote := "MVP: Stub payment - no real charge. Real payment integration planned for stage 2."
	paymentLog := models.PaymentLog{
		UserID:           &uid,
		Amount:           amount,
		SubscriptionType: string(req.SubscriptionType),
		Status:           models.PaymentStatusCompletedStub,
		PaymentMethod:    paymentMethod,
		Notes:            &stubNote,
	}

	if err := tx.Create(&paymentLog).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating payment log: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to log payment",
		})
	}

	if err := tx.Model(&user).Update("is_premium", true).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating user premium status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user status",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	log.Printf("MVP Stub Payment: User %s purchased %s for $%.2f", uid, req.SubscriptionType, amount)

	return c.JSON(models.PurchaseResponse{
		Success: true,
		Message: fmt.Sprintf("MVP: Payment stub activated. No real charge applied. Subscription active until %s.", endDate.Format("2006-01-02")),
		Subscription: subscription.ToResponse(),
		PaymentLog:   paymentLog.ToResponse(),
	})
}

func (h *SubscriptionHandler) GetMySubscription(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var user models.User
	if err := h.db.First(&user, uid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var subscription models.Subscription
	if err := h.db.Where("user_id = ? AND status = ?", uid, models.SubscriptionStatusActive).
		Order("end_date DESC").First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{
				"is_premium":   false,
				"subscription": nil,
				"message":      "No active subscription",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if subscription.EndDate.Before(time.Now()) {
		h.db.Model(&subscription).Update("status", models.SubscriptionStatusExpired)
		h.db.Model(&user).Update("is_premium", false)
		
		return c.JSON(fiber.Map{
			"is_premium":   false,
			"subscription": nil,
			"message":      "Subscription expired",
		})
	}

	return c.JSON(fiber.Map{
		"is_premium":   true,
		"subscription": subscription.ToResponse(),
	})
}
