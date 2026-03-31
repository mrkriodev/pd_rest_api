-- Insert global achievements for winning bets milestones with linked prizes.
WITH inserted_event AS (
    INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info, created_at, updated_at)
    VALUES (
        'e5',
        'Global Wins',
        'Global Wins Milestones',
        'Rewards for cumulative successful bets milestones.',
        EXTRACT(EPOCH FROM (NOW() + INTERVAL '3650 days'))::BIGINT * 1000,
        'achivements',
        '[
            {"place": "wins_10", "value": "10 USDT"},
            {"place": "wins_50", "value": "50 USDT"},
            {"place": "wins_100", "value": "100 USDT"},
            {"place": "wins_250", "value": "250 USDT"},
            {"place": "wins_500", "value": "500 USDT"}
        ]'::jsonb,
        'Global wins achievements rewards event',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    )
    ON CONFLICT (id) DO NOTHING
    RETURNING id
),
prize_rows AS (
    SELECT * FROM (VALUES
        (10, '10 USDT'),
        (50, '50 USDT'),
        (100, '100 USDT'),
        (250, '250 USDT'),
        (500, '500 USDT')
    ) AS t(value, label)
)
INSERT INTO prize_values (event_id, value, label, segment_id, created_at, updated_at)
SELECT
    'e5',
    p.value,
    p.label,
    NULL,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
FROM prize_rows p
WHERE NOT EXISTS (
    SELECT 1
    FROM prize_values pv
    WHERE pv.event_id = 'e5' AND pv.value = p.value AND pv.label = p.label
);

WITH achievements_src AS (
    SELECT * FROM (VALUES
        ('wins_10', '10 Wins', '10 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-10-wins.png', 'Awarded for 10 successful bets.', 10, 'Win 10 bets', 10),
        ('wins_50', '50 Wins', '50 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-50-wins.png', 'Awarded for 50 successful bets.', 50, 'Win 50 bets', 50),
        ('wins_100', '100 Wins', '100 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-100-wins.png', 'Awarded for 100 successful bets.', 100, 'Win 100 bets', 100),
        ('wins_250', '250 Wins', '250 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-250-wins.png', 'Awarded for 250 successful bets.', 250, 'Win 250 bets', 250),
        ('wins_500', '500 Wins', '500 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-500-wins.png', 'Awarded for 500 successful bets.', 500, 'Win 500 bets', 500)
    ) AS t(id, badge, title, image_url, desc_text, steps, step_desc, prize_value)
)
INSERT INTO achievements (id, badge, title, image_url, desc_text, tags, steps, step_desc, prize_id, created_at, updated_at)
SELECT
    a.id,
    a.badge,
    a.title,
    a.image_url,
    a.desc_text,
    'global',
    a.steps,
    a.step_desc,
    (
        SELECT pv.id
        FROM prize_values pv
        WHERE pv.event_id = 'e5' AND pv.value = a.prize_value
        ORDER BY pv.id ASC
        LIMIT 1
    ),
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
FROM achievements_src a
ON CONFLICT (id) DO NOTHING;

