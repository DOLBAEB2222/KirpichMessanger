package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/messenger/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

func createTestUser(t *testing.T, db *gorm.DB, phone string) *models.User {
	user := &models.User{
		Phone:        phone,
		PasswordHash: "hashedpassword",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestChatService_GetOrCreateDMChat(t *testing.T) {
	db := setupTestDB(t)
	
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1234567890")
	user2 := createTestUser(t, db, "+0987654321")

	t.Run("Create new DM chat", func(t *testing.T) {
		chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
		if err != nil {
			t.Fatalf("failed to create DM chat: %v", err)
		}

		if chat.Type != models.ChatTypeDM {
			t.Errorf("expected chat type to be 'dm', got %s", chat.Type)
		}

		if chat.ID == uuid.Nil {
			t.Error("expected chat ID to be set")
		}

		var members []models.ChatMember
		db.Where("chat_id = ?", chat.ID).Find(&members)
		if len(members) != 2 {
			t.Errorf("expected 2 members, got %d", len(members))
		}
	})

	t.Run("Get existing DM chat", func(t *testing.T) {
		chat1, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
		if err != nil {
			t.Fatalf("failed to get first DM chat: %v", err)
		}

		chat2, err := chatService.GetOrCreateDMChat(nil, user2.ID, user1.ID)
		if err != nil {
			t.Fatalf("failed to get second DM chat: %v", err)
		}

		if chat1.ID != chat2.ID {
			t.Error("expected same chat ID when getting existing DM")
		}
	})

	t.Run("Cannot create DM with self", func(t *testing.T) {
		_, err := chatService.GetOrCreateDMChat(nil, user1.ID, user1.ID)
		if err == nil {
			t.Error("expected error when creating DM with self")
		}
	})
}

func TestChatService_GetUserChatsWithLastMessage(t *testing.T) {
	db := setupTestDB(t)
	
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1111111111")
	user2 := createTestUser(t, db, "+2222222222")
	user3 := createTestUser(t, db, "+3333333333")

	t.Run("Get user chats", func(t *testing.T) {
		chat1, _ := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
		chat2, _ := chatService.GetOrCreateDMChat(nil, user1.ID, user3.ID)

		chats, err := chatService.GetUserChatsWithLastMessage(nil, user1.ID)
		if err != nil {
			t.Fatalf("failed to get user chats: %v", err)
		}

		if len(chats) != 2 {
			t.Errorf("expected 2 chats, got %d", len(chats))
		}

		foundChat1, foundChat2 := false, false
		for _, chat := range chats {
			if chat.ID == chat1.ID {
				foundChat1 = true
			}
			if chat.ID == chat2.ID {
				foundChat2 = true
			}
		}

		if !foundChat1 {
			t.Error("expected to find chat1")
		}
		if !foundChat2 {
			t.Error("expected to find chat2")
		}
	})

	t.Run("Empty chats for new user", func(t *testing.T) {
		newUser := createTestUser(t, db, "+4444444444")
		chats, err := chatService.GetUserChatsWithLastMessage(nil, newUser.ID)
		if err != nil {
			t.Fatalf("failed to get user chats: %v", err)
		}

		if len(chats) != 0 {
			t.Errorf("expected 0 chats for new user, got %d", len(chats))
		}
	})
}

func TestChatService_UpdateLastRead(t *testing.T) {
	db := setupTestDB(t)
	
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+5555555555")
	user2 := createTestUser(t, db, "+6666666666")

	chat, _ := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)

	t.Run("Update last read", func(t *testing.T) {
		err := chatService.UpdateLastRead(nil, chat.ID, user1.ID)
		if err != nil {
			t.Fatalf("failed to update last read: %v", err)
		}

		var member models.ChatMember
		db.Where("chat_id = ? AND user_id = ?", chat.ID, user1.ID).First(&member)

		if member.LastReadAt.IsZero() {
			t.Error("expected last_read_at to be updated")
		}
	})
}

func TestChatService_GetUnreadCount(t *testing.T) {
	db := setupTestDB(t)
	
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+7777777777")
	user2 := createTestUser(t, db, "+8888888888")

	chat, _ := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)

	message := &models.Message{
		ChatID:      chat.ID,
		SenderID:    &user2.ID,
		Content:     "Test message",
		MessageType: models.MessageTypeText,
	}
	db.Create(message)

	t.Run("Get unread count", func(t *testing.T) {
		count, err := chatService.GetUnreadCount(nil, chat.ID, user1.ID)
		if err != nil {
			t.Fatalf("failed to get unread count: %v", err)
		}

		if count != 1 {
			t.Errorf("expected 1 unread message, got %d", count)
		}
	})
}

func TestDMChatUniqueness(t *testing.T) {
	db := setupTestDB(t)
	
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+9999999999")
	user2 := createTestUser(t, db, "+0000000000")

	chat1, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create first DM chat: %v", err)
	}

	for i := 0; i < 5; i++ {
		chat, err := chatService.GetOrCreateDMChat(nil, user2.ID, user1.ID)
		if err != nil {
			t.Fatalf("failed to get DM chat on iteration %d: %v", i, err)
		}

		if chat.ID != chat1.ID {
			t.Errorf("iteration %d: expected same chat ID %s, got %s", i, chat1.ID, chat.ID)
		}
	}

	var chatCount int64
	db.Model(&models.Chat{}).Where("type = ?", models.ChatTypeDM).Count(&chatCount)
	if chatCount != 1 {
		t.Errorf("expected 1 DM chat in database, got %d", chatCount)
	}
}
