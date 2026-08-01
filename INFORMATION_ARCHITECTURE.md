# Arsitektur Informasi RT Digital

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital  
**Referensi:** `PRD.md`, `SCOPE.md`, `USER_ROLES_AND_PERMISSIONS.md`, `USER_FLOW.md`, `API_SPECIFICATION.md`

Aplikasi memiliki dua pengalaman utama:

- **Warga:** mobile-first, navigasi bawah maksimal lima tujuan utama.
- **Pengurus:** sidebar modular untuk desktop/tablet; tetap responsif saat dibuka dari telepon seluler.

Menu, halaman, tombol aksi, dan data yang terlihat wajib mengikuti RBAC serta scope data pengguna.

---

## 1. Prinsip Arsitektur Informasi

1. Tugas warga paling penting dapat dicapai dalam satu sampai dua ketukan dari Beranda.
2. Bottom navigation warga hanya memuat lima tujuan utama.
3. Navigasi pengurus dikelompokkan berdasarkan pekerjaan: kependudukan, komunikasi, keuangan, pelayanan, laporan, sistem.
4. Halaman daftar menampilkan pencarian, filter, status, serta tindakan utama sesuai izin.
5. Data sensitif dimasking secara default; pembukaan nilai lengkap memerlukan izin dan audit log.
6. Tabel pengurus beradaptasi menjadi daftar/kartu pada layar kecil; tidak memaksa tabel lebar.
7. Formulir panjang menggunakan langkah bertahap, autosave lokal aman, serta konfirmasi sebelum meninggalkan isian belum terkirim.
8. Status, kesalahan, loading, kondisi kosong, dan tindakan berikutnya selalu jelas.

---

## 2. Navigasi Warga

## 2.1 Bottom Navigation

| Urutan | Menu | Fungsi utama |
|---:|---|---|
| 1 | **Beranda** | Ringkasan layanan dan aksi cepat |
| 2 | **Tagihan** | Tagihan, pembayaran, bukti transfer, tanda terima |
| 3 | **Surat** | Pengajuan serta status surat |
| 4 | **Aduan** | Pembuatan dan pemantauan aduan |
| 5 | **Profil** | Data keluarga, notifikasi, pengaturan akun |

### Beranda

- Tagihan yang belum lunas atau menunggu verifikasi.
- Status surat yang membutuhkan tindakan.
- Status aduan aktif.
- Pengumuman penting dan terbaru.
- Agenda terdekat.
- Aksi cepat:
  - Bayar iuran.
  - Ajukan surat.
  - Buat aduan.

### Tagihan

- Daftar tagihan:
  - Belum Dibayar.
  - Menunggu Verifikasi.
  - Dibayar Sebagian.
  - Lunas.
  - Dibatalkan.
- Riwayat pembayaran keluarga sesuai izin.
- Detail tagihan:
  - Nominal, jatuh tempo, rekening tujuan, status, riwayat.
  - Aksi **Bayar dengan Transfer**.
  - Aksi unggah bukti dari kamera atau galeri.
  - Detail pembayaran dan alasan penolakan bila ada.
  - Unduh atau lihat tanda terima setelah pembayaran diverifikasi.

### Surat

- Daftar pengajuan surat keluarga sesuai izin.
- Filter status pengajuan.
- Aksi **Buat Pengajuan Surat**.
- Detail pengajuan:
  - Persyaratan.
  - Formulir.
  - Lampiran.
  - Timeline status.
  - Catatan perbaikan.
  - Unduh PDF surat terbit.

### Aduan

- Daftar aduan milik warga.
- Aksi **Buat Aduan Baru**.
- Detail aduan:
  - Nomor tiket.
  - Status dan penanggung jawab bila dapat ditampilkan.
  - Timeline pembaruan.
  - Komentar.
  - Lampiran.
  - Catatan penyelesaian.
  - Umpan balik sederhana atau tutup aduan.

### Profil

- Profil akun.
- Profil keluarga dan anggota keluarga.
- Pengajuan koreksi data.
- Notifikasi.
- Pengaturan akun:
  - Ubah email atau nomor telepon.
  - Ubah kata sandi.
  - Pengaturan MFA bila berlaku.
  - Keluar.

