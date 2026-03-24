package data

import (
	"context"
	"fmt"
	"strings"

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

func (r *PostgresUserRepository) GetUserByUUID(ctx context.Context, userUUID string) (*domain.User, error) {
	if userUUID == "" {
		return nil, fmt.Errorf("user_uuid is required")
	}

	var result domain.User
	query := `SELECT user_uuid, telegram_id FROM users WHERE user_uuid = $1`
	err := r.pool.QueryRow(ctx, query, userUUID).Scan(&result.UserID, &result.TelegramID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by uuid: %w", err)
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

// CreateOrUpdateUserWithGoogleInfoByGoogleID creates or updates a user with Google OAuth information by google_id
func (r *PostgresUserRepository) CreateOrUpdateUserWithGoogleInfoByGoogleID(ctx context.Context, googleID string) (string, error) {
	if googleID == "" {
		return "", fmt.Errorf("google_id is required")
	}

	insertQuery := `
		INSERT INTO users (google_id, authorized_fully, auth_provider, last_login_at, updated_at)
		VALUES ($1, TRUE, 'google', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
		ON CONFLICT (google_id) DO UPDATE 
		SET authorized_fully = TRUE,
		    auth_provider = 'google',
		    last_login_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
		RETURNING user_uuid
	`

	var userUUID string
	if err := r.pool.QueryRow(ctx, insertQuery, googleID).Scan(&userUUID); err != nil {
		return "", fmt.Errorf("failed to create or update user with Google info: %w", err)
	}

	return userUUID, nil
}

func (r *PostgresUserRepository) UpdateMainRefIfEmpty(ctx context.Context, userUUID string, mainRef string) error {
	if userUUID == "" {
		return fmt.Errorf("user_uuid is required")
	}
	if mainRef == "" {
		return fmt.Errorf("main_ref is required")
	}

	query := `
		UPDATE users
		SET main_ref = $2
		WHERE user_uuid = $1 AND (main_ref IS NULL OR main_ref = '')
	`
	_, err := r.pool.Exec(ctx, query, userUUID, mainRef)
	if err != nil {
		return fmt.Errorf("failed to update main_ref: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) ApplyReferralCode(ctx context.Context, userUUID string, referralCode string) error {
	if userUUID == "" {
		return fmt.Errorf("user_uuid is required")
	}
	if referralCode == "" {
		return fmt.Errorf("referral_code is required")
	}
	normalizedCode := strings.TrimSpace(referralCode)
	if normalizedCode == "" {
		return fmt.Errorf("referral_code is invalid")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin referral transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var referrerUUID string
	queryReferrer := `
		SELECT user_uuid::text
		FROM users
		WHERE main_ref = $1
		LIMIT 1
	`
	if err = tx.QueryRow(ctx, queryReferrer, normalizedCode).Scan(&referrerUUID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("referral code not found")
		}
		return fmt.Errorf("failed to lookup referrer: %w", err)
	}

	if referrerUUID == userUUID {
		return fmt.Errorf("cannot use own referral code")
	}

	querySetReferrer := `
		UPDATE users
		SET referrer_user_uuid = $2
		WHERE user_uuid = $1 AND referrer_user_uuid IS NULL
	`
	tag, execErr := tx.Exec(ctx, querySetReferrer, userUUID, referrerUUID)
	if execErr != nil {
		return fmt.Errorf("failed to set referrer: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("referrer already set")
	}

	queryAppend := `
		UPDATE users
		SET add_refs = (
			SELECT ARRAY(
				SELECT DISTINCT unnest(add_refs || $2::text[])
			)
		)
		WHERE user_uuid = $1
	`
	if _, err = tx.Exec(ctx, queryAppend, referrerUUID, userUUID); err != nil {
		return fmt.Errorf("failed to append add_refs: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit referral transaction: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) UpdateUserLanguage(ctx context.Context, userUUID string, language string) error {
	if userUUID == "" {
		return fmt.Errorf("user_uuid is required")
	}
	if strings.TrimSpace(language) == "" {
		return fmt.Errorf("language is required")
	}

	query := `
		UPDATE users
		SET language = $2
		WHERE user_uuid = $1
	`
	_, err := r.pool.Exec(ctx, query, userUUID, strings.TrimSpace(language))
	if err != nil {
		return fmt.Errorf("failed to update language: %w", err)
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

// CreateOrUpdateUserWithTelegramInfoByTelegramID creates or updates a user with Telegram OAuth information by telegram_id
func (r *PostgresUserRepository) CreateOrUpdateUserWithTelegramInfoByTelegramID(ctx context.Context, telegramID int64, telegramUsername string, telegramFirstName string, telegramLastName string) (string, error) {
	if telegramID == 0 {
		return "", fmt.Errorf("telegram_id is required")
	}

	insertQuery := `
		INSERT INTO users (telegram_id, telegram_username, telegram_first_name, telegram_last_name, authorized_fully, auth_provider, last_login_at, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, 'telegram', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
		ON CONFLICT (telegram_id) DO UPDATE 
		SET telegram_username = EXCLUDED.telegram_username,
		    telegram_first_name = EXCLUDED.telegram_first_name,
		    telegram_last_name = EXCLUDED.telegram_last_name,
		    authorized_fully = TRUE,
		    auth_provider = 'telegram',
		    last_login_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
		RETURNING user_uuid
	`

	var userUUID string
	if err := r.pool.QueryRow(ctx, insertQuery, telegramID, telegramUsername, telegramFirstName, telegramLastName).Scan(&userUUID); err != nil {
		return "", fmt.Errorf("failed to create or update user with Telegram info: %w", err)
	}

	return userUUID, nil
}
