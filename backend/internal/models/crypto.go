package models

import (
	"time"

	"github.com/google/uuid"
)

type UserDevice struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID                uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	DeviceID              string    `gorm:"not null" json:"device_id"`
	RegistrationID        uint32    `gorm:"not null" json:"registration_id"`
	IdentityKeyPublic     []byte    `gorm:"not null" json:"identity_key_public"`
	SignedPreKeyID        uint32    `gorm:"not null" json:"signed_pre_key_id"`
	SignedPreKeyPublic    []byte    `gorm:"not null" json:"signed_pre_key_public"`
	SignedPreKeySignature []byte    `gorm:"not null" json:"signed_pre_key_signature"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type UserOneTimeKey struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DeviceID  uuid.UUID `gorm:"type:uuid;not null" json:"device_id"`
	KeyID     uint32    `gorm:"not null" json:"key_id"`
	PublicKey []byte    `gorm:"not null" json:"public_key"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterDeviceRequest struct {
	DeviceID              string   `json:"device_id" validate:"required"`
	RegistrationID        uint32   `json:"registration_id" validate:"required"`
	IdentityKeyPublic     string   `json:"identity_key_public" validate:"required"` // base64
	SignedPreKeyID        uint32   `json:"signed_pre_key_id" validate:"required"`
	SignedPreKeyPublic    string   `json:"signed_pre_key_public" validate:"required"` // base64
	SignedPreKeySignature string   `json:"signed_pre_key_signature" validate:"required"` // base64
	OneTimeKeys           []OneTimeKeyDTO `json:"one_time_keys" validate:"required"`
}

type OneTimeKeyDTO struct {
	KeyID     uint32 `json:"key_id"`
	PublicKey string `json:"public_key"` // base64
}

type SendEncryptedRequest struct {
	RecipientID uuid.UUID `json:"recipient_id" validate:"required"`
	DeviceID    string    `json:"device_id" validate:"required"`
	Content     string    `json:"content" validate:"required"` // encrypted and base64 encoded
}
