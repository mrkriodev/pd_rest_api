-- Create got_prizes table
-- This table stores prizes already received by users via roulette or events

CREATE TABLE IF NOT EXISTS got_prizes (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(50),                    -- Reference to event (NULL for on_start roulette)
    user_uuid UUID NOT NULL,                 -- Reference to users table
    preauth_token_id INTEGER,                -- Reference to roulette_preauth_token (optional)
    roulette_id INTEGER,                     -- Reference to roulette table (optional, for tracking)
    prize_value_id INTEGER NOT NULL,                  -- Reference to prize_values table (optional)
    prize_value TEXT,               -- Prize description/value
    prize_type VARCHAR(50) NOT NULL,         -- Type: 'roulette_on_start', 'roulette_during_event', 'event_reward', etc.
    awarded_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    -- Foreign keys
    CONSTRAINT fk_got_prizes_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_got_prizes_preauth_token FOREIGN KEY (preauth_token_id) REFERENCES roulette_preauth_token(id) ON DELETE SET NULL,
    CONSTRAINT fk_got_prizes_roulette FOREIGN KEY (roulette_id) REFERENCES roulette(id) ON DELETE SET NULL,
    CONSTRAINT fk_got_prizes_prize_value FOREIGN KEY (prize_value_id) REFERENCES prize_values(id) ON DELETE SET NULL
);

-- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_got_prizes_event_id ON got_prizes(event_id);
-- CREATE INDEX IF NOT EXISTS idx_got_prizes_user_uuid ON got_prizes(user_uuid) WHERE user_uuid IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_got_prizes_preauth_token_id ON got_prizes(preauth_token_id) WHERE preauth_token_id IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_got_prizes_roulette_id ON got_prizes(roulette_id) WHERE roulette_id IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_got_prizes_prize_type ON got_prizes(prize_type);
-- CREATE INDEX IF NOT EXISTS idx_got_prizes_awarded_at ON got_prizes(awarded_at);
-- CREATE INDEX IF NOT EXISTS idx_got_prizes_prize_value_id ON got_prizes(prize_value_id);

-- Comments for documentation
COMMENT ON TABLE got_prizes IS 'Stores prizes already received by users via roulette or events';
COMMENT ON COLUMN got_prizes.event_id IS 'Event ID if prize is event-related (NULL for on_start roulette)';
COMMENT ON COLUMN got_prizes.user_uuid IS 'User UUID of the prize owner';
COMMENT ON COLUMN got_prizes.preauth_token_id IS 'Preauth token ID if prize was awarded preauth';
COMMENT ON COLUMN got_prizes.roulette_id IS 'Reference to roulette table for tracking which roulette session awarded the prize';
COMMENT ON COLUMN got_prizes.prize_value_id IS 'Reference to prize_values table for the prize value won';
COMMENT ON COLUMN got_prizes.prize_value IS 'Prize description/value (e.g., "100 USDT", "100 points")';
COMMENT ON COLUMN got_prizes.prize_type IS 'Type of prize: roulette_on_start, roulette_during_event, event_reward, etc.';
COMMENT ON COLUMN got_prizes.awarded_at IS 'Timestamp when prize was awarded (milliseconds)';

