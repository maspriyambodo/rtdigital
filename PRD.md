# Product Requirements Document (PRD)

## Aplikasi Manajemen RT Digital

| Informasi | Detail |
|---|---|
| Nama sementara | RT Digital |
| Dokumen | Product Requirements Document |
| Versi | 1.2 |
| Status | Draft untuk validasi |
| Tanggal | 8 Agustus 2026 |
| Pemilik produk | Pengurus RT |
| Target platform | Web responsif mobile-first / Progressive Web App (PWA) |
| Bahasa utama | Bahasa Indonesia |

---

## 1. Ringkasan Eksekutif

RT Digital adalah aplikasi web untuk membantu pengurus RT mengelola administrasi lingkungan secara terpusat, transparan, aman, dan mudah digunakan oleh warga.

Aplikasi akan menggantikan proses yang saat ini tersebar pada buku catatan, spreadsheet, grup percakapan, dokumen kertas, dan pencatatan pembayaran manual. Sistem menyediakan satu sumber data untuk informasi warga, keluarga, iuran, pengumuman, pengajuan surat, aduan, kegiatan, dokumen, dan laporan RT.

Produk dirancang **mobile-first** karena mayoritas warga akan mengakses aplikasi melalui telepon seluler. Pengurus tetap memperoleh tampilan desktop yang nyaman untuk pekerjaan administrasi dan pelaporan.

MVP ditujukan untuk digunakan oleh satu organisasi RT. Struktur data tetap menyiapkan `organization_id` agar sistem dapat dikembangkan menjadi multi-RT pada masa depan tanpa mendesain ulang seluruh basis data.

---

## 2. Latar Belakang

Pengelolaan RT umumnya menghadapi beberapa kendala:

1. Data warga dan keluarga tidak tersimpan dalam satu sistem yang konsisten.
2. Riwayat perpindahan, perubahan anggota keluarga, dan status domisili sulit dilacak.
3. Pencatatan iuran dilakukan secara manual dan rawan salah hitung.
4. Warga sulit mengetahui tagihan, status pembayaran, dan penggunaan dana RT.
5. Pengumuman penting tenggelam di grup percakapan.
6. Pengajuan surat pengantar membutuhkan komunikasi berulang dan kunjungan langsung.
7. Aduan warga tidak memiliki nomor tiket, status, atau riwayat tindak lanjut.
8. Pembuatan laporan bulanan membutuhkan penggabungan data dari banyak sumber.
9. Data pribadi warga berisiko tersebar melalui file atau perangkat yang tidak terkontrol.

RT Digital menyelesaikan masalah tersebut melalui alur kerja digital dengan kontrol akses, audit log, pencadangan, dan laporan terstruktur.

---

## 3. Visi Produk

> Menjadi pusat administrasi dan komunikasi RT yang sederhana, transparan, aman, dan dapat digunakan oleh seluruh warga melalui perangkat apa pun.

---

## 4. Tujuan Produk

### 4.1 Tujuan Utama

- Memusatkan data warga dan keluarga dalam satu sistem.
- Mengurangi pekerjaan administrasi manual pengurus RT.
- Meningkatkan transparansi pencatatan iuran dan kas RT.
- Mempercepat pelayanan surat dan permohonan administrasi warga.
- Menyediakan kanal pengumuman dan aduan yang terstruktur.
- Menghasilkan laporan operasional secara otomatis.
- Menjaga keamanan dan riwayat perubahan data.
- Mengubah pencatatan pasif menjadi layanan proaktif: sistem memberi pengingat, membentuk antrean kerja, dan menunjukkan kepastian hasil kepada warga.

### 4.2 Indikator Keberhasilan Awal

Dalam tiga bulan setelah peluncuran MVP:

- Minimal 80% kepala keluarga telah memiliki akun aktif.
- Minimal 90% data keluarga telah diverifikasi pengurus.
- Minimal 70% transaksi iuran tercatat melalui aplikasi.
- Waktu pencarian data warga kurang dari satu menit.
- Waktu pemrosesan surat standar berkurang minimal 50%.
- Seluruh perubahan penting pada data dan transaksi memiliki audit log.
- Minimal 80% pengumuman resmi RT dipublikasikan melalui aplikasi.
- Minimal 90% pengajuan surat dan aduan yang memiliki SLA diselesaikan sesuai target layanan yang disepakati pengurus.
- Status domisili seluruh warga sementara dikonfirmasi sebelum tanggal evaluasinya.

---

## 5. Ruang Lingkup Produk

### 5.1 Ruang Lingkup MVP

MVP mencakup:

1. Autentikasi dan manajemen akun.
2. Manajemen peran dan hak akses.
3. Manajemen data keluarga dan warga.
4. Pendataan rumah atau unit tempat tinggal.
5. Pengumuman RT.
6. Agenda dan kegiatan RT.
7. Manajemen jenis iuran.
8. Pembuatan tagihan warga.
9. Pencatatan pembayaran tunai atau transfer.
10. Unggah bukti pembayaran.
11. Verifikasi pembayaran oleh bendahara.
12. Buku kas pemasukan dan pengeluaran.
13. Pengajuan surat administrasi.
14. Pemrosesan dan penerbitan surat berdasarkan template.
15. Aduan atau aspirasi warga.
16. Notifikasi dalam aplikasi.
17. Notifikasi WhatsApp melalui SaungWA (`https://saungwa.com/`) untuk aktivitas penting.
18. Dashboard dan laporan dasar.
19. Penyimpanan dokumen dan lampiran menggunakan Cloudflare R2.
20. Audit log aktivitas penting.
21. Pengaturan profil RT.
22. Dukungan PWA dasar: installable, app manifest, dan halaman utama yang tetap dapat dibuka dari perangkat seluler pada koneksi tidak stabil.

### 5.2 Prioritas Implementasi Pasca-MVP: Otomatisasi dan Layanan Proaktif

Prioritas ini mengubah modul yang telah ada dari CRUD reaktif menjadi bantuan operasional terukur. Implementasi dilakukan setelah fondasi MVP stabil dan parameter bisnisnya divalidasi pengurus.

1. **Pengingat iuran, pembayaran rapel, antrean verifikasi**
   - Tagihan rutin diterbitkan terjadwal dengan idempotensi per periode.
   - Pengingat jatuh tempo/tunggakan dikirim melalui kanal yang disetujui, sesuai preferensi dan batas frekuensi.
   - Satu pembayaran dapat dialokasikan atomik ke beberapa tagihan menurut aturan alokasi yang transparan.
   - Antrean bendahara memperlihatkan bukti, nominal, sisa tagihan, dan riwayat relevan dalam satu layar; pemisahan tugas tetap berlaku.

2. **Validasi pra-pengajuan surat dan SLA antrean**
   - Sistem memeriksa data wajib, formulir, dan lampiran sebelum warga dapat mengajukan surat.
   - Setiap jenis surat dapat memiliki target waktu layanan; pengurus melihat antrean jatuh tempo/terlambat, warga melihat status dan estimasi.

3. **Timeline aduan, kategori dengan SLA, konfirmasi penyelesaian**
   - Kategori aduan memuat target respons awal dan penyelesaian.
   - Warga melihat kronologi tindak lanjut; pengurus melihat tiket yang melampaui SLA.
   - Pelapor mengonfirmasi hasil penyelesaian atau tiket ditutup otomatis setelah tenggat yang disetujui, dengan alasan dan audit log.

4. **Health score keluarga dan pengingat domisili sementara**
   - Sistem menandai data keluarga yang tidak lengkap, belum terverifikasi, kontak bermasalah, atau lama tidak diperbarui.
   - Warga sementara/kontrak memiliki tanggal evaluasi dan pengingat konfirmasi tinggal, perpanjangan, atau pindah.
   - Health score adalah daftar kerja sekretaris, bukan dasar penolakan layanan atau penilaian sosial warga.

