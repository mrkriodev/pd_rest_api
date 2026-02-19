package data

import (
	"context"
	"database/sql"
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
	GetUserAchievementByID(ctx context.Context, userUUID string, achievementID string) (*domain.UserAchievementEntry, error)
	GetAchievementByID(ctx context.Context, achievementID string) (*domain.Achievement, error)
	GetAchievementByPrizeID(ctx context.Context, prizeID int) (*domain.Achievement, error)
	GetUserAchievementStatus(ctx context.Context, userUUID string, achievementID string) (*domain.UserAchievementStatus, error)
	AddUserAchievement(ctx context.Context, userUUID string, achievementID string, stepsGot int, needSteps int) (bool, error)
	UpdateUserAchievementClaimStatus(ctx context.Context, userUUID string, achievementID string, claimed bool) error
	UpdateUserAchievementNeedSteps(ctx context.Context, userUUID string, achievementID string, needSteps int) error
	UpsertUserAchievementProgress(ctx context.Context, userUUID string, achievementID string, stepsGot int, needSteps int, claimed bool) error
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
		       a.tags,
		       CASE WHEN a.prize_id IS NULL THEN NULL ELSE pv.label END AS prize_desc,
		       a.steps, a.step_desc,
		       CASE WHEN ua.achievement_id IS NULL THEN NULL ELSE ua.steps_got END AS steps_got,
		       CASE WHEN ua.achievement_id IS NULL THEN NULL ELSE ua.need_steps END AS need_steps,
		       COALESCE(ua.claimed_status, FALSE) AS claimed_status
		FROM achievements a
		LEFT JOIN user_achievements ua
		       ON a.id = ua.achievement_id AND ua.user_uuid = $1
		LEFT JOIN prize_values pv
		       ON pv.id = a.prize_id
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
		var prizeDesc sql.NullString
		var stepsGot sql.NullInt32
		var needSteps sql.NullInt32
		if err := rows.Scan(
			&achievement.ID,
			&achievement.Badge,
			&achievement.Title,
			&achievement.ImageURL,
			&achievement.Desc,
			&achievement.Tags,
			&prizeDesc,
			&achievement.Steps,
			&achievement.StepDesc,
			&stepsGot,
			&needSteps,
			&achievement.ClaimedStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user achievement: %w", err)
		}
		if prizeDesc.Valid {
			value := prizeDesc.String
			achievement.PrizeDesc = &value
		}
		if stepsGot.Valid {
			value := int(stepsGot.Int32)
			achievement.StepsGot = &value
		}
		if needSteps.Valid {
			value := int(needSteps.Int32)
			achievement.NeedSteps = &value
		}
		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user achievements: %w", err)
	}

	return achievements, nil
}

func (r *PostgresAchievementRepository) GetUserAchievementByID(ctx context.Context, userUUID string, achievementID string) (*domain.UserAchievementEntry, error) {
	query := `
		SELECT a.id, a.badge, a.title, a.image_url,
		       CASE WHEN ua.achievement_id IS NULL THEN '' ELSE a.desc_text END AS desc_text,
		       a.tags,
		       CASE WHEN a.prize_id IS NULL THEN NULL ELSE pv.label END AS prize_desc,
		       a.steps, a.step_desc,
		       CASE WHEN ua.achievement_id IS NULL THEN NULL ELSE ua.steps_got END AS steps_got,
		       CASE WHEN ua.achievement_id IS NULL THEN NULL ELSE ua.need_steps END AS need_steps,
		       COALESCE(ua.claimed_status, FALSE) AS claimed_status
		FROM achievements a
		LEFT JOIN user_achievements ua
		       ON a.id = ua.achievement_id AND ua.user_uuid = $1
		LEFT JOIN prize_values pv
		       ON pv.id = a.prize_id
		WHERE a.id = $2
		LIMIT 1
	`

	var achievement domain.UserAchievementEntry
	var prizeDesc sql.NullString
	var stepsGot sql.NullInt32
	var needSteps sql.NullInt32
	if err := r.pool.QueryRow(ctx, query, userUUID, achievementID).Scan(
		&achievement.ID,
		&achievement.Badge,
		&achievement.Title,
		&achievement.ImageURL,
		&achievement.Desc,
		&achievement.Tags,
		&prizeDesc,
		&achievement.Steps,
		&achievement.StepDesc,
		&stepsGot,
		&needSteps,
		&achievement.ClaimedStatus,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("achievement not found")
		}
		return nil, fmt.Errorf("failed to get user achievement: %w", err)
	}

	if prizeDesc.Valid {
		value := prizeDesc.String
		achievement.PrizeDesc = &value
	}
	if stepsGot.Valid {
		value := int(stepsGot.Int32)
		achievement.StepsGot = &value
	}
	if needSteps.Valid {
		value := int(needSteps.Int32)
		achievement.NeedSteps = &value
	}

	return &achievement, nil
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

func (r *PostgresAchievementRepository) GetAchievementByPrizeID(ctx context.Context, prizeID int) (*domain.Achievement, error) {
	query := `
		SELECT id, badge, title, image_url, desc_text, tags, prize_id, steps, step_desc
		FROM achievements
		WHERE prize_id = $1
		LIMIT 1
	`

	var achievement domain.Achievement
	if err := r.pool.QueryRow(ctx, query, prizeID).Scan(
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
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get achievement by prize id: %w", err)
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

func (r *PostgresAchievementRepository) UpsertUserAchievementProgress(ctx context.Context, userUUID string, achievementID string, stepsGot int, needSteps int, claimed bool) error {
	query := `
		INSERT INTO user_achievements (user_uuid, achievement_id, steps_got, need_steps, claimed_status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_uuid, achievement_id)
		DO UPDATE SET steps_got = EXCLUDED.steps_got,
			need_steps = EXCLUDED.need_steps,
			claimed_status = EXCLUDED.claimed_status
	`

	_, err := r.pool.Exec(ctx, query, userUUID, achievementID, stepsGot, needSteps, claimed)
	if err != nil {
		return fmt.Errorf("failed to upsert user achievement progress: %w", err)
	}
	return nil
}
