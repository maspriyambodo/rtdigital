# Ruang Lingkup MVP — RT Digital

**Status:** Dikunci untuk validasi  
**Tujuan:** Mengendalikan *scope creep*. Hanya fitur pada bagian **Masuk MVP** yang boleh dibangun untuk rilis pertama. Setiap usulan lain otomatis masuk backlog pasca-MVP.

Dokumen ini melengkapi `PRD.md`. Bila terjadi perbedaan, keputusan ruang lingkup pada dokumen ini berlaku untuk prioritas rilis MVP.

---

## 1. Sasaran MVP

Menyediakan aplikasi web mobile-first untuk satu RT agar warga dapat:

1. Mengakses informasi RT.
2. Melihat tagihan serta mengirim bukti pembayaran manual.
3. Mengajukan surat.
4. Membuat dan memantau aduan.

Pengurus dapat mengelola data warga, iuran, kas dasar, surat, aduan, serta pengumuman secara aman dan terlacak.

---

## 2. Masuk MVP

## 2.1 Fondasi Akses dan Keamanan

- Login dengan nomor telepon atau email dan kata sandi.
- Undangan, aktivasi akun, lupa kata sandi, ganti kata sandi, dan penonaktifan akun.
- RBAC untuk Super Admin, Ketua RT, Sekretaris, Bendahara, Petugas RT, Kepala Keluarga, dan Warga.
- MFA untuk pengurus: Ketua RT, Sekretaris, Bendahara, dan Super Admin.
- Audit log untuk tindakan penting.
- Masking data sensitif, enkripsi data sensitif, HTTPS, validasi input, dan akses file privat melalui URL bertanda tangan.

**Batasan:** MVP hanya melayani satu organisasi RT. Kolom `organization_id` tetap digunakan sebagai kesiapan data, tanpa fitur administrasi multi-RT mandiri.

## 2.2 Data Rumah, Keluarga, dan Warga

- Data rumah atau unit, keluarga, dan anggota keluarga.
- Status domisili dan status warga.
- Pencarian dan filter dasar.
- Koreksi data oleh warga serta verifikasi oleh pengurus.
- Impor data awal CSV dengan validasi.
- Ekspor sesuai izin.

**Batasan:** Tidak ada verifikasi atau sinkronisasi otomatis dengan Dukcapil.

## 2.3 Pengumuman dan Agenda

- Pengumuman dengan kategori, target, prioritas, jadwal terbit, dan lampiran.
- Agenda kegiatan dengan waktu, lokasi, penanggung jawab, dan status.
- Notifikasi dalam aplikasi.
- Email transaksional untuk aktivitas penting melalui Resend (`https://resend.com`).
- Notifikasi WhatsApp untuk aktivitas penting melalui SaungWA (`https://saungwa.com/`).

## 2.4 Iuran, Pembayaran, dan Kas

- Jenis iuran dan pembuatan tagihan massal atau individual.
- Pembayaran tunai atau transfer manual.
- Unggah bukti pembayaran dari kamera atau galeri perangkat seluler.
- Verifikasi, penolakan, dan pembatalan pembayaran oleh bendahara dengan alasan serta audit log.
- Tanda terima pembayaran.
- Buku kas pemasukan, pengeluaran, saldo berjalan, dan transaksi pembalik.
- Rekap tagihan, tunggakan, pembayaran, dan kas.

**Batasan:**
- Tidak ada payment gateway, virtual account, atau rekonsiliasi bank otomatis.
- Transaksi keuangan tidak dapat dihapus permanen.

## 2.5 Pengajuan Surat

- Jenis surat dan template yang dikonfigurasi pengurus.
- Formulir pengajuan, persyaratan lampiran, status proses, serta catatan perbaikan.
- Pemeriksaan oleh Sekretaris dan persetujuan Ketua RT.
- Penerbitan PDF serta riwayat penerbitan dan unduhan.

**Batasan:** Tidak ada tanda tangan elektronik tersertifikasi atau QR verifikasi surat pada MVP.

## 2.6 Aduan Warga

- Pembuatan tiket aduan dengan kategori, deskripsi, lokasi umum, dan lampiran.
- Penetapan penanggung jawab oleh pengurus.
- Pembaruan status, komentar, dan catatan penyelesaian.
- Warga melihat perkembangan aduan miliknya.

**Batasan:** Aduan anonim penuh dan SLA otomatis tidak masuk MVP.

