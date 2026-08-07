# Kriteria Penerimaan (Acceptance Criteria)

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital  
**Referensi:** `PRD.md`, `DEVELOPMENT_BACKLOG.md`, `MILESTONES.md`

Dokumen ini menentukan syarat minimum agar fitur dapat diterima sebagai selesai. Kriteria ini berlaku bersama *Definition of Done* di `DEVELOPMENT_BACKLOG.md`.

## 1. Kriteria Global

Seluruh fitur harus memenuhi:

- **Mobile-first:** berfungsi pada viewport 320 px, 360 px, 390 px, tablet, dan desktop; tidak ada tombol tersembunyi, teks terpotong, atau tabel yang tidak dapat digunakan.
- **Aksesibilitas:** area sentuh minimal 44 × 44 CSS pixel, label formulir jelas, status tidak hanya dibedakan warna, serta pesan error mudah dipahami.
- **Otorisasi:** backend memvalidasi autentikasi, `organization_id`, ownership data, dan RBAC. Frontend tidak menjadi satu-satunya kontrol akses.
- **Privasi:** NIK, nomor KK, kata sandi, token, serta data sensitif tidak muncul di log, cache publik, URL, atau respons API yang tidak membutuhkan data tersebut.
- **Ketahanan:** kegagalan jaringan atau API tidak membuat aplikasi crash maupun menghilangkan isian formulir yang belum berhasil dikirim.
- **Data penting:** operasi mutasi memakai transaksi database; operasi yang dapat dikirim ulang memakai idempotency key bila relevan.
- **Pencegahan klik ganda:** setiap pemicu aksi mutasi masuk ke loading state dan tidak dapat dipicu ulang sampai request selesai, gagal, atau timeout; klik ganda serta submit ganda melalui Enter tidak menghasilkan request kedua. Guard UI ini tidak menggantikan idempotency key di backend.
- **Audit:** tindakan penting mencatat aktor, peran aktif, tindakan, entitas, waktu, request ID, serta ringkasan perubahan bila relevan.
- **Error:** API mengembalikan `code`, `message`, `details`, dan `request_id`; UI tidak menampilkan stack trace.
- **Pengujian:** test unit, integration, authorization, responsive, atau E2E relevan lulus sebelum fitur diterima.

---

## 2. Autentikasi dan Akun

- [ ] Pengguna dapat login memakai email atau nomor telepon serta kata sandi valid.
- [ ] Akun nonaktif tidak dapat login.
- [ ] Percobaan login gagal berulang memicu rate limit atau penguncian sementara.
- [ ] Access token kedaluwarsa atau tidak valid ditolak dari endpoint privat.
- [ ] Refresh token diputar aman; token lama tidak dapat dipakai kembali.
- [ ] Pengguna dapat logout dari perangkat aktif maupun seluruh perangkat.
- [ ] Pengguna dapat aktivasi akun dan reset kata sandi menggunakan token sekali pakai dengan masa berlaku.
- [ ] Email aktivasi/reset dikirim melalui Resend.
- [ ] MFA wajib dan berfungsi untuk peran pengurus.
- [ ] Login, reset kredensial, perubahan kata sandi, serta logout penting tercatat pada audit log.

---

## 3. Manajemen Pengguna dan RBAC

- [ ] Pengurus berwenang dapat melihat daftar pengguna sesuai organisasi.
- [ ] Pengurus berwenang dapat mengundang, mengaktifkan, menonaktifkan, dan menetapkan peran pengguna.
- [ ] Pengguna hanya melihat menu dan tindakan sesuai permission.
- [ ] Endpoint API mengembalikan `403` untuk permission yang tidak memadai.
- [ ] Pengurus tidak dapat menaikkan hak akses dirinya sendiri tanpa kewenangan valid.
- [ ] Super Admin tidak dapat mengakses data sensitif warga tanpa alasan operasional dan audit.
- [ ] Perubahan peran, permission, dan status pengguna tercatat pada audit log.

---

## 4. Data Rumah, Keluarga, dan Warga

