package services

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/messenger/backend/internal/models"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

type ChatService struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewChatService(db *gorm.DB, redis *redis.Client) *ChatService {
    return &ChatService{
        db:    db,
        redis: redis,
    }
}

func (s *ChatService) GetOrCreateDMChat(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Chat, error) {
    if userID1 == userID2 {
        return nil, fmt.Errorf("cannot create DM with yourself")
    }

    cacheKey := s.getDMCacheKey(userID1, userID2)

    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil && cached != "" {
        var chat models.Chat
        if err := json.Unmarshal([]byte(cached), &chat); err == nil {
            return &chat, nil
        }
    }

    var chat models.Chat
    err = s.db.Raw(`
        SELECT c.* FROM chats c
        JOIN chat_members cm1 ON c.id = cm1.chat_id AND cm1.user_id = ?
        JOIN chat_members cm2 ON c.id = cm2.chat_id AND cm2.user_id = ?
        WHERE c.type = 'dm'
        AND (SELECT COUNT(*) FROM chat_members WHERE chat_id = c.id) = 2
        LIMIT 1
    `, userID1, userID2).Scan(&chat).Error

    if err == nil && chat.ID != uuid.Nil {
        s.cacheDMChat(ctx, cacheKey, &chat)
        s.db.Preload("Members.User").First(&chat, chat.ID)
        return &chat, nil
    }

    chat = models.Chat{
        Type:    models.ChatTypeDM,
        OwnerID: nil,
    }

    tx := s.db.Begin()
    if err := tx.Create(&chat).Error; err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to create chat: %w", err)
    }

    members := []models.ChatMember{
        {
            ChatID: chat.ID,
            UserID: userID1,
            Role:   models.MemberRoleMember,
        },
        {
            ChatID: chat.ID,
            UserID: userID2,
            Role:   models.MemberRoleMember,
        },
    }

    if err := tx.Create(&members).Error; err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to add members: %w", err)
    }

    if err := tx.Commit().Error; err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    s.db.Preload("Members.User").First(&chat, chat.ID)
    s.cacheDMChat(ctx, cacheKey, &chat)

    return &chat, nil
}

func (s *ChatService) GetUserChatsWithLastMessage(ctx context.Context, userID uuid.UUID) ([]models.ChatWithLastMessageResponse, error) {
    cacheKey := fmt.Sprintf("user:chats:%s", userID.String())

    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil && cached != "" {
        var chats []models.ChatWithLastMessageResponse
        if err := json.Unmarshal([]byte(cached), &chats); err == nil {
            return chats, nil
        }
    }

    var chatMembers []models.ChatMember
    if err := s.db.Where("user_id = ?", userID).Find(&chatMembers).Error; err != nil {
        return nil, err
    }

    if len(chatMembers) == 0 {
        return []models.ChatWithLastMessageResponse{}, nil
    }

    chatIDs := make([]uuid.UUID, len(chatMembers))
    for i, cm := range chatMembers {
        chatIDs[i] = cm.ChatID
    }

    var chats []models.Chat
    if err := s.db.Where("id IN ?", chatIDs).
        Order("last_message_at DESC").
        Preload("Members.User").
        Find(&chats).Error; err != nil {
        return nil, err
    }

    responses := make([]models.ChatWithLastMessageResponse, len(chats))
    for i, chat := range chats {
        if chat.Type == models.ChatTypeDM && chat.Name == nil {
            chat = s.enrichDMChatName(chat, userID)
        }

        lastMessage, unreadCount := s.getChatMetadata(chat.ID, userID)

        responses[i] = models.ChatWithLastMessageResponse{
            ChatResponse: chat.ToResponse(),
            LastMessage:  lastMessage,
            UnreadCount:  unreadCount,
        }
    }

    chatData, _ := json.Marshal(responses)
    s.redis.Set(ctx, cacheKey, chatData, 5*time.Minute)

    return responses, nil
}

func (s *ChatService) getChatMetadata(chatID, userID uuid.UUID) (*models.MessageResponse, int64) {
    var lastMessage models.Message
    if err := s.db.Where("chat_id = ? AND is_deleted = ?", chatID, false).
        Order("created_at DESC").
        Preload("Sender").
        First(&lastMessage).Error; err != nil {
        return nil, 0
    }

    response := lastMessage.ToResponse()

    var unreadCount int64
    s.db.Model(&models.Message{}).
        Joins("JOIN chat_members cm ON messages.chat_id = cm.chat_id AND cm.user_id = ?", userID).
        Where("messages.chat_id = ? AND messages.sender_id != ? AND messages.created_at > cm.last_read_at", chatID, userID).
        Count(&unreadCount)

    return &response, unreadCount
}

func (s *ChatService) enrichDMChatName(chat models.Chat, currentUserID uuid.UUID) models.Chat {
    for _, member := range chat.Members {
        if member.UserID != currentUserID && member.User != nil {
            name := member.User.Username
            if name == nil || *name == "" {
                phone := member.User.Phone
                name = &phone
            }
            chat.Name = name
            break
        }
    }
    return chat
}

func (s *ChatService) UpdateLastRead(ctx context.Context, chatID, userID uuid.UUID) error {
    return s.db.Model(&models.ChatMember{}).
        Where("chat_id = ? AND user_id = ?", chatID, userID).
        Update("last_read_at", time.Now()).Error
}

func (s *ChatService) GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int64, error) {
    var member models.ChatMember
    if err := s.db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&member).Error; err != nil {
        return 0, err
    }

    var count int64
    err := s.db.Model(&models.Message{}).
        Where("chat_id = ? AND sender_id != ? AND created_at > ?", chatID, userID, member.LastReadAt).
        Count(&count).Error

    return count, err
}

func (s *ChatService) InvalidateUserChatsCache(ctx context.Context, userID uuid.UUID) {
    cacheKey := fmt.Sprintf("user:chats:%s", userID.String())
    s.redis.Del(ctx, cacheKey)
}

func (s *ChatService) getDMCacheKey(userID1, userID2 uuid.UUID) string {
    if userID1.String() < userID2.String() {
        return fmt.Sprintf("chat:dm:%s:%s", userID1.String(), userID2.String())
    }
    return fmt.Sprintf("chat:dm:%s:%s", userID2.String(), userID1.String())
}

func (s *ChatService) cacheDMChat(ctx context.Context, key string, chat *models.Chat) {
    chatData, _ := json.Marshal(chat)
    s.redis.Set(ctx, key, chatData, 5*time.Minute)
}
