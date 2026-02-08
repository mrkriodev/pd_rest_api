-- Insert prize value and achievement for first successful bet

WITH inserted_event AS (
    INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info, created_at, updated_at)
    VALUES (
        'e4',
        'First Win',
        'First Win',
        'Awarded for the first successful bet.',
        EXTRACT(EPOCH FROM (NOW() + INTERVAL '3650 days'))::BIGINT * 1000,
        'achivements',
        '[{"place": "any", "value": "10 USDT"}]'::jsonb,
        'First win achievement event',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    )
    ON CONFLICT (id) DO NOTHING
    RETURNING id
),
inserted_prize AS (
    INSERT INTO prize_values (event_id, value, label, segment_id, created_at, updated_at)
    VALUES ('e4', 10, '10 USDT', NULL, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
    ON CONFLICT DO NOTHING
    RETURNING id
),
prize_id AS (
    SELECT id FROM inserted_prize
    UNION ALL
    SELECT id FROM prize_values
    WHERE event_id = 'e4' AND value = 10 AND label = '10 USDT'
    LIMIT 1
)
INSERT INTO achievements (id, badge, title, image_url, desc_text, tags, steps, step_desc, prize_id, created_at, updated_at)
SELECT
    'first_bet_success',
    'First Bet',
    'First Successful Bet',
    'https://mrkriodev.github.io/mrkrio.github.io/data/1-bet-ach.svg',
    'Awarded for the first successful bet.',
    'bet',
    1,
    'Win your first bet',
    (SELECT id FROM prize_id),
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
ON CONFLICT (id) DO NOTHING;

