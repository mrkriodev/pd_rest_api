package data

import "pdrest/internal/domain"

type UserRepository interface {
	GetLastLogin(uuid string) (*domain.UserLastLogin, error)
	GetProfile(uuid string) (*domain.UserProfile, error)
	GetUserByGoogleID(googleID string) (*domain.User, error)
	GetUserByTelegramID(telegramID int64) (*domain.User, error)
	CreateOrUpdateUserBySession(sessionID string, ipAddress string) error
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

func (r *InMemoryUserRepository) CreateOrUpdateUserBySession(sessionID string, ipAddress string) error {
	// In-memory repository doesn't support user creation
	return nil
}
