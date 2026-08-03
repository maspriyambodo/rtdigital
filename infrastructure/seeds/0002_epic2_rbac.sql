-- Epic 2: system RBAC seed.
-- Idempotent. Scope ownership/assignment tetap diperiksa service.

WITH definitions(code, description) AS (
    VALUES
        ('organization.create', 'Membuat organisasi'),
        ('organization.read', 'Melihat organisasi'),
        ('organization.update', 'Mengubah organisasi'),
        ('organization.deactivate', 'Menonaktifkan organisasi'),
        ('user.invite', 'Mengundang pengguna'),
        ('user.read', 'Melihat pengguna'),
        ('user.update', 'Mengubah pengguna'),
        ('user.deactivate', 'Menonaktifkan pengguna'),
        ('role.assign', 'Menetapkan peran'),
        ('role.revoke', 'Mencabut peran'),
        ('audit.read', 'Membaca audit log'),
        ('audit.export', 'Mengekspor audit log'),
        ('house_unit.read', 'Melihat rumah/unit'),
        ('house_unit.create', 'Membuat rumah/unit'),
        ('house_unit.update', 'Mengubah rumah/unit'),
        ('house_unit.deactivate', 'Menonaktifkan rumah/unit'),
        ('household.read', 'Melihat keluarga'),
        ('household.create', 'Membuat keluarga'),
        ('household.update', 'Mengubah keluarga'),
        ('household.deactivate', 'Menonaktifkan keluarga'),
        ('household.verify', 'Memverifikasi keluarga'),
        ('household.export', 'Mengekspor keluarga'),
        ('resident.read', 'Melihat warga'),
        ('resident.create', 'Membuat warga'),
        ('resident.update', 'Mengubah warga'),
        ('resident.deactivate', 'Menonaktifkan warga'),
        ('resident.verify', 'Memverifikasi warga'),
        ('resident.export', 'Mengekspor warga'),
        ('resident.read_sensitive', 'Membuka data sensitif warga'),
        ('resident.correction.submit', 'Mengajukan koreksi warga'),
        ('resident.correction.review', 'Meninjau koreksi warga'),
        ('announcement.read', 'Melihat pengumuman'),
        ('announcement.create', 'Membuat pengumuman'),
        ('announcement.update', 'Mengubah pengumuman'),
        ('announcement.archive', 'Mengarsipkan pengumuman'),
        ('event.read', 'Melihat agenda'),
        ('event.create', 'Membuat agenda'),
        ('event.update', 'Mengubah agenda'),
        ('event.cancel', 'Membatalkan agenda'),
        ('notification.read_self', 'Melihat notifikasi sendiri'),
        ('notification.mark_read_self', 'Menandai notifikasi sendiri dibaca'),
        ('due_type.read', 'Melihat jenis iuran'),
        ('due_type.create', 'Membuat jenis iuran'),
        ('due_type.update', 'Mengubah jenis iuran'),
        ('due_type.deactivate', 'Menonaktifkan jenis iuran'),
        ('invoice.read', 'Melihat tagihan'),
        ('invoice.create', 'Membuat tagihan'),
        ('invoice.update', 'Mengubah tagihan'),
        ('invoice.cancel', 'Membatalkan tagihan'),
        ('invoice.export', 'Mengekspor tagihan'),
        ('payment.read', 'Melihat pembayaran'),
        ('payment.submit', 'Mengirim pembayaran'),
        ('payment.verify', 'Memverifikasi pembayaran'),
        ('payment.reject', 'Menolak pembayaran'),
        ('payment.cancel', 'Membatalkan pembayaran'),
        ('cash.read', 'Melihat kas'),
        ('cash.create', 'Mencatat kas'),
        ('cash.update', 'Mengubah metadata kas'),
        ('cash.reverse', 'Membuat transaksi pembalik kas'),
        ('finance.export', 'Mengekspor laporan keuangan'),
        ('letter_type.read', 'Melihat jenis surat'),
        ('letter_type.create', 'Membuat jenis surat'),
        ('letter_type.update', 'Mengubah jenis surat'),
        ('letter_type.deactivate', 'Menonaktifkan jenis surat'),
        ('letter_request.read', 'Melihat pengajuan surat'),
        ('letter_request.submit', 'Mengajukan surat'),
        ('letter_request.process', 'Memproses surat'),
        ('letter_request.request_revision', 'Meminta revisi surat'),
        ('letter_request.approve', 'Menyetujui surat'),
        ('letter_request.issue', 'Menerbitkan surat'),
        ('letter_request.download', 'Mengunduh surat'),
        ('complaint.read', 'Melihat aduan'),
        ('complaint.submit', 'Mengajukan aduan'),
        ('complaint.assign', 'Menugaskan aduan'),
        ('complaint.update_status', 'Memperbarui status aduan'),
        ('complaint.comment', 'Mengomentari aduan'),
        ('complaint.export', 'Mengekspor aduan')
)
INSERT INTO permissions (id, code, description)
SELECT gen_random_uuid(), code, description FROM definitions
ON CONFLICT (code) DO UPDATE SET description = EXCLUDED.description;