- [ ] Pengurus dapat menambah, mengubah, melihat, dan menonaktifkan rumah/unit, keluarga, warga, serta anggota keluarga.
- [ ] Satu keluarga memiliki tepat satu kepala keluarga aktif.
- [ ] Satu warga aktif tidak dapat menjadi anggota aktif lebih dari satu keluarga dalam organisasi yang sama.
- [ ] NIK dan nomor KK terenkripsi saat disimpan serta dimasking pada tampilan standar.
- [ ] Data duplikat NIK/nomor KK terdeteksi tanpa mencari langsung pada ciphertext.
- [ ] Warga hanya dapat melihat data keluarga sendiri sesuai izin.
- [ ] Warga dapat mengajukan koreksi data.
- [ ] Sekretaris dapat menyetujui, menolak, atau meminta perbaikan koreksi data.
- [ ] Perubahan NIK, nomor KK, status pindah, atau meninggal tidak berlaku sebelum verifikasi pengurus.
- [ ] Dry-run CSV mendeteksi format tidak valid dan duplikat tanpa menulis data.
- [ ] Import CSV mencatat hasil, error, serta pelaksana pada audit log.

---

## 5. Iuran dan Tagihan

- [ ] Bendahara dapat membuat dan mengelola jenis iuran tetap, fleksibel, berulang, atau sekali bayar.
- [ ] Bendahara dapat membuat tagihan individual maupun massal.
- [ ] Pembuatan tagihan massal tidak menghasilkan tagihan ganda untuk keluarga dan periode yang sama.
- [ ] Setiap tagihan memiliki nomor unik, periode, nominal, jatuh tempo, status, dan riwayat perubahan.
- [ ] Warga dapat melihat tagihan aktif serta riwayat tagihan keluarganya sendiri.
- [ ] Bendahara dapat memberi penyesuaian, diskon, atau membatalkan tagihan dengan alasan wajib.
- [ ] Daftar tunggakan dan totalnya akurat sesuai status tagihan.
- [ ] Pembuatan dan perubahan tagihan tercatat pada audit log.

---

## 6. Pembayaran

- [ ] Warga dapat melihat detail tagihan aktif dan riwayat pembayaran.
- [ ] Warga dapat memilih metode tunai atau transfer.
- [ ] Warga dapat mengunggah foto atau screenshot bukti pembayaran dari kamera atau galeri.
- [ ] Bukti pembayaran tervalidasi jenis dan ukuran filenya, tersimpan privat di Cloudflare R2, serta hanya dapat diakses melalui signed URL.
- [ ] Bendahara dapat melihat antrean pembayaran menunggu verifikasi.
- [ ] Bendahara dapat menerima atau menolak pembayaran; penolakan wajib memiliki alasan.
- [ ] Bendahara dapat membatalkan transaksi dengan alasan; transaksi tidak dihapus permanen.
- [ ] Status pembayaran dan tagihan diperbarui atomik setelah verifikasi.
- [ ] Request yang dikirim ulang tidak membuat pembayaran atau verifikasi ganda.
- [ ] Pemisahan tugas mencegah pengguna memverifikasi pembayaran yang dibuat sendiri bila kebijakan diaktifkan.
- [ ] Warga dapat melihat atau mengunduh tanda terima setelah pembayaran diverifikasi.
- [ ] Pembuatan, verifikasi, penolakan, serta pembatalan pembayaran tercatat pada audit log.

---

## 7. Buku Kas

- [ ] Pembayaran terverifikasi otomatis menghasilkan transaksi pemasukan kas.
- [ ] Bendahara dapat membuat kategori kas.
- [ ] Bendahara dapat mencatat pemasukan atau pengeluaran manual dengan nominal, kategori, tanggal, deskripsi, dan bukti bila diperlukan.
- [ ] Saldo berjalan akurat berdasarkan transaksi yang berlaku.
- [ ] Transaksi kas historis tidak dapat dihapus melalui UI.
- [ ] Koreksi transaksi dilakukan melalui transaksi pembalik dengan referensi transaksi asal.
- [ ] Warga tidak dapat melihat rincian kas kecuali ringkasan yang secara eksplisit diizinkan.
- [ ] Pencatatan, pembalikan, dan perubahan kas tercatat pada audit log.

---

## 8. Pengumuman dan Agenda

- [ ] Pengurus dapat membuat, mengubah, menjadwalkan, menerbitkan, serta mengarsipkan pengumuman.
- [ ] Pengumuman tidak tampil sebelum `publish_at` dan tidak tampil setelah masa berlakunya berakhir.
- [ ] Target pengumuman dapat dibatasi ke seluruh warga, pengurus, blok, atau keluarga tertentu.
- [ ] Warga hanya melihat pengumuman yang menjadi targetnya.
- [ ] Lampiran pengumuman tersedia aman sesuai hak akses.
- [ ] Warga dapat melihat pengumuman penting dan agenda pada layar seluler.
- [ ] Agenda memuat tanggal, waktu, lokasi, deskripsi, status, dan penanggung jawab.
- [ ] Pengurus dapat melihat statistik keterbacaan pengumuman penting.

---

## 9. Pengajuan Surat Pengantar

