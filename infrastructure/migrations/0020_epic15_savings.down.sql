-- Drop triggers
DROP TRIGGER IF EXISTS trg_prevent_savings_transactions_delete ON savings_transactions;
DROP FUNCTION IF EXISTS prevent_savings_transactions_delete();
DROP TRIGGER IF EXISTS trg_savings_transactions_updated_at ON savings_transactions;
DROP TRIGGER IF EXISTS trg_savings_accounts_updated_at ON savings_accounts;
DROP TRIGGER IF EXISTS trg_savings_products_updated_at ON savings_products;

-- Revert file_attachments check constraint
ALTER TABLE file_attachments
    DROP CONSTRAINT file_attachments_entity_type_check,
    ADD CONSTRAINT file_attachments_entity_type_check
        CHECK (entity_type IN ('announcement', 'complaint', 'letter_request', 'payment', 'cash_transaction'));

-- Drop tables
DROP TABLE IF EXISTS savings_transactions;
DROP TABLE IF EXISTS savings_accounts;
DROP TABLE IF EXISTS savings_products;