**Batas peran Warga:**

- Kepala Keluarga dapat melihat tagihan/pembayaran keluarga serta mengajukan koreksi data keluarga.
- Anggota Keluarga hanya melihat data yang diizinkan dan tidak mengelola pembayaran keluarga secara default.

---

## 3. Navigasi Pengurus

Sidebar hanya menampilkan modul yang diizinkan bagi peran aktif. Pengurus yang membuka aplikasi di layar kecil menggunakan drawer navigation; halaman daftar menjadi kartu atau tabel dengan kolom prioritas.

## 3.1 Dashboard

- Jumlah keluarga aktif dan warga aktif.
- Warga masuk, pindah, meninggal, atau nonaktif pada periode berjalan.
- Tagihan bulan berjalan dan total tunggakan.
- Pembayaran menunggu verifikasi.
- Saldo kas.
- Surat menunggu proses atau persetujuan.
- Aduan aktif berdasarkan status.
- Pengumuman terbaru.
- Pintasan cepat:
  - Buat tagihan.
  - Verifikasi pembayaran.
  - Proses surat.
  - Buat pengumuman.

## 3.2 Kependudukan

| Menu | Halaman dan fungsi |
|---|---|
| **Rumah / Unit** | Daftar, pencarian, detail, status hunian, tambah/ubah/nonaktifkan. |
| **Keluarga** | Daftar keluarga, pencarian, detail keluarga, anggota keluarga, verifikasi. |
| **Warga** | Daftar warga, pencarian, filter status/domisili, detail, tambah/ubah/verifikasi. |
| **Usulan Koreksi** | Daftar koreksi data warga, perbandingan nilai lama/usulan, setujui/tolak/minta perbaikan. |
| **Impor Data** | Validasi CSV, hasil error/duplikasi, import terkontrol. |

## 3.3 Komunikasi

| Menu | Halaman dan fungsi |
|---|---|
| **Pengumuman** | Draft, terjadwal, terbit, arsip; target pengumuman; lampiran; statistik baca. |
| **Agenda** | Daftar agenda, buat/ubah/batalkan kegiatan, lokasi, waktu, penanggung jawab. |

## 3.4 Keuangan

| Menu | Halaman dan fungsi |
|---|---|
| **Jenis Iuran** | Daftar jenis iuran, nominal, frekuensi, jatuh tempo, aktif/nonaktif. |
| **Tagihan** | Daftar tagihan, filter status/periode, buat tagihan individual atau massal, tunggakan, pembatalan beralasan. |
| **Pembayaran** | Pembayaran menunggu verifikasi, detail bukti, verifikasi/tolak, catat pembayaran tunai, riwayat. |
| **Buku Kas** | Saldo berjalan, pemasukan, pengeluaran, detail transaksi, transaksi pembalik. |
| **Kategori Kas** | Kelola kategori pemasukan dan pengeluaran. |

## 3.5 Pelayanan Warga

| Menu | Halaman dan fungsi |
|---|---|
| **Pengajuan Surat** | Daftar pengajuan, pemeriksaan lampiran, catatan perbaikan, proses, persetujuan, penerbitan PDF. |
| **Template Surat** | Jenis surat, persyaratan, formulir dinamis, template, pola nomor surat. |
| **Aduan** | Daftar tiket, filter status/kategori, penugasan petugas, komentar, pembaruan status, penyelesaian. |

## 3.6 Laporan

- Laporan keluarga dan warga.
- Rekap warga berdasarkan kelompok umur dan perubahan status.
- Rekap tagihan dan tunggakan.
- Rekap pembayaran.
- Buku kas serta pemasukan/pengeluaran.
- Rekap pengajuan surat.
- Rekap aduan.
- Ekspor CSV atau PDF sesuai izin.
- Seluruh ekspor dicatat dalam audit log.

## 3.7 Sistem

| Menu | Halaman dan fungsi |
|---|---|
| **Pengguna & Peran** | Daftar akun, undangan, status akun, penugasan peran. |
| **Pengaturan RT** | Profil RT, alamat, logo, rekening pembayaran, batas unggah, format nomor surat, zona waktu. |
| **Audit Log** | Riwayat aktivitas penting; read-only; pencarian dan filter. |

