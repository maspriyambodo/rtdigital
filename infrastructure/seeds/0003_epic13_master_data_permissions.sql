-- Epic 13: RBAC kategori aduan.
WITH definitions (code, description) AS (
    VALUES
        ('complaint_category.read', 'Melihat kategori aduan'),
        ('complaint_category.create', 'Membuat kategori aduan'),
        ('complaint_category.update', 'Mengubah atau menonaktifkan kategori aduan')
),
inserted_permissions AS (
    INSERT INTO permissions (id, code, description)
    SELECT gen_random_uuid(), code, description
    FROM definitions
    ON CONFLICT (code) DO UPDATE SET description = EXCLUDED.description
    RETURNING id, code
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
    'complaint_category.read',
    'complaint_category.create',
    'complaint_category.update'
)
WHERE r.organization_id IS NULL
  AND r.code IN ('ketua_rt', 'sekretaris')
ON CONFLICT DO NOTHING;