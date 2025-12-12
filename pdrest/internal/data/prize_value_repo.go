package data

import (
	"context"
	"fmt"
	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PrizeValueRepository provides access to prize value data
type PrizeValueRepository interface {
	GetPrizeValuesByEventID(ctx context.Context, eventID string) ([]domain.PrizeValue, error)
	GetPrizeValueByID(ctx context.Context, id int) (*domain.PrizeValue, error)
}

// PostgresPrizeValueRepository implements PrizeValueRepository with PostgreSQL
type PostgresPrizeValueRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPrizeValueRepository(pool *pgxpool.Pool) *PostgresPrizeValueRepository {
	return &PostgresPrizeValueRepository{pool: pool}
}

func (r *PostgresPrizeValueRepository) GetPrizeValuesByEventID(ctx context.Context, eventID string) ([]domain.PrizeValue, error) {
	query := `
		SELECT id, event_id, value, label, segment_id, created_at, updated_at
		FROM prize_values
		WHERE event_id = $1
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query prize values: %w", err)
	}
	defer rows.Close()

	var prizeValues []domain.PrizeValue
	for rows.Next() {
		var pv domain.PrizeValue
		var segmentID *string

		if err := rows.Scan(
			&pv.ID,
			&pv.EventID,
			&pv.Value,
			&pv.Label,
			&segmentID,
			&pv.CreatedAt,
			&pv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prize value: %w", err)
		}

		pv.SegmentID = segmentID
		prizeValues = append(prizeValues, pv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prize values: %w", err)
	}

	return prizeValues, nil
}

func (r *PostgresPrizeValueRepository) GetPrizeValueByID(ctx context.Context, id int) (*domain.PrizeValue, error) {
	query := `
		SELECT id, event_id, value, label, segment_id, created_at, updated_at
		FROM prize_values
		WHERE id = $1
	`

	var pv domain.PrizeValue
	var segmentID *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pv.ID,
		&pv.EventID,
		&pv.Value,
		&pv.Label,
		&segmentID,
		&pv.CreatedAt,
		&pv.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get prize value: %w", err)
	}

	pv.SegmentID = segmentID
	return &pv, nil
}

// InMemoryPrizeValueRepository returns empty results (used when DB is unavailable)
type InMemoryPrizeValueRepository struct{}

func NewInMemoryPrizeValueRepository() *InMemoryPrizeValueRepository {
	return &InMemoryPrizeValueRepository{}
}

func (r *InMemoryPrizeValueRepository) GetPrizeValuesByEventID(ctx context.Context, eventID string) ([]domain.PrizeValue, error) {
	return []domain.PrizeValue{}, nil
}

func (r *InMemoryPrizeValueRepository) GetPrizeValueByID(ctx context.Context, id int) (*domain.PrizeValue, error) {
	return nil, fmt.Errorf("prize value retrieval requires database connection")
}
