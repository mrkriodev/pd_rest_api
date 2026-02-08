-- Update achievements table:
-- - replace summ column with prize_id

ALTER TABLE achievements
DROP COLUMN IF EXISTS summ;

ALTER TABLE achievements
ADD COLUMN IF NOT EXISTS prize_id INTEGER;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_achievements_prize_value') THEN
        ALTER TABLE achievements
        ADD CONSTRAINT fk_achievements_prize_value FOREIGN KEY (prize_id) REFERENCES prize_values(id) ON DELETE SET NULL;
    END IF;
END $$;

