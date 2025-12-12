-- Make roulette_config.event_id NOT NULL and add foreign key constraint
-- All roulette configs must reference an event, including on_start which references 'startup' event

-- Step 1: Ensure 'startup' event exists (should already exist from migration 012)
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info, created_at, updated_at)
VALUES (
    'startup',
    'Startup Bonus',
    'Startup Roulette Event',
    'Spin the startup roulette and win small ETH rewards.',
    EXTRACT(EPOCH FROM (NOW() + INTERVAL '365 days'))::BIGINT * 1000,
    'startup',
    '[
        {"place": "any", "value": "0.01 ETH"}
    ]'::jsonb,
    'Default startup roulette event',
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
)
ON CONFLICT (id) DO NOTHING;

-- Step 2: Update all existing on_start configs to reference 'startup' event
UPDATE roulette_config
SET event_id = 'startup'
WHERE roulette_type = 'on_start' AND event_id IS NULL;

-- Step 3: Remove the old constraint that allowed NULL
ALTER TABLE roulette_config DROP CONSTRAINT IF EXISTS chk_roulette_type_event;

-- Step 4: Make event_id NOT NULL
ALTER TABLE roulette_config ALTER COLUMN event_id SET NOT NULL;

-- Step 5: Add foreign key constraint to all_events
ALTER TABLE roulette_config
ADD CONSTRAINT fk_roulette_config_event 
FOREIGN KEY (event_id) REFERENCES all_events(id) ON DELETE RESTRICT;

-- Step 6: Add new constraint to ensure event_id is always set (redundant but explicit)
ALTER TABLE roulette_config
ADD CONSTRAINT chk_roulette_event_id_not_null CHECK (event_id IS NOT NULL);

-- Update comment
COMMENT ON COLUMN roulette_config.event_id IS 'Event ID (foreign key to all_events). Required for all roulette types, including on_start which references startup event';

INSERT INTO roulette_config (id, roulette_type, event_id, max_spins, is_active, created_at, updated_at)
VALUES (
    1, 'on_start', 'startup', 3, TRUE,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
);