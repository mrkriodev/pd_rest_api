package data

import (
	"context"
	"errors"
	"fmt"
	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AchievementRepository provides access to achievement data.
type AchievementRepository interface {
	GetAllAchievements(ctx context.Context) ([]domain.Achievement, error)
	GetUserAchievements(ctx context.Context, userUUID string) ([]domain.UserAchievementEntry, error)
	GetAchievementByID(ctx context.Context, achievementID string) (*domain.Achievement, error)
	GetUserAchievementStatus(ctx context.Context, userUUID string, achievementID string) (*domain.UserAchievementStatus, error)
	AddUserAchievement(ctx context.Context, userUUID string, achievementID string, stepsGot int, needSteps int) (bool, error)
	UpdateUserAchievementClaimStatus(ctx context.Context, userUUID string, achievementID string, claimed bool) error
	UpdateUserAchievementNeedSteps(ctx context.Context, userUUID string, achievementID string, needSteps int) error
}

// PostgresAchievementRepository implements AchievementRepository with PostgreSQL.
type PostgresAchievementRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAchievementRepository(pool *pgxpool.Pool) *PostgresAchievementRepository {
	return &PostgresAchievementRepository{pool: pool}
}

func (r *PostgresAchievementRepository) GetAllAchievements(ctx context.Context) ([]domain.Achievement, error) {
	query := `
		SELECT id, badge, title, image_url, desc_text, tags, prize_id, steps, step_desc
		FROM achievements
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}
	defer rows.Close()

	var achievements []domain.Achievement
	for rows.Next() {
		var achievement domain.Achievement
		if err := rows.Scan(
			&achievement.ID,
			&achievement.Badge,
			&achievement.Title,
			&achievement.ImageURL,
			&achievement.Desc,
			&achievement.Tags,
			&achievement.PrizeID,
			&achievement.Steps,
			&achievement.StepDesc,
		); err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating achievements: %w", err)
	}

	return achievements, nil
}

func (r *PostgresAchievementRepository) GetUserAchievements(ctx context.Context, userUUID string) ([]domain.UserAchievementEntry, error) {
	query := `
		SELECT a.id, a.badge, a.title, a.image_url,
		       CASE WHEN ua.achievement_id IS NULL THEN '' ELSE a.desc_text END AS desc_text,
		       a.tags, a.prize_id, a.steps, a.step_desc,
		       COALESCE(ua.steps_got, 0) AS steps_got,
		       COALESCE(ua.need_steps, 0) AS need_steps,
		       COALESCE(ua.claimed_status, FALSE) AS claimed_status
		FROM achievements a
		LEFT JOIN user_achievements ua
		       ON a.id = ua.achievement_id AND ua.user_uuid = $1
		ORDER BY
			CASE
				WHEN ua.achievement_id IS NOT NULL
					AND COALESCE(ua.claimed_status, FALSE) = FALSE
					AND COALESCE(NULLIF(ua.need_steps, 0), a.steps) > 0
					AND COALESCE(ua.steps_got, 0) >= COALESCE(NULLIF(ua.need_steps, 0), a.steps)
				THEN 0
				WHEN ua.achievement_id IS NOT NULL
					AND COALESCE(NULLIF(ua.need_steps, 0), a.steps) > 0
					AND COALESCE(ua.steps_got, 0) < COALESCE(NULLIF(ua.need_steps, 0), a.steps)
				THEN 1
				ELSE 2
			END,
			a.id ASC
	`

	rows, err := r.pool.Query(ctx, query, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user achievements: %w", err)
	}
	defer rows.Close()

	var achievements []domain.UserAchievementEntry
	for rows.Next() {
		var achievement domain.UserAchievementEntry
		if err := rows.Scan(
			&achievement.ID,
			&achievement.Badge,
			&achievement.Title,
			&achievement.ImageURL,
			&achievement.Desc,
			&achievement.Tags,
			&achievement.PrizeID,
			&achievement.Steps,
			&achievement.StepDesc,
			&achievement.StepsGot,
			&achievement.NeedSteps,
			&achievement.ClaimedStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user achievement: %w", err)
		}
		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user achievements: %w", err)
	}

	return achievements, nil
}

func (r *PostgresAchievementRepository) GetAchievementByID(ctx context.Context, achievementID string) (*domain.Achievement, error) {
	query := `
		SELECT id, badge, title, image_url, desc_text, tags, prize_id, steps, step_desc
		FROM achievements
		WHERE id = $1
	`

	var achievement domain.Achievement
	if err := r.pool.QueryRow(ctx, query, achievementID).Scan(
		&achievement.ID,
		&achievement.Badge,
		&achievement.Title,
		&achievement.ImageURL,
		&achievement.Desc,
		&achievement.Tags,
		&achievement.PrizeID,
		&achievement.Steps,
		&achievement.StepDesc,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("achievement not found")
		}
		return nil, fmt.Errorf("failed to get achievement: %w", err)
	}

	return &achievement, nil
}

func (r *PostgresAchievementRepository) GetUserAchievementStatus(ctx context.Context, userUUID string, achievementID string) (*domain.UserAchievementStatus, error) {
	query := `
		SELECT user_uuid::text, achievement_id, steps_got, need_steps, claimed_status
		FROM user_achievements
		WHERE user_uuid = $1 AND achievement_id = $2
	`

	var status domain.UserAchievementStatus
	if err := r.pool.QueryRow(ctx, query, userUUID, achievementID).Scan(
		&status.UserID,
		&status.AchievementID,
		&status.StepsGot,
		&status.NeedSteps,
		&status.ClaimedStatus,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user achievement not found")
		}
		return nil, fmt.Errorf("failed to get user achievement: %w", err)
	}

	return &status, nil
}

func (r *PostgresAchievementRepository) AddUserAchievement(ctx context.Context, userUUID string, achievementID string, stepsGot int, needSteps int) (bool, error) {
	query := `
		INSERT INTO user_achievements (user_uuid, achievement_id, steps_got, need_steps, claimed_status)
		VALUES ($1, $2, $3, $4, FALSE)
		ON CONFLICT (user_uuid, achievement_id) DO NOTHING
	`

	result, err := r.pool.Exec(ctx, query, userUUID, achievementID, stepsGot, needSteps)
	if err != nil {
		return false, fmt.Errorf("failed to insert user achievement: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (r *PostgresAchievementRepository) UpdateUserAchievementClaimStatus(ctx context.Context, userUUID string, achievementID string, claimed bool) error {
	query := `
		UPDATE user_achievements
		SET claimed_status = $1
		WHERE user_uuid = $2 AND achievement_id = $3
	`

	_, err := r.pool.Exec(ctx, query, claimed, userUUID, achievementID)
	if err != nil {
		return fmt.Errorf("failed to update claimed_status: %w", err)
	}
	return nil
}

func (r *PostgresAchievementRepository) UpdateUserAchievementNeedSteps(ctx context.Context, userUUID string, achievementID string, needSteps int) error {
	query := `
		UPDATE user_achievements
		SET need_steps = $1
		WHERE user_uuid = $2 AND achievement_id = $3
		  AND (need_steps IS NULL OR need_steps = 0)
	`

	_, err := r.pool.Exec(ctx, query, needSteps, userUUID, achievementID)
	if err != nil {
		return fmt.Errorf("failed to update need_steps: %w", err)
	}
	return nil
}

// InMemoryAchievementRepository returns empty results (used when DB is unavailable).
type InMemoryAchievementRepository struct{}

func NewInMemoryAchievementRepository() *InMemoryAchievementRepository {
	return &InMemoryAchievementRepository{}
}

func (r *InMemoryAchievementRepository) GetAllAchievements(ctx context.Context) ([]domain.Achievement, error) {
	return []domain.Achievement{}, nil
}

func (r *InMemoryAchievementRepository) GetUserAchievements(ctx context.Context, userUUID string) ([]domain.UserAchievementEntry, error) {
	return []domain.UserAchievementEntry{}, nil
}

func (r *InMemoryAchievementRepository) GetAchievementByID(ctx context.Context, achievementID string) (*domain.Achievement, error) {
	return nil, fmt.Errorf("achievement retrieval requires database connection")
}

func (r *InMemoryAchievementRepository) GetUserAchievementStatus(ctx context.Context, userUUID string, achievementID string) (*domain.UserAchievementStatus, error) {
	return nil, fmt.Errorf("user achievement retrieval requires database connection")
}

func (r *InMemoryAchievementRepository) AddUserAchievement(ctx context.Context, userUUID string, achievementID string, stepsGot int, needSteps int) (bool, error) {
	return false, fmt.Errorf("user achievement insert requires database connection")
}

func (r *InMemoryAchievementRepository) UpdateUserAchievementClaimStatus(ctx context.Context, userUUID string, achievementID string, claimed bool) error {
	return fmt.Errorf("user achievement update requires database connection")
}

func (r *InMemoryAchievementRepository) UpdateUserAchievementNeedSteps(ctx context.Context, userUUID string, achievementID string, needSteps int) error {
	return fmt.Errorf("user achievement update requires database connection")
}
