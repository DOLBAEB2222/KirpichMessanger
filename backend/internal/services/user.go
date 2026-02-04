package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/messenger/backend/pkg/auth"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewUserService(db *gorm.DB, redis *redis.Client) *UserService {
	return &UserService{
		db:    db,
		redis: redis,
	}
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:profile:%s", userID.String())

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	userData, _ := json.Marshal(user)
	s.redis.Set(ctx, cacheKey, userData, 5*time.Minute)

	return &user, nil
}

func (s *UserService) GetPublicProfile(ctx context.Context, userID uuid.UUID) (*models.PublicUserProfile, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile := user.ToPublicProfile()
	return &profile, nil
}

func (s *UserService) GetPrivateProfile(ctx context.Context, userID uuid.UUID) (*models.PrivateUserProfile, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile := user.ToPrivateProfile()
	return &profile, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UpdateProfileRequest) (*models.UpdateProfileResponse, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}

	if req.Username != nil {
		var existingUser models.User
		if err := s.db.Where("username = ? AND id != ?", *req.Username, userID).First(&existingUser).Error; err == nil {
			return nil, fmt.Errorf("username already taken")
		}
		updates["username"] = *req.Username
	}

	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}

	if req.Avatar != nil && *req.Avatar != "" {
		avatarURL, err := s.processAvatar(*req.Avatar, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to process avatar: %w", err)
		}
		updates["avatar_url"] = avatarURL
	}

	if len(updates) > 0 {
		if err := s.db.Model(user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	s.invalidateUserCache(ctx, userID)

	updatedUser, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.UpdateProfileResponse{
		ID:        updatedUser.ID,
		Username:  updatedUser.Username,
		Bio:       updatedUser.Bio,
		AvatarURL: updatedUser.AvatarURL,
		UpdatedAt: updatedUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := auth.CheckPassword(user.PasswordHash, oldPassword); err != nil {
		return fmt.Errorf("invalid old password")
	}

	if err := auth.ValidatePassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password")
	}

	if err := s.db.Model(user).Update("password_hash", hashedPassword).Error; err != nil {
		return err
	}

	s.invalidateUserCache(ctx, userID)

	return nil
}

func (s *UserService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	if err := s.db.Delete(&models.User{}, userID).Error; err != nil {
		return err
	}

	s.invalidateUserCache(ctx, userID)

	return nil
}

func (s *UserService) FindByPhoneOrEmail(phoneOrEmail string) (*models.User, error) {
	var user models.User
	err := s.db.Where("phone = ? OR email = ?", phoneOrEmail, phoneOrEmail).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateLastSeen(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("last_seen_at", now).Error
}

func (s *UserService) invalidateUserCache(ctx context.Context, userID uuid.UUID) {
	cacheKey := fmt.Sprintf("user:profile:%s", userID.String())
	s.redis.Del(ctx, cacheKey)
}

func (s *UserService) processAvatar(avatarData string, userID uuid.UUID) (string, error) {
	return avatarData, nil
}
