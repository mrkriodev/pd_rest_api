-- Create support chat messages table
CREATE TABLE IF NOT EXISTS support_chat_messages (
  id SERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,        -- Telegram user ID
  message_id BIGINT NOT NULL,
  message TEXT NOT NULL,
  created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,
  updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT * 1000,

  CONSTRAINT fk_support_user
    FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_support_chat_messages_user_id ON support_chat_messages(user_id);
