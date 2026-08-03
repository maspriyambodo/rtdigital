DROP INDEX IF EXISTS idx_complaint_comments_organization_complaint_created_at;
DROP TABLE IF EXISTS complaint_comments;

DROP TRIGGER IF EXISTS trg_complaints_updated_at ON complaints;
DROP INDEX IF EXISTS idx_complaints_organization_assigned_to_status;
DROP INDEX IF EXISTS idx_complaints_organization_reporter_created_at;
DROP INDEX IF EXISTS idx_complaints_organization_status_created_at;
DROP TABLE IF EXISTS complaints;