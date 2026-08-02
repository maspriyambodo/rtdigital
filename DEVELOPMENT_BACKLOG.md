# Development Backlog RT Digital

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital  
**Referensi:** `PRD.md`, `SCOPE.md`, `TECHNICAL_SPECIFICATION.md`, `DATABASE_DESIGN.md`, `API_SPECIFICATION.md`, `SYSTEM_ARCHITECTURE.md`

Backlog ini memecah MVP menjadi epic dan task yang dapat dikerjakan bertahap. Semua task mengikuti standar mobile-first, RBAC, isolasi `organization_id`, validasi backend, audit untuk tindakan penting, serta pengujian yang relevan.

## Urutan Implementasi

1. Epic 0–3: fondasi, autentikasi, RBAC, data warga.
2. Epic 4–6: iuran, pembayaran, buku kas.
3. Epic 7–9: komunikasi, surat, aduan.
4. Epic 10–12: notifikasi, dashboard/laporan, hardening rilis.

---

## Epic 0: Infrastruktur dan Fondasi Proyek

Tujuan: menyiapkan monorepo, aplikasi dasar, database, CI/CD, serta lingkungan development.

- [x] **Task 0.1:** Scaffold monorepo: `apps/web`, `services/api`, `infrastructure`, pnpm workspace, Go module.
- [x] **Task 0.2:** Inisialisasi Next.js App Router, TypeScript strict, OpenNext, design token CSS, dan PWA manifest.
- [x] **Task 0.3:** Buat Go API modular monolith dengan `net/http`, `ServeMux`, `slog`, request ID, recovery, graceful shutdown.
- [x] **Task 0.4:** Konfigurasi koneksi PostgreSQL `pgxpool`, migration SQL versioned, dan seed data development.
- [x] **Task 0.5:** Konfigurasi Docker Compose untuk `web`, `api`, PostgreSQL, Redis lokal, MinIO, dan inisialisasi bucket.
- [x] **Task 0.6:** Buat konfigurasi Cloudflare R2 S3-compatible, termasuk signed upload/download URL dan MinIO emulator lokal.
- [x] **Task 0.7:** Buat CI minimum: format check, lint, type check, unit test, integration test, build frontend, build image API.
- [x] **Task 0.8:** Buat shell UI: layout warga, bottom navigation, layout pengurus, sidebar/drawer, error state, offline notice.
- [x] **Task 0.9:** Implementasi komponen dasar: `Button`, `FormField`, `TextInput`, `Select`, `DatePicker`, `StatusBadge`, `EmptyState`.

---

## Epic 1: Authentication dan Akun

Tujuan: autentikasi aman menggunakan email/nomor telepon, access token, refresh token, dan lifecycle akun.

- [x] **Task 1.1:** Buat migration `organizations`, `users`, session/refresh token, activation token, dan password reset token.
- [x] **Task 1.2:** Implementasi password hash Argon2id, normalisasi email/telepon, lockout login, dan validasi status akun.
- [x] **Task 1.3:** Implementasi `POST /auth/login`: access JWT, refresh token HttpOnly, session record, audit login.
- [x] **Task 1.4:** Implementasi refresh token rotation, logout perangkat aktif, dan logout seluruh perangkat.
- [x] **Task 1.5:** Implementasi undangan/aktivasi akun dengan token sekali pakai dan kedaluwarsa.
- [x] **Task 1.6:** Implementasi lupa/reset kata sandi menggunakan Resend API.
- [x] **Task 1.7:** Implementasi MFA pengurus dan verifikasi MFA saat login. *(Login serta penetapan peran Super Admin, Ketua RT, Sekretaris, dan Bendahara kini mewajibkan MFA.)*
- [x] **Task 1.8:** Buat UI login, aktivasi akun, lupa/reset kata sandi, profil, ganti kata sandi, dan logout.
- [x] **Task 1.9:** Tambahkan test autentikasi: lockout, token expired, refresh rotation, logout, dan akun nonaktif.


---

## Epic 2: Manajemen Pengguna dan RBAC

Tujuan: mengelola pengguna, peran, permission, serta scope akses organisasi.

