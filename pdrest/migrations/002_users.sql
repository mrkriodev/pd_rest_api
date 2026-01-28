-- Create users table
-- This table stores user information with authorization status
-- Users can be authorized_partially (by IP and X-SESSION-ID) or authorized_fully (via Google/Telegram OAuth)

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    user_uuid UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    referrer_user_uuid UUID NULL,

    -- Authorization status
    authorized_fully BOOLEAN NOT NULL DEFAULT FALSE,

    -- Partial authorization (before OAuth)
    ip_address INET,                    -- IP address for partial authorization
    session_id VARCHAR(255) UNIQUE,     -- X-SESSION-ID cookie/header value

    -- Full authorization (after OAuth)
    google_id VARCHAR(255) UNIQUE,       -- Google OAuth user ID
    google_email VARCHAR(255),           -- Google email
    google_name VARCHAR(255),            -- Google display name
    telegram_id BIGINT UNIQUE,           -- Telegram user ID
    telegram_username VARCHAR(255),      -- Telegram username
    telegram_first_name VARCHAR(255),    -- Telegram first name
    telegram_last_name VARCHAR(255),     -- Telegram last name

    -- Metadata
    auth_provider VARCHAR(50),           -- 'google', 'telegram', or NULL for partial
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    last_login_at BIGINT,                -- Last login timestamp

    -- Constraints
    CONSTRAINT chk_auth_provider CHECK (auth_provider IN ('google', 'telegram', NULL)),
    CONSTRAINT chk_authorized_fully CHECK (
        (authorized_fully = TRUE AND (google_id IS NOT NULL OR telegram_id IS NOT NULL)) OR
        (authorized_fully = FALSE)
    ),
    CONSTRAINT fk_users_referrer_user FOREIGN KEY (referrer_user_uuid) REFERENCES users(user_uuid)
);

-- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_users_authorized_fully ON users(authorized_fully);
-- CREATE INDEX IF NOT EXISTS idx_users_session_id ON users(session_id);
-- CREATE INDEX IF NOT EXISTS idx_users_ip_address ON users(ip_address);
-- CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);
-- CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
-- CREATE INDEX IF NOT EXISTS idx_users_auth_provider ON users(auth_provider);
-- CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_user_uuid ON users(user_uuid);

-- Create a composite index for partial authorization lookups
CREATE INDEX IF NOT EXISTS idx_users_partial_auth ON users(ip_address, session_id)
WHERE authorized_fully = FALSE;

-- Create a function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER trigger_update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_users_updated_at();

-- Comments for documentation
-- COMMENT ON TABLE users IS 'Stores user information with partial (IP+session) or full (OAuth) authorization';
-- COMMENT ON COLUMN users.user_uuid IS 'Unique UUID identifier for each user';
-- COMMENT ON COLUMN users.referrer_user_uuid IS 'UUID of the user who referred this user';
-- COMMENT ON COLUMN users.authorized_fully IS 'TRUE when user completed Google or Telegram OAuth, FALSE for partial auth by IP/session';
-- COMMENT ON COLUMN users.ip_address IS 'IP address used for partial authorization (before OAuth)';
-- COMMENT ON COLUMN users.session_id IS 'X-SESSION-ID value used for partial authorization';
-- COMMENT ON COLUMN users.google_id IS 'Google OAuth user ID (NULL if not authenticated via Google)';
-- COMMENT ON COLUMN users.telegram_id IS 'Telegram user ID (NULL if not authenticated via Telegram)';
-- COMMENT ON COLUMN users.auth_provider IS 'OAuth provider: google, telegram, or NULL for partial auth';