WITH definitions(code, name, description) AS (
    VALUES
        ('super_admin', 'Super Admin', 'Administrator teknis platform'),
        ('ketua_rt', 'Ketua RT', 'Penanggung jawab operasional RT'),
        ('sekretaris', 'Sekretaris', 'Administrasi dan persuratan'),
        ('bendahara', 'Bendahara', 'Keuangan dan iuran'),
        ('pengurus', 'Pengurus', 'Pelaksana tugas sesuai penugasan'),
        ('warga', 'Warga', 'Akses layanan diri/keluarga')
)
INSERT INTO roles (id, organization_id, code, name, description)
SELECT gen_random_uuid(), NULL, code, name, description FROM definitions
ON CONFLICT (code) WHERE organization_id IS NULL
DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description;

-- Reset hanya mapping peran sistem; custom role organisasi tidak tersentuh.
DELETE FROM role_permissions rp
USING roles r
WHERE rp.role_id = r.id AND r.organization_id IS NULL;

WITH mappings(role_code, permission_code) AS (
    VALUES
        ('super_admin', 'organization.create'), ('super_admin', 'organization.read'),
        ('super_admin', 'organization.update'), ('super_admin', 'organization.deactivate'),
        ('super_admin', 'user.invite'), ('super_admin', 'user.read'), ('super_admin', 'user.update'),
        ('super_admin', 'user.deactivate'), ('super_admin', 'role.assign'), ('super_admin', 'role.revoke'),
        ('super_admin', 'audit.read'), ('super_admin', 'audit.export'),

        ('ketua_rt', 'organization.read'), ('ketua_rt', 'organization.update'),
        ('ketua_rt', 'user.read'), ('ketua_rt', 'role.assign'), ('ketua_rt', 'role.revoke'),
        ('ketua_rt', 'audit.read'), ('ketua_rt', 'audit.export'),
        ('ketua_rt', 'house_unit.read'), ('ketua_rt', 'household.read'), ('ketua_rt', 'household.export'),
        ('ketua_rt', 'resident.read'), ('ketua_rt', 'resident.export'), ('ketua_rt', 'resident.correction.review'),
        ('ketua_rt', 'announcement.read'), ('ketua_rt', 'announcement.create'), ('ketua_rt', 'announcement.update'), ('ketua_rt', 'announcement.archive'),
        ('ketua_rt', 'event.read'), ('ketua_rt', 'event.create'), ('ketua_rt', 'event.update'), ('ketua_rt', 'event.cancel'),
        ('ketua_rt', 'notification.read_self'), ('ketua_rt', 'notification.mark_read_self'),
        ('ketua_rt', 'due_type.read'), ('ketua_rt', 'invoice.read'), ('ketua_rt', 'invoice.export'),
        ('ketua_rt', 'payment.read'), ('ketua_rt', 'cash.read'), ('ketua_rt', 'finance.export'),
        ('ketua_rt', 'letter_type.read'), ('ketua_rt', 'letter_request.read'), ('ketua_rt', 'letter_request.approve'),
        ('ketua_rt', 'complaint.read'), ('ketua_rt', 'complaint.assign'), ('ketua_rt', 'complaint.update_status'), ('ketua_rt', 'complaint.comment'), ('ketua_rt', 'complaint.export'),

        ('sekretaris', 'organization.read'), ('sekretaris', 'organization.update'),
        ('sekretaris', 'user.invite'), ('sekretaris', 'user.read'), ('sekretaris', 'user.update'), ('sekretaris', 'audit.read'),
        ('sekretaris', 'house_unit.create'), ('sekretaris', 'house_unit.read'), ('sekretaris', 'house_unit.update'), ('sekretaris', 'house_unit.deactivate'),
        ('sekretaris', 'household.create'), ('sekretaris', 'household.read'), ('sekretaris', 'household.update'), ('sekretaris', 'household.deactivate'), ('sekretaris', 'household.verify'), ('sekretaris', 'household.export'),
        ('sekretaris', 'resident.create'), ('sekretaris', 'resident.read'), ('sekretaris', 'resident.update'), ('sekretaris', 'resident.deactivate'), ('sekretaris', 'resident.verify'), ('sekretaris', 'resident.export'), ('sekretaris', 'resident.read_sensitive'), ('sekretaris', 'resident.correction.review'),
        ('sekretaris', 'announcement.read'), ('sekretaris', 'announcement.create'), ('sekretaris', 'announcement.update'), ('sekretaris', 'announcement.archive'),
        ('sekretaris', 'event.read'), ('sekretaris', 'event.create'), ('sekretaris', 'event.update'), ('sekretaris', 'event.cancel'),
        ('sekretaris', 'notification.read_self'), ('sekretaris', 'notification.mark_read_self'),
        ('sekretaris', 'due_type.read'), ('sekretaris', 'invoice.read'), ('sekretaris', 'payment.read'), ('sekretaris', 'cash.read'),
        ('sekretaris', 'letter_type.read'), ('sekretaris', 'letter_type.create'), ('sekretaris', 'letter_type.update'), ('sekretaris', 'letter_type.deactivate'),
        ('sekretaris', 'letter_request.read'), ('sekretaris', 'letter_request.process'), ('sekretaris', 'letter_request.request_revision'), ('sekretaris', 'letter_request.issue'), ('sekretaris', 'letter_request.download'),
        ('sekretaris', 'complaint.read'), ('sekretaris', 'complaint.assign'), ('sekretaris', 'complaint.update_status'), ('sekretaris', 'complaint.comment'),

        ('bendahara', 'organization.read'), ('bendahara', 'user.read'), ('bendahara', 'audit.read'),
        ('bendahara', 'house_unit.read'), ('bendahara', 'household.read'), ('bendahara', 'resident.read'),
        ('bendahara', 'announcement.read'), ('bendahara', 'event.read'), ('bendahara', 'notification.read_self'), ('bendahara', 'notification.mark_read_self'),
        ('bendahara', 'due_type.read'), ('bendahara', 'due_type.create'), ('bendahara', 'due_type.update'), ('bendahara', 'due_type.deactivate'),
        ('bendahara', 'invoice.read'), ('bendahara', 'invoice.create'), ('bendahara', 'invoice.update'), ('bendahara', 'invoice.cancel'), ('bendahara', 'invoice.export'),
        ('bendahara', 'payment.read'), ('bendahara', 'payment.verify'), ('bendahara', 'payment.reject'), ('bendahara', 'payment.cancel'),
        ('bendahara', 'cash.read'), ('bendahara', 'cash.create'), ('bendahara', 'cash.update'), ('bendahara', 'cash.reverse'), ('bendahara', 'finance.export'),
        ('bendahara', 'complaint.read'), ('bendahara', 'complaint.comment'),

        ('pengurus', 'organization.read'), ('pengurus', 'announcement.read'), ('pengurus', 'event.read'),
        ('pengurus', 'notification.read_self'), ('pengurus', 'notification.mark_read_self'),
        ('pengurus', 'complaint.read'), ('pengurus', 'complaint.update_status'), ('pengurus', 'complaint.comment'),

        ('warga', 'organization.read'), ('warga', 'user.update'),
        ('warga', 'house_unit.read'), ('warga', 'household.read'), ('warga', 'resident.read'), ('warga', 'resident.correction.submit'),
        ('warga', 'announcement.read'), ('warga', 'event.read'), ('warga', 'notification.read_self'), ('warga', 'notification.mark_read_self'),
        ('warga', 'due_type.read'), ('warga', 'invoice.read'), ('warga', 'payment.read'), ('warga', 'payment.submit'), ('warga', 'cash.read'),
        ('warga', 'letter_request.read'), ('warga', 'letter_request.submit'), ('warga', 'letter_request.download'),
        ('warga', 'complaint.read'), ('warga', 'complaint.submit'), ('warga', 'complaint.comment')
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM mappings m
JOIN roles r ON r.code = m.role_code AND r.organization_id IS NULL
JOIN permissions p ON p.code = m.permission_code;