5. **Transparansi kas agregat**
   - Warga melihat saldo, rekap pemasukan/pengeluaran per kategori dan periode, serta bukti yang memang boleh dipublikasikan.
   - Nama penunggak, nominal pembayaran individu, dan detail transaksi pribadi tidak ditampilkan.

6. **QR verifikasi surat**
   - Surat terbit memuat QR menuju halaman verifikasi publik berisi nomor surat, jenis, tanggal terbit, dan status valid/dibatalkan.
   - Halaman publik tidak menampilkan data pribadi maupun URL dokumen privat.

7. **Serah-terima pengurus**
   - Checklist terpandu mencakup role, akses, rekening, tagihan terbuka, kas, surat, aduan, dan dokumen.
   - Akses pengurus lama diturunkan/dinonaktifkan secara terkontrol; riwayat audit tidak dihapus.

### 5.3 Fitur Lain Setelah MVP

- Integrasi payment gateway dan virtual account.
- Tanda tangan elektronik tersertifikasi.
- Portal publik RT.
- Buku tamu dan pencatatan keamanan.
- Jadwal ronda dan absensi petugas.
- Inventaris aset RT.
- Pemesanan fasilitas bersama.
- Voting atau musyawarah digital.
- Marketplace atau direktori usaha warga.
- Integrasi data dengan RW atau kelurahan apabila tersedia dan diizinkan.
- Aplikasi mobile native.
- Dukungan banyak RT dalam satu platform.

### 5.4 Di Luar Ruang Lingkup MVP

- Sistem kependudukan resmi pemerintah.
- Verifikasi identitas langsung ke Dukcapil.
- Penggantian dokumen kependudukan resmi.
- Akuntansi berstandar perusahaan.
- Pengelolaan penggajian pegawai.
- Pembayaran online otomatis.
- CCTV, pengenalan wajah, atau pelacakan lokasi warga.
- Tanda tangan digital tersertifikasi.

---

## 6. Pengguna dan Peran

### 6.1 Super Admin Sistem

Digunakan oleh pemilik atau pengelola teknis aplikasi.

Hak utama:

- Membuat dan mengaktifkan organisasi RT.
- Mengelola konfigurasi global.
- Melihat status layanan dan audit teknis.
- Membantu pemulihan akun secara terbatas.
- Tidak boleh melihat data sensitif warga tanpa alasan operasional dan audit.

### 6.2 Ketua RT

- Mengakses dashboard keseluruhan.
- Menyetujui perubahan data penting.
- Mengelola pengumuman dan agenda.
- Menyetujui atau menandatangani surat.
- Melihat laporan warga, iuran, kas, dan aduan.
- Menetapkan pengurus dan hak akses.

### 6.3 Sekretaris

- Mengelola data keluarga dan warga.
- Memverifikasi pembaruan data warga.
- Mengelola dokumen dan template surat.
- Memproses pengajuan surat.
- Membuat pengumuman dan agenda sesuai kewenangan.
- Menyusun laporan administrasi.

### 6.4 Bendahara

- Mengelola jenis iuran dan periode tagihan.
- Membuat tagihan.
- Memverifikasi pembayaran.
- Mencatat pemasukan dan pengeluaran.
- Mengunggah bukti transaksi.
- Menyusun laporan kas dan tunggakan.

### 6.5 Petugas RT

Peran operasional yang dapat dikonfigurasi, misalnya petugas keamanan atau koordinator wilayah.

- Mengakses modul tertentu sesuai izin.
- Memperbarui status tugas atau aduan.
- Tidak otomatis memperoleh akses ke seluruh data warga atau keuangan.

### 6.6 Kepala Keluarga

- Melihat dan mengajukan koreksi data keluarganya.
- Melihat tagihan dan riwayat pembayaran keluarga.
- Mengunggah bukti pembayaran.
- Mengajukan surat.
- Melihat pengumuman dan agenda.
- Mengirim aduan atau aspirasi.
- Mengelola anggota keluarga yang diizinkan menggunakan aplikasi.

### 6.7 Anggota Keluarga atau Warga

- Melihat pengumuman dan agenda.
- Melihat informasi keluarga sesuai izin.
- Mengajukan surat atau aduan apabila diaktifkan.
- Tidak dapat mengubah data keluarga utama tanpa persetujuan kepala keluarga atau pengurus.

---

## 7. Persona Utama

### Persona A — Ketua RT

- Membutuhkan ringkasan kondisi RT tanpa membuka banyak file.
- Ingin mengetahui jumlah warga, tunggakan, surat tertunda, dan aduan aktif.
- Membutuhkan jejak persetujuan yang jelas.

### Persona B — Sekretaris RT

- Mengelola data warga dan dokumen setiap hari.
- Membutuhkan pencarian cepat, validasi data, dan ekspor laporan.
- Menginginkan template surat agar tidak mengetik ulang.

### Persona C — Bendahara RT

- Membuat tagihan dan mencatat pembayaran.
- Membutuhkan rekonsiliasi sederhana serta laporan kas yang dapat dipertanggungjawabkan.
- Harus dapat melihat siapa yang belum membayar tanpa membocorkan data kepada warga lain.

### Persona D — Kepala Keluarga

- Ingin melihat tagihan, pengumuman, dan status surat dari telepon seluler.
- Tidak ingin datang ke rumah pengurus untuk proses yang dapat dilakukan daring.
- Membutuhkan antarmuka sederhana dengan istilah yang mudah dipahami.

---

## 8. Alur Pengguna Utama

### 8.1 Aktivasi Warga

1. Pengurus memasukkan data awal keluarga.
2. Sistem mengirimkan undangan atau kode aktivasi.
3. Kepala keluarga membuat kata sandi.
4. Kepala keluarga memeriksa data keluarga.
5. Kepala keluarga mengajukan koreksi jika ada kesalahan.
6. Sekretaris menyetujui atau menolak koreksi.
7. Data berstatus terverifikasi.

### 8.2 Pembayaran Iuran Manual

1. Bendahara membuat jenis iuran dan periode.
2. Sistem menghasilkan tagihan untuk keluarga yang ditentukan.
3. Warga melihat tagihan.
4. Warga membayar tunai atau transfer.
5. Untuk transfer, warga mengunggah foto atau screenshot bukti pembayaran dari kamera atau galeri telepon seluler.
6. Bendahara memverifikasi pembayaran.
7. Tagihan berubah menjadi lunas.
8. Sistem mencatat transaksi kas dan audit log.
9. Warga dapat melihat atau mengunduh tanda terima.

### 8.3 Pengajuan Surat

1. Warga memilih jenis surat.
2. Sistem menampilkan persyaratan dan formulir.
3. Warga mengisi data dan mengunggah lampiran.
4. Sekretaris memeriksa pengajuan.
5. Jika belum lengkap, pengajuan dikembalikan dengan catatan.
6. Jika lengkap, sistem menghasilkan draft surat.
7. Ketua RT menyetujui surat.
8. Surat diterbitkan dan dapat diunduh.
9. Riwayat penerbitan tersimpan.

### 8.4 Aduan Warga

1. Warga membuat aduan dan memilih kategori.
2. Sistem membuat nomor tiket.
3. Pengurus menerima notifikasi.
4. Pengurus menetapkan penanggung jawab.
5. Penanggung jawab memberikan pembaruan status.
6. Warga dapat melihat perkembangan.
7. Aduan ditutup dengan catatan penyelesaian.
8. Warga dapat memberi umpan balik sederhana.

---

## 9. Kebutuhan Fungsional

## 9.1 Autentikasi dan Akun

### Persyaratan

- Login menggunakan nomor telepon atau email dan kata sandi.
- Undangan akun oleh pengurus.
- Aktivasi akun menggunakan token sekali pakai.
- Lupa kata sandi.
- Ganti kata sandi.
- Keluar dari seluruh perangkat.
- Rotasi refresh token.
- Penguncian sementara setelah percobaan login gagal berulang.
- MFA wajib untuk Ketua RT, Sekretaris, Bendahara, dan Super Admin.
- Pencatatan waktu dan perangkat login terakhir.
- Penonaktifan akun tanpa menghapus riwayat data.

