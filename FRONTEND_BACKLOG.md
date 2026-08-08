# Web App Backlog (Super Admin & Developer)

**Cakupan:** Web App khusus supporting dan Super Admin. Tidak ada fitur operasional warga atau pengurus RT.

## Epic 0: Infrastruktur dan Fondasi Proyek
- [ ] **Task 0.8:** Buat shell UI: layout super admin, sidebar sistem, error state, offline notice.
- [ ] **Task 0.9:** Implementasi komponen dasar: `Button`, `FormField`, `TextInput`, `Select`, `DatePicker`, `StatusBadge`, `EmptyState`.

## Epic 1: Authentication dan Akun
- [ ] **Task 1.8:** Buat UI login super admin, profil, ganti kata sandi, dan logout.

## Epic 2: Manajemen Sistem dan Tenant
- [ ] **Task 2.7:** Buat UI daftar organisasi/tenant, detail tenant, ubah status, dan kelola admin utama tenant.
- [ ] **Task 2.9:** Implementasi navigasi khusus peran Super Admin/Developer.

## Epic 11: Dashboard Sistem
- [ ] **Task 11.7:** Buat UI dashboard monitoring sistem: total tenant, status API, metrik dasar.
- [ ] **Task 11.8:** Implementasi modul statistik (Power BI style): interactive grid layout, chart metrik sistem agregat (tenant, API usage, error rates), cross-filtering, dan kemampuan ekspor laporan.
- [ ] **Task 11.8:** Implementasi modul statistik (Power BI style): interactive grid layout, chart metrik sistem agregat (tenant, API usage, error rates), cross-filtering, dan kemampuan ekspor laporan.

## Epic 12: Audit dan Observability
- [ ] **Task 12.3:** Buat UI audit log sistem untuk keperluan developer dan investigasi super admin.

## Definition of Done per Task
- [ ] UI difokuskan untuk viewport desktop.
- [ ] Validasi keamanan dan otorisasi Super Admin.
