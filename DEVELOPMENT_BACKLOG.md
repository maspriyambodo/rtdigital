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
5. Epic 13: master data.
6. Epic 14: otomatisasi dan layanan operasional proaktif.
7. Epic 15: tabungan warga (dana titipan non-kas).

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
- [ ] **Task 2.9:** Implementasi menu navigasi dinamis berbasis permission di frontend: muat permission efektif dari `GET /me`, simpan pada `AuthProvider`, filter `PengurusNavigation` dan `WargaNavigation`, serta lindungi rute modul sesuai `INFORMATION_ARCHITECTURE.md`.

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

## [x] Epic 6: Buku Kas

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

## [x] Epic 7: Pengumuman dan Agenda

Tujuan: kanal informasi RT dan agenda kegiatan yang mudah dibaca dari perangkat seluler.

- [x] **Task 7.1:** Buat migration `announcements`, `announcement_targets`, dan `events`.
- [x] **Task 7.2:** Implementasi API pengumuman: draft, target, jadwal, publish, archive, dan statistik baca.
- [x] **Task 7.3:** Implementasi API agenda: buat, ubah, batal, status, dan lampiran.
- [x] **Task 7.4:** Implementasi seleksi target pengumuman berdasarkan seluruh warga, peran, keluarga, atau unit.
- [x] **Task 7.5:** Buat UI pengurus untuk pengumuman, agenda, penjadwalan, target, dan arsip.
- [x] **Task 7.6:** Buat UI warga untuk daftar/detail pengumuman, agenda mendatang, dan penanda penting.
- [x] **Task 7.7:** Tambahkan test target visibility, status jadwal, dan akses lampiran.

---

## [x] Epic 8: Surat Pengantar

Tujuan: pengajuan surat dari formulir sampai PDF terbit dan dapat diunduh.

- [x] **Task 8.1:** Buat migration `letter_types` dan `letter_requests`.
- [x] **Task 8.2:** Implementasi API jenis surat: requirements, schema formulir, template, pola nomor, dan status.
- [x] **Task 8.3:** Implementasi validasi form dinamis, lampiran wajib, dan pengajuan surat oleh warga.
- [x] **Task 8.4:** Implementasi workflow surat: draft, diajukan, review, revisi, persetujuan, penolakan, penerbitan, pembatalan.
- [x] **Task 8.5:** Implementasi nomor surat unik, generator PDF, penyimpanan PDF di R2, dan signed download.
- [x] **Task 8.6:** Buat UI warga untuk pilih jenis surat, form bertahap, lampiran, status, revisi, dan unduh PDF.
- [x] **Task 8.7:** Buat UI sekretaris/ketua RT untuk antrean, review, catatan internal, setujui, tolak, dan terbitkan.
- [x] **Task 8.8:** Tambahkan test transisi status, kelengkapan lampiran, nomor surat, dan authorization.

---

## [x] Epic 9: Aduan Warga

Tujuan: tiket aduan dengan status, penanggung jawab, komentar, serta riwayat penyelesaian.

- [x] **Task 9.1:** Buat migration `complaints` dan `complaint_comments`.
- [x] **Task 9.2:** Implementasi API pembuatan aduan: kategori, lokasi umum, prioritas, lampiran, dan nomor tiket.
- [x] **Task 9.3:** Implementasi API daftar/detail aduan dengan scope warga, pengurus, serta petugas tertugaskan.
- [x] **Task 9.4:** Implementasi penugasan petugas, perubahan status, catatan resolusi, dan penutupan aduan.
- [x] **Task 9.5:** Implementasi komentar/pembaruan dan pemisahan komentar internal. *(Notifikasi perubahan status ditunda ke Epic 10.)*
- [x] **Task 9.6:** Buat UI warga untuk buat aduan, timeline, komentar, dan status aduan saya.
- [x] **Task 9.7:** Buat UI pengurus untuk antrean aduan, filter, assignment, pembaruan status, dan resolusi.
- [x] **Task 9.8:** Tambahkan test scope pelapor, komentar internal, state transition, dan assignment.

---

## Epic 10: Notifikasi

Tujuan: notifikasi dalam aplikasi, email Resend, dan WhatsApp SaungWA untuk aktivitas penting.

