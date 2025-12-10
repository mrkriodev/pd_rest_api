-- Create all_events table in pumpdump_db
-- This table stores event information with JSONB for rewards

CREATE TABLE IF NOT EXISTS all_events (
    id VARCHAR(50) PRIMARY KEY,
    badge TEXT NOT NULL,
    title TEXT NOT NULL,
    desc_text TEXT NOT NULL,
    deadline BIGINT NOT NULL,
    tags VARCHAR(100),
    reward JSONB NOT NULL,
    info TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
);

-- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_all_events_id ON all_events(id);
-- CREATE INDEX IF NOT EXISTS idx_all_events_deadline ON all_events(deadline);
-- CREATE INDEX IF NOT EXISTS idx_all_events_tags ON all_events(tags);

-- Example insert (for testing)
-- Note: deadline is Unix timestamp in milliseconds (UTC)
-- Example: '2025-12-01T12:00:00Z' = 1733054400000
-- INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info)
-- VALUES (
--     'e1',
--     'Get 1000+ prizes.',
--     'Ethereum Pump or Dump',
--     'Reward for 1st place, to the one who stays in the top for 7 days',
--     1733054400000,
--     'global',
--     '[
--         {"place": "1-3", "value": "0.0006 ETH"},
--         {"place": "4-5", "value": "0.0003 ETH"},
--         {"place": "6-7", "value": "0.0003 ETH"},
--         {"place": "8", "value": "0.0003 ETH"},
--         {"place": "9-10", "value": "0.0001 ETH"}
--     ]'::jsonb,
--     'Your intuition and team spirit are what matter most. Join a squad or create your own...'
-- );

