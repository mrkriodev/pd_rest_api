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
	GetBetPrizeLeaderboard(ctx context.Context, eventID string, startMs, endMs int64, limit int) ([]domain.BetPrizeLeaderboardEntry, error)
	GetUserBetNetPoints(ctx context.Context, userUUID string, startMs, endMs int64) (int64, error)
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
		INSERT INTO got_prizes (event_id, user_uuid, prize_value_id, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	// user_uuid is now mandatory
	if prize.UserID == nil {
		return fmt.Errorf("user_uuid is required")
	}
	userUUID := *prize.UserID

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
		prize.PrizeValueID,
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
		SELECT id, event_id, user_uuid, prize_value_id, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at
		FROM got_prizes
		WHERE id = $1
	`

	var prize domain.Prize
	var eventID *string
	var userID string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&prize.ID,
		&eventID,
		&userID,
		&prize.PrizeValueID,
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
	prize.UserID = &userID

	return &prize, nil
}

func (r *PostgresPrizeRepository) GetPrizesByUserID(ctx context.Context, userID string) ([]domain.Prize, error) {
	query := `
		SELECT id, event_id, user_uuid, prize_value_id, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at
		FROM got_prizes
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
		var userIDStr string

		if err := rows.Scan(
			&prize.ID,
			&eventID,
			&userIDStr,
			&prize.PrizeValueID,
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
		prize.UserID = &userIDStr
		prizes = append(prizes, prize)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prizes: %w", err)
	}

	return prizes, nil
}

func (r *PostgresPrizeRepository) GetPrizesByPreauthTokenID(ctx context.Context, preauthTokenID int) ([]domain.Prize, error) {
	query := `
		SELECT id, event_id, user_uuid, prize_value_id, preauth_token_id, roulette_id, prize_value, prize_type, awarded_at, created_at
		FROM got_prizes
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
		var userID string

		if err := rows.Scan(
			&prize.ID,
			&eventID,
			&userID,
			&prize.PrizeValueID,
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
		prize.UserID = &userID
		prizes = append(prizes, prize)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prizes: %w", err)
	}

	return prizes, nil
}

func (r *PostgresPrizeRepository) GetBetPrizeLeaderboard(ctx context.Context, eventID string, startMs, endMs int64, limit int) ([]domain.BetPrizeLeaderboardEntry, error) {
	if eventID == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if limit <= 0 {
		limit = 3
	}

	query := `
		SELECT gp.user_uuid::text AS user_uuid,
		       COUNT(*) FILTER (WHERE gp.prize_type ILIKE 'bet%win%') AS win_count,
		       COUNT(*) FILTER (WHERE gp.prize_type ILIKE 'bet%loss%') AS loss_count,
		       COALESCE(SUM(
		           CASE
		               WHEN gp.prize_type ILIKE 'bet%loss%' THEN -pv.value
		               ELSE pv.value
		           END
		       ), 0)::BIGINT AS net_points
		FROM got_prizes gp
		JOIN prize_values pv ON pv.id = gp.prize_value_id
		JOIN user_events ue ON ue.user_uuid = gp.user_uuid AND ue.event_id = $1
		WHERE gp.awarded_at >= $2
		  AND gp.awarded_at < $3
		  AND gp.prize_type ILIKE 'bet%'
		GROUP BY gp.user_uuid
		ORDER BY net_points DESC, win_count DESC, user_uuid ASC
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, eventID, startMs, endMs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get bet prize leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []domain.BetPrizeLeaderboardEntry
	for rows.Next() {
		var entry domain.BetPrizeLeaderboardEntry
		if err := rows.Scan(&entry.UserUUID, &entry.WinCount, &entry.LossCount, &entry.NetPoints); err != nil {
			return nil, fmt.Errorf("failed to scan bet prize leaderboard entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bet prize leaderboard rows: %w", err)
	}

	return entries, nil
}

func (r *PostgresPrizeRepository) GetUserBetNetPoints(ctx context.Context, userUUID string, startMs, endMs int64) (int64, error) {
	if userUUID == "" {
		return 0, fmt.Errorf("user_uuid is required")
	}

	query := `
		SELECT COALESCE(SUM(
			CASE
				WHEN gp.prize_type ILIKE 'bet%loss%' THEN -pv.value
				ELSE pv.value
			END
		), 0)::BIGINT AS net_points
		FROM got_prizes gp
		JOIN prize_values pv ON pv.id = gp.prize_value_id
		WHERE gp.user_uuid = $1
		  AND gp.awarded_at >= $2
		  AND gp.awarded_at < $3
		  AND gp.prize_type ILIKE 'bet%'
	`

	var netPoints int64
	if err := r.pool.QueryRow(ctx, query, userUUID, startMs, endMs).Scan(&netPoints); err != nil {
		return 0, fmt.Errorf("failed to get user bet net points: %w", err)
	}

	return netPoints, nil
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

func (r *InMemoryPrizeRepository) GetBetPrizeLeaderboard(ctx context.Context, eventID string, startMs, endMs int64, limit int) ([]domain.BetPrizeLeaderboardEntry, error) {
	return []domain.BetPrizeLeaderboardEntry{}, nil
}

func (r *InMemoryPrizeRepository) GetUserBetNetPoints(ctx context.Context, userUUID string, startMs, endMs int64) (int64, error) {
	return 0, nil
}
