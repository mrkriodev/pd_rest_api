package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TelegramAuthService handles Telegram Web Login verification
type TelegramAuthService struct {
	botToken string
}

// NewTelegramAuthService creates a new Telegram auth service
func NewTelegramAuthService(botToken string) *TelegramAuthService {
	return &TelegramAuthService{
		botToken: botToken,
	}
}

// TelegramAuthData represents authentication data from Telegram Web Login
type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// TelegramUserInfo represents user information from Telegram
type TelegramUserInfo struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
}

// ValidateWithTelegram validates Telegram Web Login data and returns user info
func (s *TelegramAuthService) ValidateWithTelegram(authData TelegramAuthData) (*TelegramUserInfo, error) {
	if s.botToken == "" {
		return nil, errors.New("telegram bot token not configured")
	}

	if authData.ID == 0 {
		return nil, errors.New("telegram user ID is required")
	}

	if authData.Hash == "" {
		return nil, errors.New("hash is required")
	}

	// Check if auth_date is not too old (24 hours)
	currentTime := time.Now().Unix()
	if currentTime-authData.AuthDate > 86400 {
		return nil, errors.New("authentication data expired")
	}

	// Verify hash
	if !s.verifyHash(authData) {
		return nil, errors.New("invalid hash")
	}

	return &TelegramUserInfo{
		ID:        authData.ID,
		FirstName: authData.FirstName,
		LastName:  authData.LastName,
		Username:  authData.Username,
	}, nil
}

// verifyHash verifies the HMAC-SHA256 hash of Telegram auth data
func (s *TelegramAuthService) verifyHash(authData TelegramAuthData) bool {
	// Create data string for hashing
	// Format: key=value\nkey=value\n... (sorted by key, excluding hash)
	dataCheckString := s.createDataCheckString(authData)

	// Compute secret key (SHA256 of bot token)
	secretKey := sha256.Sum256([]byte(s.botToken))

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Compare hashes
	return hmac.Equal([]byte(expectedHash), []byte(authData.Hash))
}

// createDataCheckString creates the data check string for hash verification
// Format: key=value\nkey=value\n... (sorted by key, excluding hash)
func (s *TelegramAuthService) createDataCheckString(authData TelegramAuthData) string {
	// Build map of all fields except hash
	data := make(map[string]string)
	data["id"] = strconv.FormatInt(authData.ID, 10)
	data["first_name"] = authData.FirstName
	if authData.LastName != "" {
		data["last_name"] = authData.LastName
	}
	if authData.Username != "" {
		data["username"] = authData.Username
	}
	if authData.PhotoURL != "" {
		data["photo_url"] = authData.PhotoURL
	}
	data["auth_date"] = strconv.FormatInt(authData.AuthDate, 10)

	// Sort keys
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build data check string
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, data[k]))
	}

	return strings.Join(parts, "\n")
}