### Kriteria Penerimaan

- Pengguna tidak aktif tidak dapat login.
- Token aktivasi dan reset memiliki masa berlaku.
- Pengguna hanya dapat mengakses menu sesuai izin.
- Perubahan kredensial menghasilkan audit log.

## 9.2 Manajemen Peran dan Izin

### Persyaratan

- Sistem menggunakan Role-Based Access Control.
- Satu pengguna dapat memiliki lebih dari satu peran.
- Izin dapat dibatasi per modul dan tindakan: lihat, buat, ubah, hapus, verifikasi, setujui, ekspor.
- Pengurus tidak dapat menaikkan hak akses dirinya sendiri tanpa persetujuan peran yang berwenang.

### Contoh Izin

- `resident.read`
- `resident.create`
- `resident.update`
- `resident.verify`
- `finance.read`
- `finance.manage`
- `payment.verify`
- `letter.process`
- `letter.approve`
- `complaint.assign`
- `report.export`

## 9.3 Data Rumah, Keluarga, dan Warga

### Data Rumah atau Unit

- Nomor rumah atau unit.
- Blok, jalan, gang, atau alamat detail.
- Status hunian: milik sendiri, sewa, kontrak, kosong.
- Titik penanda wilayah opsional tanpa koordinat presisi pada MVP.
- Status aktif atau tidak aktif.

### Data Keluarga

- Nomor internal keluarga.
- Nomor KK, disimpan terenkripsi atau dimasking pada tampilan.
- Kepala keluarga.
- Alamat.
- Nomor telepon utama.
- Status domisili.
- Tanggal mulai tinggal.
- Tanggal pindah jika ada.
- Catatan internal pengurus.

### Data Warga

- Nama lengkap.
- NIK, disimpan terenkripsi atau dimasking.
- Tempat dan tanggal lahir.
- Jenis kelamin.
- Hubungan dalam keluarga.
- Status perkawinan.
- Pekerjaan.
- Pendidikan opsional.
- Nomor telepon dan email.
- Status domisili.
- Status warga aktif, pindah, meninggal, atau nonaktif.
- Informasi kebutuhan khusus bersifat opsional dan sangat dibatasi aksesnya.

### Fitur

- Pencarian berdasarkan nama, nomor rumah, nomor keluarga, atau nomor telepon.
- Filter status warga dan domisili.
- Riwayat perubahan anggota keluarga.
- Pengajuan koreksi data oleh warga.
- Persetujuan perubahan oleh pengurus.
- Impor data awal melalui CSV dengan validasi.
- Ekspor data sesuai hak akses.
- Masking data sensitif pada tabel dan laporan umum.
- Deteksi data duplikat berdasarkan kombinasi atribut.

## 9.4 Pengumuman

### Persyaratan

- Membuat, mengedit, menjadwalkan, dan mengarsipkan pengumuman.
- Kategori: umum, keamanan, kesehatan lingkungan, iuran, kegiatan, darurat.
- Target: seluruh warga, pengurus, blok tertentu, atau keluarga tertentu.
- Lampiran gambar atau dokumen.
- Tanggal mulai dan berakhir.
- Penandaan pengumuman penting.
- Status draft, terjadwal, terbit, atau arsip.
- Statistik telah dibaca.

## 9.5 Agenda dan Kegiatan

- Membuat agenda dengan tanggal, waktu, lokasi, penanggung jawab, dan deskripsi.
- Mencatat peserta atau konfirmasi kehadiran sederhana.
- Pengingat kegiatan.
- Lampiran hasil rapat atau dokumentasi.
- Status direncanakan, berlangsung, selesai, atau dibatalkan.

## 9.6 Iuran dan Tagihan

### Jenis Iuran

- Nama iuran.
- Deskripsi.
- Nominal tetap atau fleksibel.
- Frekuensi: sekali, bulanan, triwulanan, tahunan.
- Tanggal jatuh tempo.
- Sasaran keluarga atau kelompok tertentu.
- Aktif atau tidak aktif.

### Tagihan

- Pembuatan tagihan massal.
- Pembuatan tagihan individual.
- Nomor tagihan unik.
- Periode tagihan.
- Nilai tagihan.
- Penyesuaian atau diskon dengan alasan.
- Status: belum dibayar, menunggu verifikasi, dibayar sebagian, lunas, dibatalkan.
- Catatan perubahan tagihan.
- Daftar tunggakan.

### Pembayaran

- Metode: tunai, transfer bank, atau lainnya.
- Pembayaran penuh atau sebagian.
- Unggah bukti pembayaran.
- Verifikasi atau penolakan oleh bendahara.
- Alasan penolakan wajib diisi.
- Nomor tanda terima.
- Tanggal transaksi dan tanggal verifikasi.
- Pembatalan transaksi membutuhkan alasan dan audit log.
- Tidak boleh menghapus transaksi keuangan secara permanen melalui antarmuka aplikasi.

## 9.7 Buku Kas

- Kategori pemasukan dan pengeluaran.
- Pencatatan transaksi kas manual.
- Relasi otomatis dari pembayaran iuran ke pemasukan kas.
- Nomor transaksi.
- Tanggal transaksi.
- Nominal.
- Deskripsi.
- Penanggung jawab.
- Bukti transaksi.
- Saldo berjalan.
- Koreksi melalui transaksi pembalik, bukan penghapusan riwayat.
- Laporan kas per periode dan kategori.

## 9.8 Pengajuan Surat

### Jenis Surat Awal

- Surat pengantar domisili.
- Surat pengantar pembuatan atau perubahan dokumen.
- Surat pengantar usaha.
- Surat keterangan warga.
- Surat pengantar pindah.
- Jenis lain yang dapat ditambahkan oleh pengurus.

### Persyaratan

- Template surat dapat dikonfigurasi.
- Nomor surat dibuat berdasarkan pola yang ditentukan.
- Formulir dinamis berdasarkan jenis surat.
- Persyaratan lampiran per jenis surat.
- Status: draft, diajukan, diperiksa, perlu perbaikan, menunggu persetujuan, disetujui, diterbitkan, ditolak, dibatalkan.
- Catatan internal pengurus terpisah dari catatan yang terlihat warga.
- Pembuatan dokumen PDF.
- Kode verifikasi atau QR pada surat sebagai fitur setelah MVP atau bila waktu memungkinkan.
- Riwayat unduhan dan penerbitan.

## 9.9 Aduan dan Aspirasi

- Nomor tiket otomatis.
- Kategori aduan.
- Judul dan deskripsi.
- Lokasi umum tanpa memaksa koordinat presisi.
- Lampiran foto atau dokumen.
- Tingkat prioritas.
- Penanggung jawab.
- Status: baru, ditinjau, diproses, menunggu informasi, selesai, ditolak, ditutup.
- Komentar dan pembaruan status.
- Opsi identitas hanya terlihat pengurus untuk kasus tertentu; aduan anonim penuh tidak masuk MVP.
- SLA internal per kategori dapat ditambahkan setelah MVP.

## 9.10 Dokumen dan Lampiran

- Penyimpanan file di Cloudflare R2.
- Akses menggunakan URL bertanda tangan dengan masa berlaku terbatas.
- Validasi jenis dan ukuran file.
- Pemindaian malware sebagai peningkatan produksi yang direkomendasikan.
- Metadata file: pemilik, modul, jenis, ukuran, checksum, dan waktu unggah.
- File pribadi tidak boleh tersedia melalui URL publik permanen.
- Penghapusan mengikuti kebijakan retensi dan audit.

## 9.11 Notifikasi

### MVP

- Notifikasi dalam aplikasi.
- Notifikasi email untuk aktivitas penting melalui Resend (`https://resend.com`).
- Notifikasi WhatsApp untuk aktivitas penting melalui SaungWA (`https://saungwa.com/`).
- Status dibaca atau belum dibaca untuk notifikasi dalam aplikasi.
- Preferensi notifikasi dasar.

