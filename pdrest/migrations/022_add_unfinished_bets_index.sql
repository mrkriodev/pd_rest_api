-- Speed up /api/user/unfinished_bets query:
-- WHERE user_uuid = ? AND (close_price IS NULL OR claimed_status = FALSE)
-- ORDER BY open_time DESC
CREATE INDEX IF NOT EXISTS idx_bets_user_unfinished_open_time
ON bets (user_uuid, open_time DESC)
WHERE (close_price IS NULL OR claimed_status = FALSE);

