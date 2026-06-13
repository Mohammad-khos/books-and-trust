CREATE TABLE IF NOT EXISTS banned_users (
    user_id UUID PRIMARY KEY,
    reason VARCHAR(500) NOT NULL,
    expired_at TIMESTAMPTZ, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_banned_users_deleted_at ON banned_users(deleted_at);