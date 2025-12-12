package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouletteRepository interface {
	// Config methods
	GetRouletteConfigByType(ctx context.Context, rouletteType domain.RouletteType, eventID string) (*domain.RouletteConfig, error)
	GetRouletteConfigByID(ctx context.Context, id int) (*domain.RouletteConfig, error)
	CreateRouletteConfig(ctx context.Context, config *domain.RouletteConfig) error
	UpdateRouletteConfig(ctx context.Context, config *domain.RouletteConfig) error

	// Preauth token methods
	CreatePreauthToken(ctx context.Context, token *domain.RoulettePreauthToken) error
	GetPreauthToken(ctx context.Context, token string) (*domain.RoulettePreauthToken, error)
	UpdatePreauthTokenUserUUID(ctx context.Context, token string, userUUID string) error
	MarkPreauthTokenAsUsed(ctx context.Context, tokenID int) error
	ValidatePreauthToken(ctx context.Context, token string) (*domain.RoulettePreauthToken, error)

	// Roulette methods
	GetRouletteByPreauthToken(ctx context.Context, preauthTokenID int) (*domain.Roulette, error)
	CreateRoulette(ctx context.Context, roulette *domain.Roulette) error
	UpdateRoulette(ctx context.Context, roulette *domain.Roulette) error
	TakePrize(ctx context.Context, rouletteID int, prize string) error
}

type PostgresRouletteRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRouletteRepository(pool *pgxpool.Pool) *PostgresRouletteRepository {
	return &PostgresRouletteRepository{pool: pool}
}

// GetRouletteConfigByType retrieves active roulette config by type and event_id
func (r *PostgresRouletteRepository) GetRouletteConfigByType(ctx context.Context, rouletteType domain.RouletteType, eventID string) (*domain.RouletteConfig, error) {
	query := `
		SELECT id, roulette_type, event_id, max_spins, is_active, created_at, updated_at
		FROM roulette_config
		WHERE roulette_type = $1 AND event_id = $2 AND is_active = TRUE
		ORDER BY id DESC
		LIMIT 1
	`

	var config domain.RouletteConfig

	err := r.pool.QueryRow(ctx, query, string(rouletteType), eventID).Scan(
		&config.ID,
		&config.Type,
		&config.EventID,
		&config.MaxSpins,
		&config.IsActive,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get roulette config: %w", err)
	}

	return &config, nil
}

// GetRouletteConfigByID retrieves roulette config by ID
func (r *PostgresRouletteRepository) GetRouletteConfigByID(ctx context.Context, id int) (*domain.RouletteConfig, error) {
	query := `
		SELECT id, roulette_type, event_id, max_spins, is_active, created_at, updated_at
		FROM roulette_config
		WHERE id = $1
	`

	var config domain.RouletteConfig

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&config.ID,
		&config.Type,
		&config.EventID,
		&config.MaxSpins,
		&config.IsActive,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get roulette config: %w", err)
	}

	return &config, nil
}

// CreateRouletteConfig creates a new roulette config
func (r *PostgresRouletteRepository) CreateRouletteConfig(ctx context.Context, config *domain.RouletteConfig) error {
	nowMs := time.Now().UTC().UnixMilli()

	query := `
		INSERT INTO roulette_config (roulette_type, event_id, max_spins, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		string(config.Type),
		config.EventID,
		config.MaxSpins,
		config.IsActive,
		nowMs,
		nowMs,
	).Scan(&config.ID)
	if err != nil {
		return fmt.Errorf("failed to create roulette config: %w", err)
	}

	config.CreatedAt = nowMs
	config.UpdatedAt = nowMs
	return nil
}

// UpdateRouletteConfig updates an existing roulette config
func (r *PostgresRouletteRepository) UpdateRouletteConfig(ctx context.Context, config *domain.RouletteConfig) error {
	nowMs := time.Now().UTC().UnixMilli()

	query := `
		UPDATE roulette_config
		SET roulette_type = $2,
		    event_id = $3,
		    max_spins = $4,
		    is_active = $5,
		    updated_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		config.ID,
		string(config.Type),
		config.EventID,
		config.MaxSpins,
		config.IsActive,
		nowMs,
	)
	if err != nil {
		return fmt.Errorf("failed to update roulette config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("roulette config with id %d not found", config.ID)
	}

	config.UpdatedAt = nowMs
	return nil
}

