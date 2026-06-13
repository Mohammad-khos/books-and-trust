CREATE UNIQUE INDEX idx_users_email
ON users(email);

CREATE UNIQUE INDEX idx_users_username
ON users(username);
CREATE INDEX idx_users_not_deleted
ON users(id)
WHERE deleted_at IS NULL;