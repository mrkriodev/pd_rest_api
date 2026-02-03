package data

import (
	"context"
	"pdrest/internal/domain"
)

type UserRepository interface {
	GetLastLogin(uuid string) (*domain.UserLastLogin, error)
	GetProfile(uuid string) (*domain.UserProfile, error)
	GetUserByGoogleID(googleID string) (*domain.User, error)
	GetUserByTelegramID(telegramID int64) (*domain.User, error)
	GetUserBySessionID(ctx context.Context, sessionID string) (*domain.User, error)
	GetUserBySessionAndIP(ctx context.Context, sessionID string, ipAddress string) (*domain.User, error)
	CreateOrUpdateUserBySession(sessionID string, ipAddress string) error
	CreateOrUpdateUserWithGoogleInfo(ctx context.Context, userUUID string, googleID string) error
	CreateOrUpdateUserWithTelegramInfo(ctx context.Context, userUUID string, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) error
	CreateOrUpdateUserWithTelegramInfoByTelegramID(ctx context.Context, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) (string, error)
}

type InMemoryUserRepository struct {
	storage map[string]*int64
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		storage: make(map[string]*int64),
	}
}

func (r *InMemoryUserRepository) GetLastLogin(uuid string) (*domain.UserLastLogin, error) {
	lastLoginAt, ok := r.storage[uuid]
	if !ok {
		return nil, nil
	}
	return &domain.UserLastLogin{UserID: uuid, LastLoginAt: lastLoginAt}, nil
}

func (r *InMemoryUserRepository) GetProfile(uuid string) (*domain.UserProfile, error) {
	// In-memory repository doesn't have profile data
	return nil, nil
}

func (r *InMemoryUserRepository) GetUserByGoogleID(googleID string) (*domain.User, error) {
	// In-memory repository doesn't have Google user data
	return nil, nil
}

func (r *InMemoryUserRepository) GetUserByTelegramID(telegramID int64) (*domain.User, error) {
	// In-memory repository doesn't have Telegram user data
	return nil, nil
}

func (r *InMemoryUserRepository) GetUserBySessionID(ctx context.Context, sessionID string) (*domain.User, error) {
	// In-memory repository doesn't support user lookup by session
	return nil, nil
}

func (r *InMemoryUserRepository) GetUserBySessionAndIP(ctx context.Context, sessionID string, ipAddress string) (*domain.User, error) {
	// In-memory repository doesn't support user lookup by session + IP
	return nil, nil
}

func (r *InMemoryUserRepository) CreateOrUpdateUserBySession(sessionID string, ipAddress string) error {
	// In-memory repository doesn't support user creation
	return nil
}

func (r *InMemoryUserRepository) CreateOrUpdateUserWithGoogleInfo(ctx context.Context, userUUID string, googleID string) error {
	// In-memory repository doesn't support user creation
	return nil
}

func (r *InMemoryUserRepository) CreateOrUpdateUserWithTelegramInfo(ctx context.Context, userUUID string, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) error {
	// In-memory repository doesn't support user creation
	return nil
}

func (r *InMemoryUserRepository) CreateOrUpdateUserWithTelegramInfoByTelegramID(ctx context.Context, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) (string, error) {
	// In-memory repository doesn't support user creation
	return "", nil
}
