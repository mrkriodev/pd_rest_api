package data

import (
	"context"
	"fmt"

	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) GetLastLogin(uuid string) (*domain.UserLastLogin, error) {
	ctx := context.Background()

	var result domain.UserLastLogin
	var lastLoginAt *int64
	query := `SELECT user_uuid, last_login_at FROM users WHERE user_uuid = $1`

	err := r.pool.QueryRow(ctx, query, uuid).Scan(&result.UserID, &lastLoginAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user last login: %w", err)
	}

	result.LastLoginAt = lastLoginAt
	return &result, nil
}

func (r *PostgresUserRepository) GetProfile(uuid string) (*domain.UserProfile, error) {
	ctx := context.Background()

	var result domain.UserProfile
	var authProvider *string
	var googleName *string
	var telegramUsername *string

	query := `SELECT user_uuid, auth_provider, google_name, telegram_username FROM users WHERE user_uuid = $1`

	err := r.pool.QueryRow(ctx, query, uuid).Scan(&result.UserID, &authProvider, &googleName, &telegramUsername)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Determine username based on auth_provider
	if authProvider != nil {
		if *authProvider == "google" && googleName != nil {
			result.Username = googleName
		} else if *authProvider == "telegram" && telegramUsername != nil {
			result.Username = telegramUsername
		}
	}

	return &result, nil
}

func (r *PostgresUserRepository) GetUserByGoogleID(googleID string) (*domain.User, error) {
	ctx := context.Background()

	var result domain.User
	query := `SELECT user_uuid, google_id, google_email, google_name FROM users WHERE google_id = $1`

	err := r.pool.QueryRow(ctx, query, googleID).Scan(&result.UserID, &result.GoogleID, &result.GoogleEmail, &result.GoogleName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by google_id: %w", err)
	}

	return &result, nil
}

func (r *PostgresUserRepository) GetUserByTelegramID(telegramID int64) (*domain.User, error) {
	ctx := context.Background()

	var result domain.User
	query := `SELECT user_uuid, telegram_id, telegram_username, telegram_first_name, telegram_last_name FROM users WHERE telegram_id = $1`

	err := r.pool.QueryRow(ctx, query, telegramID).Scan(&result.UserID, &result.TelegramID, &result.TelegramUsername, &result.TelegramFirstName, &result.TelegramLastName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by telegram_id: %w", err)
	}

	return &result, nil
}

func (r *PostgresUserRepository) GetUserBySessionID(ctx context.Context, sessionID string) (*domain.User, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	query := `
		SELECT user_uuid
		FROM users
		WHERE session_id = $1
	`

	var user domain.User

	err := r.pool.QueryRow(ctx, query, sessionID).Scan(
		&user.UserID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by session_id: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetUserBySessionAndIP(ctx context.Context, sessionID string, ipAddress string) (*domain.User, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if ipAddress == "" {
		return nil, fmt.Errorf("ip_address is required")
	}

	query := `
		SELECT user_uuid
		FROM users
		WHERE session_id = $1 AND ip_address = $2
	`

	var user domain.User

	err := r.pool.QueryRow(ctx, query, sessionID, ipAddress).Scan(
		&user.UserID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by session_id and ip: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) CreateOrUpdateUserBySession(sessionID string, ipAddress string) error {
	ctx := context.Background()

	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	// Use INSERT ... ON CONFLICT to handle race conditions
	// If user exists, update last_login_at; if not, create new user
	insertQuery := `
		INSERT INTO users (session_id, ip_address, authorized_fully, auth_provider, last_login_at)
		VALUES ($1, $2, FALSE, NULL, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
		ON CONFLICT (session_id) DO UPDATE 
		SET last_login_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
		    ip_address = EXCLUDED.ip_address
	`

	_, err := r.pool.Exec(ctx, insertQuery, sessionID, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to create or update user: %w", err)
	}

	return nil
}

// CreateOrUpdateUserWithGoogleInfo creates or updates a user with Google OAuth information
func (r *PostgresUserRepository) CreateOrUpdateUserWithGoogleInfo(ctx context.Context, userUUID string, googleID string) error {
	if userUUID == "" {
		return fmt.Errorf("user_uuid is required")
	}
	if googleID == "" {
		return fmt.Errorf("google_id is required")
	}

	// Use INSERT ... ON CONFLICT to handle both create and update cases
	// If user exists (by user_uuid), update Google info; if not, create new user with Google info
	insertQuery := `
		INSERT INTO users (user_uuid, google_id, authorized_fully, auth_provider, last_login_at, updated_at)
		VALUES ($1, $2, TRUE, 'google', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
		ON CONFLICT (user_uuid) DO UPDATE 
		SET google_id = EXCLUDED.google_id,
		    authorized_fully = TRUE,
		    auth_provider = 'google',
		    last_login_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
	`

	_, err := r.pool.Exec(ctx, insertQuery, userUUID, googleID)
	if err != nil {
		return fmt.Errorf("failed to create or update user with Google info: %w", err)
	}

	return nil
}

// CreateOrUpdateUserWithTelegramInfo creates or updates a user with Telegram OAuth information
func (r *PostgresUserRepository) CreateOrUpdateUserWithTelegramInfo(ctx context.Context, userUUID string, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) error {
	if userUUID == "" {
		return fmt.Errorf("user_uuid is required")
	}
	if telegramID == 0 {
		return fmt.Errorf("telegram_id is required")
	}

	// Use INSERT ... ON CONFLICT to handle both create and update cases
	// If user exists (by user_uuid), update Telegram info; if not, create new user with Telegram info
	insertQuery := `
		INSERT INTO users (user_uuid, telegram_id, telegram_username, telegram_first_name, telegram_last_name, authorized_fully, auth_provider, last_login_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, TRUE, 'telegram', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
		ON CONFLICT (user_uuid) DO UPDATE 
		SET telegram_id = EXCLUDED.telegram_id,
		    telegram_username = EXCLUDED.telegram_username,
		    telegram_first_name = EXCLUDED.telegram_first_name,
		    telegram_last_name = EXCLUDED.telegram_last_name,
		    authorized_fully = TRUE,
		    auth_provider = 'telegram',
		    last_login_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
	`

	_, err := r.pool.Exec(ctx, insertQuery, userUUID, telegramID, telegramUsername, telegramFirstName, telegramLastName)
	if err != nil {
		return fmt.Errorf("failed to create or update user with Telegram info: %w", err)
	}

	return nil
}
