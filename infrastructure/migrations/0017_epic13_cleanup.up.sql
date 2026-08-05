-- Epic 13 cleanup.
-- Jangan hilangkan teks legacy bila ada nilai warga yang belum termigrasi
-- ke master global. Kurasi data tersebut terlebih dahulu, lalu jalankan ulang.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM residents
        WHERE education IS NOT NULL
          AND btrim(education) <> ''
          AND education_level_id IS NULL
    ) THEN
        RAISE EXCEPTION
            'cannot drop residents.education: unmapped legacy education values exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM residents
        WHERE marital_status IS NOT NULL
          AND btrim(marital_status) <> ''
          AND marital_status_id IS NULL
    ) THEN
        RAISE EXCEPTION
            'cannot drop residents.marital_status: unmapped legacy marital status values exist';
    END IF;
END $$;

ALTER TABLE complaints
    DROP COLUMN category;

ALTER TABLE residents
    DROP COLUMN education,
    DROP COLUMN marital_status;