- [x] **Task 2.1:** Buat migration `roles`, `permissions`, `user_roles`, dan `role_permissions`.
- [x] **Task 2.2:** Seed peran sistem serta permission sesuai `USER_ROLES_AND_PERMISSIONS.md`.
- [x] **Task 2.3:** Implementasi middleware authentication, permission check, dan principal context.
- [x] **Task 2.4:** Implementasi pembatasan `organization_id`, assignment, serta separation of duties. *(Ownership keluarga diterapkan pada domain keluarga Epic 3.)*
- [x] **Task 2.5:** Implementasi API daftar pengguna, detail pengguna, undangan, perubahan status, dan penonaktifan.
- [x] **Task 2.6:** Implementasi API tambah/cabut peran dengan larangan eskalasi hak akses sendiri serta guard MFA.
- [x] **Task 2.7:** Buat UI daftar pengguna, detail akun, undang pengguna, ubah status, dan pengelolaan peran.
- [x] **Task 2.8:** Tambahkan authorization test untuk peran utama, isolasi tenant, MFA, dan separation of duties.

---

## Epic 3: Data Keluarga dan Warga

Tujuan: pendataan rumah, keluarga, warga, serta koreksi data dengan perlindungan data sensitif.

- [x] **Task 3.1:** Buat migration `house_units`, `households`, `residents`, dan `household_members` beserta constraint bisnis.
- [x] **Task 3.2:** Implementasi enkripsi NIK/nomor KK, blind index, masking respons API, dan audit akses sensitif.
- [x] **Task 3.3:** Implementasi API CRUD rumah/unit serta penonaktifan unit.
- [x] **Task 3.4:** Implementasi API CRUD keluarga, penetapan kepala keluarga, dan riwayat anggota keluarga.
- [x] **Task 3.5:** Implementasi API CRUD warga, pencarian, filter status, dan verifikasi pengurus.
- [x] **Task 3.6:** Implementasi pengajuan koreksi data warga, review, setujui, tolak, dan minta revisi.
- [x] **Task 3.7:** Implementasi dry-run validasi CSV, deteksi duplikat, import data awal, dan audit import.
- [x] **Task 3.8:** Buat UI pengurus untuk rumah/unit, keluarga, warga, filter, pencarian, dan detail data.
- [x] **Task 3.9:** Buat UI warga untuk profil keluarga, anggota keluarga, dan koreksi data.
- [x] **Task 3.10:** Tambahkan test constraint kepala keluarga, anggota aktif, masking, dan isolasi tenant.

---

## Epic 4: Iuran dan Tagihan

Tujuan: membuat jenis iuran serta menerbitkan tagihan individual atau massal.

- [x] **Task 4.1:** Buat migration `due_types` dan `invoices` beserta index, status, dan constraint nominal.
- [x] **Task 4.2:** Implementasi API CRUD jenis iuran dan penonaktifan.
- [x] **Task 4.3:** Implementasi pembuatan tagihan individual dengan nomor tagihan unik.
- [x] **Task 4.4:** Implementasi pembuatan tagihan massal dengan idempotency key, validasi sasaran, dan ringkasan hasil.
- [x] **Task 4.5:** Implementasi penyesuaian/diskon dan pembatalan tagihan dengan alasan serta audit log.
- [x] **Task 4.6:** Implementasi daftar tagihan, detail, tunggakan, filter periode, dan scope keluarga.
- [x] **Task 4.7:** Buat UI pengurus untuk jenis iuran, pembuatan tagihan, daftar tagihan, dan tunggakan.
- [x] **Task 4.8:** Buat UI warga untuk tagihan aktif, detail tagihan, dan riwayat tagihan.
- [x] **Task 4.9:** Tambahkan test pembuatan massal, idempotency, status tagihan, dan scope warga.

---

## Epic 5: Pembayaran

Tujuan: warga mengirim pembayaran manual; bendahara memverifikasi dengan alur aman dan dapat diaudit.

- [x] **Task 5.1:** Buat migration `file_objects`, `file_attachments`, dan `payments`.
- [x] **Task 5.2:** Implementasi endpoint presign upload, konfirmasi upload, validasi MIME/ukuran/purpose, dan signed download Cloudflare R2.
- [x] **Task 5.3:** Implementasi `FileUploader`: kamera/galeri, validasi lokal, progress, retry, hapus, dan fallback error.
- [x] **Task 5.4:** Implementasi API submit pembayaran tunai/transfer dengan `Idempotency-Key`.
- [x] **Task 5.5:** Implementasi API verifikasi, penolakan dengan alasan wajib, dan pembatalan pembayaran.
- [x] **Task 5.6:** Implementasi transaksi atomik: lock payment/invoice, update status invoice, dan audit. *(Pengiriman notifikasi ditunda ke Epic 10 karena layanan notifikasi belum tersedia.)*
- [x] **Task 5.7:** Buat UI warga untuk lapor pembayaran, unggah bukti, riwayat, status, dan tanda terima.
- [x] **Task 5.8:** Buat UI bendahara untuk antrean verifikasi, detail bukti, terima, tolak, dan batal.
- [x] **Task 5.9:** Tambahkan test concurrency, idempotency, pemisahan tugas, dan status pembayaran/tagihan.