### Contoh Pemicu

- Undangan akun.
- Tagihan baru.
- Pembayaran diterima atau ditolak.
- Surat membutuhkan perbaikan.
- Surat telah diterbitkan.
- Aduan mendapatkan pembaruan.
- Pengumuman penting diterbitkan.

## 9.12 Dashboard

### Dashboard Pengurus

- Jumlah keluarga aktif.
- Jumlah warga aktif.
- Warga masuk dan pindah pada periode berjalan.
- Tagihan bulan berjalan.
- Total pembayaran diterima.
- Total tunggakan.
- Saldo kas.
- Surat menunggu proses.
- Aduan aktif berdasarkan status.
- Pengumuman terbaru.

### Dashboard Warga

- Tagihan belum lunas.
- Riwayat pembayaran terbaru.
- Status pengajuan surat.
- Status aduan.
- Pengumuman penting.
- Agenda terdekat.

## 9.13 Laporan

### Laporan MVP

- Rekap keluarga dan warga.
- Rekap warga berdasarkan kelompok umur.
- Rekap warga masuk, pindah, meninggal, dan nonaktif.
- Rekap tagihan per periode.
- Daftar tunggakan.
- Rekap pembayaran.
- Buku kas.
- Laporan pemasukan dan pengeluaran.
- Rekap pengajuan surat.
- Rekap aduan berdasarkan kategori dan status.

### Format Ekspor

- CSV untuk pengolahan data.
- PDF untuk laporan formal.
- Semua ekspor mencatat pengguna, waktu, parameter, dan jenis laporan pada audit log.

## 9.14 Pengaturan RT

- Nama RT dan RW.
- Kelurahan, kecamatan, kota atau kabupaten, dan provinsi.
- Alamat sekretariat.
- Logo.
- Nama dan masa jabatan pengurus.
- Format nomor surat.
- Rekening pembayaran.
- Batas ukuran unggahan.
- Zona waktu default: Asia/Jakarta.
- Template tanda terima dan surat.

## 9.15 Audit Log

Aktivitas berikut wajib tercatat:

- Login berhasil atau gagal yang signifikan.
- Perubahan peran dan izin.
- Akses atau perubahan data sensitif.
- Penambahan, perubahan, dan verifikasi data warga.
- Pembuatan dan perubahan tagihan.
- Verifikasi, penolakan, atau pembatalan pembayaran.
- Pencatatan dan koreksi kas.
- Persetujuan atau penerbitan surat.
- Ekspor laporan.
- Perubahan konfigurasi RT.

Audit log minimal menyimpan:

- Pengguna.
- Peran aktif.
- Tindakan.
- Jenis dan ID objek.
- Ringkasan nilai sebelum dan sesudah untuk data penting.
- Waktu.
- Alamat IP yang telah diperlakukan sesuai kebijakan privasi.
- User agent.
- Request ID.

Audit log tidak dapat diedit melalui antarmuka aplikasi.

---

## 10. Aturan Bisnis

1. Satu warga aktif hanya boleh menjadi anggota aktif pada satu keluarga dalam satu organisasi RT.
2. Satu keluarga harus memiliki tepat satu kepala keluarga aktif.
3. Nomor KK dan NIK tidak boleh ditampilkan utuh kepada pengguna tanpa izin khusus.
4. Perubahan NIK, nomor KK, status meninggal, atau status pindah membutuhkan verifikasi pengurus.
5. Transaksi keuangan tidak dihapus; kesalahan diperbaiki melalui pembatalan atau transaksi pembalik.
6. Pembayaran dianggap lunas hanya setelah diverifikasi bendahara atau mekanisme pembayaran otomatis pada fase berikutnya.
7. Pengguna tidak dapat memverifikasi transaksi yang dibuatnya sendiri apabila kebijakan pemisahan tugas diaktifkan.
8. Surat hanya dapat diterbitkan setelah data wajib dan lampiran terpenuhi.
9. Nomor surat yang telah diterbitkan tidak boleh digunakan ulang.
10. Pengumuman terjadwal hanya terbit pada waktu yang telah ditentukan.
11. File sensitif hanya dapat diakses oleh pengguna yang berwenang melalui tautan sementara.
12. Penonaktifan pengguna tidak menghapus catatan aktivitas atau transaksi historis.
13. Penghapusan data pribadi harus mengikuti kebijakan retensi, kebutuhan administrasi, dan regulasi yang berlaku.

---

## 11. Kebutuhan Nonfungsional

## 11.1 Kinerja

- Halaman utama warga memiliki target Largest Contentful Paint kurang dari 2,5 detik pada koneksi seluler 4G yang wajar.
- Halaman inti warga harus tetap fungsional pada jaringan tidak stabil: tampilkan status koneksi, retry aman, dan jangan hilangkan data formulir yang belum dikirim.
- Gambar, lampiran, dan respons API dioptimalkan untuk bandwidth seluler melalui kompresi, ukuran respons minimum, lazy loading, serta pagination atau cursor.
- Respons API untuk operasi baca sederhana memiliki target p95 kurang dari 500 ms, tidak termasuk unggah file dan pembuatan laporan besar.
- Pencarian warga untuk skala satu RT harus memberikan hasil kurang dari satu detik.
- Operasi massal dan ekspor besar dijalankan sebagai pekerjaan asinkron pada fase lanjutan jika diperlukan.

## 11.2 Ketersediaan

- Target ketersediaan MVP: 99,5% per bulan, tidak termasuk pemeliharaan terjadwal.
- Backend harus memiliki health check dan dapat melakukan restart otomatis.
- Database produksi menggunakan backup otomatis.
- Produksi direkomendasikan menggunakan Multi-AZ setelah penggunaan dan anggaran membenarkan.

## 11.3 Skalabilitas

Baseline per organisasi:

- Hingga 5.000 warga.
- Hingga 2.000 keluarga.
- Hingga 100 pengguna aktif bersamaan.
- Hingga 1 juta baris audit dan transaksi sebelum strategi arsip diperlukan.

Sistem harus dapat meningkatkan kapasitas backend secara horizontal tanpa menyimpan session state hanya pada memori satu container.

## 11.4 Keamanan

- TLS wajib untuk seluruh koneksi publik.
- Database tidak memiliki akses langsung dari internet.
- Backend ditempatkan pada private subnet bila arsitektur AWS memungkinkan.
- Kata sandi di-hash menggunakan algoritma modern seperti Argon2id.
- Refresh token disimpan dalam bentuk hash dan dapat dicabut.
- CORS hanya mengizinkan domain frontend resmi.
- Rate limiting untuk login, reset kata sandi, unggah, dan endpoint sensitif.
- Validasi dan sanitasi seluruh input.
- Query database harus menggunakan parameter binding.
- Proteksi CSRF diterapkan sesuai mekanisme autentikasi.
- Header keamanan browser diaktifkan.
- Secret tidak disimpan di repository atau image container.
- AWS Secrets Manager atau Parameter Store digunakan untuk rahasia produksi.
- Enkripsi at-rest untuk RDS, Cloudflare R2, dan backup.
- MFA wajib bagi peran pengurus.
- Dependency dan container image dipindai pada CI.
- Audit keamanan dilakukan sebelum produksi.

## 11.5 Privasi

- Mengumpulkan data minimum yang dibutuhkan.
- Menampilkan tujuan pengumpulan data kepada warga.
- Membatasi data sensitif berdasarkan peran.
- Menerapkan masking untuk NIK, nomor KK, nomor telepon, dan informasi sensitif lain.
- Mencatat ekspor dan akses data sensitif.
- Menyediakan proses koreksi data.
- Menetapkan masa retensi dokumen dan data akun.
- Tidak menggunakan data warga untuk iklan atau tujuan komersial tanpa dasar dan persetujuan yang sah.
- Kebijakan privasi dan persetujuan penggunaan aplikasi harus tersedia sebelum peluncuran.

