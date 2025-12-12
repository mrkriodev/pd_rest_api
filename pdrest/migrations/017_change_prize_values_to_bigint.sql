-- Change prize_values.value from TEXT to BIGINT
-- This represents exact points to add to user balance in ratings without string decoding

-- First, we need to convert existing string values to bigint
-- For values like "0.01 ETH", we'll convert to points (1 ETH = 10^9 points)
-- Example: "0.01 ETH" = 0.01 * 10^9 = 10,000,000 points

-- Step 1: Add a temporary column for the numeric value
ALTER TABLE prize_values ADD COLUMN IF NOT EXISTS value_numeric BIGINT;

-- Step 2: Convert existing TEXT values to BIGINT
-- This handles values like "0.01 ETH", "0.005 ETH", "0.001 ETH"
-- Pattern: extract number, multiply by 10^9 if ETH, otherwise use as-is
UPDATE prize_values
SET value_numeric = CASE
    WHEN value LIKE '%ETH%' THEN
        CAST(REPLACE(REPLACE(value, ' ETH', ''), ' ', '') AS NUMERIC) * 1000000000
    ELSE
        CAST(REPLACE(value, ' ', '') AS NUMERIC)
END
WHERE value_numeric IS NULL;

-- Step 3: Set default for any NULL values (shouldn't happen, but safety)
UPDATE prize_values SET value_numeric = 0 WHERE value_numeric IS NULL;

-- Step 4: Drop the old TEXT column
ALTER TABLE prize_values DROP COLUMN IF EXISTS value;

-- Step 5: Rename the new column to value
ALTER TABLE prize_values RENAME COLUMN value_numeric TO value;

-- Step 6: Make value NOT NULL
ALTER TABLE prize_values ALTER COLUMN value SET NOT NULL;

-- Update comment
COMMENT ON COLUMN prize_values.value IS 'Prize value in points (exact points to add to user balance)';