- [ ] Warga dapat memilih jenis surat aktif dan melihat persyaratannya.
- [ ] Formulir dinamis memvalidasi data wajib serta lampiran yang diperlukan.
- [ ] Warga dapat mengunggah lampiran dari perangkat seluler dan mengirim pengajuan.
- [ ] Sekretaris dapat memeriksa, meminta revisi, meneruskan, atau menolak pengajuan dengan catatan.
- [ ] Ketua RT berwenang dapat menyetujui surat.
- [ ] Surat tidak dapat diterbitkan sebelum data dan lampiran wajib terpenuhi.
- [ ] Sistem menghasilkan PDF surat dengan nomor unik yang tidak dapat digunakan ulang.
- [ ] Warga dapat melihat status, catatan revisi, dan mengunduh PDF surat terbit.
- [ ] Riwayat pengajuan, persetujuan, penerbitan, dan unduhan tercatat.

---

## 10. Aduan Warga

- [ ] Warga dapat membuat aduan berisi kategori, judul, deskripsi, lokasi umum, dan lampiran opsional.
- [ ] Sistem membuat nomor tiket unik saat aduan berhasil dikirim.
- [ ] Warga hanya dapat melihat aduan miliknya sendiri.
- [ ] Pengurus dapat melihat, memfilter, menetapkan penanggung jawab, serta mengubah status aduan.
- [ ] Petugas tertugaskan dapat memberi pembaruan dan komentar sesuai permission.
- [ ] Warga dapat melihat timeline serta pembaruan status aduannya.
- [ ] Catatan internal pengurus tidak terlihat oleh warga pelapor.
- [ ] Aduan hanya dapat ditutup dengan catatan penyelesaian.
- [ ] Perubahan status, assignment, dan penutupan tercatat pada audit log.

---

## 11. Notifikasi

- [ ] Sistem menyimpan notifikasi dalam aplikasi dengan status dibaca atau belum dibaca.
- [ ] Pengguna dapat melihat dan menandai notifikasi sebagai telah dibaca.
- [ ] Pemicu penting mengirim notifikasi sesuai preferensi melalui aplikasi, Resend, atau SaungWA.
- [ ] Kegagalan Resend atau SaungWA dicatat tanpa membatalkan transaksi utama.
- [ ] Notifikasi tidak dikirim kepada pengguna di luar target atau organisasi.
- [ ] Payload notifikasi tidak memuat data sensitif berlebihan.

---

## 12. Dashboard dan Laporan

- [ ] Dashboard warga menampilkan tagihan aktif, pembayaran terbaru, status surat, status aduan, pengumuman penting, dan agenda terdekat.
- [ ] Dashboard pengurus menampilkan jumlah warga/keluarga aktif, tagihan, tunggakan, kas, surat, aduan, dan pengumuman terbaru sesuai permission.
- [ ] Nilai agregat dashboard sesuai data sumber dan filter periode.
- [ ] Pengurus berwenang dapat membuat laporan warga, tagihan, tunggakan, pembayaran, kas, surat, dan aduan.
- [ ] Ekspor CSV menghormati permission, scope organisasi, serta parameter filter.
- [ ] PDF laporan hanya diterbitkan untuk format yang telah disetujui.
- [ ] Setiap ekspor mencatat aktor, waktu, parameter, jenis laporan, request ID, dan informasi audit yang diizinkan.

---

## 13. Pengaturan dan Audit Log

- [ ] Pengurus berwenang dapat mengatur identitas RT, alamat, logo, rekening, zona waktu, batas unggahan, format nomor surat, dan template.
- [ ] Perubahan pengaturan RT berlaku hanya untuk organisasi terkait.
- [ ] Audit log bersifat append-only; tidak dapat diubah atau dihapus melalui antarmuka aplikasi.
- [ ] Audit log menyimpan minimal aktor, peran aktif, tindakan, jenis/ID entitas, waktu, alamat IP sesuai kebijakan privasi, user agent, serta request ID.
- [ ] Audit log tidak menyimpan kata sandi, token, NIK, nomor KK, atau file sensitif secara utuh.

---

## 14. Kesiapan Rilis Fitur

Fitur dapat dinyatakan diterima bila:

- [ ] Semua kriteria modul yang relevan telah lulus.
- [ ] Test relevan lulus pada CI.
- [ ] Dokumentasi API, migration, serta konfigurasi diperbarui bila berubah.
- [ ] Tidak ada bug severity kritis atau tinggi terbuka pada fitur.
- [ ] Fitur lulus pengujian staging oleh perwakilan pengurus RT.