---

## Epic 6: Buku Kas

Tujuan: mencatat pemasukan/pengeluaran RT tanpa menghapus riwayat transaksi.

- [x] **Task 6.1:** Buat migration `cash_categories` dan `cash_transactions`.
- [x] **Task 6.2:** Implementasi pencatatan pemasukan kas otomatis saat pembayaran diverifikasi.
- [x] **Task 6.3:** Implementasi API CRUD kategori kas.
- [x] **Task 6.4:** Implementasi API transaksi kas manual dengan validasi nominal, kategori, bukti, dan audit.
- [x] **Task 6.5:** Implementasi transaksi pembalik untuk koreksi; larang penghapusan transaksi historis.
- [x] **Task 6.6:** Implementasi API buku kas, saldo berjalan, filter periode/kategori, dan detail transaksi.
- [x] **Task 6.7:** Buat UI bendahara untuk kategori kas, catat transaksi, buku kas, dan pembalikan.
- [x] **Task 6.8:** Tambahkan test relasi payment-kas, pembalikan, saldo, dan larangan delete.

---

## Epic 7: Pengumuman dan Agenda

Tujuan: kanal informasi RT dan agenda kegiatan yang mudah dibaca dari perangkat seluler.

- [ ] **Task 7.1:** Buat migration `announcements`, `announcement_targets`, dan `events`.
- [ ] **Task 7.2:** Implementasi API pengumuman: draft, target, jadwal, publish, archive, dan statistik baca.
- [ ] **Task 7.3:** Implementasi API agenda: buat, ubah, batal, status, dan lampiran.
- [ ] **Task 7.4:** Implementasi seleksi target pengumuman berdasarkan seluruh warga, peran, keluarga, atau unit.
- [ ] **Task 7.5:** Buat UI pengurus untuk pengumuman, agenda, penjadwalan, target, dan arsip.
- [ ] **Task 7.6:** Buat UI warga untuk daftar/detail pengumuman, agenda mendatang, dan penanda penting.
- [ ] **Task 7.7:** Tambahkan test target visibility, status jadwal, dan akses lampiran.

---

## Epic 8: Surat Pengantar

Tujuan: pengajuan surat dari formulir sampai PDF terbit dan dapat diunduh.

- [ ] **Task 8.1:** Buat migration `letter_types` dan `letter_requests`.
- [ ] **Task 8.2:** Implementasi API jenis surat: requirements, schema formulir, template, pola nomor, dan status.
- [ ] **Task 8.3:** Implementasi validasi form dinamis, lampiran wajib, dan pengajuan surat oleh warga.
- [ ] **Task 8.4:** Implementasi workflow surat: draft, diajukan, review, revisi, persetujuan, penolakan, penerbitan, pembatalan.
- [ ] **Task 8.5:** Implementasi nomor surat unik, generator PDF, penyimpanan PDF di R2, dan signed download.
- [ ] **Task 8.6:** Buat UI warga untuk pilih jenis surat, form bertahap, lampiran, status, revisi, dan unduh PDF.
- [ ] **Task 8.7:** Buat UI sekretaris/ketua RT untuk antrean, review, catatan internal, setujui, tolak, dan terbitkan.
- [ ] **Task 8.8:** Tambahkan test transisi status, kelengkapan lampiran, nomor surat, dan authorization.

---

## Epic 9: Aduan Warga

Tujuan: tiket aduan dengan status, penanggung jawab, komentar, serta riwayat penyelesaian.

- [ ] **Task 9.1:** Buat migration `complaints` dan `complaint_comments`.
- [ ] **Task 9.2:** Implementasi API pembuatan aduan: kategori, lokasi umum, prioritas, lampiran, dan nomor tiket.
- [ ] **Task 9.3:** Implementasi API daftar/detail aduan dengan scope warga, pengurus, serta petugas tertugaskan.
- [ ] **Task 9.4:** Implementasi penugasan petugas, perubahan status, catatan resolusi, dan penutupan aduan.
- [ ] **Task 9.5:** Implementasi komentar/pembaruan, pemisahan komentar internal, serta notifikasi perubahan status.
- [ ] **Task 9.6:** Buat UI warga untuk buat aduan, timeline, komentar, dan status aduan saya.
- [ ] **Task 9.7:** Buat UI pengurus untuk antrean aduan, filter, assignment, pembaruan status, dan resolusi.
- [ ] **Task 9.8:** Tambahkan test scope pelapor, komentar internal, state transition, dan assignment.

