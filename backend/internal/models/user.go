package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Phone        string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	Email        *string    `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Username     *string    `gorm:"type:varchar(50);uniqueIndex" json:"username,omitempty"`
	AvatarURL    *string    `gorm:"type:text" json:"avatar_url,omitempty"`
	Bio          *string    `gorm:"type:text" json:"bio,omitempty"`
	IsPremium    bool       `gorm:"default:false" json:"is_premium"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Contact struct {
	UserID      uuid.UUID `gorm:"primaryKey" json:"user_id"`
	ContactID   uuid.UUID `gorm:"primaryKey" json:"contact_id"`
	DisplayName *string   `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	AddedAt     time.Time `json:"added_at"`
}

type BlockedUser struct {
	UserID        uuid.UUID `gorm:"primaryKey" json:"user_id"`
	BlockedUserID uuid.UUID `gorm:"primaryKey" json:"blocked_user_id"`
	BlockedAt     time.Time `json:"blocked_at"`
}

type RegisterRequest struct {
	Phone    string  `json:"phone" validate:"required,e164"`
	Email    *string `json:"email" validate:"omitempty,email"`
	Password string  `json:"password" validate:"required,min=8"`
	Username *string `json:"username" validate:"omitempty,min=3,max=50"`
}

type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Phone     string     `json:"phone"`
	Email     *string    `json:"email,omitempty"`
	Username  *string    `json:"username,omitempty"`
	AvatarURL *string    `json:"avatar_url,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	IsPremium bool       `json:"is_premium"`
	CreatedAt time.Time  `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Phone:     u.Phone,
		Email:     u.Email,
		Username:  u.Username,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		IsPremium: u.IsPremium,
		CreatedAt: u.CreatedAt,
	}
}
