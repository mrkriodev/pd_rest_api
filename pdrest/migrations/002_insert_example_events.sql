-- Insert example events into all_events table
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
        {"place": "1-3", "value": "0.0006 ETH"},
        {"place": "4-5", "value": "0.0003 ETH"},
        {"place": "6-7", "value": "0.0003 ETH"},
        {"place": "8", "value": "0.0003 ETH"},
        {"place": "9-10", "value": "0.0001 ETH"}
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
        {"place": "1", "value": "0.01 BTC"},
        {"place": "2", "value": "0.005 BTC"},
        {"place": "3", "value": "0.002 BTC"},
        {"place": "4-10", "value": "0.001 BTC"}
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

