-- Add start_time column to all_events and backfill existing rows

ALTER TABLE all_events
ADD COLUMN IF NOT EXISTS start_time BIGINT;

-- Backfill existing events with created_at if start_time is missing
UPDATE all_events
SET start_time = created_at
WHERE start_time IS NULL;

-- Add best_of_the_week event and prize values
-- Start: 2026-02-09T00:00:00Z (1770595200000)
-- End:   2026-02-15T00:00:00Z (1771113600000)

INSERT INTO all_events (id, badge, title, desc_text, start_time, deadline, tags, reward, info, created_at, updated_at)
VALUES (
    'best_of_the_week_09_02_2026',
    'https://mrkriodev.github.io/mrkrio.github.io/data/events/event-weekly-3.png',
    'Best of the Week',
    'Weekly competition for the best betting results.',
    1770595200000,
    1771113600000,
    'competition',
    '[
        {"place": "1", "prize": "50 USDT", "image_url": "https://mrkriodev.github.io/mrkrio.github.io/data/events/event-weekly-3.png"},
        {"place": "2", "prize": "30 USDT", "image_url": "https://mrkriodev.github.io/mrkrio.github.io/data/events/event-weekly-1.png"},
        {"place": "3", "prize": "10 USDT", "image_url": "https://mrkriodev.github.io/mrkrio.github.io/data/events/event-weekly-1.png"}
    ]'::jsonb,
    'Start: 2026-02-09T00:00:00Z',
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
)
ON CONFLICT (id) DO NOTHING;

-- Prize values for placements (value is used only for ordering/lookup)
INSERT INTO prize_values (event_id, value, label, segment_id, created_at, updated_at)
VALUES
    ('best_of_the_week_09_02_2026', 50, '50 USDT', NULL, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
    ('best_of_the_week_09_02_2026', 30, '30 USDT', NULL, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000),
    ('best_of_the_week_09_02_2026', 10, '10 USDT', NULL, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000)
ON CONFLICT DO NOTHING;

-- Achievements for best_of_the_week event placements
INSERT INTO achievements (id, badge, title, image_url, desc_text, tags, prize_id, steps, step_desc, created_at, updated_at)
VALUES
    (
        'best_of_the_week_09_02_2026_first',
        'Best of the Week',
        'Best of the Week: 1st Place',
        'https://mrkriodev.github.io/mrkrio.github.io/data/events/event-weekly-3.png',
        'Finish 1st place in Best of the Week competition.',
        'event',
        (SELECT id FROM prize_values WHERE event_id = 'best_of_the_week_09_02_2026' AND value = 50 LIMIT 1),
        1,
        'Claim event prize',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'best_of_the_week_09_02_2026_second',
        'Best of the Week',
        'Best of the Week: 2nd Place',
        'https://mrkriodev.github.io/mrkrio.github.io/data/events/event-weekly-1.png',
        'Finish 2nd place in Best of the Week competition.',
        'event',
        (SELECT id FROM prize_values WHERE event_id = 'best_of_the_week_09_02_2026' AND value = 30 LIMIT 1),
        1,
        'Claim event prize',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'best_of_the_week_09_02_2026_third',
        'Best of the Week',
        'Best of the Week: 3rd Place',
        'https://mrkriodev.github.io/mrkrio.github.io/data/events/event-weekly-1.png',
        'Finish 3rd place in Best of the Week competition.',
        'event',
        (SELECT id FROM prize_values WHERE event_id = 'best_of_the_week_09_02_2026' AND value = 10 LIMIT 1),
        1,
        'Claim event prize',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    )
ON CONFLICT (id) DO NOTHING;

-- Ensure best_of_the_week_09_02_2026 has correct start_time
UPDATE all_events
SET start_time = 1770595200000
WHERE id = 'best_of_the_week_09_02_2026';