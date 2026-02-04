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
	db          *gorm.DB
	redis       *redis.Client
	clients     sync.Map
	typingMu    sync.RWMutex
	typingUsers map[string]map[string]time.Time
}

type WSClient struct {
	UserID    string
	Conn      *websocket.Conn
	Send      chan []byte
	ChatRooms sync.Map
}

type WSMessage struct {
	Type      string      `json:"type"`
	ChatID    string      `json:"chat_id,omitempty"`
	Content   string      `json:"content,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

type TypingEvent struct {
	Type      string `json:"type"`
	ChatID    string `json:"chat_id"`
	UserID    string `json:"user_id"`
	IsTyping  bool   `json:"is_typing"`
	Timestamp int64  `json:"timestamp"`
}

type ReadReceiptEvent struct {
	Type           string    `json:"type"`
	ChatID         string    `json:"chat_id"`
	UserID         string    `json:"user_id"`
	LastReadAt     time.Time `json:"last_read_at"`
	UnreadCount    int64     `json:"unread_count"`
	MessageID      *string   `json:"message_id,omitempty"`
}

type OnlineStatusEvent struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	IsOnline  bool   `json:"is_online"`
	LastSeen  string `json:"last_seen,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type ChatPresenceEvent struct {
	Type     string `json:"type"`
	ChatID   string `json:"chat_id"`
	UserID   string `json:"user_id"`
	IsJoined bool   `json:"is_joined"`
}

func NewWebSocketHandler(db *gorm.DB, redisClient *redis.Client) *WebSocketHandler {
	return &WebSocketHandler{
		db:          db,
		redis:       redisClient,
		typingUsers: make(map[string]map[string]time.Time),
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
		h.broadcastOnlineStatus(client.UserID, false)
		close(client.Send)
	}()

	h.broadcastOnlineStatus(client.UserID, true)

	go h.writePump(client)
	go h.subscribeToUserChats(client)
	go h.sendUserChats(client)

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

		wsMsg.Timestamp = time.Now().Unix()
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
	case "join_chat":
		h.handleJoinChat(client, msg)
	case "leave_chat":
		h.handleLeaveChat(client, msg)
	case "ping":
		h.handlePing(client)
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

	h.clearTypingIndicator(chatID.String(), client.UserID)
}

func (h *WebSocketHandler) handleTypingIndicator(client *WSClient, msg *WSMessage) {
	chatID := msg.ChatID
	if chatID == "" {
		return
	}

	h.typingMu.Lock()
	if h.typingUsers[chatID] == nil {
		h.typingUsers[chatID] = make(map[string]time.Time)
	}
	h.typingUsers[chatID][client.UserID] = time.Now()
	h.typingMu.Unlock()

	typingMsg, _ := json.Marshal(TypingEvent{
		Type:      "typing",
		ChatID:    chatID,
		UserID:    client.UserID,
		IsTyping:  true,
		Timestamp: time.Now().Unix(),
	})

	channelName := "chat:" + chatID
	h.redis.Publish(context.Background(), channelName, typingMsg)

	time.AfterFunc(3*time.Second, func() {
		h.clearTypingIndicator(chatID, client.UserID)
	})
}

