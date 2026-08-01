# Milestones RT Digital

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital  
**Referensi:** `DEVELOPMENT_BACKLOG.md`, `PRD.md`

Dokumen ini memetakan epic backlog menjadi enam milestone berurutan. Setiap milestone menghasilkan inkremen yang dapat diuji oleh pengurus RT sebelum melanjutkan ke milestone berikutnya.

## Prinsip

- Setiap milestone wajib memenuhi standar mobile-first, RBAC, isolasi `organization_id`, validasi backend, dan perlindungan data sensitif.
- Deployment ke staging dilakukan sebelum validasi milestone.
- Task audit log, test, dokumentasi API, serta error handling dikerjakan bersamaan dengan modul terkait; tidak ditunda hingga akhir.
- Milestone berikutnya dimulai setelah kriteria selesai milestone sebelumnya dipenuhi.

---

## Milestone 1: Fondasi Project dan Authentication

**Tujuan:** Infrastruktur dasar siap; pengguna dapat masuk dengan aman; peran dan akses dasar berlaku.

**Epic terkait:**
- Epic 0: Infrastruktur dan Fondasi Proyek
- Epic 1: Authentication dan Akun
- Epic 2: Manajemen Pengguna dan RBAC

### Target Pengiriman

- Monorepo Next.js dan Go API tersedia.
- Docker development berjalan dengan PostgreSQL dan MinIO untuk pengujian file.
- Database awal untuk organisasi, pengguna, peran, permission, dan session tersedia.
- Login, logout, refresh token, aktivasi akun, reset password, dan penguncian login berfungsi.
- MFA aktif untuk peran pengurus.
- RBAC membatasi menu, endpoint, serta data sesuai peran dan `organization_id`.
- Super Admin dapat menyiapkan organisasi; Ketua RT dapat mengundang dan menetapkan peran pengurus.
- Shell UI mobile-first, bottom navigation warga, sidebar/drawer pengurus, loading/error/empty state tersedia.
- CI minimum berjalan: format, lint, type check, test, dan build.

### Kriteria Selesai

- Akun tidak aktif, token kedaluwarsa, atau permission tidak memadai tidak dapat mengakses data.
- Perubahan peran serta peristiwa login penting tercatat pada audit log.
- Login diuji pada viewport 320 px, 360 px, 390 px, Chrome Android, dan Safari iOS.
- Staging dapat diakses dan health check API lulus.

---

## Milestone 2: Data Keluarga dan Warga

**Tujuan:** Pengurus dapat membangun basis data warga RT; warga dapat melihat dan mengoreksi data keluarganya.

**Epic terkait:**
- Epic 3: Data Keluarga dan Warga

### Target Pengiriman

- Pengurus dapat mengelola rumah/unit, keluarga, warga, dan anggota keluarga.
- Satu keluarga memiliki tepat satu kepala keluarga aktif.
- Pencarian, filter, status domisili, dan riwayat perubahan tersedia.
- NIK dan nomor KK dienkripsi, dimasking, serta tidak dapat dicari langsung dari ciphertext.
- Warga hanya dapat melihat data keluarga yang diizinkan.
- Warga dapat mengajukan koreksi; Sekretaris dapat menyetujui, menolak, atau meminta perbaikan.
- Import CSV memiliki dry-run, validasi, deteksi duplikasi, hasil error, dan audit import.
- UI pengurus menyediakan tampilan tabel desktop serta kartu/list ringkas pada layar kecil.

### Kriteria Selesai

- Data sensitif tidak muncul utuh tanpa permission khusus.
- Isolasi data antar organisasi dan ownership keluarga telah diuji.
- Import CSV staging dapat dijalankan, diverifikasi, dan diulang secara aman.
- Alur lihat profil dan pengajuan koreksi dapat selesai dari perangkat seluler.

---

## Milestone 3: Iuran dan Pembayaran

**Tujuan:** Bendahara dapat menerbitkan tagihan; warga dapat melaporkan pembayaran; pembayaran dapat diverifikasi dengan aman.

**Epic terkait:**
- Epic 4: Iuran dan Tagihan
- Epic 5: Pembayaran

### Target Pengiriman

- Master jenis iuran tersedia: nominal, frekuensi, jatuh tempo, sasaran, dan status aktif.
- Bendahara dapat membuat tagihan individual dan massal.
- Warga dapat melihat tagihan aktif dan riwayat tagihan dari telepon seluler.
- Warga dapat mengunggah foto/screenshot bukti transfer dari kamera atau galeri.
- File bukti disimpan privat di Cloudflare R2 melalui signed URL.
- Bendahara dapat melihat antrean pembayaran dan menerima, menolak, atau membatalkan transaksi dengan alasan.
- Status invoice diperbarui melalui transaksi atomik setelah pembayaran diverifikasi.
- Tanda terima dan riwayat pembayaran tersedia bagi warga.

### Kriteria Selesai

- Pengiriman ulang request tidak menimbulkan pembayaran ganda melalui `Idempotency-Key`.
- Lock transaksi mencegah verifikasi bersamaan mengubah tagihan secara salah.
- Bendahara tidak dapat memverifikasi transaksi sendiri bila pemisahan tugas diaktifkan.
- Bukti transfer tidak tersedia melalui URL publik permanen.
- Alur tagihan hingga verifikasi dapat diuji lengkap pada perangkat seluler.