// CreatePreauthToken creates a new preauth token
func (r *PostgresRouletteRepository) CreatePreauthToken(ctx context.Context, token *domain.RoulettePreauthToken) error {
	nowMs := time.Now().UTC().UnixMilli()

	query := `
		INSERT INTO roulette_preauth_token (token, user_uuid, roulette_config_id, is_used, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		token.Token,
		token.UserUUID,
		token.RouletteConfigID,
		token.IsUsed,
		token.ExpiresAt,
		nowMs,
	).Scan(&token.ID)
	if err != nil {
		return fmt.Errorf("failed to create preauth token: %w", err)
	}

	token.CreatedAt = nowMs
	return nil
}

// GetPreauthToken retrieves a preauth token by token string
func (r *PostgresRouletteRepository) GetPreauthToken(ctx context.Context, token string) (*domain.RoulettePreauthToken, error) {
	query := `
		SELECT id, token, user_uuid, roulette_config_id, is_used, expires_at, created_at
		FROM roulette_preauth_token
		WHERE token = $1
	`

	var preauthToken domain.RoulettePreauthToken
	var userUUIDPtr *string

	err := r.pool.QueryRow(ctx, query, token).Scan(
		&preauthToken.ID,
		&preauthToken.Token,
		&userUUIDPtr,
		&preauthToken.RouletteConfigID,
		&preauthToken.IsUsed,
		&preauthToken.ExpiresAt,
		&preauthToken.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get preauth token: %w", err)
	}

	preauthToken.UserUUID = userUUIDPtr
	return &preauthToken, nil
}

// UpdatePreauthTokenUserUUID updates the user_uuid of a preauth token
func (r *PostgresRouletteRepository) UpdatePreauthTokenUserUUID(ctx context.Context, token string, userUUID string) error {
	query := `UPDATE roulette_preauth_token SET user_uuid = $1 WHERE token = $2`

	result, err := r.pool.Exec(ctx, query, userUUID, token)
	if err != nil {
		return fmt.Errorf("failed to update preauth token user_uuid: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("preauth token not found")
	}

	return nil
}

// MarkPreauthTokenAsUsed marks a preauth token as used
func (r *PostgresRouletteRepository) MarkPreauthTokenAsUsed(ctx context.Context, tokenID int) error {
	query := `UPDATE roulette_preauth_token SET is_used = TRUE WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark preauth token as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("preauth token with id %d not found", tokenID)
	}

	return nil
}

// ValidatePreauthToken validates a preauth token and returns it if valid
func (r *PostgresRouletteRepository) ValidatePreauthToken(ctx context.Context, token string) (*domain.RoulettePreauthToken, error) {
	preauthToken, err := r.GetPreauthToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if preauthToken == nil {
		return nil, fmt.Errorf("preauth token not found")
	}

	// Check if token is already used
	if preauthToken.IsUsed {
		return nil, fmt.Errorf("preauth token already used")
	}

	// Check if token is expired
	nowMs := time.Now().UTC().UnixMilli()
	if preauthToken.ExpiresAt < nowMs {
		return nil, fmt.Errorf("preauth token expired")
	}

	return preauthToken, nil
}

