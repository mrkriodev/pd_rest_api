package services

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/idtoken"
)

// GoogleAuthService handles Google ID token verification
type GoogleAuthService struct {
	clientID string
}

// NewGoogleAuthService creates a new Google auth service
func NewGoogleAuthService(clientID string) (*GoogleAuthService, error) {
	if clientID == "" {
		return nil, errors.New("Google client ID is required")
	}
	return &GoogleAuthService{
		clientID: clientID,
	}, nil
}

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID    string
	Email string
	Name  string
}

// ValidateWithGoogle validates a Google ID token and returns user info
func (s *GoogleAuthService) ValidateWithGoogle(token string) (*GoogleUserInfo, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	ctx := context.Background()

	// Validate the ID token using Google's idtoken package
	// This verifies the token signature, expiration, and audience (client ID)
	payload, err := idtoken.Validate(ctx, token, s.clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate Google ID token: %w", err)
	}

	// Extract user information from the token payload
	userID, ok := payload.Claims["sub"].(string)
	if !ok || userID == "" {
		return nil, errors.New("invalid token: missing user ID (sub claim)")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)

	return &GoogleUserInfo{
		ID:    userID,
		Email: email,
		Name:  name,
	}, nil
}
