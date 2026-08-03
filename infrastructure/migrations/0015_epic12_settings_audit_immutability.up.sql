-- Epic 12: pengaturan RT dan audit log append-only.
-- `logo_file_id` sudah tersedia sejak migration organisasi; validasi kepemilikan file dilakukan service.

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS bank_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS bank_account_number VARCHAR(100),
    ADD COLUMN IF NOT EXISTS bank_account_holder VARCHAR(255),
    ADD COLUMN IF NOT EXISTS max_upload_size_bytes BIGINT NOT NULL DEFAULT 10485760
        CHECK (max_upload_size_bytes BETWEEN 1 AND 52428800),
    ADD COLUMN IF NOT EXISTS default_letter_number_pattern VARCHAR(100) NOT NULL
        DEFAULT 'SK/{YEAR}/{MONTH}/{SEQ}',
    ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS actor_role_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS before_data JSONB,
    ADD COLUMN IF NOT EXISTS after_data JSONB,
    ADD COLUMN IF NOT EXISTS request_id VARCHAR(100);

CREATE OR REPLACE FUNCTION enforce_audit_logs_immutable()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is append-only and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_audit_logs_no_update ON audit_logs;

CREATE TRIGGER trg_audit_logs_no_update
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION enforce_audit_logs_immutable();