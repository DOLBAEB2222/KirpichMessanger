package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type LastSeenMiddleware struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewLastSeenMiddleware(db *gorm.DB, redis *redis.Client) *LastSeenMiddleware {
	return &LastSeenMiddleware{
		db:    db,
		redis: redis,
	}
}

func (m *LastSeenMiddleware) UpdateLastSeen() fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDVal := c.Locals("userID")
		if userIDVal == nil {
			return c.Next()
		}

		userIDStr, ok := userIDVal.(string)
		if !ok {
			return c.Next()
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Next()
		}

		go func() {
			now := time.Now().UTC()

			cacheKey := "user:last_seen:" + userIDStr
			ctx := context.Background()

			cached, err := m.redis.Get(ctx, cacheKey).Result()
			if err == nil && cached != "" {
				var lastUpdate time.Time
				if err := json.Unmarshal([]byte(cached), &lastUpdate); err == nil {
					if now.Sub(lastUpdate) < time.Minute {
						return
					}
				}
			}

			m.db.Model(&struct {
				ID         uuid.UUID `gorm:"primaryKey"`
				LastSeenAt *time.Time
			}{}).Where("id = ?", userID).Update("last_seen_at", now)

			cacheData, _ := json.Marshal(now)
			m.redis.Set(ctx, cacheKey, cacheData, time.Minute)
		}()

		return c.Next()
	}
}
