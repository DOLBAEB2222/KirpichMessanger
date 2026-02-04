package middleware

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type ValidationErrors map[string]string

func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

func (ve ValidationErrors) Add(field, message string) {
	ve[field] = message
}

func ValidateRegisterRequest() fiber.Handler {
	return func(c fiber.Ctx) error {
		var body struct {
			Phone    string  `json:"phone"`
			Email    *string `json:"email"`
			Password string  `json:"password"`
			Username *string `json:"username"`
		}

		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		errors := make(ValidationErrors)

		if body.Phone == "" && body.Email == nil {
			errors.Add("phone_or_email", "Either phone or email is required")
		}

		if body.Phone != "" && !isValidPhone(body.Phone) {
			errors.Add("phone", "Invalid phone format. Use E.164 format (e.g., +79991234567)")
		}

		if body.Email != nil && *body.Email != "" && !isValidEmail(*body.Email) {
			errors.Add("email", "Invalid email format")
		}

		if err := validatePassword(body.Password); err != nil {
			errors.Add("password", err.Error())
		}

		if body.Username != nil && !isValidUsername(*body.Username) {
			errors.Add("username", "Username must be 3-50 characters and contain only letters, numbers, and underscores")
		}

		if errors.HasErrors() {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": errors,
			})
		}

		return c.Next()
	}
}

func ValidateLoginRequest() fiber.Handler {
	return func(c fiber.Ctx) error {
		var body struct {
			PhoneOrEmail string `json:"phone_or_email"`
			Password     string `json:"password"`
		}

		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		errors := make(ValidationErrors)

		if body.PhoneOrEmail == "" {
			errors.Add("phone_or_email", "Phone or email is required")
		}

		if body.Password == "" {
			errors.Add("password", "Password is required")
		}

		if errors.HasErrors() {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": errors,
			})
		}

		return c.Next()
	}
}

func ValidateUpdateProfileRequest() fiber.Handler {
	return func(c fiber.Ctx) error {
		var body struct {
			Username *string `json:"username"`
			Bio      *string `json:"bio"`
			Avatar   *string `json:"avatar"`
		}

		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		errors := make(ValidationErrors)

		if body.Username != nil && !isValidUsername(*body.Username) {
			errors.Add("username", "Username must be 3-50 characters and contain only letters, numbers, and underscores")
		}

		if body.Bio != nil && len(*body.Bio) > 500 {
			errors.Add("bio", "Bio must not exceed 500 characters")
		}

		if errors.HasErrors() {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": errors,
			})
		}

		return c.Next()
	}
}

func ValidateChangePasswordRequest() fiber.Handler {
	return func(c fiber.Ctx) error {
		var body struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		errors := make(ValidationErrors)

		if body.OldPassword == "" {
			errors.Add("old_password", "Old password is required")
		}

		if err := validatePassword(body.NewPassword); err != nil {
			errors.Add("new_password", err.Error())
		}

		if errors.HasErrors() {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": errors,
			})
		}

		return c.Next()
	}
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "Password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	var missing []string
	if !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if !hasDigit {
		missing = append(missing, "digit")
	}

	if len(missing) > 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Password must contain at least one "+strings.Join(missing, ", "))
	}

	return nil
}
