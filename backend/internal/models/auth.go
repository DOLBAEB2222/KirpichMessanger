package models

import (
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Phone    string  `json:"phone"`
	Email    *string `json:"email"`
	Password string  `json:"password"`
	Username *string `json:"username"`
}

type LoginRequest struct {
	PhoneOrEmail string `json:"phone_or_email"`
	Password     string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
}

type RefreshResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type PasswordChangeRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type PasswordChangeResponse struct {
	Message string `json:"message"`
}

type DeleteAccountResponse struct {
	Message string `json:"message"`
}

type UpdateProfileRequest struct {
	Username *string `json:"username"`
	Bio      *string `json:"bio"`
	Avatar   *string `json:"avatar"`
}

type UpdateProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  *string   `json:"username"`
	Bio       *string   `json:"bio"`
	AvatarURL *string   `json:"avatar_url"`
	UpdatedAt string    `json:"updated_at"`
}

type PublicUserProfile struct {
	ID        uuid.UUID `json:"id"`
	Username  *string   `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Bio       *string   `json:"bio"`
	IsPremium bool      `json:"is_premium"`
	LastSeen  *string   `json:"last_seen,omitempty"`
}

type PrivateUserProfile struct {
	ID        uuid.UUID `json:"id"`
	Phone     string    `json:"phone"`
	Email     *string   `json:"email,omitempty"`
	Username  *string   `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Bio       *string   `json:"bio"`
	IsPremium bool      `json:"is_premium"`
	CreatedAt string    `json:"created_at"`
	LastSeen  *string   `json:"last_seen,omitempty"`
}
