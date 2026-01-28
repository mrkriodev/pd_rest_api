-- Create prize_values table
-- This table stores available prize values for each event
-- Each prize_value is linked to an event via foreign key

CREATE TABLE IF NOT EXISTS prize_values (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(50) NOT NULL,           -- Foreign key to all_events
    value BIGINT NOT NULL,                   -- Prize value in points (1 USDT = 1 point)
    label TEXT NOT NULL,                      -- Display label (e.g., "100 USDT")
    segment_id VARCHAR(50),                  -- Optional segment ID for roulette wheel
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    -- Foreign key
    CONSTRAINT fk_prize_values_event FOREIGN KEY (event_id) REFERENCES all_events(id) ON DELETE CASCADE
);

-- Create indexes
-- CREATE INDEX IF NOT EXISTS idx_prize_values_event_id ON prize_values(event_id);

-- Comments
COMMENT ON TABLE prize_values IS 'Stores available prize values for each event';
COMMENT ON COLUMN prize_values.event_id IS 'Reference to event in all_events table';
COMMENT ON COLUMN prize_values.value IS 'Prize value in points (1 USDT = 1 point)';
COMMENT ON COLUMN prize_values.label IS 'Display label for the prize';
COMMENT ON COLUMN prize_values.segment_id IS 'Optional segment ID for roulette wheel visualization';

-- Insert prize values for startup event
INSERT INTO prize_values (event_id, value, label, segment_id, created_at, updated_at)
VALUES 
    ('startup', 100, '100 USDT', '1', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
    ('startup', 50, '50 USDT', '2', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
    ('startup', 10, '10 USDT', '3', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
ON CONFLICT DO NOTHING;