- [x] **Task 10.1:** Buat migration `notifications` dan API daftar notifikasi, tandai dibaca, serta tandai semua dibaca.
- [x] **Task 10.2:** Buat komponen notifikasi warga: indikator belum dibaca, daftar, detail, dan empty state.
- [x] **Task 10.3:** Implementasi adapter Resend untuk email transaksional dan template pesan dasar.
- [x] **Task 10.4:** Implementasi adapter SaungWA untuk notifikasi WhatsApp dan validasi konfigurasi provider.
- [x] **Task 10.5:** Tambahkan trigger notifikasi untuk undangan, reset password, tagihan, pembayaran, surat, aduan, dan pengumuman penting. *(Tagihan, pembayaran, surat, aduan, dan pengumuman penting sudah terhubung; undangan, reset password, serta pengajuan surat awal kini melalui dispatcher.)*
- [x] **Task 10.6:** Pastikan kegagalan provider dicatat tanpa membatalkan transaksi utama.
- [x] **Task 10.7:** Tambahkan mekanisme retry terkontrol jika kebutuhan reliabilitas terbukti. *(Belum terbukti diperlukan pada MVP; dispatcher best-effort mencatat kegagalan. Upgrade: durable outbox dengan retry terbatas.)*
- [x] **Task 10.8:** Tambahkan test payload provider, kegagalan provider, dan permission notifikasi.

---

## Epic 11: Dashboard dan Laporan

Tujuan: ringkasan operasional yang relevan serta ekspor laporan sesuai permission.

- [x] **Task 11.1:** Implementasi API dashboard warga: tagihan aktif, pembayaran terbaru, surat, aduan, pengumuman, agenda.
- [x] **Task 11.2:** Implementasi API dashboard pengurus: keluarga/warga aktif, tagihan, pembayaran, tunggakan, kas, surat, aduan.
- [x] **Task 11.3:** Implementasi endpoint laporan keluarga/warga, mutasi warga, tagihan, tunggakan, pembayaran, kas, surat, dan aduan.
- [x] **Task 11.4:** Implementasi ekspor CSV sesuai filter, permission, scope, dan audit log.
- [x] **Task 11.5:** Implementasi ekspor PDF laporan formal. *(PDF formal dasar multi-halaman tersedia pada seluruh endpoint laporan; CSV tetap tersedia. Upgrade: template HTML/CSS setelah format branding disahkan.)*
- [x] **Task 11.6:** Buat UI dashboard warga mobile-first.
- [x] **Task 11.7:** Buat UI dashboard pengurus, laporan, filter periode, dan ekspor.
- [x] **Task 11.8:** Tambahkan test akurasi agregat, scope ekspor, dan audit ekspor.

---

## Epic 12: Pengaturan, Audit, Hardening, dan Rilis

Tujuan: menyelesaikan kesiapan operasional, keamanan, observabilitas, serta UAT sebelum peluncuran.

- [x] **Task 12.1:** Implementasi pengaturan RT: identitas, logo, rekening, zona waktu, nomor surat, batas unggahan, dan template.
- [x] **Task 12.2:** Buat migration dan layanan `audit_logs` append-only untuk autentikasi, data sensitif, keuangan, surat, ekspor, peran, dan konfigurasi.
- [x] **Task 12.3:** Buat UI audit log dengan filter dan detail yang disanitasi.
- [x] **Task 12.4:** Terapkan CORS, CSRF sesuai mekanisme cookie, rate limit, security headers, body limit, dan sanitasi log.
- [ ] **Task 12.5:** Konfigurasi monitoring: health/readiness, CloudWatch log/metric/alarm, request ID, serta error tracking yang lolos evaluasi privasi. *(Health/readiness, request ID, log JSON, dan runbook alarm tersedia; konfigurasi CloudWatch produksi belum dapat divalidasi tanpa akun/infrastruktur target.)*
- [ ] **Task 12.6:** Konfigurasi backup RDS, retensi/pemulihan R2, dan uji restore di staging. *(Prosedur dan checklist tersedia di `docs/RELEASE_RUNBOOK.md`; konfigurasi cloud serta restore staging membutuhkan akses lingkungan target.)*
- [ ] **Task 12.7:** Jalankan responsive, accessibility, authorization, integration, E2E, security header, dan smoke test staging. *(Go test, lint, dan build lulus; pengujian viewport, E2E, serta staging belum dijalankan.)*
- [ ] **Task 12.8:** Jalankan UAT pengurus, migrasi CSV final, pelatihan, runbook, kebijakan privasi, dan soft launch. *(Runbook tersedia; aktivitas operasional/UAT belum dapat dinyatakan selesai.)*

