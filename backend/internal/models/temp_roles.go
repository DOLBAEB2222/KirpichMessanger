package models

import (
	"time"

	"github.com/google/uuid"
)

type TempRoleType string

const (
	TempRoleTypeModerator TempRoleType = "moderator"
	TempRoleTypeAdmin     TempRoleType = "admin"
	TempRoleTypeCustom    TempRoleType = "custom"
)

type TempRoleTargetType string

const (
	TempRoleTargetChat    TempRoleTargetType = "chat"
	TempRoleTargetChannel TempRoleTargetType = "channel"
)

type TempRole struct {
	ID           uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TargetID     uuid.UUID           `gorm:"type:uuid;not null;index:idx_temp_role_target" json:"target_id"`
	TargetType   TempRoleTargetType  `gorm:"type:varchar(20);not null;index:idx_temp_role_target" json:"target_type"`
	UserID       uuid.UUID           `gorm:"type:uuid;not null;index:idx_temp_role_user" json:"user_id"`
	RoleType     TempRoleType        `gorm:"type:varchar(50);not null" json:"role_type"`
	CustomName   *string             `gorm:"type:varchar(255)" json:"custom_name,omitempty"`
	Permissions []string            `gorm:"type:text[];serializer:json" json:"permissions"`
	GrantedByID  uuid.UUID           `gorm:"type:uuid;not null" json:"granted_by_id"`
	ExpiresAt    time.Time           `gorm:"not null;index:idx_temp_role_expires" json:"expires_at"`
	IsActive     bool                `gorm:"default:true;index" json:"is_active"`
	CreatedAt    time.Time           `json:"created_at"`

	GrantedBy   *User  `gorm:"foreignKey:GrantedByID" json:"granted_by,omitempty"`
	User        *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type CreateTempRoleRequest struct {
	TargetID    uuid.UUID          `json:"target_id" validate:"required,uuid"`
	TargetType  TempRoleTargetType `json:"target_type" validate:"required,oneof=chat channel"`
	UserID      uuid.UUID          `json:"user_id" validate:"required,uuid"`
	RoleType    TempRoleType       `json:"role_type" validate:"required,oneof=moderator admin custom"`
	CustomName  *string            `json:"custom_name" validate:"omitempty,max=255"`
	Permissions []string          `json:"permissions" validate:"required,min=1"`
	DurationHours int              `json:"duration_hours" validate:"required,min=1,max=8760"`
}

type UpdateTempRoleRequest struct {
	IsEnabled  *bool      `json:"is_enabled"`
	DurationHours *int     `json:"duration_hours" validate:"omitempty,min=1,max=8760"`
}

type TempRoleResponse struct {
	ID           uuid.UUID           `json:"id"`
	TargetID     uuid.UUID           `json:"target_id"`
	TargetType   TempRoleTargetType  `json:"target_type"`
	UserID       uuid.UUID           `json:"user_id"`
	RoleType     TempRoleType        `json:"role_type"`
	CustomName   *string             `json:"custom_name,omitempty"`
	Permissions  []string            `json:"permissions"`
	GrantedByID  uuid.UUID           `json:"granted_by_id"`
	ExpiresAt    time.Time           `json:"expires_at"`
	IsActive     bool                `json:"is_active"`
	CreatedAt    time.Time           `json:"created_at"`
	GrantedBy    *UserResponse       `json:"granted_by,omitempty"`
	User         *UserResponse       `json:"user,omitempty"`
	IsExpired    bool                `json:"is_expired"`
}

func (t *TempRole) ToResponse() TempRoleResponse {
	resp := TempRoleResponse{
		ID:           t.ID,
		TargetID:     t.TargetID,
		TargetType:   t.TargetType,
		UserID:       t.UserID,
		RoleType:     t.RoleType,
		CustomName:   t.CustomName,
		Permissions:  t.Permissions,
		GrantedByID:  t.GrantedByID,
		ExpiresAt:    t.ExpiresAt,
		IsActive:     t.IsActive,
		CreatedAt:    t.CreatedAt,
		IsExpired:    time.Now().After(t.ExpiresAt),
	}

	if t.GrantedBy != nil {
		userResp := t.GrantedBy.ToResponse()
		resp.GrantedBy = &userResp
	}

	if t.User != nil {
		userResp := t.User.ToResponse()
		resp.User = &userResp
	}

	return resp
}
