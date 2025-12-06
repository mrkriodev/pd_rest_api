package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pdrest/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	GetAllEvents(ctx context.Context) ([]domain.Event, error)
	GetEventByID(ctx context.Context, id string) (*domain.Event, error)
	CreateEvent(ctx context.Context, event *domain.Event) error
	UpdateEvent(ctx context.Context, event *domain.Event) error
	DeleteEvent(ctx context.Context, id string) error
}

type PostgresEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEventRepository(pool *pgxpool.Pool) *PostgresEventRepository {
	return &PostgresEventRepository{pool: pool}
}

// GetAllEvents retrieves all events from the database
func (r *PostgresEventRepository) GetAllEvents(ctx context.Context) ([]domain.Event, error) {
	query := `
		SELECT id, badge, title, desc_text, deadline, tags, reward, info
		FROM all_events
		ORDER BY deadline ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		var rewardJSON []byte
		var deadlineMs int64

		err := rows.Scan(
			&event.ID,
			&event.Badge,
			&event.Title,
			&event.Desc,
			&deadlineMs,
			&event.Tags,
			&rewardJSON,
			&event.Info,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Convert Unix milliseconds (UTC) to time.Time
		event.Deadline = time.Unix(0, deadlineMs*int64(time.Millisecond)).UTC()

		// Unmarshal JSONB reward array
		if err := json.Unmarshal(rewardJSON, &event.Reward); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reward: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventByID retrieves a single event by ID
func (r *PostgresEventRepository) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	query := `
		SELECT id, badge, title, desc_text, deadline, tags, reward, info
		FROM all_events
		WHERE id = $1
	`

	var event domain.Event
	var rewardJSON []byte
	var deadlineMs int64

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Badge,
		&event.Title,
		&event.Desc,
		&deadlineMs,
		&event.Tags,
		&rewardJSON,
		&event.Info,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Convert Unix milliseconds (UTC) to time.Time
	event.Deadline = time.Unix(0, deadlineMs*int64(time.Millisecond)).UTC()

	// Unmarshal JSONB reward array
	if err := json.Unmarshal(rewardJSON, &event.Reward); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reward: %w", err)
	}

	return &event, nil
}

// CreateEvent creates a new event
func (r *PostgresEventRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	// Marshal reward array to JSON
	rewardJSON, err := json.Marshal(event.Reward)
	if err != nil {
		return fmt.Errorf("failed to marshal reward: %w", err)
	}

	// Convert time.Time to Unix milliseconds (UTC)
	deadlineMs := event.Deadline.UTC().UnixMilli()
	nowMs := time.Now().UTC().UnixMilli()

	query := `
		INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			badge = EXCLUDED.badge,
			title = EXCLUDED.title,
			desc_text = EXCLUDED.desc_text,
			deadline = EXCLUDED.deadline,
			tags = EXCLUDED.tags,
			reward = EXCLUDED.reward,
			info = EXCLUDED.info,
			updated_at = $10
	`

	_, err = r.pool.Exec(ctx, query,
		event.ID,
		event.Badge,
		event.Title,
		event.Desc,
		deadlineMs,
		event.Tags,
		rewardJSON,
		event.Info,
		nowMs,
		nowMs,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// UpdateEvent updates an existing event
func (r *PostgresEventRepository) UpdateEvent(ctx context.Context, event *domain.Event) error {
	// Marshal reward array to JSON
	rewardJSON, err := json.Marshal(event.Reward)
	if err != nil {
		return fmt.Errorf("failed to marshal reward: %w", err)
	}

	// Convert time.Time to Unix milliseconds (UTC)
	deadlineMs := event.Deadline.UTC().UnixMilli()
	nowMs := time.Now().UTC().UnixMilli()

	query := `
		UPDATE all_events
		SET badge = $2,
		    title = $3,
		    desc_text = $4,
		    deadline = $5,
		    tags = $6,
		    reward = $7,
		    info = $8,
		    updated_at = $9
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		event.ID,
		event.Badge,
		event.Title,
		event.Desc,
		deadlineMs,
		event.Tags,
		rewardJSON,
		event.Info,
		nowMs,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event with id %s not found", event.ID)
	}

	return nil
}

// DeleteEvent deletes an event by ID
func (r *PostgresEventRepository) DeleteEvent(ctx context.Context, id string) error {
	query := `DELETE FROM all_events WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event with id %s not found", id)
	}

	return nil
}
