-- Revert Epic 14: Otomatisasi dan Layanan Operasional Proaktif.
-- Hanya gunakan pada lingkungan tanpa data Epic 14 yang perlu dipertahankan.

DROP TRIGGER IF EXISTS trg_office_handovers_updated_at ON office_handovers;
DROP TABLE IF EXISTS office_handovers;

DROP INDEX IF EXISTS idx_public_cash_summaries_organization_period;
DROP TABLE IF EXISTS public_cash_summary_categories;
DROP TABLE IF EXISTS public_cash_summaries;
DROP TABLE IF EXISTS cash_publication_categories;
DROP TABLE IF EXISTS cash_publication_policies;

DROP INDEX IF EXISTS idx_households_organization_domicile_review_due_at;
ALTER TABLE households
    DROP COLUMN IF EXISTS domicile_review_due_at,
    DROP COLUMN IF EXISTS domicile_last_confirmed_at;

DROP INDEX IF EXISTS idx_complaints_organization_resolution_due_at;
DROP INDEX IF EXISTS idx_complaints_organization_response_due_at;
DROP INDEX IF EXISTS idx_complaint_events_organization_complaint_created_at;
DROP TABLE IF EXISTS complaint_events;
ALTER TABLE complaints
    DROP COLUMN IF EXISTS response_due_at,
    DROP COLUMN IF EXISTS responded_at,
    DROP COLUMN IF EXISTS resolution_due_at,
    DROP COLUMN IF EXISTS reporter_confirmation_due_at,
    DROP COLUMN IF EXISTS reporter_confirmed_at,
    DROP COLUMN IF EXISTS closure_reason,
    DROP COLUMN IF EXISTS closed_by;
ALTER TABLE complaint_categories
    DROP COLUMN IF EXISTS target_response_hours,
    DROP COLUMN IF EXISTS target_resolution_hours,
    DROP COLUMN IF EXISTS target_reporter_confirmation_hours;

DROP INDEX IF EXISTS idx_letter_requests_organization_sla_due_at;
ALTER TABLE letter_requests
    DROP CONSTRAINT IF EXISTS chk_letter_requests_cancellation_data,
    DROP CONSTRAINT IF EXISTS fk_letter_requests_cancelled_by_tenant,
    DROP CONSTRAINT IF EXISTS uq_letter_requests_public_verification_code,
    DROP COLUMN IF EXISTS sla_due_at,
    DROP COLUMN IF EXISTS sla_escalated_at,
    DROP COLUMN IF EXISTS public_verification_code,
    DROP COLUMN IF EXISTS cancelled_by,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS cancellation_reason;
ALTER TABLE letter_types
    DROP COLUMN IF EXISTS sla_hours;

DROP INDEX IF EXISTS idx_payment_allocations_organization_invoice;
DROP TABLE IF EXISTS payment_allocations;
ALTER TABLE payments
    ALTER COLUMN invoice_id SET NOT NULL;

DROP INDEX IF EXISTS idx_invoice_reminder_deliveries_organization_created_at;
DROP TABLE IF EXISTS invoice_reminder_deliveries;
DROP TABLE IF EXISTS notification_preferences;

DROP INDEX IF EXISTS idx_invoice_generation_runs_organization_status_started_at;
DROP TABLE IF EXISTS invoice_generation_runs;
ALTER TABLE due_types
    DROP COLUMN IF EXISTS automatic_generation_enabled,
    DROP COLUMN IF EXISTS generation_lead_days,
    DROP COLUMN IF EXISTS reminder_lead_days;