---

## Epic 10: Notifikasi

Tujuan: notifikasi dalam aplikasi, email Resend, dan WhatsApp SaungWA untuk aktivitas penting.

- [ ] **Task 10.1:** Buat migration `notifications` dan API daftar notifikasi, tandai dibaca, serta tandai semua dibaca.
- [ ] **Task 10.2:** Buat komponen notifikasi warga: indikator belum dibaca, daftar, detail, dan empty state.
- [ ] **Task 10.3:** Implementasi adapter Resend untuk email transaksional dan template pesan dasar.
- [ ] **Task 10.4:** Implementasi adapter SaungWA untuk notifikasi WhatsApp dan validasi konfigurasi provider.
- [ ] **Task 10.5:** Tambahkan trigger notifikasi untuk undangan, reset password, tagihan, pembayaran, surat, aduan, dan pengumuman penting.
- [ ] **Task 10.6:** Pastikan kegagalan provider dicatat tanpa membatalkan transaksi utama.
- [ ] **Task 10.7:** Tambahkan mekanisme retry terkontrol jika kebutuhan reliabilitas terbukti.
- [ ] **Task 10.8:** Tambahkan test payload provider, kegagalan provider, dan permission notifikasi.

---

## Epic 11: Dashboard dan Laporan

Tujuan: ringkasan operasional yang relevan serta ekspor laporan sesuai permission.

- [ ] **Task 11.1:** Implementasi API dashboard warga: tagihan aktif, pembayaran terbaru, surat, aduan, pengumuman, agenda.
- [ ] **Task 11.2:** Implementasi API dashboard pengurus: keluarga/warga aktif, tagihan, pembayaran, tunggakan, kas, surat, aduan.
- [ ] **Task 11.3:** Implementasi endpoint laporan keluarga/warga, mutasi warga, tagihan, tunggakan, pembayaran, kas, surat, dan aduan.
- [ ] **Task 11.4:** Implementasi ekspor CSV sesuai filter, permission, scope, dan audit log.
- [ ] **Task 11.5:** Implementasi ekspor PDF laporan formal bila formatnya telah disetujui.
- [ ] **Task 11.6:** Buat UI dashboard warga mobile-first.
- [ ] **Task 11.7:** Buat UI dashboard pengurus, laporan, filter periode, dan ekspor.
- [ ] **Task 11.8:** Tambahkan test akurasi agregat, scope ekspor, dan audit ekspor.

---

## Epic 12: Pengaturan, Audit, Hardening, dan Rilis

Tujuan: menyelesaikan kesiapan operasional, keamanan, observabilitas, serta UAT sebelum peluncuran.

- [ ] **Task 12.1:** Implementasi pengaturan RT: identitas, logo, rekening, zona waktu, nomor surat, batas unggahan, dan template.
- [ ] **Task 12.2:** Buat migration dan layanan `audit_logs` append-only untuk autentikasi, data sensitif, keuangan, surat, ekspor, peran, dan konfigurasi.
- [ ] **Task 12.3:** Buat UI audit log dengan filter dan detail yang disanitasi.
- [ ] **Task 12.4:** Terapkan CORS, CSRF sesuai mekanisme cookie, rate limit, security headers, body limit, dan sanitasi log.
- [ ] **Task 12.5:** Konfigurasi monitoring: health/readiness, CloudWatch log/metric/alarm, request ID, serta error tracking yang lolos evaluasi privasi.
- [ ] **Task 12.6:** Konfigurasi backup RDS, retensi/pemulihan R2, dan uji restore di staging.
- [ ] **Task 12.7:** Jalankan responsive, accessibility, authorization, integration, E2E, security header, dan smoke test staging.
- [ ] **Task 12.8:** Jalankan UAT pengurus, migrasi CSV final, pelatihan, runbook, kebijakan privasi, dan soft launch.

---

## Definition of Done per Task

- [ ] Acceptance criteria task terpenuhi.
- [ ] UI diuji pada viewport 320 px, 360 px, 390 px, lalu desktop bila relevan.
- [ ] Validasi frontend untuk UX dan validasi backend untuk keamanan diterapkan.
- [ ] Authorization, `organization_id`, ownership, dan transisi status diuji bila task privat.
- [ ] Error API memakai format standar `code`, `message`, `details`, `request_id`.
- [ ] Audit log dibuat untuk tindakan penting.
- [ ] Data sensitif tidak masuk log, cache publik, atau penyimpanan browser tidak aman.
- [ ] Unit/integration/E2E test relevan lulus.
- [ ] Dokumentasi API, migration, dan konfigurasi diperbarui bila berubah.