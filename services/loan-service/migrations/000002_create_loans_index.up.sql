CREATE INDEX IF NOT EXISTS idx_loans_owner_id ON loans(owner_id);
CREATE INDEX IF NOT EXISTS idx_loans_deleted_at ON loans(deleted_at);