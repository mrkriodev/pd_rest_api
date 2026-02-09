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
	GetAllEvents(ctx context.Context, tag string) ([]domain.Event, error)
	GetEventByID(ctx context.Context, id string) (*domain.Event, error)
	CreateEvent(ctx context.Context, event *domain.Event) error
	UpdateEvent(ctx context.Context, event *domain.Event) error
	DeleteEvent(ctx context.Context, id string) error
	AddUserEvent(ctx context.Context, userUUID string, eventID string, status string) (bool, error)
	GetUserEventsWithAvailable(ctx context.Context, userUUID string, tag string, nowMs int64) ([]domain.UserEventEntry, error)
	GetUserEventPrizeStatus(ctx context.Context, userUUID string, eventID string) (*bool, *int, *bool, error)
	UpdateUserEventPrizeStatusIfUnknown(ctx context.Context, userUUID string, eventID string, hasPrise *bool, prizeValueID *int) (bool, error)
	UpdateUserEventPrizeTakenStatusIfNotTaken(ctx context.Context, userUUID string, eventID string, taken bool) (bool, error)
}

type PostgresEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEventRepository(pool *pgxpool.Pool) *PostgresEventRepository {
	return &PostgresEventRepository{pool: pool}
}

// GetAllEvents retrieves all events from the database, optionally filtered by tag
func (r *PostgresEventRepository) GetAllEvents(ctx context.Context, tag string) ([]domain.Event, error) {
	query := `
		SELECT id, badge, title, desc_text, deadline, tags, reward, info
		FROM all_events
		WHERE ($1 = '' OR tags ILIKE '%' || $1 || '%')
		ORDER BY deadline ASC
	`

	rows, err := r.pool.Query(ctx, query, tag)
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

func (r *PostgresEventRepository) AddUserEvent(ctx context.Context, userUUID string, eventID string, status string) (bool, error) {
	query := `
		INSERT INTO user_events (user_uuid, event_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
		ON CONFLICT (user_uuid, event_id) DO NOTHING
	`

	result, err := r.pool.Exec(ctx, query, userUUID, eventID, status)
	if err != nil {
		return false, fmt.Errorf("failed to insert user event: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (r *PostgresEventRepository) GetUserEventsWithAvailable(ctx context.Context, userUUID string, tag string, nowMs int64) ([]domain.UserEventEntry, error) {
	query := `
		WITH user_events_cte AS (
			SELECT e.id, e.badge, e.title, e.desc_text, e.deadline, e.tags, e.reward, e.info,
			       ue.status, ue.created_at AS joined_at, ue.has_prise_status, ue.prize_taken_status
			FROM user_events ue
			JOIN all_events e ON e.id = ue.event_id
			WHERE ue.user_uuid = $1
		),
		available_events AS (
			SELECT e.id, e.badge, e.title, e.desc_text, e.deadline, e.tags, e.reward, e.info,
			       'available' AS status, NULL::BIGINT AS joined_at, NULL::BOOL AS has_prise_status, FALSE AS prize_taken_status
			FROM all_events e
			LEFT JOIN user_events ue ON ue.event_id = e.id AND ue.user_uuid = $1
			WHERE ue.event_id IS NULL
			  AND e.tags ILIKE '%' || $2 || '%'
			  AND e.deadline > $3
		)
		SELECT * FROM user_events_cte
		UNION ALL
		SELECT * FROM available_events
		ORDER BY deadline ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, userUUID, tag, nowMs)
	if err != nil {
		return nil, fmt.Errorf("failed to query user events: %w", err)
	}
	defer rows.Close()

	var events []domain.UserEventEntry
	for rows.Next() {
		var event domain.UserEventEntry
		var rewardJSON []byte
		var deadlineMs int64
		var joinedAtMs *int64

		if err := rows.Scan(
			&event.ID,
			&event.Badge,
			&event.Title,
			&event.Desc,
			&deadlineMs,
			&event.Tags,
			&rewardJSON,
			&event.Info,
			&event.Status,
			&joinedAtMs,
			&event.HasPrise,
			&event.PrizeTakenStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user event: %w", err)
		}

		event.Deadline = time.Unix(0, deadlineMs*int64(time.Millisecond)).UTC()
		if joinedAtMs != nil {
			joinedAt := time.Unix(0, (*joinedAtMs)*int64(time.Millisecond)).UTC()
			event.JoinedAt = &joinedAt
		}

		if err := json.Unmarshal(rewardJSON, &event.Reward); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reward: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user events: %w", err)
	}

	return events, nil
}

func (r *PostgresEventRepository) GetUserEventPrizeStatus(ctx context.Context, userUUID string, eventID string) (*bool, *int, *bool, error) {
	query := `
		SELECT has_prise_status, prize_value_id, prize_taken_status
		FROM user_events
		WHERE user_uuid = $1 AND event_id = $2
	`

	var hasPrise *bool
	var prizeValueID *int
	var prizeTaken *bool
	if err := r.pool.QueryRow(ctx, query, userUUID, eventID).Scan(&hasPrise, &prizeValueID, &prizeTaken); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil, fmt.Errorf("user event not found")
		}
		return nil, nil, nil, fmt.Errorf("failed to get user event prize status: %w", err)
	}

	return hasPrise, prizeValueID, prizeTaken, nil
}

func (r *PostgresEventRepository) UpdateUserEventPrizeStatusIfUnknown(ctx context.Context, userUUID string, eventID string, hasPrise *bool, prizeValueID *int) (bool, error) {
	query := `
		UPDATE user_events
		SET has_prise_status = $1,
		    prize_value_id = $2,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
		WHERE user_uuid = $3 AND event_id = $4
		  AND has_prise_status IS NULL
	`

	result, err := r.pool.Exec(ctx, query, hasPrise, prizeValueID, userUUID, eventID)
	if err != nil {
		return false, fmt.Errorf("failed to update user event prize status: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (r *PostgresEventRepository) UpdateUserEventPrizeTakenStatusIfNotTaken(ctx context.Context, userUUID string, eventID string, taken bool) (bool, error) {
	query := `
		UPDATE user_events
		SET prize_taken_status = $1,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
		WHERE user_uuid = $2 AND event_id = $3
		  AND prize_taken_status = FALSE
	`

	result, err := r.pool.Exec(ctx, query, taken, userUUID, eventID)
	if err != nil {
		return false, fmt.Errorf("failed to update user event prize taken status: %w", err)
	}

	return result.RowsAffected() > 0, nil
}
