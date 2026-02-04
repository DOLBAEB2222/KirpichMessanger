package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/messenger/backend/internal/services"
	"github.com/messenger/backend/pkg/auth"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db          *gorm.DB
	userService *services.UserService
}

func NewAuthHandler(db *gorm.DB, redis *redis.Client) *AuthHandler {
	return &AuthHandler{
		db:          db,
		userService: services.NewUserService(db, redis),
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Phone != "" {
		var existingUser models.User
		if err := h.db.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Phone number already registered",
			})
		}
	}

	if req.Email != nil && *req.Email != "" {
		var existingUser models.User
		if err := h.db.Where("email = ?", *req.Email).First(&existingUser).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already registered",
			})
		}
	}

	if err := auth.ValidatePassword(req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process password",
		})
	}

	user := models.User{
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Username:     req.Username,
		IsPremium:    false,
	}

	if err := h.db.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	tokenPair, err := auth.GenerateTokenPair(user.ID.String())
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.AuthResponse{
		UserID:       user.ID,
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var user models.User
	err := h.db.Where("phone = ? OR email = ?", req.PhoneOrEmail, req.PhoneOrEmail).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		log.Printf("Error finding user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	now := time.Now().UTC()
	user.LastSeenAt = &now
	h.db.Save(&user)

	tokenPair, err := auth.GenerateTokenPair(user.ID.String())
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	return c.JSON(models.AuthResponse{
		UserID:       user.ID,
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	var req models.RefreshTokenRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	tokenPair, err := auth.GenerateTokenPair(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	return c.JSON(models.RefreshResponse{
		Token:     tokenPair.AccessToken,
		ExpiresIn: tokenPair.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	return c.JSON(models.LogoutResponse{
		Message: "Logged out successfully",
	})
}

func (h *AuthHandler) GetMe(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var user models.User
	if err := h.db.First(&user, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.JSON(user.ToPrivateProfile())
}
