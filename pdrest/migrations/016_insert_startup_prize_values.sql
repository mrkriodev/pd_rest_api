-- Insert three test prize values for startup roulette event

-- Ensure startup event exists (should already exist from migration 012)
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info, created_at, updated_at)
VALUES (
    'startup',
    'Startup Bonus',
    'Startup Roulette Event',
    'Spin the startup roulette and win small ETH rewards.',
    EXTRACT(EPOCH FROM (NOW() + INTERVAL '365 days'))::BIGINT * 1000,
    'startup',
    '[
        {"place": "any", "value": "0.01 ETH"}
    ]'::jsonb,
    'Default startup roulette event',
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
)
ON CONFLICT (id) DO NOTHING;

-- Insert three prize values for startup event
-- Values are in points: 1 ETH = 10^9 points
-- 0.01 ETH = 10,000,000 points
-- 0.005 ETH = 5,000,000 points
-- 0.001 ETH = 1,000,000 points
INSERT INTO prize_values (event_id, value, label, segment_id, created_at, updated_at)
VALUES 
    ('startup', 10000000, '0.01 ETH', '1', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
    ('startup', 5000000, '0.005 ETH', '2', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
    ('startup', 1000000, '0.001 ETH', '3', EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
ON CONFLICT DO NOTHING;