// GetRouletteByPreauthToken retrieves roulette by preauth token ID
func (r *PostgresRouletteRepository) GetRouletteByPreauthToken(ctx context.Context, preauthTokenID int) (*domain.Roulette, error) {
	query := `
		SELECT id, roulette_config_id, preauth_token_id, spin_number, prize, 
		       prize_taken, spin_result, created_at, updated_at, prize_taken_at
		FROM roulette
		WHERE preauth_token_id = $1
	`

	var roulette domain.Roulette
	var prizePtr *string
	var spinResultJSON []byte
	var prizeTakenAtPtr *int64

	err := r.pool.QueryRow(ctx, query, preauthTokenID).Scan(
		&roulette.ID,
		&roulette.RouletteConfigID,
		&roulette.PreauthTokenID,
		&roulette.SpinNumber,
		&prizePtr,
		&roulette.PrizeTaken,
		&spinResultJSON,
		&roulette.CreatedAt,
		&roulette.UpdatedAt,
		&prizeTakenAtPtr,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get roulette: %w", err)
	}

	roulette.Prize = prizePtr
	roulette.PrizeTakenAt = prizeTakenAtPtr

	// Unmarshal spin_result JSONB
	if len(spinResultJSON) > 0 {
		if err := json.Unmarshal(spinResultJSON, &roulette.SpinResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spin_result: %w", err)
		}
	}

	return &roulette, nil
}

// CreateRoulette creates a new roulette entry
func (r *PostgresRouletteRepository) CreateRoulette(ctx context.Context, roulette *domain.Roulette) error {
	nowMs := time.Now().UTC().UnixMilli()

	// Marshal spin_result to JSON
	var spinResultJSON []byte
	var err error
	if roulette.SpinResult != nil {
		spinResultJSON, err = json.Marshal(roulette.SpinResult)
		if err != nil {
			return fmt.Errorf("failed to marshal spin_result: %w", err)
		}
	}

	query := `
		INSERT INTO roulette (roulette_config_id, preauth_token_id, spin_number, 
		                     prize, prize_taken, spin_result, created_at, updated_at, prize_taken_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err = r.pool.QueryRow(ctx, query,
		roulette.RouletteConfigID,
		roulette.PreauthTokenID,
		roulette.SpinNumber,
		roulette.Prize,
		roulette.PrizeTaken,
		spinResultJSON,
		nowMs,
		nowMs,
		roulette.PrizeTakenAt,
	).Scan(&roulette.ID)
	if err != nil {
		return fmt.Errorf("failed to create roulette: %w", err)
	}

	roulette.CreatedAt = nowMs
	roulette.UpdatedAt = nowMs
	return nil
}

// UpdateRoulette updates an existing roulette entry
func (r *PostgresRouletteRepository) UpdateRoulette(ctx context.Context, roulette *domain.Roulette) error {
	nowMs := time.Now().UTC().UnixMilli()

	// Marshal spin_result to JSON
	var spinResultJSON []byte
	var err error
	if roulette.SpinResult != nil {
		spinResultJSON, err = json.Marshal(roulette.SpinResult)
		if err != nil {
			return fmt.Errorf("failed to marshal spin_result: %w", err)
		}
	}

	query := `
		UPDATE roulette
		SET spin_number = $2,
		    prize = $3,
		    prize_taken = $4,
		    spin_result = $5,
		    updated_at = $6,
		    prize_taken_at = $7
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		roulette.ID,
		roulette.SpinNumber,
		roulette.Prize,
		roulette.PrizeTaken,
		spinResultJSON,
		nowMs,
		roulette.PrizeTakenAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update roulette: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("roulette with id %d not found", roulette.ID)
	}

	roulette.UpdatedAt = nowMs
	return nil
}

// TakePrize marks the prize as taken
func (r *PostgresRouletteRepository) TakePrize(ctx context.Context, rouletteID int, prize string) error {
	nowMs := time.Now().UTC().UnixMilli()

	query := `
		UPDATE roulette
		SET prize = $2,
		    prize_taken = TRUE,
		    prize_taken_at = $3,
		    updated_at = $3
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, rouletteID, prize, nowMs)
	if err != nil {
		return fmt.Errorf("failed to take prize: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("roulette with id %d not found", rouletteID)
	}

	return nil
}
