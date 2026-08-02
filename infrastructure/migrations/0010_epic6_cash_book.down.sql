-- Epic 6: Rollback buku kas append-only.

DROP TRIGGER IF EXISTS trg_prevent_cash_transactions_delete ON cash_transactions;
DROP FUNCTION IF EXISTS prevent_cash_transactions_delete();
DROP TRIGGER IF EXISTS trg_cash_transactions_updated_at ON cash_transactions;
DROP INDEX IF EXISTS uq_cash_transactions_reversal_of;
DROP INDEX IF EXISTS uq_cash_transactions_active_payment_reference;
DROP INDEX IF EXISTS idx_cash_transactions_organization_reference;
DROP INDEX IF EXISTS idx_cash_transactions_organization_category;
DROP INDEX IF EXISTS idx_cash_transactions_organization_date;
DROP TABLE IF EXISTS cash_transactions;

DROP TRIGGER IF EXISTS trg_cash_categories_updated_at ON cash_categories;
DROP INDEX IF EXISTS idx_cash_categories_organization_type_status;
DROP TABLE IF EXISTS cash_categories;

ALTER TABLE file_attachments
    DROP CONSTRAINT IF EXISTS file_attachments_entity_type_check,
    ADD CONSTRAINT file_attachments_entity_type_check
        CHECK (entity_type IN ('announcement', 'complaint', 'letter_request', 'payment'));
