-- Rollback Epic 13 cleanup.
-- Nilai teks dipulihkan dari master aktif saat rollback; ejaan legacy asli
-- tidak dapat direkonstruksi setelah kolom legacy dihapus.

ALTER TABLE complaints
    ADD COLUMN IF NOT EXISTS category VARCHAR(100);

UPDATE complaints c
SET category = cc.name
FROM complaint_categories cc
WHERE cc.organization_id = c.organization_id
  AND cc.id = c.complaint_category_id;

ALTER TABLE residents
    ADD COLUMN IF NOT EXISTS education VARCHAR(100),
    ADD COLUMN IF NOT EXISTS marital_status VARCHAR(50);

UPDATE residents r
SET education = el.name
FROM education_levels el
WHERE el.id = r.education_level_id;

UPDATE residents r
SET marital_status = ms.name
FROM marital_statuses ms
WHERE ms.id = r.marital_status_id;