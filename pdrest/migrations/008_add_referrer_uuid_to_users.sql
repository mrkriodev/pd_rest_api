-- Add referrer linkage to users table

ALTER TABLE users
ADD COLUMN IF NOT EXISTS referrer_user_uuid UUID NULL,
ADD CONSTRAINT fk_users_referrer_user FOREIGN KEY (referrer_user_uuid) REFERENCES users(user_uuid);

COMMENT ON COLUMN users.referrer_user_uuid IS 'UUID of the user who referred this user';

