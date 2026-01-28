-- Create user_achievements table
-- This table stores which achievements users have earned

CREATE TABLE IF NOT EXISTS user_achievements (
    id SERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,
    achievement_id VARCHAR(50) NOT NULL,
    earned_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    -- Foreign keys
    CONSTRAINT fk_user_achievements_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_user_achievements_achievement FOREIGN KEY (achievement_id) REFERENCES achievements(id) ON DELETE CASCADE,

    -- Unique constraint: a user can only earn an achievement once
    CONSTRAINT uq_user_achievement UNIQUE (user_uuid, achievement_id)
);

-- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_user_achievements_user_uuid ON user_achievements(user_uuid);
-- CREATE INDEX IF NOT EXISTS idx_user_achievements_achievement_id ON user_achievements(achievement_id);
-- CREATE INDEX IF NOT EXISTS idx_user_achievements_earned_at ON user_achievements(earned_at);

-- -- Comments for documentation
-- COMMENT ON TABLE user_achievements IS 'Stores achievements earned by users';
-- COMMENT ON COLUMN user_achievements.user_uuid IS 'Reference to users.user_uuid';
-- COMMENT ON COLUMN user_achievements.achievement_id IS 'Reference to achievements.id';
-- COMMENT ON COLUMN user_achievements.earned_at IS 'Timestamp when achievement was earned (milliseconds)';

