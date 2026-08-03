-- Rollback Epic 12.
-- Data pada kolom yang dihapus tidak dapat dipulihkan.

DROP TRIGGER IF EXISTS trg_audit_logs_no_update ON audit_logs;
DROP FUNCTION IF EXISTS enforce_audit_logs_immutable();

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS actor_role_codes,
    DROP COLUMN IF EXISTS before_data,
    DROP COLUMN IF EXISTS after_data,
    DROP COLUMN IF EXISTS request_id;

ALTER TABLE organizations
    DROP COLUMN IF EXISTS bank_name,
    DROP COLUMN IF EXISTS bank_account_number,
    DROP COLUMN IF EXISTS bank_account_holder,
    DROP COLUMN IF EXISTS max_upload_size_bytes,
    DROP COLUMN IF EXISTS default_letter_number_pattern,
    DROP COLUMN IF EXISTS settings;