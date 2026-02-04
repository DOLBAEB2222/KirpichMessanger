package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type User struct {
    ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    Phone        string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
    Email        *string        `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty"`
    PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
    Username     *string        `gorm:"type:varchar(50);uniqueIndex" json:"username,omitempty"`
    AvatarURL    *string        `gorm:"type:text" json:"avatar_url,omitempty"`
    Bio          *string        `gorm:"type:text" json:"bio,omitempty"`
    IsPremium    bool           `gorm:"default:false" json:"is_premium"`
    LastSeenAt   *time.Time     `json:"last_seen_at,omitempty"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
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

func (u *User) ToResponse() UserResponse {
    return UserResponse{
        ID:        u.ID,
        Phone:     u.Phone,
        Email:     u.Email,
        Username:  u.Username,
        AvatarURL: u.AvatarURL,
        Bio:       u.Bio,
        IsPremium: u.IsPremium,
        CreatedAt: u.CreatedAt.Format(time.RFC3339),
    }
}

func (u *User) ToPrivateProfile() PrivateUserProfile {
    profile := PrivateUserProfile{
        ID:        u.ID,
        Phone:     u.Phone,
        Email:     u.Email,
        Username:  u.Username,
        AvatarURL: u.AvatarURL,
        Bio:       u.Bio,
        IsPremium: u.IsPremium,
        CreatedAt: u.CreatedAt.Format(time.RFC3339),
    }
    if u.LastSeenAt != nil {
        formatted := u.LastSeenAt.Format(time.RFC3339)
        profile.LastSeen = &formatted
    }
    return profile
}

func (u *User) ToPublicProfile() PublicUserProfile {
    profile := PublicUserProfile{
        ID:        u.ID,
        Username:  u.Username,
        AvatarURL: u.AvatarURL,
        Bio:       u.Bio,
        IsPremium: u.IsPremium,
    }
    if u.LastSeenAt != nil {
        formatted := u.LastSeenAt.Format(time.RFC3339)
        profile.LastSeen = &formatted
    }
    return profile
}
