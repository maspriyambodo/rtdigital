-- Epic 13: Master data.
-- Kolom teks legacy dipertahankan selama transisi agar rollback dan data tak dikenal aman.

CREATE TABLE complaint_categories (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL CHECK (code <> ''),
    name VARCHAR(100) NOT NULL CHECK (name <> ''),
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_complaint_categories_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_complaint_categories_organization_code UNIQUE (organization_id, code)
);

CREATE INDEX idx_complaint_categories_organization_status_name
    ON complaint_categories (organization_id, status, name);

CREATE TRIGGER trg_complaint_categories_updated_at
    BEFORE UPDATE ON complaint_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Kategori awal tersedia bagi setiap organisasi, termasuk organisasi baru yang
-- dibuat setelah migration ini melalui service provisioning.
WITH definitions (code, name) AS (
    VALUES
        ('keamanan', 'Keamanan'),
        ('kebersihan', 'Kebersihan'),
        ('infrastruktur', 'Infrastruktur'),
        ('fasilitas_umum', 'Fasilitas umum'),
        ('lainnya', 'Lainnya')
)
INSERT INTO complaint_categories (id, organization_id, code, name)
SELECT gen_random_uuid(), o.id, d.code, d.name
FROM organizations o
CROSS JOIN definitions d;

-- Nilai historis di luar kategori awal tetap dipertahankan sebagai kategori
-- tenant sendiri. Hash mencegah benturan kode dari ejaan legacy berbeda.
INSERT INTO complaint_categories (id, organization_id, code, name)
SELECT gen_random_uuid(),
       c.organization_id,
       'legacy-' || substr(md5(lower(btrim(c.category))), 1, 12),
       btrim(c.category)
FROM complaints c
WHERE NOT EXISTS (
    SELECT 1
    FROM complaint_categories cc
    WHERE cc.organization_id = c.organization_id
      AND lower(cc.name) = lower(btrim(c.category))
)
GROUP BY c.organization_id, btrim(c.category);

ALTER TABLE complaints
    ADD COLUMN complaint_category_id UUID;

UPDATE complaints c
SET complaint_category_id = cc.id
FROM complaint_categories cc
WHERE cc.organization_id = c.organization_id
  AND lower(cc.name) = lower(btrim(c.category));

ALTER TABLE complaints
    ALTER COLUMN complaint_category_id SET NOT NULL,
    ADD CONSTRAINT fk_complaints_category_tenant
        FOREIGN KEY (organization_id, complaint_category_id)
        REFERENCES complaint_categories (organization_id, id)
        ON DELETE RESTRICT;

CREATE INDEX idx_complaints_organization_category_created_at
    ON complaints (organization_id, complaint_category_id, created_at DESC);

CREATE TABLE education_levels (
    id UUID PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE CHECK (code <> ''),
    name VARCHAR(100) NOT NULL UNIQUE CHECK (name <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO education_levels (id, code, name) VALUES
    (gen_random_uuid(), 'none', 'Tidak Sekolah'),
    (gen_random_uuid(), 'sd', 'SD/Sederajat'),
    (gen_random_uuid(), 'smp', 'SMP/Sederajat'),
    (gen_random_uuid(), 'sma', 'SMA/Sederajat'),
    (gen_random_uuid(), 'diploma', 'Diploma'),
    (gen_random_uuid(), 's1', 'Sarjana (S1)'),
    (gen_random_uuid(), 's2', 'Magister (S2)'),
    (gen_random_uuid(), 's3', 'Doktor (S3)');

CREATE TABLE marital_statuses (
    id UUID PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE CHECK (code <> ''),
    name VARCHAR(100) NOT NULL UNIQUE CHECK (name <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO marital_statuses (id, code, name) VALUES
    (gen_random_uuid(), 'single', 'Belum Kawin'),
    (gen_random_uuid(), 'married', 'Kawin'),
    (gen_random_uuid(), 'divorced', 'Cerai Hidup'),
    (gen_random_uuid(), 'widowed', 'Cerai Mati');

ALTER TABLE residents
    ADD COLUMN education_level_id UUID,
    ADD COLUMN marital_status_id UUID,
    ADD CONSTRAINT fk_residents_education_level
        FOREIGN KEY (education_level_id) REFERENCES education_levels (id) ON DELETE RESTRICT,
    ADD CONSTRAINT fk_residents_marital_status
        FOREIGN KEY (marital_status_id) REFERENCES marital_statuses (id) ON DELETE RESTRICT;

-- Hanya nilai yang dapat dipetakan tanpa ambiguitas dipindahkan. Teks legacy
-- tidak dikenal tetap tersimpan untuk kurasi, tanpa memblokir data warga.
UPDATE residents r
SET education_level_id = el.id
FROM education_levels el
WHERE lower(btrim(r.education)) IN (
    lower(el.code),
    lower(el.name),
    CASE el.code
        WHEN 'none' THEN 'tidak sekolah'
        WHEN 'sd' THEN 'sd'
        WHEN 'smp' THEN 'smp'
        WHEN 'sma' THEN 'sma'
        WHEN 'diploma' THEN 'd1'
        WHEN 's1' THEN 's1'
        WHEN 's2' THEN 's2'
        WHEN 's3' THEN 's3'
    END
);

UPDATE residents r
SET marital_status_id = ms.id
FROM marital_statuses ms
WHERE lower(btrim(r.marital_status)) IN (
    lower(ms.code),
    lower(ms.name),
    CASE ms.code
        WHEN 'single' THEN 'belum menikah'
        WHEN 'married' THEN 'menikah'
        WHEN 'divorced' THEN 'cerai'
        WHEN 'widowed' THEN 'janda/duda'
    END
);

CREATE INDEX idx_residents_education_level_id
    ON residents (education_level_id)
    WHERE education_level_id IS NOT NULL;

CREATE INDEX idx_residents_marital_status_id
    ON residents (marital_status_id)
    WHERE marital_status_id IS NOT NULL;