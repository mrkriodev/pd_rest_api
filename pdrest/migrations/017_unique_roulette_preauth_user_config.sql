-- Ensure a user cannot have multiple preauth tokens for the same roulette config
CREATE UNIQUE INDEX IF NOT EXISTS idx_roulette_preauth_user_config_unique
ON roulette_preauth_token (user_uuid, roulette_config_id)
WHERE user_uuid IS NOT NULL;

