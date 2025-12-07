package data

import (
	"context"
	"fmt"
	"pdrest/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BetRepository interface {
	CreateBet(ctx context.Context, bet *domain.Bet) error
	GetBetByID(ctx context.Context, betID int, userUUID string) (*domain.Bet, error)
	UpdateBetClosePrice(ctx context.Context, betID int, closePrice float64, closeTime time.Time) error
}

type PostgresBetRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBetRepository(pool *pgxpool.Pool) *PostgresBetRepository {
	return &PostgresBetRepository{pool: pool}
}

func (r *PostgresBetRepository) CreateBet(ctx context.Context, bet *domain.Bet) error {
	query := `
		INSERT INTO bets (user_uuid, side, sum, pair, timeframe, open_price, close_price, open_time, close_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	var closePrice interface{}
	if bet.ClosePrice != nil {
		closePrice = *bet.ClosePrice
	} else {
		closePrice = nil
	}

	var closeTime interface{}
	if bet.CloseTime != nil {
		closeTime = *bet.CloseTime
	} else {
		closeTime = nil
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		bet.UserUUID,
		bet.Side,
		bet.Sum,
		bet.Pair,
		bet.Timeframe,
		bet.OpenPrice,
		closePrice,
		bet.OpenTime,
		closeTime,
	).Scan(&bet.ID, &bet.CreatedAt, &bet.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create bet: %w", err)
	}

	return nil
}

func (r *PostgresBetRepository) GetBetByID(ctx context.Context, betID int, userUUID string) (*domain.Bet, error) {
	query := `
		SELECT id, user_uuid, side, sum, pair, timeframe, open_price, close_price, open_time, close_time, created_at, updated_at
		FROM bets
		WHERE id = $1 AND user_uuid = $2
	`

	var bet domain.Bet
	var closePrice *float64
	var closeTime *time.Time

	err := r.pool.QueryRow(ctx, query, betID, userUUID).Scan(
		&bet.ID,
		&bet.UserUUID,
		&bet.Side,
		&bet.Sum,
		&bet.Pair,
		&bet.Timeframe,
		&bet.OpenPrice,
		&closePrice,
		&bet.OpenTime,
		&closeTime,
		&bet.CreatedAt,
		&bet.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get bet: %w", err)
	}

	bet.ClosePrice = closePrice
	bet.CloseTime = closeTime

	return &bet, nil
}

func (r *PostgresBetRepository) UpdateBetClosePrice(ctx context.Context, betID int, closePrice float64, closeTime time.Time) error {
	query := `
		UPDATE bets
		SET close_price = $1, close_time = $2, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
		WHERE id = $3
	`

	_, err := r.pool.Exec(ctx, query, closePrice, closeTime, betID)
	if err != nil {
		return fmt.Errorf("failed to update bet close price: %w", err)
	}

	return nil
}

type InMemoryBetRepository struct{}

func NewInMemoryBetRepository() *InMemoryBetRepository {
	return &InMemoryBetRepository{}
}

func (r *InMemoryBetRepository) CreateBet(ctx context.Context, bet *domain.Bet) error {
	// In-memory repository doesn't support bet creation
	return fmt.Errorf("bet creation requires database connection")
}

func (r *InMemoryBetRepository) GetBetByID(ctx context.Context, betID int, userUUID string) (*domain.Bet, error) {
	// In-memory repository doesn't support bet retrieval
	return nil, fmt.Errorf("bet retrieval requires database connection")
}

func (r *InMemoryBetRepository) UpdateBetClosePrice(ctx context.Context, betID int, closePrice float64, closeTime time.Time) error {
	// In-memory repository doesn't support bet updates
	return fmt.Errorf("bet update requires database connection")
}
