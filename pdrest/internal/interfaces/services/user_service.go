package services

import (
	"context"
	"errors"
	"pdrest/internal/data"
	"pdrest/internal/domain"
)

type UserService struct {
	repo data.UserRepository
}

func NewUserService(r data.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) GetLastLogin(uuid string) (*domain.UserLastLogin, error) {
	result, err := s.repo.GetLastLogin(uuid)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("user not found")
	}
	return result, nil
}

func (s *UserService) GetProfile(uuid string) (*domain.UserProfile, error) {
	result, err := s.repo.GetProfile(uuid)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("user not found")
	}
	return result, nil
}

func (s *UserService) GetUserByGoogleID(googleID string) (*domain.User, error) {
	result, err := s.repo.GetUserByGoogleID(googleID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("user not found")
	}
	return result, nil
}

func (s *UserService) GetUserByTelegramID(telegramID int64) (*domain.User, error) {
	result, err := s.repo.GetUserByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("user not found")
	}
	return result, nil
}

func (s *UserService) GetUserBySessionID(ctx context.Context, sessionID string) (*domain.User, error) {
	result, err := s.repo.GetUserBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("user not found")
	}
	return result, nil
}

func (s *UserService) GetUserBySessionAndIP(ctx context.Context, sessionID string, ipAddress string) (*domain.User, error) {
	result, err := s.repo.GetUserBySessionAndIP(ctx, sessionID, ipAddress)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("user not found")
	}
	return result, nil
}

func (s *UserService) CreateOrUpdateUserBySession(sessionID string, ipAddress string) error {
	return s.repo.CreateOrUpdateUserBySession(sessionID, ipAddress)
}

// RegisterUserWithGoogle registers or updates a user with Google OAuth information
func (s *UserService) RegisterUserWithGoogle(ctx context.Context, userUUID string, googleID string) error {
	if userUUID == "" {
		return errors.New("user_uuid is required")
	}
	if googleID == "" {
		return errors.New("google_id is required")
	}
	return s.repo.CreateOrUpdateUserWithGoogleInfo(ctx, userUUID, googleID)
}

// RegisterUserWithTelegram registers or updates a user with Telegram OAuth information
func (s *UserService) RegisterUserWithTelegram(ctx context.Context, userUUID string, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) error {
	if userUUID == "" {
		return errors.New("user_uuid is required")
	}
	if telegramID == 0 {
		return errors.New("telegram_id is required")
	}
	return s.repo.CreateOrUpdateUserWithTelegramInfo(ctx, userUUID, telegramID, telegramUsername, telegramFirstName, telegramLastName)
}

// RegisterUserWithTelegramByTelegramID creates or updates a user by telegram_id and returns user UUID
func (s *UserService) RegisterUserWithTelegramByTelegramID(ctx context.Context, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) (string, error) {
	if telegramID == 0 {
		return "", errors.New("telegram_id is required")
	}
	return s.repo.CreateOrUpdateUserWithTelegramInfoByTelegramID(ctx, telegramID, telegramUsername, telegramFirstName, telegramLastName)
}
