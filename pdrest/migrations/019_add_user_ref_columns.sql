-- Add referral tracking columns to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS main_ref TEXT;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS add_refs TEXT[] DEFAULT ARRAY[]::TEXT[];

CREATE INDEX IF NOT EXISTS idx_users_main_ref ON users(main_ref);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS language VARCHAR(10) DEFAULT 'en';