## 11.6 Aksesibilitas dan UX

- Antarmuka mobile-first: desain dan pengujian dimulai dari layar seluler, kemudian ditingkatkan untuk tablet dan desktop.
- Ukuran area sentuh minimum 44 × 44 CSS pixel, dengan jarak antarelemen yang mencegah salah tekan.
- Navigasi warga menggunakan bottom navigation dengan maksimal lima tujuan utama; tindakan utama mudah dijangkau ibu jari.
- Tampilan data pada seluler menggunakan kartu atau daftar ringkas; tabel lebar hanya untuk desktop pengurus atau menyediakan kolom prioritas dan scroll horizontal yang jelas.
- Formulir seluler menggunakan keyboard dan tipe input yang sesuai, pilihan terstruktur, validasi per langkah, autosave lokal untuk isian panjang, serta unggah dari kamera atau galeri bila relevan.
- Kontras warna memadai.
- Formulir memiliki label dan pesan kesalahan yang jelas.
- Navigasi keyboard untuk fungsi utama.
- Istilah menggunakan bahasa administrasi yang mudah dipahami warga.
- Status tidak hanya dibedakan melalui warna.
- Target aksesibilitas: WCAG 2.2 level AA untuk halaman inti.

## 11.7 Kompatibilitas

- Chrome, Edge, Firefox, dan Safari versi modern.
- Android dan iOS melalui browser modern.
- Tampilan seluler adalah prioritas utama, diikuti tablet dan desktop.
- PWA dasar disertakan pada MVP: dapat diinstal melalui browser, memiliki app manifest, ikon, dan halaman offline atau fallback yang aman tanpa menampilkan data pribadi yang kedaluwarsa.

## 11.8 Observabilitas

- Structured logging dalam format JSON.
- Correlation atau request ID dari frontend hingga backend.
- Metrik CPU, memori, error rate, latency, dan jumlah request.
- Log aplikasi dikirim ke Amazon CloudWatch.
- Alert untuk error rate tinggi, backend tidak sehat, ruang penyimpanan, dan koneksi database.
- Error tracking frontend dapat menggunakan layanan eksternal yang disetujui.
- Data sensitif tidak boleh ditulis ke log.

## 11.9 Backup dan Pemulihan

- Backup otomatis RDS harian.
- Point-in-time recovery diaktifkan untuk produksi.
- Retensi backup awal minimal tujuh hari; target produksi 14–30 hari sesuai anggaran.
- Object versioning untuk dokumen penting bila didukung konfigurasi bucket.
- Uji pemulihan dilakukan secara berkala.
- Target awal RPO: maksimal 24 jam.
- Target awal RTO: maksimal 4 jam.

---

## 12. Arsitektur Teknis

## 12.1 Baseline Versi

Versi berikut adalah baseline stabil yang diverifikasi pada 1 Agustus 2026. Patch version harus diperbarui melalui dependency management dan pengujian otomatis selama development.

| Komponen | Baseline |
|---|---|
| Go | 1.26.5 |
| Next.js | 16.2 stable line |
| React | Versi stabil yang kompatibel dengan Next.js |
| TypeScript | Versi stabil terbaru yang kompatibel |
| PostgreSQL | 18.4 |
| Docker Engine | 29.7.1 |
| Docker Compose | Plugin Compose stabil terbaru yang kompatibel |
| Cloudflare Wrangler | Versi stabil terbaru yang kompatibel dengan OpenNext |

Next.js 16.3 masih berstatus preview pada saat baseline ini dibuat dan tidak digunakan untuk MVP sampai rilis stabil serta lolos pengujian kompatibilitas.

## 12.2 Arsitektur Tingkat Tinggi

```text
Pengguna
   |
   v
Cloudflare DNS, CDN, WAF, TLS
   |
   v
Next.js di Cloudflare Workers
   |
   | HTTPS REST API
   v
AWS Application Load Balancer
   |
   v
Go API di Amazon ECS Fargate
   |                |
   |                +--> Cloudflare R2 (dokumen dan lampiran)
   |                +--> Resend (email transaksional)
   |                +--> SaungWA (notifikasi WhatsApp)
   |                +--> CloudWatch (log dan metrik)
   |
   v
Amazon RDS for PostgreSQL 18.4
```

## 12.3 Frontend

- Next.js App Router.
- TypeScript strict mode.
- Server Components dan Client Components digunakan sesuai kebutuhan.
- Cloudflare Workers menggunakan adapter OpenNext.
- Cloudflare Pages hanya digunakan jika proyek diputuskan menjadi static export penuh; bukan pilihan default MVP.
- UI responsif dan mobile-first.
- Komunikasi ke backend melalui REST API HTTPS.
- Domain yang direkomendasikan:
  - `app.domain-rt.id` untuk frontend.
  - `api.domain-rt.id` untuk backend.
- Cookie dan CORS dikonfigurasi agar hanya berlaku untuk domain resmi.

## 12.4 Backend

- Go 1.26.5.
- Arsitektur modular monolith untuk MVP.
- REST API dengan prefix `/api/v1`.
- OpenAPI sebagai kontrak API.
- Pemisahan lapisan:
  - handler atau transport;
  - service atau use case;
  - repository;
  - domain model;
  - infrastructure.
- Dependency injection sederhana melalui constructor.
- Graceful shutdown.
- Structured logging.
- Database migration otomatis sebagai langkah terkontrol saat deployment, bukan pada setiap replica API secara bersamaan.
- Background job dapat dimulai dengan worker terpisah pada ECS apabila kebutuhan notifikasi dan laporan bertambah.

## 12.5 Database

- Amazon RDS for PostgreSQL 18.4.
- Database berada di private subnet.
- Public access dinonaktifkan.
- Enkripsi at-rest diaktifkan.
- Backup otomatis dan point-in-time recovery.
- Migration menggunakan tool migration yang dipilih tim.
- UUID digunakan sebagai primary key untuk entitas bisnis.
- Timestamp disimpan dalam UTC dan ditampilkan menggunakan zona waktu organisasi.
- Soft delete hanya digunakan untuk entitas yang memerlukannya; transaksi keuangan dan audit menggunakan status atau pembalikan.

## 12.6 AWS Deployment

### Komponen Minimum

- Amazon ECR untuk image backend.
- Amazon ECS Fargate untuk menjalankan container Go.
- Application Load Balancer untuk HTTPS dan health check.
- Amazon RDS for PostgreSQL.
- Cloudflare R2 untuk file dan lampiran.
- AWS Secrets Manager untuk secret.
- Amazon CloudWatch untuk log dan alarm.
- Resend untuk email transaksional.
- SaungWA untuk notifikasi WhatsApp sejak MVP.
- IAM dengan prinsip least privilege.

### Rekomendasi Region

- Default: `ap-southeast-3` atau Region Jakarta, setelah memeriksa ketersediaan seluruh layanan dan biaya.
- Alternatif: `ap-southeast-1` atau Region Singapura apabila layanan, biaya, atau reliabilitas lebih sesuai.
- Keputusan final harus mempertimbangkan lokasi mayoritas pengguna, residensi data, biaya egress, dan ketersediaan layanan.

## 12.7 Cloudflare Deployment

- Cloudflare DNS.
- Cloudflare Workers untuk Next.js.
- OpenNext adapter.
- WAF dan rate limiting pada endpoint yang relevan.
- Bot protection sesuai paket yang digunakan.
- Preview deployment untuk pull request jika pipeline mendukung.
- Environment production dan staging dipisahkan.
- Cache tidak boleh menyimpan respons pengguna yang mengandung data pribadi.

## 12.8 Development dengan Docker

Docker Compose digunakan untuk lingkungan lokal.

Service minimum:

```text
web       Next.js development server
api       Go API
postgres  PostgreSQL 18.4
```

Prinsip development:

