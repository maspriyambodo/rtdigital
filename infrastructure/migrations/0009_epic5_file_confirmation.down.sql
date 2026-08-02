-- Epic 5: Rollback konfirmasi objek file.

DROP INDEX IF EXISTS idx_file_objects_organization_confirmed_at;

ALTER TABLE file_objects
    DROP COLUMN IF EXISTS confirmed_at;