---

## 4. Sitemap dan Rute Next.js

```text
/                                      Landing atau redirect berdasarkan sesi
├── /login
├── /activate?token=...
├── /forgot-password
└── /reset-password?token=...

/app                                   Area Warga
├── /app/beranda
├── /app/tagihan
│   └── /app/tagihan/[id]
├── /app/surat
│   ├── /app/surat/baru
│   └── /app/surat/[id]
├── /app/aduan
│   ├── /app/aduan/baru
│   └── /app/aduan/[id]
└── /app/profil
    ├── /app/profil/keluarga
    ├── /app/profil/koreksi-data
    ├── /app/profil/notifikasi
    └── /app/profil/pengaturan

/admin                                 Area Pengurus
├── /admin/dashboard
├── /admin/kependudukan
│   ├── /admin/kependudukan/rumah-unit
│   ├── /admin/kependudukan/keluarga
│   ├── /admin/kependudukan/keluarga/[id]
│   ├── /admin/kependudukan/warga
│   ├── /admin/kependudukan/warga/[id]
│   ├── /admin/kependudukan/usulan-koreksi
│   └── /admin/kependudukan/impor
├── /admin/komunikasi
│   ├── /admin/komunikasi/pengumuman
│   └── /admin/komunikasi/agenda
├── /admin/keuangan
│   ├── /admin/keuangan/jenis-iuran
│   ├── /admin/keuangan/tagihan
│   ├── /admin/keuangan/pembayaran
│   ├── /admin/keuangan/buku-kas
│   └── /admin/keuangan/kategori-kas
├── /admin/pelayanan
│   ├── /admin/pelayanan/pengajuan-surat
│   ├── /admin/pelayanan/template-surat
│   └── /admin/pelayanan/aduan
├── /admin/laporan
└── /admin/pengaturan
    ├── /admin/pengaturan/rt
    ├── /admin/pengaturan/pengguna
    ├── /admin/pengaturan/peran
    └── /admin/pengaturan/audit-log
```

---

## 5. Pola Halaman

### Daftar Data Pengurus

- Judul, deskripsi singkat, jumlah hasil, dan tombol aksi utama.
- Pencarian berdasarkan nama, nomor rumah, nomor internal, atau nomor dokumen sesuai modul.
- Filter: status, tanggal/periode, kategori, dan penanggung jawab bila relevan.
- Sorting dan pagination.
- Tindakan per baris: lihat, ubah, verifikasi, setujui, nonaktifkan, batalkan, atau ekspor sesuai izin.
- Data sensitif tetap dimasking dalam daftar.

### Halaman Detail

- Breadcrumb pada desktop.
- Ringkasan status dan tindakan utama di bagian atas.
- Riwayat/timeline untuk surat, aduan, pembayaran, dan koreksi data.
- Data sensitif menggunakan tombol **Tampilkan Data Sensitif**; akses memerlukan alasan dan audit log.
- Tindakan destruktif atau tidak dapat dibatalkan memerlukan dialog konfirmasi.

### Status UI

| Kondisi | Perilaku |
|---|---|
| Loading | Skeleton loading; hindari layar kosong. |
| Kosong | Jelaskan kondisi dan tampilkan tindakan berikutnya. |
| Error | Pesan jelas, opsi coba lagi, tampilkan `request_id` bila perlu. |
| Offline/jaringan lemah | Indikator koneksi, retry aman, dan jangan hilangkan isian belum dikirim. |
| Tidak berizin | Halaman `403` tanpa membocorkan data atau keberadaan objek. |

---

## 6. Aturan Akses Navigasi

1. Frontend menyembunyikan menu tanpa permission efektif untuk mengurangi kebingungan.
2. Backend tetap memvalidasi permission dan scope setiap request.
3. Rute `/admin/*` hanya tersedia bagi peran pengurus yang memiliki izin modul terkait.
4. Warga hanya dapat membuka rute `/app/*` dalam scope diri atau keluarga sesuai perannya.
5. Pengguna multi-peran dapat berpindah konteks Warga/Pengurus bila memiliki kedua akses; konteks aktif harus terlihat jelas.