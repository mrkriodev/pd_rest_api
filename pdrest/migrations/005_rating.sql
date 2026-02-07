-- Create rating table to store user points (1 USDT = 1 point)

CREATE TABLE IF NOT EXISTS rating (
    id SERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,
    points BIGINT NOT NULL,
    got_prize_id INTEGER,
    bet_id INTEGER,
    description TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    CONSTRAINT fk_rating_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_rating_got_prize FOREIGN KEY (got_prize_id) REFERENCES got_prizes(id) ON DELETE SET NULL,
    CONSTRAINT fk_rating_bet FOREIGN KEY (bet_id) REFERENCES bets(id) ON DELETE SET NULL
);

-- CREATE INDEX IF NOT EXISTS idx_rating_user_uuid ON rating(user_uuid);
-- CREATE INDEX IF NOT EXISTS idx_rating_got_prize_id ON rating(got_prize_id);
-- CREATE INDEX IF NOT EXISTS idx_rating_bet_id ON rating(bet_id);

COMMENT ON TABLE rating IS 'Stores user rating points for prizes and bets';
COMMENT ON COLUMN rating.user_uuid IS 'Reference to users.user_uuid';
COMMENT ON COLUMN rating.points IS 'Number of points awarded (USDT)';
COMMENT ON COLUMN rating.got_prize_id IS 'Reference to got_prizes.id';
COMMENT ON COLUMN rating.bet_id IS 'Reference to bets.id';

