DROP TRIGGER IF EXISTS trg_letter_requests_updated_at ON letter_requests;
DROP INDEX IF EXISTS idx_letter_requests_org_resident_created_at;
DROP INDEX IF EXISTS idx_letter_requests_org_requester_created_at;
DROP INDEX IF EXISTS idx_letter_requests_org_status_created_at;
DROP INDEX IF EXISTS uq_letter_requests_organization_letter_number;
DROP TABLE IF EXISTS letter_requests;

DROP TRIGGER IF EXISTS trg_letter_types_updated_at ON letter_types;
DROP INDEX IF EXISTS idx_letter_types_organization_status;
DROP TABLE IF EXISTS letter_types;