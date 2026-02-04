package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func CorrelationID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get("X-Correlation-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("X-Correlation-ID", id)
		c.Locals("correlationID", id)
		return c.Next()
	}
}
