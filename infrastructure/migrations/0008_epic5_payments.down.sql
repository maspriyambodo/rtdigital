-- Epic 5: Rollback pembayaran dan lampiran privat.

DROP TRIGGER IF EXISTS trg_payments_updated_at ON payments;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS file_attachments;
DROP TABLE IF EXISTS file_objects;