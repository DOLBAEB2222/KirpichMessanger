package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte
var refreshSecret []byte

const (
	DefaultAccessTokenExpiry  = time.Hour
	DefaultRefreshTokenExpiry = 30 * 24 * time.Hour
)

type Claims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func Initialize(secret string) {
	jwtSecret = []byte(secret)
	refreshSecret = []byte(secret + "_refresh")
}

func getAccessTokenExpiry() time.Duration {
	if hoursStr := os.Getenv("JWT_EXPIRE_HOURS"); hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil && hours > 0 {
			return time.Duration(hours) * time.Hour
		}
	}
	return DefaultAccessTokenExpiry
}

func GenerateTokenPair(userID string) (*TokenPair, error) {
	accessToken, err := generateToken(userID, "access", getAccessTokenExpiry())
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(userID, "refresh", DefaultRefreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(getAccessTokenExpiry().Seconds()),
	}, nil
}

func generateToken(userID, tokenType string, expiry time.Duration) (string, error) {
	var secret []byte
	if tokenType == "refresh" {
		secret = refreshSecret
	} else {
		secret = jwtSecret
	}

	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	return validateTokenWithSecret(tokenString, jwtSecret)
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := validateTokenWithSecret(tokenString, refreshSecret)
	if err != nil {
		return nil, err
	}
	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

func validateTokenWithSecret(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		if claims.Type != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token type",
			})
		}

		c.Locals("userID", claims.UserID)
		return c.Next()
	}
}