- Satu perintah untuk menjalankan lingkungan lokal.
- Hot reload frontend dan backend.
- Volume persisten untuk database lokal.
- `.env.example` tanpa rahasia.
- Seed data untuk akun dan contoh keluarga.
- Health check untuk API dan database.
- Container berjalan sebagai non-root apabila memungkinkan.
- Multi-stage build untuk image produksi.
- Image produksi tidak membawa source code dan tool build yang tidak diperlukan.

## 12.9 Struktur Repository yang Direkomendasikan

```text
/
├── apps/
│   └── web/                 # Next.js
├── services/
│   └── api/                 # Go API
├── packages/
│   ├── api-client/          # Generated atau typed API client
│   └── shared-config/
├── infrastructure/
│   ├── docker/
│   ├── aws/
│   └── cloudflare/
├── docs/
│   ├── PRD.md
│   ├── API.md
│   ├── ARCHITECTURE.md
│   └── SECURITY.md
├── docker-compose.yml
├── .env.example
└── README.md
```

Monorepo dapat digunakan, tetapi deployment frontend dan backend harus tetap independen.

---

## 13. Model Data Awal

## 13.1 Entitas Inti

### Organization

- `id`
- `name`
- `rt_number`
- `rw_number`
- `address`
- `timezone`
- `logo_file_id`
- `status`
- `created_at`
- `updated_at`

### User

- `id`
- `organization_id`
- `resident_id`, nullable
- `email`, nullable
- `phone`, nullable
- `password_hash`
- `status`
- `last_login_at`
- `created_at`
- `updated_at`

### Role

- `id`
- `organization_id`, nullable untuk role sistem
- `name`
- `description`

### Permission

- `id`
- `code`
- `description`

### UserRole

- `user_id`
- `role_id`

### RolePermission

- `role_id`
- `permission_id`

### HouseUnit

- `id`
- `organization_id`
- `code`
- `address_detail`
- `occupancy_status`
- `status`

### Household

- `id`
- `organization_id`
- `house_unit_id`
- `internal_number`
- `family_card_number_encrypted`
- `head_resident_id`
- `domicile_status`
- `move_in_date`
- `move_out_date`
- `verification_status`

### Resident

- `id`
- `organization_id`
- `national_id_encrypted`
- `full_name`
- `birth_place`
- `birth_date`
- `gender`
- `marital_status`
- `occupation`
- `phone`
- `email`
- `resident_status`
- `verification_status`

### HouseholdMember

- `id`
- `household_id`
- `resident_id`
- `relationship`
- `is_active`
- `started_at`
- `ended_at`

### Announcement

- `id`
- `organization_id`
- `author_user_id`
- `title`
- `content`
- `category`
- `priority`
- `publish_at`
- `expire_at`
- `status`

### Event

- `id`
- `organization_id`
- `title`
- `description`
- `location`
- `starts_at`
- `ends_at`
- `status`

### DueType

- `id`
- `organization_id`
- `name`
- `amount`
- `frequency`
- `due_day`
- `status`

### Invoice

- `id`
- `organization_id`
- `household_id`
- `due_type_id`
- `invoice_number`
- `period_start`
- `period_end`
- `due_date`
- `amount`
- `paid_amount`
- `status`

### Payment

- `id`
- `organization_id`
- `invoice_id`
- `payment_number`
- `method`
- `amount`
- `paid_at`
- `verification_status`
- `verified_by`
- `verified_at`
- `proof_file_id`

### CashTransaction

- `id`
- `organization_id`
- `transaction_number`
- `type`
- `category_id`
- `amount`
- `transaction_date`
- `reference_type`
- `reference_id`
- `status`

### LetterType

- `id`
- `organization_id`
- `name`
- `requirements`
- `template`
- `number_pattern`
- `status`

### LetterRequest

- `id`
- `organization_id`
- `requester_user_id`
- `resident_id`
- `letter_type_id`
- `request_number`
- `form_data`
- `status`
- `submitted_at`
- `approved_by`
- `approved_at`
- `issued_file_id`

### Complaint

- `id`
- `organization_id`
- `reporter_user_id`
- `ticket_number`
- `category`
- `title`
- `description`
- `priority`
- `status`
- `assigned_to`
- `resolved_at`

### FileObject

- `id`
- `organization_id`
- `storage_key`
- `original_name`
- `mime_type`
- `size_bytes`
- `checksum`
- `visibility`
- `uploaded_by`

### Notification

- `id`
- `organization_id`
- `user_id`
- `type`
- `title`
- `body`
- `reference_type`
- `reference_id`
- `read_at`

### AuditLog

- `id`
- `organization_id`
- `actor_user_id`
- `action`
- `entity_type`
- `entity_id`
- `before_data`
- `after_data`
- `ip_address`
- `user_agent`
- `request_id`
- `created_at`

## 13.2 Index Awal

Index minimum direkomendasikan untuk:

- `organization_id` pada seluruh tabel tenant.
- Email dan nomor telepon pengguna yang dinormalisasi.
- Status pengguna.
- Nama warga untuk pencarian.
- Nomor internal keluarga.
- Nomor rumah atau unit.
- Status dan jatuh tempo tagihan.
- Waktu transaksi pembayaran.
- Status pengajuan surat.
- Status dan penanggung jawab aduan.
- `created_at` pada audit log.

Pencarian NIK dan nomor KK dilakukan menggunakan blind index atau nilai hash terpisah jika diperlukan, bukan mencari langsung pada ciphertext.

---

## 14. Desain API Awal

Prefix API:

```text
/api/v1
```

### Endpoint Utama

```text
POST   /auth/login
POST   /auth/refresh
POST   /auth/logout
POST   /auth/forgot-password
POST   /auth/reset-password
GET    /me

GET    /households
POST   /households
GET    /households/{id}
PATCH  /households/{id}

GET    /residents
POST   /residents
GET    /residents/{id}
PATCH  /residents/{id}
POST   /residents/{id}/verify

GET    /announcements
POST   /announcements
PATCH  /announcements/{id}
POST   /announcements/{id}/publish

GET    /events
POST   /events
PATCH  /events/{id}

GET    /due-types
POST   /due-types
POST   /invoices/generate
GET    /invoices
GET    /invoices/{id}

POST   /payments
POST   /payments/{id}/verify
POST   /payments/{id}/reject
POST   /payments/{id}/cancel

GET    /cash-transactions
POST   /cash-transactions
POST   /cash-transactions/{id}/reverse

GET    /letter-types
POST   /letter-requests
GET    /letter-requests
GET    /letter-requests/{id}
POST   /letter-requests/{id}/request-revision
POST   /letter-requests/{id}/approve
POST   /letter-requests/{id}/issue

POST   /complaints
GET    /complaints
GET    /complaints/{id}
PATCH  /complaints/{id}
POST   /complaints/{id}/comments

POST   /files/presign-upload
GET    /files/{id}/download

GET    /notifications
POST   /notifications/{id}/read

GET    /dashboard/admin
GET    /dashboard/resident
GET    /reports/{type}
GET    /audit-logs
```

### Konvensi API

- JSON menggunakan `snake_case` atau `camelCase`; pilih satu dan gunakan konsisten.
- Pagination berbasis cursor untuk data besar, pagination halaman dapat digunakan pada MVP untuk data sederhana.
- Semua error memiliki `code`, `message`, `details`, dan `request_id`.
- Idempotency key digunakan pada endpoint pembayaran dan operasi sensitif yang berpotensi dikirim ulang.
- OpenAPI menjadi sumber kontrak untuk frontend dan backend.

Contoh error:

```json
{
  "error": {
    "code": "PAYMENT_ALREADY_VERIFIED",
    "message": "Pembayaran sudah diverifikasi.",
    "details": {},
    "request_id": "req_01..."
  }
}
```

---

## 15. Halaman dan Navigasi

## 15.1 Warga

- Login.
- Aktivasi akun.
- Lupa kata sandi.
- Beranda.
- Profil keluarga.
- Data anggota keluarga.
- Tagihan.
- Detail pembayaran.
- Pengumuman.
- Agenda.
- Pengajuan surat.
- Detail status surat.
- Aduan saya.
- Buat aduan.
- Notifikasi.
- Pengaturan akun.

