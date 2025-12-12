-- Create prize_values table
-- This table stores available prize values for each event
-- Each prize_value is linked to an event via foreign key

CREATE TABLE IF NOT EXISTS prize_values (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(50) NOT NULL,           -- Foreign key to all_events
    value TEXT NOT NULL,                     -- Prize value (e.g., "0.01 ETH", "0.005 ETH")
    label TEXT NOT NULL,                      -- Display label (e.g., "0.01 ETH")
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
COMMENT ON COLUMN prize_values.value IS 'Prize value (e.g., "0.01 ETH")';
COMMENT ON COLUMN prize_values.label IS 'Display label for the prize';
COMMENT ON COLUMN prize_values.segment_id IS 'Optional segment ID for roulette wheel visualization';