---

## Epic 13: Master Data

Tujuan: memformalkan data referensi yang saat ini berupa teks bebas agar input konsisten dan laporan akurat.

- [x] **Task 13.1:** Buat migration `complaint_categories` per organisasi: `id`, `organization_id`, `code`, `name`, `status`, timestamps; unique `(organization_id, code)`.
- [x] **Task 13.2:** Seed kategori aduan awal per organisasi dan migrasikan nilai `complaints.category` historis ke `complaint_categories`.
- [x] **Task 13.3:** Tambahkan `complaint_category_id` pada `complaints`, validasi tenant melalui FK komposit, ubah API/daftar/filter/laporan, lalu hapus kolom teks lama pada migration lanjutan. *(FK, API, daftar, filter, dan laporan selesai; cleanup guarded tersedia pada `0017_epic13_cleanup.up.sql`.)*
- [x] **Task 13.4:** Implementasi API CRUD kategori aduan, RBAC pengurus, penonaktifan tanpa menghapus kategori yang telah dipakai, serta audit log.
- [x] **Task 13.5:** Perbarui UI pengurus pengelolaan kategori aduan dan form/filter aduan agar memakai data master.
- [x] **Task 13.6:** Buat lookup global read-only `education_levels` dan `marital_statuses`; seed nilai standar nasional; migrasikan nilai teks warga ke FK.
- [x] **Task 13.7:** Evaluasi normalisasi `occupations` dari data produksi. Buat master global hanya bila variasi penulisan mengganggu laporan; jangan blokir input pekerjaan bebas pada MVP. *(Pekerjaan tetap teks bebas pada MVP.)*
- [x] **Task 13.8:** Perbarui API, UI, import CSV, serta laporan warga untuk memakai lookup pendidikan/status perkawinan; pekerjaan tetap teks sampai Task 13.7 disetujui. *(API lookup, UI detail warga, import CSV, dan laporan selesai.)*
- [ ] **Task 13.9:** Tambahkan integration test FK, isolasi tenant, kategori nonaktif, migrasi data lama, RBAC, dan konsistensi filter/laporan. *(Test kategori aktif/nonaktif, ID tidak valid, RBAC, isolasi tenant, filter, scope warga, dan lookup warga tersedia; test migrasi legacy serta kontrak laporan/import lookup masih diperlukan.)*
- [x] **Task 13.10:** Evaluasi kebutuhan `announcement_categories`. Pertahankan `CHECK` global saat kategori seragam; buat master per organisasi hanya bila ada kebutuhan kategori kustom yang disetujui. *(Tetap `CHECK` global pada MVP.)*

---

## Epic 14: Otomatisasi dan Layanan Operasional Proaktif

Tujuan: mengubah data dan transaksi menjadi pengingat, antrean kerja, kepastian status warga, serta transparansi yang aman.

