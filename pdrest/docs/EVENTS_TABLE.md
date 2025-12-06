# All Events Table Schema

## Database: `pumpdump_db`
## Table: `all_events`

This table stores event information for the Pump or Dump application.

## Table Structure

```sql
CREATE TABLE all_events (
    id VARCHAR(50) PRIMARY KEY,
    badge TEXT NOT NULL,
    title TEXT NOT NULL,
    desc_text TEXT NOT NULL,
    deadline BIGINT NOT NULL,
    tags VARCHAR(100),
    reward JSONB NOT NULL,
    info TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000
);
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | VARCHAR(50) | Primary key, unique event identifier (e.g., "e1") |
| `badge` | TEXT | Badge text for the event |
| `title` | TEXT | Event title |
| `desc_text` | TEXT | Event description |
| `deadline` | BIGINT | Event deadline as Unix timestamp in milliseconds (UTC) |
| `tags` | VARCHAR(100) | Event tags (e.g., "global") |
| `reward` | JSONB | Array of reward objects (see Reward Structure below) |
| `info` | TEXT | Additional event information |
| `created_at` | BIGINT | Auto-generated creation timestamp as Unix milliseconds (UTC) |
| `updated_at` | BIGINT | Auto-generated update timestamp as Unix milliseconds (UTC) |

## Reward Structure (JSONB)

The `reward` column stores a JSON array of reward objects:

```json
[
  { "place": "1-3", "value": "0.0006 ETH" },
  { "place": "4-5", "value": "0.0003 ETH" },
  { "place": "6-7", "value": "0.0003 ETH" },
  { "place": "8", "value": "0.0003 ETH" },
  { "place": "9-10", "value": "0.0001 ETH" }
]
```

## Indexes

- `idx_all_events_id` - Index on `id` column
- `idx_all_events_deadline` - Index on `deadline` column
- `idx_all_events_tags` - Index on `tags` column

## Example Insert

```sql
-- Note: deadline is Unix timestamp in milliseconds (UTC)
-- Example: '2025-12-01T12:00:00Z' = 1733054400000
INSERT INTO all_events (id, badge, title, desc_text, deadline, tags, reward, info)
VALUES (
    'e1',
    'Get 1000+ prizes.',
    'Ethereum Pump or Dump',
    'Reward for 1st place, to the one who stays in the top for 7 days',
    1733054400000,
    'global',
    '[
        {"place": "1-3", "value": "0.0006 ETH"},
        {"place": "4-5", "value": "0.0003 ETH"},
        {"place": "6-7", "value": "0.0003 ETH"},
        {"place": "8", "value": "0.0003 ETH"},
        {"place": "9-10", "value": "0.0001 ETH"}
    ]'::jsonb,
    'Your intuition and team spirit are what matter most. Join a squad or create your own...'
);
```

## Timestamp Format

All timestamp columns (`deadline`, `created_at`, `updated_at`) store Unix timestamps in **milliseconds** (UTC). 

- To convert from ISO 8601 to Unix milliseconds: Use `EXTRACT(EPOCH FROM '2025-12-01T12:00:00Z'::timestamptz)::BIGINT * 1000`
- To convert from Unix milliseconds to timestamp: Use `to_timestamp(1733054400000 / 1000.0)`

## Querying Rewards (JSONB)

### Get all events
```sql
SELECT * FROM all_events ORDER BY deadline ASC;
```

### Query events by tag
```sql
SELECT * FROM all_events WHERE tags = 'global';
```

### Query events by deadline (convert BIGINT to timestamp for readability)
```sql
SELECT id, title, to_timestamp(deadline / 1000.0) as deadline_ts
FROM all_events
WHERE deadline > EXTRACT(EPOCH FROM NOW())::BIGINT * 1000;
```

### Query rewards from a specific event
```sql
SELECT reward FROM all_events WHERE id = 'e1';
```

### Query specific reward place
```sql
SELECT reward->0->>'place' as place, reward->0->>'value' as value
FROM all_events
WHERE id = 'e1';
```

### Find events with specific reward value
```sql
SELECT * FROM all_events
WHERE reward @> '[{"place": "1-3"}]'::jsonb;
```

## API Response Format

The table structure matches the API response format:

```json
{
  "events": [
    {
      "id": "e1",
      "badge": "Get 1000+ prizes.",
      "title": "Ethereum Pump or Dump",
      "desc": "Reward for 1st place, to the one who stays in the top for 7 days",
      "deadline": "2025-12-01T12:00:00Z",
      "tags": "global",
      "reward": [
        { "place": "1-3", "value": "0.0006 ETH" },
        { "place": "4-5", "value": "0.0003 ETH" },
        { "place": "6-7", "value": "0.0003 ETH" },
        { "place": "8", "value": "0.0003 ETH" },
        { "place": "9-10", "value": "0.0001 ETH" }
      ],
      "info": "Your intuition and team spirit are what matter most. Join a squad or create your own..."
    }
  ]
}
```