## 2.7 Dashboard dan Laporan

- Dashboard warga: tagihan, pembayaran, surat, aduan, pengumuman penting, agenda.
- Dashboard pengurus: keluarga, warga, tagihan, tunggakan, kas, surat, aduan, pengumuman.
- Laporan dasar dalam CSV dan PDF.
- Pencatatan audit untuk ekspor laporan.

## 2.8 Mobile-First dan PWA Dasar

- UI dioptimalkan untuk viewport 320 px, 360 px, dan 390 px.
- Area sentuh minimum 44 × 44 CSS pixel.
- Bottom navigation warga maksimal lima tujuan utama.
- Tampilan kartu atau daftar ringkas untuk data seluler; tabel lebar diprioritaskan untuk desktop pengurus.
- Formulir bertahap, tipe input sesuai perangkat, validasi jelas, serta unggah kamera/galeri bila relevan.
- PWA installable dengan manifest, ikon, dan fallback offline aman.
- Indikator koneksi, retry aman, dan perlindungan isian formulir belum terkirim.

**Batasan:** PWA bukan aplikasi offline penuh. Data pribadi tidak boleh dicache atau ditampilkan kedaluwarsa saat offline.

---

## 3. Ditunda Pasca-MVP

Fitur berikut bernilai, tetapi tidak dibangun sebelum MVP rilis dan tervalidasi:

### 3.1 Integrasi dan Otomasi

- Payment gateway dan virtual account.
- Integrasi RW, kelurahan, atau instansi pemerintah.
- Validasi QR surat.
- Tanda tangan elektronik tersertifikasi.
- Pemindaian malware lampiran.
- Background job skala besar untuk ekspor dan notifikasi, bila kebutuhan volume membenarkan.

### 3.2 Produk dan Platform

- Aplikasi Android atau iOS native.
- Multi-RT sebagai platform mandiri.
- Portal publik RT.
- PWA offline penuh dengan sinkronisasi data.

### 3.3 Fitur Operasional Tambahan

- Buku tamu.
- Jadwal ronda dan absensi petugas.
- Inventaris aset.
- Pemesanan fasilitas bersama.
- Voting atau musyawarah digital.
- Marketplace atau direktori usaha warga.
- Statistik dashboard lanjutan.
- SLA aduan otomatis.
- Konfirmasi kehadiran kegiatan lanjutan.

---

## 4. Di Luar Ruang Lingkup

Fitur berikut tidak menjadi bagian produk RT Digital:

- Sistem kependudukan resmi pemerintah.
- Penerbitan atau penggantian dokumen kependudukan resmi.
- Akses langsung atau verifikasi identitas ke Dukcapil.
- Akuntansi perusahaan lengkap atau payroll.
- CCTV, pengenalan wajah, geolokasi presisi, atau pelacakan warga.
- Penggunaan data warga untuk iklan atau tujuan komersial.

---

## 5. Aturan Kontrol Perubahan

1. Fitur baru tidak langsung dikerjakan.
2. Pemohon mencatat masalah, pengguna terdampak, manfaat, risiko bila ditunda, dan estimasi dampak jadwal.
3. Product Owner menentukan salah satu hasil: **tolak**, **masuk backlog**, atau **perubahan darurat MVP**.
4. Perubahan darurat hanya dapat diterima bila tanpa fitur tersebut alur inti warga atau operasional pengurus tidak dapat digunakan.
5. Perubahan darurat wajib memperbarui `SCOPE.md`, `PRD.md`, jadwal, dan anggaran sebelum implementasi.
6. Penambahan fitur MVP harus disertai penghapusan atau penundaan fitur lain dengan beban setara, kecuali Product Owner menyetujui perubahan jadwal.
7. Tidak ada perubahan ruang lingkup setelah fase UAT dimulai, kecuali perbaikan keamanan, kehilangan data, kegagalan alur inti, atau kewajiban hukum.

---

## 6. Kriteria Feature Freeze

MVP memasuki *feature freeze* saat seluruh fitur bagian 2 telah memiliki:

- Kriteria penerimaan.
- Desain mobile-first.
- Hak akses yang ditentukan.
- Pengujian relevan.
- Status rilis atau keputusan eksplisit untuk dipindahkan ke pasca-MVP.

Setelah *feature freeze*, pekerjaan hanya dibatasi pada perbaikan bug, keamanan, performa, aksesibilitas, dokumentasi, migrasi data, dan UAT.