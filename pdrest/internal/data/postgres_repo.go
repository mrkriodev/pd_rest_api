package data

import (
	"context"
	"fmt"

	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClientRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresClientRepository(pool *pgxpool.Pool) *PostgresClientRepository {
	return &PostgresClientRepository{pool: pool}
}

func (r *PostgresClientRepository) GetStatus(id int) (*domain.ClientStatus, error) {
	ctx := context.Background()

	var status domain.ClientStatus
	query := `SELECT id, status FROM client_status WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(&status.ID, &status.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get client status: %w", err)
	}

	return &status, nil
}

// Seed initial data (optional, for development)
func (r *PostgresClientRepository) SeedData(ctx context.Context) error {
	query := `
	INSERT INTO client_status (id, status) 
	VALUES ($1, $2)
	ON CONFLICT (id) DO NOTHING;
	`

	seedData := []struct {
		id     int
		status string
	}{
		{1, "active"},
		{2, "pending"},
		{3, "blocked"},
	}

	for _, data := range seedData {
		_, err := r.pool.Exec(ctx, query, data.id, data.status)
		if err != nil {
			return fmt.Errorf("failed to seed data: %w", err)
		}
	}

	return nil
}
