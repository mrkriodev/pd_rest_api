-- Update user_achievements table for claiming flow

ALTER TABLE user_achievements
    ADD COLUMN IF NOT EXISTS claimed_status BOOL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS steps_got INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS need_steps INTEGER DEFAULT 0;

UPDATE user_achievements ua
SET need_steps = a.steps
FROM achievements a
WHERE ua.achievement_id = a.id
  AND (ua.need_steps IS NULL OR ua.need_steps = 0);

