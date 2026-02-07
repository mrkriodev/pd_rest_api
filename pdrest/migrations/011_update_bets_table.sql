-- Update bets table:
-- - add claimed_status flag

ALTER TABLE bets
ADD COLUMN IF NOT EXISTS claimed_status BOOLEAN NOT NULL DEFAULT FALSE;

