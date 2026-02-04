package handlers

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/google/uuid"
    "github.com/messenger/backend/internal/models"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

type CallHandler struct {
    db        *gorm.DB
    redis     *redis.Client
    wsHandler *WebSocketHandler
}

func NewCallHandler(db *gorm.DB, redisClient *redis.Client, wsHandler *WebSocketHandler) *CallHandler {
    return &CallHandler{
        db:        db,
        redis:     redisClient,
        wsHandler: wsHandler,
    }
}

func (h *CallHandler) InitiateCall(c fiber.Ctx) error {
    userID := c.Locals("userID").(string)

    uid, err := uuid.Parse(userID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }

    var req models.InitiateCallRequest
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

    recipientID, err := uuid.Parse(req.RecipientID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid recipient ID",
        })
    }

    if uid == recipientID {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Cannot call yourself",
        })
    }

    var chatMember models.ChatMember
    if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, uid).First(&chatMember).Error; err != nil {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "You are not a member of this chat",
        })
    }

    var recipientMember models.ChatMember
    if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, recipientID).First(&recipientMember).Error; err != nil {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Recipient is not a member of this chat",
        })
    }

    var existingCall models.Call
    h.db.Where("(initiator_id = ? OR recipient_id = ?) AND (initiator_id = ? OR recipient_id = ?) AND status IN ?",
        uid, uid, recipientID, recipientID, []models.CallStatus{models.CallStatusRinging, models.CallStatusAccepted}).
        First(&existingCall)

    if existingCall.ID != uuid.Nil {
        return c.Status(fiber.StatusConflict).JSON(fiber.Map{
            "error":   "Active call already exists",
            "call_id": existingCall.ID,
        })
    }

    call := models.Call{
        ChatID:      chatID,
        InitiatorID: uid,
        RecipientID: recipientID,
        Type:        req.CallType,
        Status:      models.CallStatusRinging,
    }

    if err := h.db.Create(&call).Error; err != nil {
        log.Printf("Error creating call: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to initiate call",
        })
    }

    h.db.Preload("Initiator").Preload("Recipient").First(&call, call.ID)

    callJSON, _ := json.Marshal(map[string]interface{}{
        "type":      "call:initiate",
        "call":      call.ToResponse(),
        "chat_id":   chatID.String(),
        "initiator": call.Initiator.ToResponse(),
    })

    h.wsHandler.BroadcastToUser(recipientID.String(), callJSON)

    return c.Status(fiber.StatusCreated).JSON(call.ToResponse())
}

func (h *CallHandler) RespondToCall(c fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    callID := c.Params("call_id")

    uid, err := uuid.Parse(userID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }

    cid, err := uuid.Parse(callID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid call ID",
        })
    }

    var req models.CallResponseRequest
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    var call models.Call
    if err := h.db.First(&call, cid).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Call not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Database error",
        })
    }

    if call.RecipientID != uid {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "You are not the recipient of this call",
        })
    }

    if call.Status != models.CallStatusRinging {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Call is no longer active",
        })
    }

    if req.Accept {
        call.SetStarted()
    } else {
        call.Status = models.CallStatusRejected
        now := time.Now()
        call.EndedAt = &now
    }

    if err := h.db.Save(&call).Error; err != nil {
        log.Printf("Error updating call: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to respond to call",
        })
    }

    responseType := "call:rejected"
    if req.Accept {
        responseType = "call:accepted"
    }

    callJSON, _ := json.Marshal(map[string]interface{}{
        "type":    responseType,
        "call":    call.ToResponse(),
        "chat_id": call.ChatID.String(),
    })

    h.wsHandler.BroadcastToUser(call.InitiatorID.String(), callJSON)

    return c.JSON(call.ToResponse())
}

func (h *CallHandler) GetCall(c fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    callID := c.Params("call_id")

    uid, err := uuid.Parse(userID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }

    cid, err := uuid.Parse(callID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid call ID",
        })
    }

    var call models.Call
    if err := h.db.Preload("Initiator").Preload("Recipient").First(&call, cid).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Call not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Database error",
        })
    }

    if call.InitiatorID != uid && call.RecipientID != uid {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Access denied",
        })
    }

    return c.JSON(call.ToResponse())
}

func (h *CallHandler) EndCall(c fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    callID := c.Params("call_id")

    uid, err := uuid.Parse(userID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }

    cid, err := uuid.Parse(callID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid call ID",
        })
    }

    var call models.Call
    if err := h.db.First(&call, cid).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Call not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Database error",
        })
    }

    if call.InitiatorID != uid && call.RecipientID != uid {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Access denied",
        })
    }

    if call.Status == models.CallStatusEnded || call.Status == models.CallStatusRejected || call.Status == models.CallStatusMissed {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Call has already ended",
        })
    }

    call.SetEnded()

    if err := h.db.Save(&call).Error; err != nil {
        log.Printf("Error ending call: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to end call",
        })
    }

    callJSON, _ := json.Marshal(map[string]interface{}{
        "type":     "call:ended",
        "call":     call.ToResponse(),
        "chat_id":  call.ChatID.String(),
        "ended_by": uid.String(),
    })

    h.wsHandler.BroadcastToUser(call.InitiatorID.String(), callJSON)
    h.wsHandler.BroadcastToUser(call.RecipientID.String(), callJSON)

    return c.JSON(call.ToResponse())
}

func (h *CallHandler) GetICEServers(c fiber.Ctx) error {
    turnHost := getEnv("TURN_HOST", "localhost")
    turnPort := getEnv("TURN_PORT", "3478")
    turnUser := getEnv("TURN_USER", "user")
    turnPass := getEnv("TURN_PASS", "password")

    iceServers := models.ICEServersResponse{
        ICEServers: []models.ICEServer{
            {
                URLs:       []string{fmt.Sprintf("turn:%s:%s?transport=udp", turnHost, turnPort)},
                Username:   turnUser,
                Credential: turnPass,
            },
            {
                URLs: []string{fmt.Sprintf("stun:%s:%s", turnHost, turnPort)},
            },
            {
                URLs: []string{"stun:stun.l.google.com:19302"},
            },
        },
    }

    return c.JSON(iceServers)
}

func (h *CallHandler) SaveCallSignal(c fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    callID := c.Params("call_id")

    uid, err := uuid.Parse(userID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }

    cid, err := uuid.Parse(callID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid call ID",
        })
    }

    var req struct {
        Type string          `json:"type" validate:"required"`
        Data json.RawMessage `json:"data" validate:"required"`
    }
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    var call models.Call
    if err := h.db.First(&call, cid).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Call not found",
        })
    }

    if !call.CanJoin(uid) {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Access denied",
        })
    }

    signal := models.CallSignal{
        CallID: cid,
        Type:   req.Type,
        Data:   string(req.Data),
    }

    if err := h.db.Create(&signal).Error; err != nil {
        log.Printf("Error saving call signal: %v", err)
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message": "Signal saved",
    })
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
