package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

// FetchUserInfo requests user info from Google OAuth userinfo endpoint.
func (s *GoogleAuthService) FetchUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	endpoint, err := url.Parse("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to build userinfo URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("client_id", s.clientID)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch userinfo: status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &GoogleUserInfo{
		ID:    payload.Sub,
		Email: payload.Email,
		Name:  payload.Name,
	}, nil
}
