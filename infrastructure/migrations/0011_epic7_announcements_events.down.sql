-- Rollback Epic 7.

DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS announcement_read_receipts;
DROP TABLE IF EXISTS announcement_targets;
DROP TABLE IF EXISTS announcements;

ALTER TABLE file_attachments
    DROP CONSTRAINT file_attachments_entity_type_check,
    ADD CONSTRAINT file_attachments_entity_type_check
        CHECK (entity_type IN ('announcement', 'complaint', 'letter_request', 'payment', 'cash_transaction'));