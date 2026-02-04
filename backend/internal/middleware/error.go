package middleware

import (
	"github.com/gofiber/fiber/v3"
)

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
	Code    string            `json:"code,omitempty"`
}

func ErrorHandler() fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal server error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		switch code {
		case fiber.StatusBadRequest:
			message = "Bad request"
		case fiber.StatusUnauthorized:
			message = "Unauthorized"
		case fiber.StatusForbidden:
			message = "Forbidden"
		case fiber.StatusNotFound:
			message = "Not found"
		case fiber.StatusConflict:
			message = "Conflict"
		case fiber.StatusTooManyRequests:
			message = "Too many requests"
		case fiber.StatusInternalServerError:
			message = "Internal server error"
		}

		return c.Status(code).JSON(ErrorResponse{
			Error: message,
		})
	}
}

func NotFoundHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "Route not found",
		})
	}
}