func (h *WebSocketHandler) clearTypingIndicator(chatID, userID string) {
	h.typingMu.Lock()
	if users, ok := h.typingUsers[chatID]; ok {
		if lastTime, exists := users[userID]; exists {
			if time.Since(lastTime) >= 3*time.Second {
				delete(users, userID)
				h.typingMu.Unlock()

				typingMsg, _ := json.Marshal(TypingEvent{
					Type:      "typing",
					ChatID:    chatID,
					UserID:    userID,
					IsTyping:  false,
					Timestamp: time.Now().Unix(),
				})

				channelName := "chat:" + chatID
				h.redis.Publish(context.Background(), channelName, typingMsg)
				return
			}
		}
	}
	h.typingMu.Unlock()
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

	now := time.Now()
	h.db.Model(&models.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", cid, uid).
		Update("last_read_at", now)

	var unreadCount int64
	h.db.Model(&models.Message{}).
		Where("chat_id = ? AND sender_id != ? AND created_at > ?", cid, uid, now).
		Count(&unreadCount)

	var lastMessage models.Message
	h.db.Where("chat_id = ?", cid).Order("created_at DESC").First(&lastMessage)

	var messageID *string
	if lastMessage.ID != uuid.Nil {
		id := lastMessage.ID.String()
		messageID = &id
	}

	readMsg, _ := json.Marshal(ReadReceiptEvent{
		Type:        "read",
		ChatID:      chatID,
		UserID:      client.UserID,
		LastReadAt:  now,
		UnreadCount: unreadCount,
		MessageID:   messageID,
	})

	channelName := "chat:" + chatID
	h.redis.Publish(context.Background(), channelName, readMsg)
}

func (h *WebSocketHandler) handleJoinChat(client *WSClient, msg *WSMessage) {
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

	var chatMember models.ChatMember
	if err := h.db.Where("chat_id = ? AND user_id = ?", cid, uid).First(&chatMember).Error; err != nil {
		return
	}

	client.ChatRooms.Store(chatID, true)

	presenceMsg, _ := json.Marshal(ChatPresenceEvent{
		Type:     "chat_presence",
		ChatID:   chatID,
		UserID:   client.UserID,
		IsJoined: true,
	})

	channelName := "chat:" + chatID
	h.redis.Publish(context.Background(), channelName, presenceMsg)
}

func (h *WebSocketHandler) handleLeaveChat(client *WSClient, msg *WSMessage) {
	chatID := msg.ChatID
	if chatID == "" {
		return
	}

	client.ChatRooms.Delete(chatID)

	presenceMsg, _ := json.Marshal(ChatPresenceEvent{
		Type:     "chat_presence",
		ChatID:   chatID,
		UserID:   client.UserID,
		IsJoined: false,
	})

	channelName := "chat:" + chatID
	h.redis.Publish(context.Background(), channelName, presenceMsg)
}

func (h *WebSocketHandler) handlePing(client *WSClient) {
	pongMsg, _ := json.Marshal(map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now().Unix(),
	})
	
	select {
	case client.Send <- pongMsg:
	default:
	}
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

	personalChannel := "user:" + client.UserID
	if err := pubsub.Subscribe(ctx, personalChannel); err != nil {
		log.Printf("Error subscribing to personal channel: %v", err)
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

func (h *WebSocketHandler) sendUserChats(client *WSClient) {
	time.Sleep(100 * time.Millisecond)

	uid, err := uuid.Parse(client.UserID)
	if err != nil {
		return
	}

	var chatMembers []models.ChatMember
	if err := h.db.Where("user_id = ?", uid).Find(&chatMembers).Error; err != nil {
		return
	}

	chatIDs := make([]string, len(chatMembers))
	for i, cm := range chatMembers {
		chatIDs[i] = cm.ChatID.String()
	}

	chatsMsg, _ := json.Marshal(map[string]interface{}{
		"type":     "user_chats",
		"chat_ids": chatIDs,
	})

	select {
	case client.Send <- chatsMsg:
	default:
	}
}

func (h *WebSocketHandler) broadcastOnlineStatus(userID string, isOnline bool) {
	statusMsg, _ := json.Marshal(OnlineStatusEvent{
		Type:      "online_status",
		UserID:    userID,
		IsOnline:  isOnline,
		Timestamp: time.Now().Unix(),
	})

	uid, err := uuid.Parse(userID)
	if err != nil {
		return
	}

	var chatMembers []models.ChatMember
	if err := h.db.Where("user_id = ?", uid).Find(&chatMembers).Error; err != nil {
		return
	}

	broadcastTo := make(map[string]bool)
	for _, cm := range chatMembers {
		var otherMembers []models.ChatMember
		h.db.Where("chat_id = ? AND user_id != ?", cm.ChatID, uid).Find(&otherMembers)
		for _, om := range otherMembers {
			broadcastTo[om.UserID.String()] = true
		}
	}

	ctx := context.Background()
	for otherUserID := range broadcastTo {
		channelName := "user:" + otherUserID
		h.redis.Publish(ctx, channelName, statusMsg)
	}
}

func (h *WebSocketHandler) BroadcastToChat(chatID string, message interface{}) {
	msgJSON, err := json.Marshal(message)
	if err != nil {
		return
	}

	channelName := "chat:" + chatID
	h.redis.Publish(context.Background(), channelName, msgJSON)
}
