-- Create all_events table in pumpdump_db
-- This table stores event information with JSONB for rewards (USDT integer values)

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

-- Startup event (used by roulette on_start)
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info, created_at, updated_at)
VALUES (
    'startup',
    'Startup Bonus',
    'Startup Roulette Event',
    'Spin the startup roulette and win USDT rewards.',
    EXTRACT(EPOCH FROM (NOW() + INTERVAL '365 days'))::BIGINT * 1000,
    'startup',
    '[
        {"place": "any", "value": "100 USDT"}
    ]'::jsonb,
    'Default startup roulette event',
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
)
ON CONFLICT (id) DO NOTHING;

-- Example events (for testing)
-- All timestamps are Unix milliseconds (UTC)

-- Event 1: Ethereum Pump or Dump
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info)
VALUES (
    'e1',
    'Get 1000+ prizes.',
    'Ethereum Pump or Dump',
    'Reward for 1st place, to the one who stays in the top for 7 days',
    1733054400000, -- 2025-12-01T12:00:00Z
    'global',
    '[
        {"place": "1-3", "value": "100 USDT"},
        {"place": "4-5", "value": "50 USDT"},
        {"place": "6-7", "value": "50 USDT"},
        {"place": "8", "value": "50 USDT"},
        {"place": "9-10", "value": "10 USDT"}
    ]'::jsonb,
    'Your intuition and team spirit are what matter most. Join a squad or create your own...'
)
ON CONFLICT (id) DO NOTHING;

-- Event 2: Bitcoin Trading Challenge
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info)
VALUES (
    'e2',
    'Win big with Bitcoin!',
    'Bitcoin Trading Challenge',
    'Compete in the ultimate Bitcoin trading competition. Show your skills and win amazing prizes!',
    1733140800000, -- 2025-12-02T12:00:00Z
    'crypto',
    '[
        {"place": "1", "value": "500 USDT"},
        {"place": "2", "value": "250 USDT"},
        {"place": "3", "value": "100 USDT"},
        {"place": "4-10", "value": "50 USDT"}
    ]'::jsonb,
    'Test your trading strategies and compete against the best traders. Daily leaderboards and weekly prizes!'
)
ON CONFLICT (id) DO NOTHING;

-- Event 3: Weekly Prediction Contest
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info)
VALUES (
    'e3',
    'Predict and win!',
    'Weekly Prediction Contest',
    'Make accurate predictions about market movements and earn rewards based on your accuracy.',
    1733241600000, -- 2025-12-03T16:00:00Z
    'prediction',
    '[
        {"place": "1-5", "value": "500 USDT"},
        {"place": "6-15", "value": "200 USDT"},
        {"place": "16-30", "value": "100 USDT"},
        {"place": "31-50", "value": "50 USDT"}
    ]'::jsonb,
    'Join our weekly prediction contest! Make predictions on various market indicators and compete for weekly prizes. The more accurate your predictions, the higher your rank!'
)
ON CONFLICT (id) DO NOTHING;

