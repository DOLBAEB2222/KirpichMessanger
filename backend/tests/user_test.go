package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
)

func TestUserToResponse(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	bio := "Test bio"

	user := models.User{
		ID:        userID,
		Phone:     "+79991234567",
		Email:     &email,
		Username:  &username,
		Bio:       &bio,
		IsPremium: true,
	}

	response := user.ToResponse()

	if response.ID != userID {
		t.Errorf("Expected ID %v, got %v", userID, response.ID)
	}

	if response.Phone != "+79991234567" {
		t.Errorf("Expected Phone +79991234567, got %s", response.Phone)
	}

	if response.Email == nil || *response.Email != email {
		t.Error("Expected Email to match")
	}

	if response.Username == nil || *response.Username != username {
		t.Error("Expected Username to match")
	}

	if response.Bio == nil || *response.Bio != bio {
		t.Error("Expected Bio to match")
	}

	if !response.IsPremium {
		t.Error("Expected IsPremium to be true")
	}
}

func TestUserToPublicProfile(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	bio := "Test bio"

	user := models.User{
		ID:        userID,
		Phone:     "+79991234567",
		Username:  &username,
		Bio:       &bio,
		IsPremium: false,
	}

	profile := user.ToPublicProfile()

	if profile.ID != userID {
		t.Errorf("Expected ID %v, got %v", userID, profile.ID)
	}

	if profile.Username == nil || *profile.Username != username {
		t.Error("Expected Username to match")
	}

	if profile.Bio == nil || *profile.Bio != bio {
		t.Error("Expected Bio to match")
	}

	if profile.IsPremium {
		t.Error("Expected IsPremium to be false")
	}
}

func TestUserToPrivateProfile(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"

	user := models.User{
		ID:       userID,
		Phone:    "+79991234567",
		Email:    &email,
		Username: &username,
	}

	profile := user.ToPrivateProfile()

	if profile.ID != userID {
		t.Errorf("Expected ID %v, got %v", userID, profile.ID)
	}

	if profile.Phone != "+79991234567" {
		t.Errorf("Expected Phone +79991234567, got %s", profile.Phone)
	}

	if profile.Email == nil || *profile.Email != email {
		t.Error("Expected Email to match")
	}

	if profile.Username == nil || *profile.Username != username {
		t.Error("Expected Username to match")
	}
}

func TestRegisterRequestValidation(t *testing.T) {
	email := "test@example.com"
	username := "testuser"

	req := models.RegisterRequest{
		Phone:    "+79991234567",
		Email:    &email,
		Password: "SecurePass123",
		Username: &username,
	}

	if req.Phone == "" && req.Email == nil {
		t.Error("Either phone or email should be provided")
	}

	if req.Password == "" {
		t.Error("Password should not be empty")
	}
}

func TestLoginRequestValidation(t *testing.T) {
	req := models.LoginRequest{
		PhoneOrEmail: "+79991234567",
		Password:     "SecurePass123",
	}

	if req.PhoneOrEmail == "" {
		t.Error("PhoneOrEmail should not be empty")
	}

	if req.Password == "" {
		t.Error("Password should not be empty")
	}
}

func TestPasswordChangeRequest(t *testing.T) {
	req := models.PasswordChangeRequest{
		OldPassword: "OldPass123",
		NewPassword: "NewPass123",
	}

	if req.OldPassword == "" {
		t.Error("OldPassword should not be empty")
	}

	if req.NewPassword == "" {
		t.Error("NewPassword should not be empty")
	}
}

func TestUpdateProfileRequest(t *testing.T) {
	username := "newusername"
	bio := "New bio"

	req := models.UpdateProfileRequest{
		Username: &username,
		Bio:      &bio,
	}

	if req.Username == nil || *req.Username != username {
		t.Error("Username should match")
	}

	if req.Bio == nil || *req.Bio != bio {
		t.Error("Bio should match")
	}
}

func TestAuthResponse(t *testing.T) {
	userID := uuid.New()

	resp := models.AuthResponse{
		UserID:       userID,
		Token:        "access_token",
		RefreshToken: "refresh_token",
		ExpiresIn:    3600,
	}

	if resp.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, resp.UserID)
	}

	if resp.Token != "access_token" {
		t.Errorf("Expected Token access_token, got %s", resp.Token)
	}

	if resp.RefreshToken != "refresh_token" {
		t.Errorf("Expected RefreshToken refresh_token, got %s", resp.RefreshToken)
	}

	if resp.ExpiresIn != 3600 {
		t.Errorf("Expected ExpiresIn 3600, got %d", resp.ExpiresIn)
	}
}

func TestRefreshResponse(t *testing.T) {
	resp := models.RefreshResponse{
		Token:     "new_token",
		ExpiresIn: 3600,
	}

	if resp.Token != "new_token" {
		t.Errorf("Expected Token new_token, got %s", resp.Token)
	}

	if resp.ExpiresIn != 3600 {
		t.Errorf("Expected ExpiresIn 3600, got %d", resp.ExpiresIn)
	}
}

func TestPublicUserProfile(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	bio := "Test bio"
	lastSeen := "2026-02-02T17:00:00Z"

	profile := models.PublicUserProfile{
		ID:        userID,
		Username:  &username,
		Bio:       &bio,
		IsPremium: true,
		LastSeen:  &lastSeen,
	}

	if profile.ID != userID {
		t.Errorf("Expected ID %v, got %v", userID, profile.ID)
	}

	if profile.Username == nil || *profile.Username != username {
		t.Error("Expected Username to match")
	}

	if !profile.IsPremium {
		t.Error("Expected IsPremium to be true")
	}
}

func TestPrivateUserProfile(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	createdAt := "2026-02-02T10:00:00Z"
	lastSeen := "2026-02-02T17:00:00Z"

	profile := models.PrivateUserProfile{
		ID:        userID,
		Phone:     "+79991234567",
		Email:     &email,
		Username:  &username,
		IsPremium: false,
		CreatedAt: createdAt,
		LastSeen:  &lastSeen,
	}

	if profile.ID != userID {
		t.Errorf("Expected ID %v, got %v", userID, profile.ID)
	}

	if profile.Phone != "+79991234567" {
		t.Errorf("Expected Phone +79991234567, got %s", profile.Phone)
	}

	if profile.CreatedAt != createdAt {
		t.Errorf("Expected CreatedAt %s, got %s", createdAt, profile.CreatedAt)
	}
}
