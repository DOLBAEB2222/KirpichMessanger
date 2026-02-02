package handlers

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v3"
	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/messenger/backend/pkg/auth"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type WebSocketHandler struct {
	db      *gorm.DB
	redis   *redis.Client
	clients sync.Map
}

type WSClient struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

type WSMessage struct {
	Type    string      `json:"type"`
	ChatID  string      `json:"chat_id,omitempty"`
	Content string      `json:"content,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewWebSocketHandler(db *gorm.DB, redisClient *redis.Client) *WebSocketHandler {
	return &WebSocketHandler{
		db:    db,
		redis: redisClient,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	tokenString := c.Query("token")
	if tokenString == "" {
		tokenString = c.Headers("Authorization")
	}

	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		log.Printf("WebSocket auth failed: %v", err)
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "Authentication required"}`))
		c.Close()
		return
	}

	client := &WSClient{
		UserID: claims.UserID,
		Conn:   c,
		Send:   make(chan []byte, 256),
	}

	h.clients.Store(client.UserID, client)
	defer func() {
		h.clients.Delete(client.UserID)
		close(client.Send)
	}()

	go h.writePump(client)
	go h.subscribeToUserChats(client)

	h.readPump(client)
}

func (h *WebSocketHandler) readPump(client *WSClient) {
	defer client.Conn.Close()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Invalid WebSocket message: %v", err)
			continue
		}

		h.handleIncomingMessage(client, &wsMsg)
	}
}

func (h *WebSocketHandler) writePump(client *WSClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) handleIncomingMessage(client *WSClient, msg *WSMessage) {
	switch msg.Type {
	case "message":
		h.handleSendMessage(client, msg)
	case "typing":
		h.handleTypingIndicator(client, msg)
	case "read":
		h.handleReadReceipt(client, msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *WebSocketHandler) handleSendMessage(client *WSClient, msg *WSMessage) {
	uid, err := uuid.Parse(client.UserID)
	if err != nil {
		return
	}

	chatID, err := uuid.Parse(msg.ChatID)
	if err != nil {
		return
	}

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, uid).First(&chatMember).Error; err != nil {
		return
	}

	message := models.Message{
		SenderID:    &uid,
		ChatID:      chatID,
		Content:     msg.Content,
		MessageType: models.MessageTypeText,
	}

	if err := h.db.Create(&message).Error; err != nil {
		log.Printf("Error creating message: %v", err)
		return
	}

	h.db.Preload("Sender").First(&message, message.ID)

	messageJSON, err := json.Marshal(map[string]interface{}{
		"type":    "new_message",
		"message": message.ToResponse(),
	})
	if err != nil {
		return
	}

	channelName := "chat:" + chatID.String()
	h.redis.Publish(context.Background(), channelName, messageJSON)
}

func (h *WebSocketHandler) handleTypingIndicator(client *WSClient, msg *WSMessage) {
	chatID := msg.ChatID
	if chatID == "" {
		return
	}

	typingMsg, _ := json.Marshal(map[string]interface{}{
		"type":    "typing",
		"chat_id": chatID,
		"user_id": client.UserID,
	})

	channelName := "chat:" + chatID
	h.redis.Publish(context.Background(), channelName, typingMsg)
}

func (h *WebSocketHandler) handleReadReceipt(client *WSClient, msg *WSMessage) {
	chatID := msg.ChatID
	if chatID == "" {
		return
	}

	uid, err := uuid.Parse(client.UserID)
	if err != nil {
		return
	}

	cid, err := uuid.Parse(chatID)
	if err != nil {
		return
	}

	h.db.Model(&models.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", cid, uid).
		Update("last_read_at", time.Now())

	readMsg, _ := json.Marshal(map[string]interface{}{
		"type":    "read",
		"chat_id": chatID,
		"user_id": client.UserID,
	})

	channelName := "chat:" + chatID
	h.redis.Publish(context.Background(), channelName, readMsg)
}

func (h *WebSocketHandler) subscribeToUserChats(client *WSClient) {
	uid, err := uuid.Parse(client.UserID)
	if err != nil {
		return
	}

	var chatMembers []models.ChatMember
	if err := h.db.Where("user_id = ?", uid).Find(&chatMembers).Error; err != nil {
		log.Printf("Error finding user chats: %v", err)
		return
	}

	ctx := context.Background()
	pubsub := h.redis.Subscribe(ctx)
	defer pubsub.Close()

	channels := make([]string, len(chatMembers))
	for i, cm := range chatMembers {
		channels[i] = "chat:" + cm.ChatID.String()
	}

	if len(channels) > 0 {
		if err := pubsub.Subscribe(ctx, channels...); err != nil {
			log.Printf("Error subscribing to Redis channels: %v", err)
			return
		}
	}

	ch := pubsub.Channel()
	for message := range ch {
		select {
		case client.Send <- []byte(message.Payload):
		default:
			log.Printf("Client send buffer full, dropping message")
		}
	}
}
