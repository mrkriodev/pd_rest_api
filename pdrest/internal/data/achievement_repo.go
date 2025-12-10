package data

import (
	"context"
	"fmt"
	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AchievementRepository provides access to achievement data.
type AchievementRepository interface {
	GetAllAchievements(ctx context.Context) ([]domain.Achievement, error)
	GetUserAchievements(ctx context.Context, userUUID string) ([]domain.Achievement, error)
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
		SELECT id, badge, title, image_url, desc_text, tags, summ, steps, step_desc
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
			&achievement.Summ,
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

func (r *PostgresAchievementRepository) GetUserAchievements(ctx context.Context, userUUID string) ([]domain.Achievement, error) {
	query := `
		SELECT a.id, a.badge, a.title, a.image_url, a.desc_text, a.tags, a.summ, a.steps, a.step_desc
		FROM achievements a
		INNER JOIN user_achievements ua ON a.id = ua.achievement_id
		WHERE ua.user_uuid = $1
		ORDER BY ua.earned_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user achievements: %w", err)
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
			&achievement.Summ,
			&achievement.Steps,
			&achievement.StepDesc,
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

// InMemoryAchievementRepository returns empty results (used when DB is unavailable).
type InMemoryAchievementRepository struct{}

func NewInMemoryAchievementRepository() *InMemoryAchievementRepository {
	return &InMemoryAchievementRepository{}
}

func (r *InMemoryAchievementRepository) GetAllAchievements(ctx context.Context) ([]domain.Achievement, error) {
	return []domain.Achievement{}, nil
}

func (r *InMemoryAchievementRepository) GetUserAchievements(ctx context.Context, userUUID string) ([]domain.Achievement, error) {
	return []domain.Achievement{}, nil
}
