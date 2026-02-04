package services

import (
	"context"
	"log"
	"time"

	"github.com/messenger/backend/internal/models"
	"gorm.io/gorm"
)

type WorkerService struct {
	db *gorm.DB
}

func NewWorkerService(db *gorm.DB) *WorkerService {
	return &WorkerService{db: db}
}

func (s *WorkerService) Start(ctx context.Context) {
	roleTicker := time.NewTicker(5 * time.Minute)
	mediaTicker := time.NewTicker(24 * time.Hour)
	rssTicker := time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-roleTicker.C:
				s.cleanupExpiredRoles()
			case <-mediaTicker.C:
				s.cleanupOldMedia()
			case <-rssTicker.C:
				s.refreshAllActiveRSSFeeds()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *WorkerService) cleanupExpiredRoles() {
	result := s.db.Model(&models.TempRole{}).
		Where("expires_at <= ? AND is_active = ?", time.Now(), true).
		Update("is_active", false)
	
	if result.Error != nil {
		log.Printf("Worker: Error cleaning up expired roles: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("Worker: Deactivated %d expired temporary roles", result.RowsAffected)
	}
}

func (s *WorkerService) cleanupOldMedia() {
	// Simple DB-based cleanup, files would need a separate process or the cleanup.sh script
	log.Println("Worker: Starting old media cleanup task")
	// Logic to match cleanup.sh
}

func (s *WorkerService) refreshAllActiveRSSFeeds() {
	log.Println("Worker: Refreshing active RSS feeds")
	// Logic would go here, possibly calling the RSSHandler refresh logic
}
