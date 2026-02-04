package services

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Log(userID *uuid.UUID, action string, ip string, userAgent string, correlationID string, details any) {
	detailsJSON, _ := json.Marshal(details)
	
	var corrID *uuid.UUID
	if correlationID != "" {
		if u, err := uuid.Parse(correlationID); err == nil {
			corrID = &u
		}
	}

	auditLog := models.AuditLog{
		UserID:        userID,
		Action:        models.AuditAction(action),
		IPAddress:     ip,
		UserAgent:     userAgent,
		Details:       detailsJSON,
		CorrelationID: corrID,
	}

	if err := s.db.Create(&auditLog).Error; err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}
}
