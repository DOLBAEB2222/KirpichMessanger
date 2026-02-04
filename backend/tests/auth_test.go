package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/messenger/backend/pkg/auth"
)

func TestHashPassword(t *testing.T) {
	password := "SecurePass123"

	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashed == "" {
		t.Error("Expected non-empty hash")
	}

	if hashed == password {
		t.Error("Hash should not be the same as plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "SecurePass123"

	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = auth.CheckPassword(hashed, password)
	if err != nil {
		t.Errorf("Expected password to match, got error: %v", err)
	}

	err = auth.CheckPassword(hashed, "WrongPass123")
	if err == nil {
		t.Error("Expected password mismatch error")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "SecurePass123",
			wantErr:  false,
		},
		{
			name:     "Too short",
			password: "Short1",
			wantErr:  true,
		},
		{
			name:     "No uppercase",
			password: "securepass123",
			wantErr:  true,
		},
		{
			name:     "No lowercase",
			password: "SECUREPASS123",
			wantErr:  true,
		},
		{
			name:     "No digit",
			password: "SecurePass",
			wantErr:  true,
		},
		{
			name:     "Only 8 chars",
			password: "Secure1a",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email   string
		isValid bool
	}{
		{"user@example.com", true},
		{"test.email@domain.co.uk", true},
		{"user+tag@example.com", true},
		{"invalid", false},
		{"@example.com", false},
		{"user@", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := auth.IsValidEmail(tt.email)
			if got != tt.isValid {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.isValid)
			}
		})
	}
}

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		phone   string
		isValid bool
	}{
		{"+79991234567", true},
		{"+1234567890", true},
		{"+441234567890", true},
		{"79991234567", false},
		{"+abc", false},
		{"", false},
		{"123", false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			got := auth.IsValidPhone(tt.phone)
			if got != tt.isValid {
				t.Errorf("IsValidPhone(%q) = %v, want %v", tt.phone, got, tt.isValid)
			}
		})
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		username string
		isValid  bool
	}{
		{"john_doe", true},
		{"user123", true},
		{"abc", true},
		{"ab", false},
		{"a very long username that exceeds the fifty character limit", false},
		{"user-name", false},
		{"user.name", false},
		{"user@email", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			got := auth.IsValidUsername(tt.username)
			if got != tt.isValid {
				t.Errorf("IsValidUsername(%q) = %v, want %v", tt.username, got, tt.isValid)
			}
		})
	}
}

func TestGenerateTokenPair(t *testing.T) {
	auth.Initialize("test_secret_key")

	userID := uuid.New().String()

	tokenPair, err := auth.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}

	if tokenPair.ExpiresIn <= 0 {
		t.Error("Expected positive expires_in")
	}
}

func TestValidateToken(t *testing.T) {
	auth.Initialize("test_secret_key")

	userID := uuid.New().String()

	tokenPair, err := auth.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := auth.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Errorf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Type != "access" {
		t.Errorf("Expected token type 'access', got %s", claims.Type)
	}
}

func TestValidateRefreshToken(t *testing.T) {
	auth.Initialize("test_secret_key")

	userID := uuid.New().String()

	tokenPair, err := auth.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := auth.ValidateRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Errorf("Failed to validate refresh token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Type != "refresh" {
		t.Errorf("Expected token type 'refresh', got %s", claims.Type)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	auth.Initialize("test_secret_key")

	_, err := auth.ValidateToken("invalid_token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	auth.Initialize("test_secret_key")

	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCIsInR5cGUiOiJhY2Nlc3MiLCJleHAiOjE2MDAwMDAwMDB9.invalid"

	_, err := auth.ValidateToken(expiredToken)
	if err == nil {
		t.Error("Expected error for invalid token signature")
	}
}

func TestAccessTokenAsRefreshToken(t *testing.T) {
	auth.Initialize("test_secret_key")

	userID := uuid.New().String()

	tokenPair, err := auth.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	_, err = auth.ValidateRefreshToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("Expected error when using access token as refresh token")
	}
}

func TestTokenExpiry(t *testing.T) {
	auth.Initialize("test_secret_key")

	userID := uuid.New().String()

	tokenPair, err := auth.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := auth.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.ExpiresAt == nil {
		t.Error("Expected ExpiresAt to be set")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired immediately after creation")
	}
}
