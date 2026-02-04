package middleware

import (
    "context"
    "fmt"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
    redis      *redis.Client
    maxAttempts int
    window      time.Duration
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
    return &RateLimiter{
        redis:       redis,
        maxAttempts: 5,
        window:      15 * time.Minute,
    }
}

func (rl *RateLimiter) LoginRateLimit() fiber.Handler {
    return func(c fiber.Ctx) error {
        identifier := getIdentifier(c)
        key := fmt.Sprintf("ratelimit:login:%s", identifier)

        ctx := context.Background()

        current, err := rl.redis.Get(ctx, key).Int()
        if err != nil && err != redis.Nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        if current >= rl.maxAttempts {
            ttl, _ := rl.redis.TTL(ctx, key).Result()
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error":       "Too many login attempts",
                "retry_after": int(ttl.Seconds()),
            })
        }

        pipe := rl.redis.Pipeline()
        pipe.Incr(ctx, key)
        pipe.Expire(ctx, key, rl.window)
        _, err = pipe.Exec(ctx)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        return c.Next()
    }
}

func (rl *RateLimiter) ResetLoginAttempts(identifier string) error {
    key := fmt.Sprintf("ratelimit:login:%s", identifier)
    ctx := context.Background()
    return rl.redis.Del(ctx, key).Err()
}

func (rl *RateLimiter) UploadRateLimit() fiber.Handler {
    return func(c fiber.Ctx) error {
        userID := c.Locals("userID")
        if userID == nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Unauthorized",
            })
        }

        userIDStr, ok := userID.(string)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid user ID",
            })
        }

        key := fmt.Sprintf("ratelimit:upload:%s", userIDStr)
        ctx := context.Background()

        maxUploads := 10
        window := time.Hour

        current, err := rl.redis.Get(ctx, key).Int()
        if err != nil && err != redis.Nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        if current >= maxUploads {
            ttl, _ := rl.redis.TTL(ctx, key).Result()
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error":       "Upload limit exceeded",
                "retry_after": int(ttl.Seconds()),
                "limit":       maxUploads,
                "window":      "1h",
            })
        }

        pipe := rl.redis.Pipeline()
        pipe.Incr(ctx, key)
        pipe.Expire(ctx, key, window)
        _, err = pipe.Exec(ctx)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        return c.Next()
    }
}

func (rl *RateLimiter) IPRateLimit() fiber.Handler {
    return func(c fiber.Ctx) error {
        key := fmt.Sprintf("ratelimit:ip:%s", c.IP())
        ctx := context.Background()

        maxRequests := 100 // per minute
        window := time.Minute

        current, err := rl.redis.Get(ctx, key).Int()
        if err != nil && err != redis.Nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        if current >= maxRequests {
            ttl, _ := rl.redis.TTL(ctx, key).Result()
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error":       "IP rate limit exceeded",
                "retry_after": int(ttl.Seconds()),
            })
        }

        pipe := rl.redis.Pipeline()
        pipe.Incr(ctx, key)
        pipe.Expire(ctx, key, window)
        _, err = pipe.Exec(ctx)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        return c.Next()
    }
}

func (rl *RateLimiter) UserRateLimit() fiber.Handler {
    return func(c fiber.Ctx) error {
        userID := c.Locals("userID")
        if userID == nil {
            return c.Next()
        }

        userIDStr := userID.(string)
        key := fmt.Sprintf("ratelimit:user:%s", userIDStr)
        ctx := context.Background()

        maxRequests := 200 // per minute
        window := time.Minute

        current, err := rl.redis.Get(ctx, key).Int()
        if err != nil && err != redis.Nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        if current >= maxRequests {
            ttl, _ := rl.redis.TTL(ctx, key).Result()
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error":       "User rate limit exceeded",
                "retry_after": int(ttl.Seconds()),
            })
        }

        pipe := rl.redis.Pipeline()
        pipe.Incr(ctx, key)
        pipe.Expire(ctx, key, window)
        _, err = pipe.Exec(ctx)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Rate limiting error",
            })
        }

        return c.Next()
    }
}

func getIdentifier(c fiber.Ctx) string {
    body := struct {
        PhoneOrEmail string `json:"phone_or_email"`
        Phone        string `json:"phone"`
        Email        string `json:"email"`
    }{}

    if err := c.Bind().JSON(&body); err == nil {
        if body.PhoneOrEmail != "" {
            return body.PhoneOrEmail
        }
        if body.Phone != "" {
            return body.Phone
        }
        if body.Email != "" {
            return body.Email
        }
    }

    return c.IP()
}
