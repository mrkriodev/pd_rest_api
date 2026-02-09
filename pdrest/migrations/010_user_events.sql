-- Create user_events table
-- This table stores user participation in events

CREATE TABLE IF NOT EXISTS user_events (
    id SERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,
    event_id VARCHAR(50) NOT NULL,
    status TEXT NOT NULL DEFAULT 'joined',
    has_prise_status BOOL DEFAULT NULL,
    prize_value_id INTEGER DEFAULT NULL,
    prize_taken_status BOOL DEFAULT FALSE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    -- Foreign keys
    CONSTRAINT fk_user_events_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_user_events_event FOREIGN KEY (event_id) REFERENCES all_events(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_events_prize_value FOREIGN KEY (prize_value_id) REFERENCES prize_values(id) ON DELETE SET NULL,

    -- Unique constraint: a user can only join an event once
    CONSTRAINT uq_user_events UNIQUE (user_uuid, event_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_events_user_uuid ON user_events(user_uuid);
-- CREATE INDEX IF NOT EXISTS idx_user_events_event_id ON user_events(event_id);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_user_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER trigger_update_user_events_updated_at
BEFORE UPDATE ON user_events
FOR EACH ROW
EXECUTE FUNCTION update_user_events_updated_at();

