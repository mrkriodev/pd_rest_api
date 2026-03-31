-- Make rating.bet_id cascade on bet deletion.
-- Old behavior was ON DELETE SET NULL.

ALTER TABLE rating
    DROP CONSTRAINT IF EXISTS fk_rating_bet;

ALTER TABLE rating
    ADD CONSTRAINT fk_rating_bet
    FOREIGN KEY (bet_id) REFERENCES bets(id) ON DELETE CASCADE;

