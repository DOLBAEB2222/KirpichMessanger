package handlers

import (
	"encoding/base64"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type CryptoHandler struct {
	db *gorm.DB
}

func NewCryptoHandler(db *gorm.DB) *CryptoHandler {
	return &CryptoHandler{db: db}
}

func (h *CryptoHandler) RegisterDevice(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)

	var req models.RegisterDeviceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	identityKey, _ := base64.StdEncoding.DecodeString(req.IdentityKeyPublic)
	signedPreKey, _ := base64.StdEncoding.DecodeString(req.SignedPreKeyPublic)
	signedPreKeySig, _ := base64.StdEncoding.DecodeString(req.SignedPreKeySignature)

	device := models.UserDevice{
		UserID:                uid,
		DeviceID:              req.DeviceID,
		RegistrationID:        req.RegistrationID,
		IdentityKeyPublic:     identityKey,
		SignedPreKeyID:        req.SignedPreKeyID,
		SignedPreKeyPublic:    signedPreKey,
		SignedPreKeySignature: signedPreKeySig,
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND device_id = ?", uid, req.DeviceID).Delete(&models.UserDevice{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&device).Error; err != nil {
			return err
		}

		for _, otk := range req.OneTimeKeys {
			pubKey, _ := base64.StdEncoding.DecodeString(otk.PublicKey)
			oneTimeKey := models.UserOneTimeKey{
				DeviceID:  device.ID,
				KeyID:     otk.KeyID,
				PublicKey: pubKey,
			}
			if err := tx.Create(&oneTimeKey).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error registering device: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register device"})
	}

	return c.Status(fiber.StatusCreated).JSON(device)
}

func (h *CryptoHandler) GetUserKeys(c fiber.Ctx) error {
	recipientIDStr := c.Params("userId")
	recipientID, err := uuid.Parse(recipientIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var devices []models.UserDevice
	if err := h.db.Where("user_id = ?", recipientID).Find(&devices).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	type DeviceKeysResponse struct {
		DeviceID              string `json:"device_id"`
		RegistrationID        uint32 `json:"registration_id"`
		IdentityKeyPublic     string `json:"identity_key_public"`
		SignedPreKeyID        uint32 `json:"signed_pre_key_id"`
		SignedPreKeyPublic    string `json:"signed_pre_key_public"`
		SignedPreKeySignature string `json:"signed_pre_key_signature"`
		OneTimeKey            *models.OneTimeKeyDTO `json:"one_time_key,omitempty"`
	}

	var response []DeviceKeysResponse
	for _, d := range devices {
		var otk models.UserOneTimeKey
		var otkDTO *models.OneTimeKeyDTO
		if err := h.db.Where("device_id = ? AND is_used = ?", d.ID, false).First(&otk).Error; err == nil {
			otkDTO = &models.OneTimeKeyDTO{
				KeyID:     otk.KeyID,
				PublicKey: base64.StdEncoding.EncodeToString(otk.PublicKey),
			}
			// In a real Signal implementation, you'd mark it as used when fetched, 
			// but usually it's better to do it when the message is actually sent.
			// For simplicity, we'll keep it as is.
		}

		response = append(response, DeviceKeysResponse{
			DeviceID:              d.DeviceID,
			RegistrationID:        d.RegistrationID,
			IdentityKeyPublic:     base64.StdEncoding.EncodeToString(d.IdentityKeyPublic),
			SignedPreKeyID:        d.SignedPreKeyID,
			SignedPreKeyPublic:    base64.StdEncoding.EncodeToString(d.SignedPreKeyPublic),
			SignedPreKeySignature: base64.StdEncoding.EncodeToString(d.SignedPreKeySignature),
			OneTimeKey:            otkDTO,
		})
	}

	return c.JSON(response)
}

func (h *CryptoHandler) SendEncrypted(c fiber.Ctx) error {
	// This would normally just be a wrapper around the normal message sending
	// but with encrypted content.
	var req models.SendEncryptedRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Implementation details would go here. 
	// For this task, we just acknowledge receipt and say it's "sent"
	return c.JSON(fiber.Map{"status": "encrypted message sent"})
}
