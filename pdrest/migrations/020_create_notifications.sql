-- Create notifications status enum
CREATE TYPE status_notification AS ENUM (
  'CREATED',
  'DELIVERED',
  'UNKNOWN',
  'ERROR',
  'ERROR_USER_BLOCK'
);

-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
  id SERIAL PRIMARY KEY,
  user_id UUID NULL,
  producer VARCHAR NOT NULL,
  message VARCHAR NOT NULL,
  status status_notification NOT NULL DEFAULT 'CREATED',
  created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
  updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
  CONSTRAINT fk_notifications_user
    FOREIGN KEY (user_id) REFERENCES users(user_uuid) ON DELETE SET NULL
);
