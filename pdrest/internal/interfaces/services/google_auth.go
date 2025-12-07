package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// GoogleAuthService handles Google OAuth token verification
type GoogleAuthService struct {
	httpClient *http.Client
}

// NewGoogleAuthService creates a new Google auth service
func NewGoogleAuthService() (*GoogleAuthService, error) {
	return &GoogleAuthService{
		httpClient: &http.Client{},
	}, nil
}

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID    string
	Email string
	Name  string
}

// ValidateWithGoogle validates a Google access token and returns user info
func (s *GoogleAuthService) ValidateWithGoogle(token string) (*GoogleUserInfo, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	ctx := context.Background()

	// Create OAuth2 config (no client ID/secret needed for tokeninfo)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := oauth2.NewClient(ctx, ts)

	// Create OAuth2 service with authenticated client
	service, err := googleoauth2.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google OAuth2 service: %w", err)
	}

	// Use the tokeninfo endpoint to validate the token
	tokenInfo, err := service.Tokeninfo().AccessToken(token).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to validate Google token: %w", err)
	}

	if tokenInfo.UserId == "" {
		return nil, errors.New("invalid token: missing user ID")
	}

	// Get user info using the authenticated client
	userInfo, err := service.Userinfo.V2.Me.Get().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &GoogleUserInfo{
		ID:    tokenInfo.UserId,
		Email: userInfo.Email,
		Name:  userInfo.Name,
	}, nil
}
