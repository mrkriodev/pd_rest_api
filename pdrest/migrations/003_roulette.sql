-- Create roulette tables
-- This migration creates tables for roulette functionality with preauth tokens

-- roulette_config: Stores configuration for roulette (N spins, type, event_id)
CREATE TABLE IF NOT EXISTS roulette_config (
    id SERIAL PRIMARY KEY,
    roulette_type VARCHAR(20) NOT NULL CHECK (roulette_type IN ('on_start', 'during_event')),
    event_id VARCHAR(50) NOT NULL,           -- Always required, references all_events
    max_spins INTEGER NOT NULL DEFAULT 1,    -- N spins allowed per user
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    -- Constraints
    CONSTRAINT chk_max_spins_positive CHECK (max_spins > 0),
    CONSTRAINT fk_roulette_config_event FOREIGN KEY (event_id) REFERENCES all_events(id) ON DELETE RESTRICT
);

-- roulette_preauth_token: Stores preauth tokens generated in user browser
CREATE TABLE IF NOT EXISTS roulette_preauth_token (
    id SERIAL PRIMARY KEY,
    token VARCHAR(255) UNIQUE NOT NULL,      -- Token generated in browser
    user_uuid UUID,                          -- Optional reference to users table (NULL for unauthenticated users)
    roulette_config_id INTEGER NOT NULL,     -- Reference to roulette_config
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at BIGINT NOT NULL,              -- Expiration timestamp (milliseconds)
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

    -- Foreign keys
    CONSTRAINT fk_roulette_preauth_user FOREIGN KEY (user_uuid) REFERENCES users(user_uuid) ON DELETE SET NULL,
    CONSTRAINT fk_roulette_preauth_config FOREIGN KEY (roulette_config_id) REFERENCES roulette_config(id) ON DELETE CASCADE
);

-- roulette: Tracks spins and prizes (linked to preauth token, not user directly)
CREATE TABLE IF NOT EXISTS roulette (
    id SERIAL PRIMARY KEY,
    roulette_config_id INTEGER NOT NULL,     -- Reference to roulette_config
    preauth_token_id INTEGER NOT NULL UNIQUE, -- Reference to roulette_preauth_token (one roulette per token)
    spin_number INTEGER NOT NULL,            -- Current spin number (1 to max_spins)
    prize TEXT,                              -- Prize won (NULL if not yet taken)
    prize_taken BOOLEAN NOT NULL DEFAULT FALSE, -- Whether prize has been taken
    spin_result JSONB,                       -- Store spin result details (optional)
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    prize_taken_at BIGINT,                   -- Timestamp when prize was taken

    -- Foreign keys
    CONSTRAINT fk_roulette_config FOREIGN KEY (roulette_config_id) REFERENCES roulette_config(id) ON DELETE CASCADE,
    CONSTRAINT fk_roulette_preauth FOREIGN KEY (preauth_token_id) REFERENCES roulette_preauth_token(id) ON DELETE CASCADE,

    -- Constraints
    CONSTRAINT chk_spin_number_positive CHECK (spin_number > 0),
    CONSTRAINT chk_prize_taken CHECK (
        (prize_taken = TRUE AND prize IS NOT NULL AND prize_taken_at IS NOT NULL) OR
        (prize_taken = FALSE)
    )
);

-- Create indexes for better query performance
-- CREATE INDEX IF NOT EXISTS idx_roulette_config_type ON roulette_config(roulette_type);
-- CREATE INDEX IF NOT EXISTS idx_roulette_config_event_id ON roulette_config(event_id);
-- CREATE INDEX IF NOT EXISTS idx_roulette_config_active ON roulette_config(is_active);

-- CREATE INDEX IF NOT EXISTS idx_roulette_preauth_token ON roulette_preauth_token(token);
-- CREATE INDEX IF NOT EXISTS idx_roulette_preauth_user ON roulette_preauth_token(user_uuid) WHERE user_uuid IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_roulette_preauth_config ON roulette_preauth_token(roulette_config_id);
-- CREATE INDEX IF NOT EXISTS idx_roulette_preauth_used ON roulette_preauth_token(is_used);
-- CREATE INDEX IF NOT EXISTS idx_roulette_preauth_expires ON roulette_preauth_token(expires_at);

-- CREATE INDEX IF NOT EXISTS idx_roulette_config_id ON roulette(roulette_config_id);
-- CREATE INDEX IF NOT EXISTS idx_roulette_preauth_token_id ON roulette(preauth_token_id);
-- CREATE INDEX IF NOT EXISTS idx_roulette_prize_taken ON roulette(prize_taken);

-- Create function to update updated_at timestamp for roulette_config
CREATE OR REPLACE FUNCTION update_roulette_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at for roulette_config
CREATE TRIGGER trigger_update_roulette_config_updated_at
    BEFORE UPDATE ON roulette_config
    FOR EACH ROW
    EXECUTE FUNCTION update_roulette_config_updated_at();

-- Create function to update updated_at timestamp for roulette
CREATE OR REPLACE FUNCTION update_roulette_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT * 1000;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at for roulette
CREATE TRIGGER trigger_update_roulette_updated_at
    BEFORE UPDATE ON roulette
    FOR EACH ROW
    EXECUTE FUNCTION update_roulette_updated_at();

-- Seed roulette_config id=1 for startup roulette
INSERT INTO roulette_config (id, roulette_type, event_id, max_spins, is_active, created_at, updated_at)
VALUES (
    1, 'on_start', 'startup', 3, TRUE,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
)
ON CONFLICT (id) DO UPDATE
SET roulette_type = EXCLUDED.roulette_type,
    event_id = EXCLUDED.event_id,
    max_spins = EXCLUDED.max_spins,
    is_active = EXCLUDED.is_active,
    updated_at = EXCLUDED.updated_at;

-- Comments for documentation
-- COMMENT ON TABLE roulette_config IS 'Stores roulette configuration (N spins, type: on_start or during_event)';
-- COMMENT ON COLUMN roulette_config.roulette_type IS 'Type: on_start (once) or during_event (per event_id)';
-- COMMENT ON COLUMN roulette_config.event_id IS 'Event ID (foreign key to all_events). Required for all roulette types';
-- COMMENT ON COLUMN roulette_config.max_spins IS 'Maximum number of spins allowed per user (N)';

-- COMMENT ON TABLE roulette_preauth_token IS 'Stores preauth tokens generated in user browser';
-- COMMENT ON COLUMN roulette_preauth_token.token IS 'Token generated in browser for authentication';
-- COMMENT ON COLUMN roulette_preauth_token.is_used IS 'Whether token has been used for spinning';
-- COMMENT ON COLUMN roulette_preauth_token.expires_at IS 'Token expiration timestamp in milliseconds';

-- COMMENT ON TABLE roulette IS 'Tracks spins and prizes (linked to preauth token, works for unauthenticated users)';
-- COMMENT ON COLUMN roulette_preauth_token.user_uuid IS 'Optional user UUID (NULL for unauthenticated users)';
-- COMMENT ON COLUMN roulette.spin_number IS 'Current spin number (1 to max_spins)';
-- COMMENT ON COLUMN roulette.prize IS 'Prize won (NULL if not yet taken)';
-- COMMENT ON COLUMN roulette.prize_taken IS 'Whether prize has been taken (no more spins after this)';
-- COMMENT ON COLUMN roulette.spin_result IS 'Optional JSONB field for spin result details';

