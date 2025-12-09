-- Create rating table to store user points

CREATE TABLE IF NOT EXISTS rating (
    id SERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,
    points BIGINT NOT NULL CHECK (points >= 0),
    source VARCHAR(50) NOT NULL CHECK (source IN ('from_event', 'bet_bonus', 'promo_bonus', 'servivce_bonus')),
    description TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    CONSTRAINT fk_rating_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE CASCADE
);

-- CREATE INDEX IF NOT EXISTS idx_rating_user_uuid ON rating(user_uuid);
-- CREATE INDEX IF NOT EXISTS idx_rating_source ON rating(source);

COMMENT ON TABLE rating IS 'Stores user rating points from multiple sources';
COMMENT ON COLUMN rating.user_uuid IS 'Reference to users.user_uuid';
COMMENT ON COLUMN rating.points IS 'Number of points awarded';
COMMENT ON COLUMN rating.source IS 'Point source: from_event, bet_bonus, promo_bonus, servivce_bonus';