## 15.2 Pengurus

- Dashboard.
- Rumah atau unit.
- Keluarga.
- Warga.
- Permintaan koreksi data.
- Pengumuman.
- Agenda.
- Jenis iuran.
- Tagihan.
- Pembayaran.
- Buku kas.
- Pengajuan surat.
- Template surat.
- Aduan.
- Laporan.
- Pengguna dan peran.
- Audit log.
- Pengaturan RT.

---

## 16. Desain UX

### Prinsip

1. Mobile-first.
2. Tugas utama dapat selesai dalam sedikit langkah.
3. Status selalu terlihat dan mudah dipahami.
4. Formulir panjang dibagi menjadi beberapa bagian.
5. Konfirmasi diwajibkan untuk tindakan sensitif.
6. Warga hanya melihat data yang relevan dengan dirinya atau keluarganya.
7. Dashboard tidak menampilkan terlalu banyak grafik yang tidak dapat ditindaklanjuti.
8. Tabel pengurus menyediakan pencarian, filter, sorting, dan ekspor.
9. Setiap error memberikan solusi, bukan hanya kode teknis.
10. Gunakan format tanggal, mata uang, dan nomor yang sesuai Indonesia.

### Komponen Inti

- Sidebar desktop dan bottom navigation mobile untuk warga.
- Tampilan kartu atau daftar ringkas pada seluler; data table untuk desktop pengurus, dengan kolom adaptif atau scroll horizontal pada layar kecil.
- Bottom sheet untuk pilihan atau tindakan sekunder pada seluler; dialog digunakan hanya untuk konfirmasi kritis.
- Status badge dengan teks.
- Stepper vertikal atau per langkah untuk pengajuan surat pada layar kecil.
- Timeline untuk aduan dan surat.
- Upload dengan progress, validasi, kompresi gambar bila aman, serta akses kamera atau galeri telepon seluler.
- Empty state yang memberikan tindakan berikutnya.
- Skeleton loading.
- Indikator jaringan, retry, dan status pengiriman untuk tindakan yang bergantung koneksi.

---

## 17. Analitik Produk

Analitik harus menghormati privasi dan tidak mengirim data pribadi ke platform analitik tanpa evaluasi.

Event awal:

- `account_activated`
- `login_succeeded`
- `household_profile_viewed`
- `resident_correction_submitted`
- `invoice_viewed`
- `payment_proof_uploaded`
- `letter_request_submitted`
- `complaint_submitted`
- `announcement_opened`

Metrik produk:

- Pengguna aktif harian dan bulanan.
- Persentase aktivasi kepala keluarga.
- Rasio tagihan dilihat dan dibayar.
- Rata-rata waktu verifikasi pembayaran.
- Rata-rata waktu penyelesaian surat.
- Rata-rata waktu penyelesaian aduan.
- Tingkat pembacaan pengumuman penting.

---

## 18. Strategi Pengujian

## 18.1 Backend

- Unit test untuk domain dan service.
- Integration test repository dengan PostgreSQL container.
- API contract test.
- Authorization test untuk setiap role.
- Test transaksi dan idempotency pembayaran.
- Test migration naik dan turun bila migration tool mendukung.
- Race test untuk modul yang memiliki concurrency.
- Static analysis, lint, dan vulnerability scan.

## 18.2 Frontend

- Unit test untuk utility dan komponen penting.
- Component test untuk formulir utama.
- End-to-end test untuk alur:
  - login;
  - verifikasi data warga;
  - pembuatan tagihan;
  - pembayaran dan verifikasi;
  - pengajuan dan penerbitan surat;
  - pembuatan dan penyelesaian aduan.
- Accessibility test otomatis dan manual.
- Responsive test pada ukuran layar utama, dengan prioritas viewport seluler 320 px, 360 px, dan 390 px serta Safari iOS dan Chrome Android.
- Uji alur inti warga pada jaringan seluler lambat atau tidak stabil: login, melihat tagihan, unggah bukti pembayaran, pengajuan surat, dan aduan.
- Uji instalasi PWA serta perilaku offline atau fallback tanpa menyimpan data pribadi secara tidak aman.

## 18.3 Infrastruktur

- Health check.
- Deployment rollback.
- Backup dan restore test.
- Secret rotation test.
- Load test sederhana sebelum produksi.
- Security header dan TLS test.
- Penetration test ringan sebelum peluncuran.

---

## 19. CI/CD

### Pull Request

1. Install dependency.
2. Lint.
3. Type check.
4. Unit test.
5. Integration test.
6. Build frontend dan backend.
7. Scan dependency.
8. Scan container image.
9. Preview frontend apabila tersedia.

### Deployment Staging

1. Build image backend.
2. Push ke ECR.
3. Jalankan migration job.
4. Deploy ECS service.
5. Health check dan smoke test.
6. Deploy frontend Cloudflare Workers.
7. Jalankan end-to-end test utama.

### Deployment Production

- Memerlukan approval manual.
- Menggunakan image yang sama dengan staging.
- Backup atau snapshot dilakukan sebelum migration berisiko.
- Deployment backend menggunakan rolling atau blue/green sesuai kematangan.
- Rollback procedure terdokumentasi.
- Frontend dan backend memiliki versi rilis yang dapat dilacak.

---

## 20. Lingkungan

| Environment | Frontend | Backend | Database | Tujuan |
|---|---|---|---|---|
| Local | Docker/local dev | Docker | Docker PostgreSQL | Development |
| Test | CI runner | CI container | Ephemeral PostgreSQL | Automated test |
| Staging | Cloudflare Workers staging | AWS ECS staging | RDS staging | UAT |
| Production | Cloudflare Workers production | AWS ECS production | RDS production | Pengguna nyata |

Data produksi tidak boleh disalin ke development tanpa proses anonymization.

---

## 21. Migrasi Data Awal

1. Pengurus membersihkan spreadsheet sumber.
2. Tim menyediakan template CSV.
3. Sistem menjalankan validasi tanpa menyimpan data.
4. Sistem menampilkan error dan duplikasi.
5. Pengurus memperbaiki data.
6. Import dijalankan ke staging.
7. Pengurus melakukan sampling dan validasi.
8. Import final dilakukan ke produksi.
9. Hasil import dan pengguna pelaksana dicatat.

Data minimum untuk aktivasi:

- Nomor rumah atau alamat.
- Nama kepala keluarga.
- Nama anggota keluarga.
- Hubungan keluarga.
- Nomor telepon kepala keluarga bila tersedia.

NIK dan nomor KK dapat dilengkapi setelah alur keamanan dan persetujuan telah disiapkan.

---

## 22. Tahapan Pengembangan

## Fase 0 — Discovery dan Persiapan

- Validasi proses kerja pengurus.
- Inventaris jenis iuran.
- Inventaris jenis surat.
- Audit data spreadsheet.
- Penentuan kebijakan akses dan privasi.
- Wireframe.
- Desain sistem.
- Setup repository, Docker, CI, AWS, dan Cloudflare.

## Fase 1 — Fondasi

- Autentikasi.
- Role dan permission.
- Profil organisasi.
- Data rumah, keluarga, dan warga.
- Audit log.
- Import CSV.

## Fase 2 — Komunikasi

- Pengumuman.
- Agenda.
- Notifikasi dalam aplikasi.
- Email transaksional.

## Fase 3 — Keuangan

- Jenis iuran.
- Tagihan.
- Pembayaran manual.
- Verifikasi pembayaran.
- Buku kas.
- Laporan keuangan dasar.

## Fase 4 — Pelayanan Warga

- Jenis dan template surat.
- Pengajuan surat.
- Persetujuan dan PDF.
- Aduan dan tindak lanjut.

## Fase 5 — UAT dan Peluncuran

- Migrasi data awal.
- UAT bersama pengurus.
- Perbaikan prioritas tinggi.
- Pelatihan pengurus.
- Panduan warga.
- Soft launch.
- Monitoring dan evaluasi.

