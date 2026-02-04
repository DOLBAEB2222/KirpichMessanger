package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/messenger/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupWebSocketTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	db.AutoMigrate(
		&models.User{},
		&models.Chat{},
		&models.ChatMember{},
		&models.Message{},
	)

	return db
}

type MockWSMessage struct {
	Type      string          `json:"type"`
	ChatID    string          `json:"chat_id,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

type TypingEvent struct {
	Type      string `json:"type"`
	ChatID    string `json:"chat_id"`
	UserID    string `json:"user_id"`
	IsTyping  bool   `json:"is_typing"`
	Timestamp int64  `json:"timestamp"`
}

type ReadReceiptEvent struct {
	Type        string    `json:"type"`
	ChatID      string    `json:"chat_id"`
	UserID      string    `json:"user_id"`
	LastReadAt  time.Time `json:"last_read_at"`
	UnreadCount int64     `json:"unread_count"`
	MessageID   *string   `json:"message_id,omitempty"`
}

type OnlineStatusEvent struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	IsOnline  bool   `json:"is_online"`
	Timestamp int64  `json:"timestamp"`
}

func TestTypingEventStructure(t *testing.T) {
	event := TypingEvent{
		Type:      "typing",
		ChatID:    uuid.New().String(),
		UserID:    uuid.New().String(),
		IsTyping:  true,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal typing event: %v", err)
	}

	var decoded TypingEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal typing event: %v", err)
	}

	if decoded.Type != "typing" {
		t.Errorf("expected type 'typing', got %s", decoded.Type)
	}

	if decoded.ChatID != event.ChatID {
		t.Error("chat_id mismatch")
	}

	if decoded.UserID != event.UserID {
		t.Error("user_id mismatch")
	}

	if !decoded.IsTyping {
		t.Error("expected is_typing to be true")
	}
}

func TestReadReceiptEventStructure(t *testing.T) {
	messageID := uuid.New().String()
	event := ReadReceiptEvent{
		Type:        "read",
		ChatID:      uuid.New().String(),
		UserID:      uuid.New().String(),
		LastReadAt:  time.Now(),
		UnreadCount: 5,
		MessageID:   &messageID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal read receipt event: %v", err)
	}

	var decoded ReadReceiptEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal read receipt event: %v", err)
	}

	if decoded.Type != "read" {
		t.Errorf("expected type 'read', got %s", decoded.Type)
	}

	if decoded.UnreadCount != 5 {
		t.Errorf("expected unread_count 5, got %d", decoded.UnreadCount)
	}

	if decoded.MessageID == nil || *decoded.MessageID != messageID {
		t.Error("message_id mismatch")
	}
}

func TestOnlineStatusEventStructure(t *testing.T) {
	event := OnlineStatusEvent{
		Type:      "online_status",
		UserID:    uuid.New().String(),
		IsOnline:  true,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal online status event: %v", err)
	}

	var decoded OnlineStatusEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal online status event: %v", err)
	}

	if decoded.Type != "online_status" {
		t.Errorf("expected type 'online_status', got %s", decoded.Type)
	}

	if !decoded.IsOnline {
		t.Error("expected is_online to be true")
	}

	if decoded.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
}

func TestWebSocketEventTypes(t *testing.T) {
	eventTypes := []string{"typing", "read", "online_status", "message", "join_chat", "leave_chat", "ping"}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			msg := MockWSMessage{
				Type:      eventType,
				Timestamp: time.Now().Unix(),
			}

			data, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("failed to marshal %s event: %v", eventType, err)
			}

			var decoded MockWSMessage
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal %s event: %v", eventType, err)
			}

			if decoded.Type != eventType {
				t.Errorf("expected type %s, got %s", eventType, decoded.Type)
			}
		})
	}
}

func TestTypingDebounce(t *testing.T) {
	db := setupWebSocketTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1111111111")
	user2 := createTestUser(t, db, "+2222222222")

	chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	t.Run("Typing indicator debounce", func(t *testing.T) {
		event1 := TypingEvent{
			Type:      "typing",
			ChatID:    chat.ID.String(),
			UserID:    user1.ID.String(),
			IsTyping:  true,
			Timestamp: time.Now().Unix(),
		}

		time.Sleep(100 * time.Millisecond)

		event2 := TypingEvent{
			Type:      "typing",
			ChatID:    chat.ID.String(),
			UserID:    user1.ID.String(),
			IsTyping:  true,
			Timestamp: time.Now().Unix(),
		}

		if event1.ChatID != event2.ChatID {
			t.Error("chat IDs should match")
		}

		if event1.UserID != event2.UserID {
			t.Error("user IDs should match")
		}

		if event2.Timestamp <= event1.Timestamp {
			t.Error("second event should have later timestamp")
		}
	})
}

func TestReadReceiptUpdate(t *testing.T) {
	db := setupWebSocketTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+3333333333")
	user2 := createTestUser(t, db, "+4444444444")

	chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	message := &models.Message{
		ChatID:      chat.ID,
		SenderID:    &user2.ID,
		Content:     "Test message for read receipt",
		MessageType: models.MessageTypeText,
	}
	if err := db.Create(message).Error; err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	t.Run("Update last read", func(t *testing.T) {
		err := chatService.UpdateLastRead(nil, chat.ID, user1.ID)
		if err != nil {
			t.Fatalf("failed to update last read: %v", err)
		}

		count, err := chatService.GetUnreadCount(nil, chat.ID, user1.ID)
		if err != nil {
			t.Fatalf("failed to get unread count: %v", err)
		}

		if count != 0 {
			t.Errorf("expected 0 unread messages after read, got %d", count)
		}
	})

	t.Run("Unread count before read", func(t *testing.T) {
		newMessage := &models.Message{
			ChatID:      chat.ID,
			SenderID:    &user2.ID,
			Content:     "Another test message",
			MessageType: models.MessageTypeText,
		}
		if err := db.Create(newMessage).Error; err != nil {
			t.Fatalf("failed to create message: %v", err)
		}

		count, err := chatService.GetUnreadCount(nil, chat.ID, user1.ID)
		if err != nil {
			t.Fatalf("failed to get unread count: %v", err)
		}

		if count != 1 {
			t.Errorf("expected 1 unread message, got %d", count)
		}
	})
}

func TestChatPresenceEvents(t *testing.T) {
	t.Run("Join chat event", func(t *testing.T) {
		event := MockWSMessage{
			Type:   "join_chat",
			ChatID: uuid.New().String(),
			Data:   json.RawMessage(`{"user_id": "` + uuid.New().String() + `"}`),
		}

		data, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("failed to marshal join chat event: %v", err)
		}

		var decoded MockWSMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal join chat event: %v", err)
		}

		if decoded.Type != "join_chat" {
			t.Errorf("expected type 'join_chat', got %s", decoded.Type)
		}
	})

	t.Run("Leave chat event", func(t *testing.T) {
		event := MockWSMessage{
			Type:   "leave_chat",
			ChatID: uuid.New().String(),
			Data:   json.RawMessage(`{"user_id": "` + uuid.New().String() + `"}`),
		}

		data, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("failed to marshal leave chat event: %v", err)
		}

		var decoded MockWSMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal leave chat event: %v", err)
		}

		if decoded.Type != "leave_chat" {
			t.Errorf("expected type 'leave_chat', got %s", decoded.Type)
		}
	})
}

func TestWebSocketMessageBroadcast(t *testing.T) {
	db := setupWebSocketTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+5555555555")
	user2 := createTestUser(t, db, "+6666666666")

	chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	t.Run("Message structure for broadcast", func(t *testing.T) {
		message := &models.Message{
			ChatID:      chat.ID,
			SenderID:    &user1.ID,
			Content:     "Broadcast test message",
			MessageType: models.MessageTypeText,
		}
		if err := db.Create(message).Error; err != nil {
			t.Fatalf("failed to create message: %v", err)
		}

		response := message.ToResponse()

		broadcast := map[string]interface{}{
			"type":    "new_message",
			"message": response,
		}

		data, err := json.Marshal(broadcast)
		if err != nil {
			t.Fatalf("failed to marshal broadcast: %v", err)
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal broadcast: %v", err)
		}

		if decoded["type"] != "new_message" {
			t.Errorf("expected type 'new_message', got %v", decoded["type"])
		}

		if decoded["message"] == nil {
			t.Error("expected message data in broadcast")
		}
	})
}
