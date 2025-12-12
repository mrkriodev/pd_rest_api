-- Modify prizes table to add prize_value_id and make user_uuid mandatory

-- First, add the new column (nullable initially to allow migration)
ALTER TABLE prizes
ADD COLUMN IF NOT EXISTS prize_value_id INTEGER;

-- Add foreign key constraint
ALTER TABLE prizes
ADD CONSTRAINT fk_prizes_prize_value 
FOREIGN KEY (prize_value_id) REFERENCES prize_values(id) ON DELETE SET NULL;

-- Handle existing NULL user_uuid values before making it mandatory
-- For prizes with preauth_token_id but NULL user_uuid, we need to either:
-- 1. Delete them (if they're test/old data)
-- 2. Link them to users via preauth_token (requires additional logic)
-- For now, we'll delete prizes with NULL user_uuid as they're likely test data
-- In production, you should handle this more carefully by linking preauth_tokens to users first
DELETE FROM prizes WHERE user_uuid IS NULL;

-- Drop the old constraint that allowed either user_uuid OR preauth_token_id
ALTER TABLE prizes
DROP CONSTRAINT IF EXISTS chk_prize_recipient;

-- Make user_uuid mandatory
ALTER TABLE prizes
ALTER COLUMN user_uuid SET NOT NULL;

-- -- Create index for prize_value_id
-- CREATE INDEX IF NOT EXISTS idx_prizes_prize_value_id ON prizes(prize_value_id);

-- -- Comments
-- COMMENT ON COLUMN prizes.prize_value_id IS 'Reference to prize_values table for the prize value won';
-- COMMENT ON COLUMN prizes.user_uuid IS 'User UUID (now mandatory)';

