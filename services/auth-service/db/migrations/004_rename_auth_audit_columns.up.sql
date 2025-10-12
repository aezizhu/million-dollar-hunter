
ALTER TABLE auth_audit RENAME COLUMN event TO event_type;
ALTER TABLE auth_audit RENAME COLUMN ip TO ip_address;
ALTER TABLE auth_audit RENAME COLUMN created_at TO timestamp;

DROP INDEX IF EXISTS idx_auth_audit_event;
CREATE INDEX IF NOT EXISTS idx_auth_audit_event_type ON auth_audit(event_type);

DROP INDEX IF EXISTS idx_auth_audit_user_created;
CREATE INDEX IF NOT EXISTS idx_auth_audit_user_timestamp ON auth_audit(user_id, timestamp DESC);
