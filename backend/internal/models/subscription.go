package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionType string
type SubscriptionStatus string
type PaymentStatus string

const (
	SubscriptionTypeMonthly SubscriptionType = "premium_monthly"
	SubscriptionTypeYearly  SubscriptionType = "premium_yearly"

	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"

	PaymentStatusCompletedStub PaymentStatus = "completed_stub"
	PaymentStatusPending       PaymentStatus = "pending"
	PaymentStatusFailed        PaymentStatus = "failed"
	PaymentStatusCompleted     PaymentStatus = "completed"
	PaymentStatusRefunded      PaymentStatus = "refunded"
)

type Subscription struct {
	ID        uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID          `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      SubscriptionType   `gorm:"type:varchar(50);not null" json:"type"`
	StartDate time.Time          `gorm:"type:date;not null" json:"start_date"`
	EndDate   time.Time          `gorm:"type:date;not null" json:"end_date"`
	AutoRenew bool               `gorm:"default:false" json:"auto_renew"`
	Status    SubscriptionStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	
	User      *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type PaymentLog struct {
	ID               uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID           *uuid.UUID    `gorm:"type:uuid" json:"user_id"`
	Amount           float64       `gorm:"type:decimal(10,2);not null" json:"amount"`
	SubscriptionType string        `gorm:"type:varchar(50);not null" json:"subscription_type"`
	Status           PaymentStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	PaymentMethod    string        `gorm:"type:varchar(50);default:'stub'" json:"payment_method"`
	TransactionID    *string       `gorm:"type:varchar(255)" json:"transaction_id,omitempty"`
	Notes            *string       `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	
	User             *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type PurchaseSubscriptionRequest struct {
	SubscriptionType SubscriptionType `json:"subscription_type" validate:"required,oneof=premium_monthly premium_yearly"`
	PaymentMethod    *string          `json:"payment_method" validate:"omitempty"`
}

type SubscriptionResponse struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	Type      SubscriptionType   `json:"type"`
	StartDate time.Time          `json:"start_date"`
	EndDate   time.Time          `json:"end_date"`
	Status    SubscriptionStatus `json:"status"`
	AutoRenew bool               `json:"auto_renew"`
	CreatedAt time.Time          `json:"created_at"`
}

type PaymentLogResponse struct {
	ID               uuid.UUID     `json:"id"`
	Amount           float64       `json:"amount"`
	SubscriptionType string        `json:"subscription_type"`
	Status           PaymentStatus `json:"status"`
	PaymentMethod    string        `json:"payment_method"`
	CreatedAt        time.Time     `json:"created_at"`
}

type PurchaseResponse struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	Subscription SubscriptionResponse   `json:"subscription"`
	PaymentLog   PaymentLogResponse     `json:"payment_log"`
}

func (s *Subscription) ToResponse() SubscriptionResponse {
	return SubscriptionResponse{
		ID:        s.ID,
		UserID:    s.UserID,
		Type:      s.Type,
		StartDate: s.StartDate,
		EndDate:   s.EndDate,
		Status:    s.Status,
		AutoRenew: s.AutoRenew,
		CreatedAt: s.CreatedAt,
	}
}

func (p *PaymentLog) ToResponse() PaymentLogResponse {
	return PaymentLogResponse{
		ID:               p.ID,
		Amount:           p.Amount,
		SubscriptionType: p.SubscriptionType,
		Status:           p.Status,
		PaymentMethod:    p.PaymentMethod,
		CreatedAt:        p.CreatedAt,
	}
}

func GetSubscriptionPrice(subType SubscriptionType) float64 {
	switch subType {
	case SubscriptionTypeMonthly:
		return 4.99
	case SubscriptionTypeYearly:
		return 49.99
	default:
		return 0.0
	}
}

func GetSubscriptionDuration(subType SubscriptionType) time.Duration {
	switch subType {
	case SubscriptionTypeMonthly:
		return 30 * 24 * time.Hour
	case SubscriptionTypeYearly:
		return 365 * 24 * time.Hour
	default:
		return 0
	}
}
