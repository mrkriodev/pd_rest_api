-- Add start_time column to all_events and backfill existing rows

ALTER TABLE all_events
ADD COLUMN IF NOT EXISTS start_time BIGINT;

-- Backfill existing events with created_at if start_time is missing
UPDATE all_events
SET start_time = created_at
WHERE start_time IS NULL;

-- Ensure best_of_the_week_09_02_2026 has correct start_time
UPDATE all_events
SET start_time = 1770595200000
WHERE id = 'best_of_the_week_09_02_2026';

