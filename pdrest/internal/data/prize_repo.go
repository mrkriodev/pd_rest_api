package data

import (
	"context"
	"fmt"
	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PrizeRepository provides access to prize data.
type PrizeRepository interface {
	CreatePrize(ctx context.Context, prize *domain.Prize) error
	GetPrizeByID(ctx context.Context, id int) (*domain.Prize, error)
	GetPrizesByUserID(ctx context.Context, userID string) ([]domain.Prize, error)
	GetPrizesByPreauthTokenID(ctx context.Context, preauthTokenID int) ([]domain.Prize, error)
}

// PostgresPrizeRepository implements PrizeRepository with PostgreSQL.
type PostgresPrizeRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPrizeRepository(pool *pgxpool.Pool) *PostgresPrizeRepository {
	return &PostgresPrizeRepository{pool: pool}
}

func (r *PostgresPrizeRepository) CreatePrize(ctx context.Context, prize *domain.Prize) error {
	query := `
		INSERT INTO prizes (event_id, user_uuid, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var userUUID interface{}
	if prize.UserID != nil {
		userUUID = *prize.UserID
	} else {
		userUUID = nil
	}

	var eventID interface{}
	if prize.EventID != nil {
		eventID = *prize.EventID
	} else {
		eventID = nil
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		eventID,
		userUUID,
		prize.PreauthTokenID,
		prize.RouletteID,
		prize.PrizeValue,
		prize.PrizeType,
		prize.AwardedAt,
		prize.CreatedAt,
	).Scan(&prize.ID)

	if err != nil {
		return fmt.Errorf("failed to create prize: %w", err)
	}

	return nil
}

func (r *PostgresPrizeRepository) GetPrizeByID(ctx context.Context, id int) (*domain.Prize, error) {
	query := `
		SELECT id, event_id, user_uuid, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at
		FROM prizes
		WHERE id = $1
	`

	var prize domain.Prize
	var eventID *string
	var userID *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&prize.ID,
		&eventID,
		&userID,
		&prize.PreauthTokenID,
		&prize.RouletteID,
		&prize.PrizeValue,
		&prize.PrizeType,
		&prize.AwardedAt,
		&prize.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get prize: %w", err)
	}

	prize.EventID = eventID
	prize.UserID = userID

	return &prize, nil
}

func (r *PostgresPrizeRepository) GetPrizesByUserID(ctx context.Context, userID string) ([]domain.Prize, error) {
	query := `
		SELECT id, event_id, user_uuid, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at
		FROM prizes
		WHERE user_uuid = $1
		ORDER BY awarded_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prizes by user ID: %w", err)
	}
	defer rows.Close()

	var prizes []domain.Prize
	for rows.Next() {
		var prize domain.Prize
		var eventID *string
		var userIDPtr *string

		if err := rows.Scan(
			&prize.ID,
			&eventID,
			&userIDPtr,
			&prize.PreauthTokenID,
			&prize.RouletteID,
			&prize.PrizeValue,
			&prize.PrizeType,
			&prize.AwardedAt,
			&prize.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prize: %w", err)
		}

		prize.EventID = eventID
		prize.UserID = userIDPtr
		prizes = append(prizes, prize)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prizes: %w", err)
	}

	return prizes, nil
}

func (r *PostgresPrizeRepository) GetPrizesByPreauthTokenID(ctx context.Context, preauthTokenID int) ([]domain.Prize, error) {
	query := `
		SELECT id, event_id, user_uuid, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at
		FROM prizes
		WHERE preauth_token_id = $1
		ORDER BY awarded_at DESC
	`

	rows, err := r.pool.Query(ctx, query, preauthTokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prizes by preauth token ID: %w", err)
	}
	defer rows.Close()

	var prizes []domain.Prize
	for rows.Next() {
		var prize domain.Prize
		var eventID *string
		var userID *string

		if err := rows.Scan(
			&prize.ID,
			&eventID,
			&userID,
			&prize.PreauthTokenID,
			&prize.RouletteID,
			&prize.PrizeValue,
			&prize.PrizeType,
			&prize.AwardedAt,
			&prize.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prize: %w", err)
		}

		prize.EventID = eventID
		prize.UserID = userID
		prizes = append(prizes, prize)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prizes: %w", err)
	}

	return prizes, nil
}

// InMemoryPrizeRepository returns empty results (used when DB is unavailable).
type InMemoryPrizeRepository struct{}

func NewInMemoryPrizeRepository() *InMemoryPrizeRepository {
	return &InMemoryPrizeRepository{}
}

func (r *InMemoryPrizeRepository) CreatePrize(ctx context.Context, prize *domain.Prize) error {
	return fmt.Errorf("prize creation requires database connection")
}

func (r *InMemoryPrizeRepository) GetPrizeByID(ctx context.Context, id int) (*domain.Prize, error) {
	return nil, fmt.Errorf("prize retrieval requires database connection")
}

func (r *InMemoryPrizeRepository) GetPrizesByUserID(ctx context.Context, userID string) ([]domain.Prize, error) {
	return []domain.Prize{}, nil
}

func (r *InMemoryPrizeRepository) GetPrizesByPreauthTokenID(ctx context.Context, preauthTokenID int) ([]domain.Prize, error) {
	return []domain.Prize{}, nil
}