- [ ] **Task 14.1:** Implementasi penerbitan tagihan rutin terjadwal untuk `due_types` aktif, target keluarga yang sah, idempotensi periode, ringkasan hasil, dan audit.
- [ ] **Task 14.2:** Implementasi pengingat iuran melalui dispatcher sesuai preferensi kanal, jadwal yang dapat dikonfigurasi, pembatasan frekuensi, dan pencatatan kegagalan tanpa menggagalkan transaksi utama.
- [ ] **Task 14.3:** Implementasi pembayaran rapel: satu pelaporan pembayaran dialokasikan atomik ke beberapa invoice menurut aturan alokasi yang eksplisit, dengan perlindungan overpayment dan audit.
- [ ] **Task 14.4:** Buat antrean verifikasi bendahara satu layar: bukti, nominal, invoice, sisa tagihan, riwayat relevan, alasan penolakan standar, dan separation of duties.
- [ ] **Task 14.5:** Implementasi validasi pra-pengajuan surat untuk data formulir, persyaratan lampiran, dan status data yang diwajibkan oleh jenis surat; tampilkan kekurangan sebelum pengajuan.
- [ ] **Task 14.6:** Tambahkan SLA antrean surat per jenis surat, indikator jatuh tempo/terlambat bagi pengurus, estimasi status bagi warga, dan notifikasi eskalasi yang tidak berlebihan.
- [ ] **Task 14.7:** Perluas `complaint_categories` dengan target respons dan penyelesaian; buat timeline aduan, indikator SLA, serta pengelompokan/dukungan aduan serupa bila kebutuhan volume terbukti.
- [ ] **Task 14.8:** Implementasi konfirmasi penyelesaian aduan oleh pelapor, tenggat penutupan otomatis yang dapat dikonfigurasi, alasan penutupan, dan jejak audit.
- [ ] **Task 14.9:** Implementasi health score keluarga berbasis kelengkapan, verifikasi, usia pembaruan, dan kontak; tampilkan daftar kerja sekretaris tanpa menghukum warga.
- [ ] **Task 14.10:** Tambahkan tanggal evaluasi domisili sementara/kontrak dan pengingat konfirmasi tinggal/pindah bagi warga serta sekretaris.
- [ ] **Task 14.11:** Buat transparansi kas agregat untuk warga: saldo, pemasukan/pengeluaran per kategori dan periode, bukti yang diizinkan, tanpa nama penunggak atau detail transaksi pribadi.
- [ ] **Task 14.12:** Tambahkan QR dan halaman verifikasi publik surat yang hanya menampilkan nomor, jenis, tanggal terbit, dan status valid/dibatalkan; tanpa data pribadi atau URL dokumen privat.
- [ ] **Task 14.13:** Implementasi serah-terima jabatan: checklist role, akses, rekening, tagihan terbuka, kas, surat, aduan, dokumen; penurunan akses pengurus lama; audit historis tetap utuh.
- [ ] **Task 14.14:** Tambahkan integration test scheduler/idempotensi, alokasi rapel, SLA, isolasi tenant, otorisasi, privasi kas/QR, dan serah-terima akses.

---

## Epic 15: Tabungan Warga (Dana Titipan Non-Kas)

Tujuan: memfasilitasi tabungan terarah warga, misalnya Qurban/Idul Adha, tanpa mencampurkan dana titipan dengan kas operasional RT.

- [ ] **Task 15.1:** Buat migration `savings_products`, `savings_accounts`, dan `savings_transactions` dengan index tenant, FK komposit, constraint saldo, serta mutasi append-only.
- [ ] **Task 15.2:** Implementasi API CRUD `savings_products` sebagai master jenis tabungan per organisasi: kode, periode, sasaran, setoran minimum, aturan penarikan, tujuan alokasi, dan status; penonaktifan tanpa penghapusan.
- [ ] **Task 15.3:** Implementasi pembukaan/penutupan akun tabungan per keluarga pada produk aktif; satu akun aktif per keluarga dan produk.
- [ ] **Task 15.4:** Implementasi setoran: laporan setoran, unggah bukti, idempotensi, verifikasi bendahara, dan mutasi kredit setelah verifikasi.
- [ ] **Task 15.5:** Implementasi penarikan atau pengembalian: permintaan, persetujuan pemilik saldo, verifikasi bendahara, bukti, batas saldo, mutasi debit, dan audit log.
- [ ] **Task 15.6:** Implementasi alokasi dana untuk tujuan produk, misalnya pembelian hewan qurban, dengan persetujuan kebijakan yang jelas dan jejak mutasi; tidak menjadi pendapatan/pengeluaran kas operasional RT.
- [ ] **Task 15.7:** Implementasi koreksi melalui mutasi pembalik; larang `UPDATE` nominal dan penghapusan mutasi historis.
- [ ] **Task 15.8:** Buat UI warga untuk memilih produk, melihat saldo dan riwayat mutasi, melapor setoran, serta mengajukan penarikan bila diizinkan.
- [ ] **Task 15.9:** Buat UI bendahara untuk master produk, akun, antrean verifikasi, persetujuan mutasi debit, dan rekonsiliasi rekening penampungan.
- [ ] **Task 15.10:** Tambahkan laporan dana titipan per produk/per keluarga dan rekonsiliasi saldo sistem dengan kas fisik atau rekening penampungan khusus; pisahkan dari laporan kas operasional.
- [ ] **Task 15.11:** Tambahkan integration test perhitungan saldo dari mutasi, idempotensi setoran, pencegahan saldo negatif, otorisasi debit, append-only, rekonsiliasi, dan isolasi tenant.

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