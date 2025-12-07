-- Add UUID column to users table
-- This migration adds a unique UUID identifier for each user

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Add UUID column with unique constraint
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS user_uuid UUID UNIQUE DEFAULT uuid_generate_v4();

-- Generate UUIDs for existing rows that don't have one
UPDATE users 
SET user_uuid = uuid_generate_v4() 
WHERE user_uuid IS NULL;

-- Make the column NOT NULL after populating existing rows
ALTER TABLE users 
ALTER COLUMN user_uuid SET NOT NULL;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_users_user_uuid ON users(user_uuid);

-- Add comment for documentation
COMMENT ON COLUMN users.user_uuid IS 'Unique UUID identifier for each user';

