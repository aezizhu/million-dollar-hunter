ALTER TABLE auth_audit RENAME COLUMN event_type TO event;
ALTER TABLE auth_audit RENAME COLUMN ip_address TO ip;
ALTER TABLE auth_audit RENAME COLUMN timestamp TO created_at;

DROP INDEX IF EXISTS idx_auth_audit_event_type;
CREATE INDEX IF NOT EXISTS idx_auth_audit_event ON auth_audit(event);

DROP INDEX IF EXISTS idx_auth_audit_user_timestamp;
CREATE INDEX IF NOT EXISTS idx_auth_audit_user_created ON auth_audit(user_id, created_at DESC);
