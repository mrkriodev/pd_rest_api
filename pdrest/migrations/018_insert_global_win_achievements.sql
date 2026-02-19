-- Insert global achievements for winning bets milestones
INSERT INTO achievements (id, badge, title, image_url, desc_text, tags, steps, step_desc, created_at, updated_at)
VALUES
    (
        'wins_10',
        '10 Wins',
        '10 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-10-wins.png',
        'Awarded for 10 successful bets.',
        'global',
        10,
        'Win 10 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'wins_50',
        '50 Wins',
        '50 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-50-wins.png',
        'Awarded for 50 successful bets.',
        'global',
        50,
        'Win 50 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'wins_100',
        '100 Wins',
        '100 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-100-wins.png',
        'Awarded for 100 successful bets.',
        'global',
        100,
        'Win 100 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'wins_250',
        '250 Wins',
        '250 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-250-wins.png',
        'Awarded for 250 successful bets.',
        'global',
        250,
        'Win 250 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'wins_500',
        '500 Wins',
        '500 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-500-wins.png',
        'Awarded for 500 successful bets.',
        'global',
        500,
        'Win 500 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'wins_1000',
        '1000 Wins',
        '1000 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-1000-wins.png',
        'Awarded for 1000 successful bets.',
        'global',
        1000,
        'Win 1000 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'wins_5000',
        '5000 Wins',
        '5000 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-5000-wins.png',
        'Awarded for 5000 successful bets.',
        'global',
        5000,
        'Win 5000 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    ),
    (
        'wins_10000',
        '10000 Wins',
        '10000 Successful Bets',
        'https://mrkriodev.github.io/mrkrio.github.io/data/achieves/achieve-10000-wins.png',
        'Awarded for 10000 successful bets.',
        'global',
        10000,
        'Win 10000 bets',
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
        EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
    )
ON CONFLICT (id) DO NOTHING;

