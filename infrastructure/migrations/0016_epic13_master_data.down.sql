-- Rollback Epic 13.
-- Data pada kolom yang dihapus tidak dapat dipulihkan.

DROP INDEX IF EXISTS idx_residents_marital_status_id;
DROP INDEX IF EXISTS idx_residents_education_level_id;

ALTER TABLE residents
    DROP CONSTRAINT IF EXISTS fk_residents_marital_status,
    DROP CONSTRAINT IF EXISTS fk_residents_education_level,
    DROP COLUMN IF EXISTS marital_status_id,
    DROP COLUMN IF EXISTS education_level_id;

DROP TABLE IF EXISTS marital_statuses;
DROP TABLE IF EXISTS education_levels;

DROP INDEX IF EXISTS idx_complaints_organization_category_created_at;

ALTER TABLE complaints
    DROP CONSTRAINT IF EXISTS fk_complaints_category_tenant,
    DROP COLUMN IF EXISTS complaint_category_id;

DROP TRIGGER IF EXISTS trg_complaint_categories_updated_at ON complaint_categories;
DROP INDEX IF EXISTS idx_complaint_categories_organization_status_name;
DROP TABLE IF EXISTS complaint_categories;