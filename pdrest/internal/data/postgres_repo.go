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
