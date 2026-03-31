-- Insert global achievements for winning bets milestones.
-- Each milestone has its own event (e5..e9), own prize_value and linked achievement.

WITH events_src AS (
    SELECT * FROM (VALUES
        ('e5', '10 Wins',  '10 Successful Bets',  'Awarded for 10 successful bets.',  'wins_10',  10,  '10 USDT'),
        ('e6', '50 Wins',  '50 Successful Bets',  'Awarded for 50 successful bets.',  'wins_50',  50,  '50 USDT'),
        ('e7', '100 Wins', '100 Successful Bets', 'Awarded for 100 successful bets.', 'wins_100', 100, '100 USDT'),
        ('e8', '250 Wins', '250 Successful Bets', 'Awarded for 250 successful bets.', 'wins_250', 250, '250 USDT'),
        ('e9', '500 Wins', '500 Successful Bets', 'Awarded for 500 successful bets.', 'wins_500', 500, '500 USDT')
    ) AS t(event_id, badge, title, desc_text, place_key, prize_value, prize_label)
)
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info, created_at, updated_at)
SELECT
    e.event_id,
    e.badge,
    e.title,
    e.desc_text,
    EXTRACT(EPOCH FROM (NOW() + INTERVAL '3650 days'))::BIGINT * 1000,
    'achivements',
    jsonb_build_array(jsonb_build_object('place', 'any', 'value', e.prize_label)),
    'Global wins achievement event: ' || e.place_key,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
FROM events_src e
ON CONFLICT (id) DO NOTHING;

WITH prizes_src AS (
    SELECT * FROM (VALUES
        ('e5', 10,  '10 USDT'),
        ('e6', 50,  '50 USDT'),
        ('e7', 100, '100 USDT'),
        ('e8', 250, '250 USDT'),
        ('e9', 500, '500 USDT')
    ) AS t(event_id, value, label)
)
INSERT INTO prize_values (event_id, value, label, segment_id, created_at, updated_at)
SELECT
    p.event_id,
    p.value,
    p.label,
    NULL,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
FROM prizes_src p
WHERE NOT EXISTS (
    SELECT 1
    FROM prize_values pv
    WHERE pv.event_id = p.event_id AND pv.value = p.value AND pv.label = p.label
);

WITH achievements_src AS (
    SELECT * FROM (VALUES
        ('wins_10',  '10 Wins',  '10 Successful Bets',  'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-10-wins.png',  'Awarded for 10 successful bets.',  10,  'Win 10 bets',  'e5', 10),
        ('wins_50',  '50 Wins',  '50 Successful Bets',  'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-50-wins.png',  'Awarded for 50 successful bets.',  50,  'Win 50 bets',  'e6', 50),
        ('wins_100', '100 Wins', '100 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-100-wins.png', 'Awarded for 100 successful bets.', 100, 'Win 100 bets', 'e7', 100),
        ('wins_250', '250 Wins', '250 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-250-wins.png', 'Awarded for 250 successful bets.', 250, 'Win 250 bets', 'e8', 250),
        ('wins_500', '500 Wins', '500 Successful Bets', 'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-500-wins.png', 'Awarded for 500 successful bets.', 500, 'Win 500 bets', 'e9', 500)
    ) AS t(id, badge, title, image_url, desc_text, steps, step_desc, event_id, prize_value)
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
        WHERE pv.event_id = a.event_id AND pv.value = a.prize_value
        ORDER BY pv.id ASC
        LIMIT 1
    ),
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
FROM achievements_src a
ON CONFLICT (id) DO NOTHING;

