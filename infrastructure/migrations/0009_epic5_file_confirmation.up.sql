-- Epic 5: Objek file hanya dapat digunakan setelah diverifikasi di object storage.

ALTER TABLE file_objects
    ADD COLUMN confirmed_at TIMESTAMPTZ;

CREATE INDEX idx_file_objects_organization_confirmed_at
    ON file_objects (organization_id, confirmed_at)
    WHERE deleted_at IS NULL;