---

## 23. Prioritas MoSCoW

### Must Have

- Login dan role-based access.
- Data keluarga dan warga.
- Pengumuman.
- Tagihan dan pembayaran manual.
- Buku kas dasar.
- Pengajuan surat.
- Aduan.
- Dashboard pengurus dan warga.
- Audit log.
- Backup dan keamanan dasar.

### Should Have

- Impor CSV.
- Ekspor PDF dan CSV.
- Email transaksional.
- Agenda kegiatan.
- Pengajuan koreksi data oleh warga.
- MFA pengurus.
- Tanda terima pembayaran.

### Could Have

- Konfirmasi kehadiran kegiatan.
- Pemindaian malware file.
- Dashboard statistik tambahan.

### Won't Have pada MVP

- Payment gateway.
- Mobile native.
- Multi-RT.
- Integrasi pemerintah.
- Voting digital.
- CCTV atau facial recognition.

---

## 24. Risiko dan Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Data awal tidak rapi | Import gagal atau data duplikat | Template CSV, dry run, validasi, dan sampling |
| Warga enggan menggunakan aplikasi | Adopsi rendah | UI sederhana, onboarding, bantuan pengurus, soft launch |
| Data sensitif bocor | Dampak hukum dan kepercayaan | RBAC, masking, enkripsi, audit, private storage, security review |
| Pengurus salah mencatat transaksi | Laporan kas tidak akurat | Approval, reversal, audit log, rekonsiliasi |
| Biaya AWS lebih tinggi dari perkiraan | Operasional tidak berkelanjutan | Budget alert, right-sizing, staging kecil, evaluasi Fargate dan RDS |
| Ketergantungan pada satu pengurus | Operasional terhenti | Lebih dari satu admin, dokumentasi, pemulihan akun |
| Scope terlalu luas | MVP terlambat | MoSCoW, fase pengembangan, change request |
| Notifikasi email masuk spam | Informasi tidak diterima | Verifikasi domain, SPF/DKIM/DMARC, in-app notification |
| Konfigurasi cache salah | Data pribadi tercache | Cache-control ketat, test authenticated response, no-store default |
| Migration database gagal | Downtime atau data rusak | Backup, migration review, staging rehearsal, rollback plan |

---

## 25. Kriteria Rilis MVP

MVP dapat diluncurkan jika:

- Semua fitur Must Have selesai dan lolos UAT.
- Tidak ada bug severity kritis atau tinggi yang terbuka.
- Hak akses seluruh peran telah diuji.
- Backup dan restore telah diuji.
- Audit log bekerja pada seluruh transaksi penting.
- Data sensitif telah dimasking dan dienkripsi sesuai desain.
- Pengurus dapat mengimpor dan memverifikasi data awal.
- Alur tagihan sampai tanda terima berjalan penuh.
- Alur surat sampai PDF berjalan penuh.
- Alur aduan sampai selesai berjalan penuh.
- Monitoring dan alert dasar aktif.
- Dokumentasi operasional tersedia.
- Pengurus telah mengikuti pelatihan.
- Kebijakan privasi dan ketentuan penggunaan tersedia.

---

## 26. Dokumentasi Pendukung yang Harus Dibuat

- `README.md`
- `ARCHITECTURE.md`
- `DESIGN_SYSTEM.md`
- `API.md` atau OpenAPI specification
- `DATABASE.md`
- `SECURITY.md`
- `DEPLOYMENT.md`
- `RUNBOOK.md`
- `BACKUP_RESTORE.md`
- `DATA_MIGRATION.md`
- `TESTING.md`
- `CONTRIBUTING.md`
- `CHANGELOG.md`
- Panduan pengguna pengurus.
- Panduan singkat warga.
- Kebijakan privasi.
- Ketentuan penggunaan.

---

## 27. Pertanyaan yang Perlu Divalidasi Sebelum Development

1. Berapa jumlah keluarga dan warga saat ini?
2. Siapa saja pengurus yang akan menggunakan sistem?
3. Apakah satu rumah dapat dihuni lebih dari satu keluarga?
4. Apakah warga non-KTP setempat tetap dicatat?
5. Jenis iuran apa saja dan bagaimana aturan nominalnya?
6. Apakah pembayaran sebagian diperbolehkan?
7. Apakah tunggakan lama akan dimigrasikan?
8. Jenis surat apa saja yang paling sering dibuat?
9. Siapa yang berhak menyetujui dan menandatangani surat?
10. Apakah tanda tangan pada MVP berupa gambar tanda tangan, tanda tangan manual setelah dicetak, atau tanpa tanda tangan digital?
11. Apakah laporan kas dapat dilihat seluruh warga atau hanya ringkasannya?
12. Apakah identitas pelapor aduan dapat disembunyikan dari petugas tertentu?
13. Berapa batas ukuran dan jenis file yang diizinkan?
14. Apakah email warga cukup tersedia untuk notifikasi?
15. Apakah seluruh warga memiliki nomor WhatsApp aktif, dan berapa kuota atau anggaran bulanan SaungWA yang disetujui?
16. Berapa anggaran operasional bulanan untuk AWS, Cloudflare, domain, Resend, SaungWA, dan monitoring?
17. Apakah aplikasi menggunakan domain milik RT atau subdomain dari domain lain?
18. Berapa lama dokumen, audit log, dan data warga nonaktif disimpan?
19. Siapa penanggung jawab data dan permintaan koreksi atau penghapusan?
20. Apakah sistem nantinya akan digunakan oleh RT lain?

Jawaban atas pertanyaan ini dapat memperbarui ruang lingkup, aturan bisnis, biaya, dan desain teknis tanpa mengubah tujuan utama produk.

---

## 28. Keputusan Awal yang Direkomendasikan

1. Gunakan modular monolith, bukan microservices, untuk MVP.
2. Gunakan REST API dengan OpenAPI.
3. Gunakan Cloudflare Workers + OpenNext untuk frontend Next.js.
4. Gunakan Amazon ECS Fargate untuk backend Go.
5. Gunakan Amazon RDS PostgreSQL 18.4 untuk database produksi.
6. Gunakan Cloudflare R2 melalui API S3-compatible untuk seluruh dokumen privat.
7. Mulai dengan pembayaran manual dan verifikasi bendahara.
8. Gunakan satu organisasi RT pada MVP dengan kolom `organization_id` untuk kesiapan masa depan.
9. Prioritaskan data warga, iuran, surat, dan pengumuman sebelum fitur tambahan.
10. Terapkan keamanan dan audit sejak awal, bukan setelah aplikasi selesai.

---

## 29. Definisi Selesai

Sebuah fitur dianggap selesai apabila:

- Kebutuhan dan acceptance criteria dipenuhi.
- UI mobile-first telah diuji pada viewport 320 px, 360 px, dan 390 px, lalu desktop.
- Alur warga inti dapat diselesaikan dengan satu tangan tanpa tabel lebar, salah tekan, atau kehilangan data saat koneksi terganggu.
- Authorization telah diuji.
- Unit atau integration test yang relevan tersedia.
- Error handling dan empty state tersedia.
- Audit log tersedia jika tindakan bersifat penting.
- Dokumentasi API diperbarui.
- Tidak menulis data sensitif ke log.
- Lolos code review.
- Lolos CI.
- Lolos pengujian di staging.
- Disetujui pemilik produk atau perwakilan pengurus RT.

---

## 30. Catatan Perubahan

| Versi | Tanggal | Perubahan |
|---|---|---|
| 1.2 | 8 Agustus 2026 | Menambahkan prioritas otomatisasi dan layanan operasional proaktif pasca-MVP |
| 1.1 | 1 Agustus 2026 | Memperkuat kebutuhan mobile-first, PWA dasar, UX seluler, ketahanan koneksi, dan pengujian perangkat seluler |
| 1.0 | 1 Agustus 2026 | Draft awal PRD aplikasi manajemen RT digital |
