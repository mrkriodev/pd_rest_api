-- Create bets table
-- This table stores user bets for pump/dump predictions

CREATE TABLE IF NOT EXISTS bets (
    id SERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,                -- Reference to users table
    side VARCHAR(10) NOT NULL CHECK (side IN ('pump', 'dump')),
    sum NUMERIC(18, 0) NOT NULL CHECK (sum > 0), -- Bet amount in whole USDT
    pair VARCHAR(20) NOT NULL,              -- Trading pair (e.g., 'ETH/USDT')
    timeframe INTEGER NOT NULL,             -- Timeframe in seconds
    open_price NUMERIC(18, 8) NOT NULL,     -- Opening price
    close_price NUMERIC(18, 8),            -- Closing price (NULL if bet is still open)
    open_time TIMESTAMP NOT NULL,           -- Opening time
    close_time TIMESTAMP,                  -- Closing time (NULL if bet is still open)
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    -- Foreign key
    CONSTRAINT fk_bets_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE CASCADE
);

-- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_bets_user_uuid ON bets(user_uuid);
-- CREATE INDEX IF NOT EXISTS idx_bets_pair ON bets(pair);
-- CREATE INDEX IF NOT EXISTS idx_bets_open_time ON bets(open_time);
-- CREATE INDEX IF NOT EXISTS idx_bets_side ON bets(side);
-- CREATE INDEX IF NOT EXISTS idx_bets_created_at ON bets(created_at);

-- Add comment for documentation
COMMENT ON TABLE bets IS 'Stores user bets for pump/dump predictions';
COMMENT ON COLUMN bets.side IS 'Bet side: pump or dump';
COMMENT ON COLUMN bets.sum IS 'Bet amount in whole USDT';
COMMENT ON COLUMN bets.pair IS 'Trading pair (e.g., ETH/USDT)';
COMMENT ON COLUMN bets.timeframe IS 'Timeframe in minutes';
COMMENT ON COLUMN bets.close_price IS 'NULL if bet is still open';

