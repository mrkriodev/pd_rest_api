package data

import (
	"context"
	"fmt"
	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RatingRepository provides access to rating data.
type RatingRepository interface {
	GetUserRatingTotals(ctx context.Context, userUUID string) (*domain.RatingTotals, error)
	GetGlobalRating(ctx context.Context, limit, offset int) ([]domain.GlobalRatingEntry, error)
	GetFriendsRatings(ctx context.Context, userUUID string, limit, offset int) ([]domain.FriendRatingEntry, error)
	AddPoints(ctx context.Context, userUUID string, points int64, source domain.RatingSource, description string) error
	GetMaxCreatedAt(ctx context.Context, userUUID string) (*int64, error)
}

// PostgresRatingRepository implements RatingRepository with PostgreSQL.
type PostgresRatingRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRatingRepository(pool *pgxpool.Pool) *PostgresRatingRepository {
	return &PostgresRatingRepository{pool: pool}
}

func (r *PostgresRatingRepository) GetUserRatingTotals(ctx context.Context, userUUID string) (*domain.RatingTotals, error) {
	query := `
		SELECT
			COALESCE(SUM(points) FILTER (WHERE source = 'from_event'), 0)::BIGINT AS from_event_points,
			COALESCE(SUM(points) FILTER (WHERE source = 'bet_bonus'), 0)::BIGINT AS bet_bonus_points,
			COALESCE(SUM(points) FILTER (WHERE source = 'promo_bonus'), 0)::BIGINT AS promo_bonus_points,
			COALESCE(SUM(points) FILTER (WHERE source = 'servivce_bonus'), 0)::BIGINT AS service_bonus_points
		FROM rating
		WHERE user_uuid = $1
	`

	var totals domain.RatingTotals

	if err := r.pool.QueryRow(ctx, query, userUUID).Scan(
		&totals.FromEvent,
		&totals.BetBonus,
		&totals.PromoBonus,
		&totals.ServiceBonus,
	); err != nil {
		return nil, fmt.Errorf("failed to get user rating totals: %w", err)
	}

	return &totals, nil
}

func (r *PostgresRatingRepository) GetGlobalRating(ctx context.Context, limit, offset int) ([]domain.GlobalRatingEntry, error) {
	query := `
		SELECT 
			user_uuid::text AS user_uuid,
			COALESCE(SUM(points), 0)::BIGINT AS total_points
		FROM rating
		GROUP BY user_uuid
		ORDER BY total_points DESC, user_uuid ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get global rating: %w", err)
	}
	defer rows.Close()

	var entries []domain.GlobalRatingEntry
	for rows.Next() {
		var entry domain.GlobalRatingEntry
		if err := rows.Scan(&entry.UserID, &entry.Value); err != nil {
			return nil, fmt.Errorf("failed to scan global rating entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating global rating rows: %w", err)
	}

	return entries, nil
}

func (r *PostgresRatingRepository) GetFriendsRatings(ctx context.Context, userUUID string, limit, offset int) ([]domain.FriendRatingEntry, error) {
	query := `
		SELECT 
			u.user_uuid::text AS friend_uuid,
			COALESCE(SUM(r.points), 0)::BIGINT AS total_points
		FROM users u
		LEFT JOIN rating r ON r.user_uuid = u.user_uuid
		WHERE u.referrer_user_uuid = $1
		GROUP BY u.user_uuid
		ORDER BY total_points DESC, friend_uuid ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userUUID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends ratings: %w", err)
	}
	defer rows.Close()

	var entries []domain.FriendRatingEntry
	for rows.Next() {
		var entry domain.FriendRatingEntry
		if err := rows.Scan(&entry.UserID, &entry.Value); err != nil {
			return nil, fmt.Errorf("failed to scan friends rating entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating friends rating rows: %w", err)
	}

	return entries, nil
}

func (r *PostgresRatingRepository) AddPoints(ctx context.Context, userUUID string, points int64, source domain.RatingSource, description string) error {
	if points <= 0 {
		return nil // Don't add zero or negative points
	}

	query := `
		INSERT INTO rating (user_uuid, points, source, description, created_at)
		VALUES ($1, $2, $3, $4, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
	`

	_, err := r.pool.Exec(ctx, query, userUUID, points, source, description)
	if err != nil {
		return fmt.Errorf("failed to add points: %w", err)
	}

	return nil
}

func (r *PostgresRatingRepository) GetMaxCreatedAt(ctx context.Context, userUUID string) (*int64, error) {
	query := `
		SELECT MAX(created_at)
		FROM rating
		WHERE user_uuid = $1
	`

	var maxCreatedAt *int64
	err := r.pool.QueryRow(ctx, query, userUUID).Scan(&maxCreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get max created_at: %w", err)
	}

	return maxCreatedAt, nil
}

// InMemoryRatingRepository returns zeroed totals (used when DB is unavailable).
type InMemoryRatingRepository struct{}

func NewInMemoryRatingRepository() *InMemoryRatingRepository {
	return &InMemoryRatingRepository{}
}

func (r *InMemoryRatingRepository) GetUserRatingTotals(ctx context.Context, userUUID string) (*domain.RatingTotals, error) {
	return &domain.RatingTotals{}, nil
}

func (r *InMemoryRatingRepository) GetGlobalRating(ctx context.Context, limit, offset int) ([]domain.GlobalRatingEntry, error) {
	return []domain.GlobalRatingEntry{}, nil
}

func (r *InMemoryRatingRepository) GetFriendsRatings(ctx context.Context, userUUID string, limit, offset int) ([]domain.FriendRatingEntry, error) {
	return []domain.FriendRatingEntry{}, nil
}

func (r *InMemoryRatingRepository) AddPoints(ctx context.Context, userUUID string, points int64, source domain.RatingSource, description string) error {
	return nil
}

func (r *InMemoryRatingRepository) GetMaxCreatedAt(ctx context.Context, userUUID string) (*int64, error) {
	return nil, nil
}
