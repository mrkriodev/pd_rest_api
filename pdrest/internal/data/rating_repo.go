package data

import (
	"context"
	"database/sql"
	"fmt"
	"pdrest/internal/domain"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RatingRepository provides access to rating data.
type RatingRepository interface {
	GetUserRatingTotals(ctx context.Context, userUUID string) (*domain.RatingTotals, error)
	GetGlobalRating(ctx context.Context, limit, offset int) ([]domain.GlobalRatingEntry, error)
	GetFriendsRatings(ctx context.Context, userUUID string, limit, offset int) ([]domain.FriendRatingEntry, error)
	AddPoints(ctx context.Context, userUUID string, points int64, gotPrizeID *int, betID *int, description string) error
	GetMaxCreatedAt(ctx context.Context, userUUID string) (*int64, error)
	GetUserBetPointsInRange(ctx context.Context, userUUID string, startMs, endMs int64) (int64, error)
	GetBetPointsLeaderboard(ctx context.Context, startMs, endMs int64, limit int) ([]domain.BetPrizeLeaderboardEntry, error)
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
			COALESCE(SUM(points), 0)::BIGINT AS total_points
		FROM rating
		WHERE user_uuid = $1
	`

	var totals domain.RatingTotals

	if err := r.pool.QueryRow(ctx, query, userUUID).Scan(
		&totals.FromEvent,
	); err != nil {
		return nil, fmt.Errorf("failed to get user rating totals: %w", err)
	}

	// Backward-compatible totals: store total points in FromEvent bucket.
	totals.BetBonus = 0
	totals.PromoBonus = 0
	totals.ServiceBonus = 0

	return &totals, nil
}

func (r *PostgresRatingRepository) GetGlobalRating(ctx context.Context, limit, offset int) ([]domain.GlobalRatingEntry, error) {
	query := `
		SELECT
			r.user_uuid::text AS user_uuid,
			COALESCE(SUM(r.points), 0)::BIGINT AS total_points,
			u.google_name,
			u.telegram_username,
			u.telegram_first_name,
			u.telegram_last_name
		FROM rating r
		LEFT JOIN users u ON u.user_uuid = r.user_uuid
		GROUP BY r.user_uuid, u.google_name, u.telegram_username, u.telegram_first_name, u.telegram_last_name
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
		var userUUID string
		var totalPoints int64
		var googleName sql.NullString
		var telegramUsername sql.NullString
		var telegramFirstName sql.NullString
		var telegramLastName sql.NullString

		if err := rows.Scan(&userUUID, &totalPoints, &googleName, &telegramUsername, &telegramFirstName, &telegramLastName); err != nil {
			return nil, fmt.Errorf("failed to scan global rating entry: %w", err)
		}

		displayName := ""
		if googleName.Valid && googleName.String != "" {
			displayName = googleName.String
		} else if telegramUsername.Valid && telegramUsername.String != "" {
			displayName = telegramUsername.String
		} else {
			first := ""
			last := ""
			if telegramFirstName.Valid {
				first = telegramFirstName.String
			}
			if telegramLastName.Valid {
				last = telegramLastName.String
			}
			combined := strings.TrimSpace(strings.TrimSpace(first) + " " + strings.TrimSpace(last))
			if combined != "" {
				displayName = combined
			}
		}

		if displayName == "" {
			displayName = "Unknown"
		}
		entry.UserName = displayName
		entry.Value = totalPoints
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

func (r *PostgresRatingRepository) AddPoints(ctx context.Context, userUUID string, points int64, gotPrizeID *int, betID *int, description string) error {
	if points == 0 {
		return nil // Don't add zero points
	}

	query := `
		INSERT INTO rating (user_uuid, points, got_prize_id, bet_id, description, created_at)
		VALUES ($1, $2, $3, $4, $5, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
	`

	_, err := r.pool.Exec(ctx, query, userUUID, points, gotPrizeID, betID, description)
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

func (r *PostgresRatingRepository) GetUserBetPointsInRange(ctx context.Context, userUUID string, startMs, endMs int64) (int64, error) {
	query := `
		SELECT COALESCE(SUM(points), 0)::BIGINT
		FROM rating
		WHERE user_uuid = $1
		  AND bet_id IS NOT NULL
		  AND got_prize_id IS NULL
		  AND created_at >= $2
		  AND created_at < $3
	`

	var total int64
	if err := r.pool.QueryRow(ctx, query, userUUID, startMs, endMs).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to get bet points in range: %w", err)
	}

	return total, nil
}

func (r *PostgresRatingRepository) GetBetPointsLeaderboard(ctx context.Context, startMs, endMs int64, limit int) ([]domain.BetPrizeLeaderboardEntry, error) {
	query := `
		SELECT
			user_uuid::text AS user_uuid,
			COALESCE(SUM(points), 0)::BIGINT AS net_points
		FROM rating
		WHERE bet_id IS NOT NULL
		  AND got_prize_id IS NULL
		  AND created_at >= $1
		  AND created_at < $2
		GROUP BY user_uuid
		ORDER BY net_points DESC, user_uuid ASC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, startMs, endMs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get bet points leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []domain.BetPrizeLeaderboardEntry
	for rows.Next() {
		var entry domain.BetPrizeLeaderboardEntry
		if err := rows.Scan(&entry.UserUUID, &entry.NetPoints); err != nil {
			return nil, fmt.Errorf("failed to scan bet points leaderboard: %w", err)
		}
		entry.WinCount = 0
		entry.LossCount = 0
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bet points leaderboard rows: %w", err)
	}

	return entries, nil
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

func (r *InMemoryRatingRepository) AddPoints(ctx context.Context, userUUID string, points int64, gotPrizeID *int, betID *int, description string) error {
	return nil
}

func (r *InMemoryRatingRepository) GetMaxCreatedAt(ctx context.Context, userUUID string) (*int64, error) {
	return nil, nil
}

func (r *InMemoryRatingRepository) GetUserBetPointsInRange(ctx context.Context, userUUID string, startMs, endMs int64) (int64, error) {
	return 0, nil
}

func (r *InMemoryRatingRepository) GetBetPointsLeaderboard(ctx context.Context, startMs, endMs int64, limit int) ([]domain.BetPrizeLeaderboardEntry, error) {
	return []domain.BetPrizeLeaderboardEntry{}, nil
}
