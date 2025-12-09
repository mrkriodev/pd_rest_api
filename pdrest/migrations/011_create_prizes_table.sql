-- Create prizes table
-- This table stores prizes awarded to users via roulette or events

CREATE TABLE IF NOT EXISTS prizes (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(50),                    -- Reference to event (NULL for on_start roulette)
    user_uuid UUID,                          -- Reference to users table (NULL if awarded by preauth_token)
    preauth_token_id INTEGER,                -- Reference to roulette_preauth_token (NULL if awarded by user_uuid)
    roulette_id INTEGER,                     -- Reference to roulette table (optional, for tracking)
    prize_value TEXT NOT NULL,               -- Prize description/value
    prize_type VARCHAR(50) NOT NULL,         -- Type: 'roulette_on_start', 'roulette_during_event', 'event_reward', etc.
    awarded_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    
    -- Foreign keys
    CONSTRAINT fk_prizes_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_prizes_preauth_token FOREIGN KEY (preauth_token_id) REFERENCES roulette_preauth_token(id) ON DELETE SET NULL,
    CONSTRAINT fk_prizes_roulette FOREIGN KEY (roulette_id) REFERENCES roulette(id) ON DELETE SET NULL,
    
    -- Constraints: Either user_uuid or preauth_token_id must be provided
    CONSTRAINT chk_prize_recipient CHECK (
        (user_uuid IS NOT NULL) OR (preauth_token_id IS NOT NULL)
    )
);

-- -- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_prizes_event_id ON prizes(event_id);
-- CREATE INDEX IF NOT EXISTS idx_prizes_user_uuid ON prizes(user_uuid) WHERE user_uuid IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_prizes_preauth_token_id ON prizes(preauth_token_id) WHERE preauth_token_id IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_prizes_roulette_id ON prizes(roulette_id) WHERE roulette_id IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_prizes_prize_type ON prizes(prize_type);
-- CREATE INDEX IF NOT EXISTS idx_prizes_awarded_at ON prizes(awarded_at);

-- -- Comments for documentation
-- COMMENT ON TABLE prizes IS 'Stores prizes awarded to users via roulette or events';
-- COMMENT ON COLUMN prizes.event_id IS 'Event ID if prize is event-related (NULL for on_start roulette)';
-- COMMENT ON COLUMN prizes.user_uuid IS 'User UUID if prize is awarded to authenticated user (NULL if by preauth_token)';
-- COMMENT ON COLUMN prizes.preauth_token_id IS 'Preauth token ID if prize is awarded to unauthenticated user (NULL if by user_uuid)';
-- COMMENT ON COLUMN prizes.roulette_id IS 'Reference to roulette table for tracking which roulette session awarded the prize';
-- COMMENT ON COLUMN prizes.prize_value IS 'Prize description/value (e.g., "0.0001 ETH", "100 points")';
-- COMMENT ON COLUMN prizes.prize_type IS 'Type of prize: roulette_on_start, roulette_during_event, event_reward, etc.';
-- COMMENT ON COLUMN prizes.awarded_at IS 'Timestamp when prize was awarded (milliseconds)';

