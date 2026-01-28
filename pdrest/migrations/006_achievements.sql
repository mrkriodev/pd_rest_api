-- Create achievements table
-- This table stores achievement definitions

CREATE TABLE IF NOT EXISTS achievements (
    id VARCHAR(50) PRIMARY KEY,
    badge TEXT NOT NULL,
    title TEXT NOT NULL,
    image_url TEXT NOT NULL,
    desc_text TEXT NOT NULL,
    tags VARCHAR(100),
    summ NUMERIC(20, 8) NOT NULL DEFAULT 0,
    steps INTEGER NOT NULL DEFAULT 0,
    step_desc TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
);

-- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_achievements_id ON achievements(id);
-- CREATE INDEX IF NOT EXISTS idx_achievements_tags ON achievements(tags);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_achievements_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER trigger_update_achievements_updated_at
    BEFORE UPDATE ON achievements
    FOR EACH ROW
    EXECUTE FUNCTION update_achievements_updated_at();

-- Comments for documentation
-- COMMENT ON TABLE achievements IS 'Stores achievement definitions';
-- COMMENT ON COLUMN achievements.id IS 'Unique achievement identifier';
-- COMMENT ON COLUMN achievements.badge IS 'Badge name/category';
-- COMMENT ON COLUMN achievements.title IS 'Achievement title';
-- COMMENT ON COLUMN achievements.image_url IS 'URL to achievement image/icon';
-- COMMENT ON COLUMN achievements.desc_text IS 'Achievement description';
-- COMMENT ON COLUMN achievements.tags IS 'Tags for filtering (e.g., "global")';
-- COMMENT ON COLUMN achievements.summ IS 'Reward sum/amount';
-- COMMENT ON COLUMN achievements.steps IS 'Number of steps required';
-- COMMENT ON COLUMN achievements.step_desc IS 'Description of steps required';