---

## Milestone 4: Buku Kas dan Laporan

**Tujuan:** Keuangan RT tercatat konsisten, dapat dikoreksi tanpa menghapus riwayat, serta dapat dilaporkan.

**Epic terkait:**
- Epic 6: Buku Kas
- Epic 11: Dashboard dan Laporan

### Target Pengiriman

- Pembayaran terverifikasi otomatis menghasilkan transaksi pemasukan kas.
- Bendahara dapat membuat kategori kas serta mencatat pemasukan/pengeluaran manual beserta bukti.
- Kesalahan dikoreksi dengan transaksi pembalik; transaksi historis tidak dapat dihapus.
- Buku kas menampilkan saldo berjalan dan filter periode/kategori.
- Dashboard warga menampilkan tagihan, pembayaran terbaru, surat, aduan, pengumuman, dan agenda relevan.
- Dashboard pengurus menampilkan warga aktif, tagihan, tunggakan, kas, surat, dan aduan.
- Laporan warga, tagihan, pembayaran, tunggakan, kas, surat, dan aduan dapat diekspor CSV.
- Ekspor PDF ditambahkan untuk laporan formal yang formatnya telah disetujui.

### Kriteria Selesai

- Saldo kas, invoice, payment, dan transaksi kas konsisten pada test integrasi.
- Setiap pembalikan, ekspor, dan perubahan keuangan memiliki audit log.
- Warga tidak dapat melihat rincian keuangan keluarga lain.
- Laporan hanya dapat diekspor oleh peran yang memiliki permission.

---

## Milestone 5: Pengumuman, Surat, dan Aduan

**Tujuan:** Aplikasi menjadi kanal komunikasi, pelayanan administrasi, dan tindak lanjut masalah lingkungan.

**Epic terkait:**
- Epic 7: Pengumuman dan Agenda
- Epic 8: Surat Pengantar
- Epic 9: Aduan Warga
- Epic 10: Notifikasi

### Target Pengiriman

- Pengurus dapat membuat, menjadwalkan, menerbitkan, menargetkan, dan mengarsipkan pengumuman.
- Agenda kegiatan tersedia bagi warga dan pengurus.
- Warga dapat menerima notifikasi dalam aplikasi, email melalui Resend, dan WhatsApp melalui SaungWA untuk pemicu penting.
- Warga dapat memilih jenis surat, mengisi formulir bertahap, serta mengunggah lampiran dari perangkat seluler.
- Sekretaris dapat review; Ketua RT dapat menyetujui; sistem menerbitkan PDF surat dengan nomor unik.
- Warga dapat membuat aduan, menerima nomor tiket, melihat timeline, dan memberi komentar.
- Pengurus dapat menetapkan penanggung jawab, memperbarui status, menulis catatan resolusi, serta menutup aduan.

### Kriteria Selesai

- Target pengumuman hanya dapat dilihat penerima yang berhak.
- Surat tidak dapat diterbitkan bila data/lampiran wajib belum lengkap.
- Nomor surat terbit tidak dapat digunakan ulang.
- Pelapor hanya melihat aduannya sendiri; komentar internal tidak terlihat warga.
- Kegagalan Resend atau SaungWA dicatat tanpa membatalkan transaksi inti.

---

## Milestone 6: Deployment dan Uji Coba RT

**Tujuan:** Aplikasi siap operasi, diuji bersama pengurus, dan diluncurkan sebagai MVP.

**Epic terkait:**
- Epic 12: Pengaturan, Audit, Hardening, dan Rilis

### Target Pengiriman

- Pengaturan RT selesai: identitas, logo, rekening, zona waktu, format nomor surat, batas unggahan, dan template.
- Audit log append-only tersedia untuk tindakan sensitif, transaksi, ekspor, otorisasi, dan konfigurasi.
- Staging dan production terpisah pada Cloudflare Workers, ECS Fargate, RDS, dan Cloudflare R2.
- CORS, CSRF, security headers, rate limiting, body limit, secret management, dan sanitasi log diterapkan.
- Monitoring, request ID, health/readiness check, CloudWatch logs/metrics/alarm aktif.
- Backup RDS, pemulihan R2, serta prosedur restore diuji di staging.
- UAT pengurus, responsive test, accessibility test, authorization test, E2E, smoke test, dan security test selesai.
- Data awal diimpor, pengurus dilatih, panduan operasional tersedia, lalu soft launch dilakukan.

### Kriteria Selesai

- Tidak ada bug severity kritis atau tinggi terbuka.
- Seluruh fitur Must Have PRD lulus UAT.
- Backup dan restore terbukti berjalan.
- Monitoring dan alert dasar aktif.
- Kebijakan privasi, ketentuan penggunaan, runbook, dan panduan pengurus tersedia.
- Pemilik produk atau perwakilan pengurus RT menyetujui rilis MVP.

---

## Ringkasan Dependensi

| Milestone | Bergantung pada | Hasil utama |
|---|---|---|
| 1 | — | Infrastruktur, authentication, RBAC |
| 2 | 1 | Data rumah, keluarga, warga |
| 3 | 1, 2 | Iuran, tagihan, pembayaran |
| 4 | 3 | Kas, dashboard, laporan |
| 5 | 1, 2 | Pengumuman, surat, aduan, notifikasi |
| 6 | 1–5 | Hardening, UAT, production launch |