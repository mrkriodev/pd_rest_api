package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(secretKey string, accessTokenTTLHours, refreshTokenTTLHours int) *AuthService {
	return &AuthService{
		secretKey:       secretKey,
		accessTokenTTL:  time.Duration(accessTokenTTLHours) * time.Hour,
		refreshTokenTTL: time.Duration(refreshTokenTTLHours) * time.Hour,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until expiration
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (s *AuthService) GenerateTokenPair(userUUID string) (*TokenPair, error) {
	accessToken, accessExpiresAt, err := s.generateAccessToken(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := s.generateRefreshToken(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresIn := int64(accessExpiresAt.Sub(time.Now()).Seconds())

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken validates a refresh token and generates a new token pair
func (s *AuthService) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	// Parse and validate refresh token
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Extract user UUID
	userUUID, ok := claims["uuid"].(string)
	if !ok || userUUID == "" {
		return nil, errors.New("invalid token: missing user UUID")
	}

	// Generate new token pair
	return s.GenerateTokenPair(userUUID)
}

// generateAccessToken generates an access token
func (s *AuthService) generateAccessToken(userUUID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTokenTTL)
	claims := jwt.MapClaims{
		"uuid": userUUID,
		"type": "access",
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// generateRefreshToken generates a refresh token
func (s *AuthService) generateRefreshToken(userUUID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.refreshTokenTTL)
	claims := jwt.MapClaims{
		"uuid": userUUID,
		"type": "refresh",
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
