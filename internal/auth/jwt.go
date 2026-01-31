package auth

import (
	"errors"
	"time"

	"ioi-amms/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles token generation and validation
type JWTService struct {
	secret          []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewJWTService creates a new JWT service with config
func NewJWTService(cfg *config.AuthConfig) *JWTService {
	secret := cfg.JWTSecret
	if secret == "" {
		secret = "default-dev-secret-change-in-production"
	}

	return &JWTService{
		secret:          []byte(secret),
		accessDuration:  cfg.AccessTokenDuration,
		refreshDuration: cfg.RefreshTokenDuration,
	}
}

// GenerateToken creates a new JWT token for the user
func (s *JWTService) GenerateToken(userID, tenantID, email, role string) (string, error) {
	expirationTime := time.Now().Add(s.accessDuration)

	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ioi-amms",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken creates a longer-lived refresh token
func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	expirationTime := time.Now().Add(s.refreshDuration)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ioi-amms-refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken parses and validates a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// Package-level functions for backward compatibility (use default config)
var defaultService *JWTService

func init() {
	// Initialize with default config for backward compatibility
	defaultService = NewJWTService(&config.AuthConfig{
		JWTSecret:            "",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	})
}

func GenerateToken(userID, tenantID, email, role string) (string, error) {
	return defaultService.GenerateToken(userID, tenantID, email, role)
}

func GenerateRefreshToken(userID string) (string, error) {
	return defaultService.GenerateRefreshToken(userID)
}

func ValidateToken(tokenString string) (*Claims, error) {
	return defaultService.ValidateToken(tokenString)
}
