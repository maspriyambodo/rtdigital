# Sistem Desain RT Digital

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital  
**Prinsip:** Mobile-first, sederhana, konsisten, aksesibel sesuai WCAG 2.2 AA, dan mengutamakan kejelasan tugas warga maupun pengurus.

Dokumen ini menjadi acuan UI/UX frontend Next.js. Gunakan CSS custom properties/design tokens. Tailwind CSS dapat digunakan sebagai implementasi, bukan sebagai sumber keputusan desain.

---

## 1. Prinsip Dasar

1. Prioritaskan aksi warga: melihat tagihan, mengunggah bukti pembayaran, mengajukan surat, dan membuat aduan.
2. Desain dimulai dari viewport 320 px, lalu ditingkatkan untuk layar lebih besar.
3. Status, kesalahan, loading, serta langkah berikutnya selalu ditampilkan dengan teks; warna tidak boleh menjadi satu-satunya penanda.
4. Area sentuh minimum seluruh kontrol interaktif adalah **44 × 44 CSS pixel**.
5. Data sensitif dimasking secara default.
6. Jangan gunakan dekorasi, animasi, atau grafik yang tidak membantu penyelesaian tugas.
7. Semua komponen mendukung state: default, hover, focus-visible, active, disabled, loading, error bila relevan.

---

## 2. Design Tokens

```css
:root {
  /* Brand */
  --color-primary-50: #eff6ff;
  --color-primary-100: #dbeafe;
  --color-primary-600: #2563eb;
  --color-primary-700: #1d4ed8;
  --color-primary-800: #1e40af;

  /* Netral */
  --color-bg: #ffffff;
  --color-bg-subtle: #f8fafc;
  --color-surface: #ffffff;
  --color-surface-muted: #f1f5f9;
  --color-border: #e2e8f0;
  --color-border-strong: #cbd5e1;
  --color-text: #0f172a;
  --color-text-secondary: #475569;
  --color-text-muted: #64748b;
  --color-text-disabled: #94a3b8;

  /* Semantik */
  --color-success: #15803d;
  --color-success-bg: #dcfce7;
  --color-warning: #b45309;
  --color-warning-bg: #fef3c7;
  --color-danger: #b91c1c;
  --color-danger-bg: #fee2e2;
  --color-info: #0369a1;
  --color-info-bg: #e0f2fe;

  /* Fokus */
  --color-focus: #2563eb;

  /* Radius */
  --radius-sm: 0.375rem;
  --radius-md: 0.5rem;
  --radius-lg: 0.75rem;
  --radius-full: 9999px;

  /* Elevasi */
  --shadow-sm: 0 1px 2px rgb(15 23 42 / 0.08);
  --shadow-md: 0 4px 12px rgb(15 23 42 / 0.12);
}
```

### 2.1 Aturan Warna

| Token | Penggunaan |
|---|---|
| `primary-600` | Tombol utama, tautan, focus ring, aksi penting |
| `primary-700` | Hover tombol utama |
| `success` | Lunas, disetujui, selesai, berhasil |
| `warning` | Menunggu verifikasi, perlu perbaikan, tunggakan |
| `danger` | Ditolak, dibatalkan, gagal, tindakan destruktif |
| `info` | Informasi, draft, status baru bila tidak memerlukan peringatan |
| Netral | Teks, border, surface, struktur halaman |

- Kontras teks normal terhadap latar minimal **4.5:1**.
- Kontras teks besar minimal **3:1**.
- Jangan menggunakan kuning terang sebagai warna teks pada latar putih.
- Jangan menyampaikan status hanya dengan warna; sertakan teks dan ikon bila membantu.

---

## 3. Tipografi

Gunakan system font stack agar cepat dimuat pada koneksi seluler:

```css
font-family:
  Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont,
  "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
```

| Gaya | Mobile | Desktop | Berat | Line-height | Penggunaan |
|---|---:|---:|---:|---:|---|
| H1 | 24 px | 32 px | 700 | 1.2 | Judul halaman utama |
| H2 | 20 px | 24 px | 600 | 1.25 | Judul section |
| H3 | 18 px | 20 px | 600 | 1.3 | Judul card/dialog |
| Body | 16 px | 16 px | 400 | 1.5 | Isi utama dan input |
| Body kecil | 14 px | 14 px | 400 | 1.5 | Metadata dan bantuan |
| Caption | 12 px | 12 px | 500 | 1.4 | Label pendukung/status |

Aturan:

- Ukuran teks pada input minimal **16 px** untuk mencegah auto-zoom Safari iOS.
- Jangan gunakan teks kurang dari 12 px.
- Gunakan angka tabular untuk nominal uang, nomor tagihan, serta data finansial bila font mendukung `font-variant-numeric: tabular-nums`.
- Format tanggal dan nominal mengikuti Indonesia, misalnya `1 Agustus 2026` dan `Rp150.000`.

---

## 4. Spacing dan Grid

Gunakan kelipatan 4 px.

| Token | Nilai | Penggunaan |
|---|---:|---|
| `space-1` | 4 px | Ikon dengan teks |
| `space-2` | 8 px | Elemen berdekatan |
| `space-3` | 12 px | Gap kontrol/form ringkas |
| `space-4` | 16 px | Padding card dan halaman mobile |
| `space-5` | 20 px | Jarak antar kelompok |
| `space-6` | 24 px | Jarak antar section |
| `space-8` | 32 px | Jarak section besar |
| `space-10` | 40 px | Jarak layout desktop |

Aturan layout:

- Padding halaman mobile: 16 px.
- Padding halaman tablet/desktop: 24–32 px.
- Gap standar antar input formulir: 16 px.
- Card memakai padding minimum 16 px.
- Lebar konten desktop dibatasi `max-width: 1280px`; jangan membentangkan form atau teks terlalu lebar.

---

## 5. Responsive Layout

| Breakpoint | Lebar | Aturan utama |
|---|---:|---|
| Base | `< 360 px` | Dukungan 320 px; satu kolom; prioritas konten utama |
| `sm` | `≥ 360 px` | Mobile umum; target 360 px dan 390 px |
| `md` | `≥ 768 px` | Tablet; form dapat dua kolom bila aman |
| `lg` | `≥ 1024 px` | Desktop; sidebar pengurus, data table penuh |
| `xl` | `≥ 1280 px` | Desktop besar; batasi lebar konten |

### 5.1 Navigasi

- **Warga mobile:** bottom navigation maksimal lima tujuan utama.
- **Pengurus mobile:** drawer navigation; jangan memaksakan sidebar permanen.
- **Pengurus desktop:** sidebar tetap, konten utama responsif.
- Bottom navigation harus menghormati safe area perangkat:
  `padding-bottom: env(safe-area-inset-bottom)`.

### 5.2 Layout Data

- Mobile: kartu atau daftar ringkas dengan informasi paling penting di atas.
- Desktop: data table untuk kebutuhan administrasi dan pelaporan.
- Bila tabel wajib tersedia di mobile, tampilkan kolom prioritas dan scroll horizontal yang jelas; jangan mengecilkan teks sampai sulit dibaca.

---

## 6. Tombol dan Kontrol Interaktif

### 6.1 Varian Tombol

| Varian | Penggunaan | Visual |
|---|---|---|
| Primary | Simpan, Bayar, Ajukan, Konfirmasi | Latar `primary-600`, teks putih |
| Secondary | Batal, Kembali, aksi alternatif | Surface/border netral |
| Outline | Aksi sekunder yang tetap terlihat | Border dan teks primary |
| Ghost | Aksi ringan: tutup, kembali, lihat detail | Tanpa border/surface |
| Danger | Tolak, batalkan, nonaktifkan | Merah; gunakan setelah konfirmasi |

### 6.2 Aturan

- Tinggi minimum: 44 px.
- Tombol aksi utama mobile menggunakan lebar penuh bila berada di akhir form atau alur utama.
- Hanya satu tombol primary dominan per section atau dialog.
- Tombol icon-only wajib memiliki `aria-label` dan tooltip pada desktop.
- `focus-visible` memakai ring minimal 2 px dengan warna `--color-focus`.
- State `disabled` harus tidak interaktif dan tetap memiliki alasan bila tindakan tertahan.
- State `loading` menampilkan spinner serta mencegah klik ganda.
- Tombol bahaya tidak boleh menjadi aksi default.

---

## 7. Formulir

### 7.1 Struktur

- Label berada di atas input.
- Label wajib terhubung ke input melalui `for` dan `id`.
- Tandai field wajib dengan teks/indikator yang dapat dibaca screen reader.
- Bantuan field ditampilkan sebelum error; error tampil di bawah input.
- Form panjang dibagi menjadi langkah bertahap dengan indikator langkah dan tombol Kembali/Lanjut.
- Jangan gunakan placeholder sebagai pengganti label.

### 7.2 Input

| Elemen | Standar |
|---|---|
| Text/select | Tinggi minimum 44 px, font 16 px |
| Textarea | Minimum 3 baris, resize vertikal bila relevan |
| Date | Gunakan input native bila cocok agar mobile menampilkan picker perangkat |
| Telepon | `type="tel"` dan format validasi yang jelas |
| Email | `type="email"` serta `autocomplete="email"` |
| Nominal | `inputmode="numeric"`; format Rupiah saat aman tanpa mengubah nilai tersimpan |
| Password | `autocomplete` tepat dan toggle tampilkan/sembunyikan password |
| File | Kamera/galeri di mobile, drag-and-drop opsional di desktop |

### 7.3 Validasi

- Validasi ringan boleh dilakukan client-side; backend tetap menjadi sumber validasi akhir.
- Tampilkan error spesifik, contoh: `Nominal harus lebih besar dari Rp0.`
- Jangan hanya mengubah border menjadi merah.
- Fokuskan field error pertama saat submit gagal, tanpa mengganggu pembaca layar.
- Tampilkan ringkasan error untuk formulir panjang bila banyak field gagal.

### 7.4 Upload File

- Tampilkan jenis file yang didukung, ukuran maksimum, serta alasan penolakan.
- Tampilkan nama file, preview gambar bila aman, progress, gagal/coba lagi, serta hapus sebelum submit.
- Untuk mobile gunakan tombol **Ambil Foto** dan **Pilih dari Galeri** bila browser/perangkat mendukung.
- Jangan menampilkan file privat melalui URL permanen.
- Form surat dan aduan dapat menyimpan draft lokal secara aman; jangan simpan token, NIK, nomor KK, atau file dokumen ke `localStorage`.

---

## 8. Tabel dan Daftar

### 8.1 Mobile

Gunakan **Card List** untuk daftar warga, tagihan, pembayaran, surat, dan aduan:

- Status di posisi mudah terlihat.
- Informasi utama: nama/nomor, nominal atau tanggal, status.
- Informasi sekunder di bawahnya.
- Aksi lanjutan berada di menu overflow dengan label jelas.
- Satu kartu harus dapat diketuk untuk membuka detail.

### 8.2 Desktop

Gunakan tabel HTML semantik:

- Header jelas, `scope="col"`, dan sticky bila daftar panjang.
- Header `surface-muted`, teks sekunder, berat medium.
- Border baris `--color-border`.
- Hover row halus; jangan gunakan hover sebagai satu-satunya cara menemukan aksi.
- Sorting memiliki indikator dan label aksesibel.
- Pagination/cursor navigation jelas di bawah tabel.
- Tindakan per baris memakai menu overflow bila jumlah aksi lebih dari dua.

---

## 9. Modal, Dialog, dan Bottom Sheet

| Konteks | Mobile | Desktop |
|---|---|---|
| Filter/pilihan singkat | Bottom sheet | Popover atau dialog kecil |
| Aksi sekunder | Bottom sheet | Dialog/popover |
| Konfirmasi kritis | Alert dialog | Alert dialog |
| Form panjang | Halaman penuh | Halaman atau dialog besar bila sederhana |

Aturan:

- Bottom sheet memiliki handle, judul, tombol tutup, serta area sentuh aman.
- Dialog memakai focus trap, `Escape` untuk tutup bila aman, dan focus kembali ke pemicu saat ditutup.
- Alert dialog untuk pembatalan pembayaran, penolakan, nonaktifkan akun, atau tindakan kritis selalu memiliki tombol Batal yang jelas.
- Hindari modal bertumpuk.
- Jangan gunakan modal untuk workflow multi-langkah yang panjang.

---

## 10. Status Badge

Badge memuat teks status; ikon kecil dapat ditambahkan.

| Kategori | Latar | Teks | Contoh |
|---|---|---|---|
| Success | `success-bg` | `success` | Lunas, Disetujui, Diterbitkan, Selesai |
| Warning | `warning-bg` | `warning` | Menunggu Verifikasi, Perlu Perbaikan, Dibayar Sebagian |
| Info | `info-bg` | `info` | Baru, Diajukan, Diproses |
| Neutral | `surface-muted` | `text-secondary` | Draft, Nonaktif |
| Danger | `danger-bg` | `danger` | Ditolak, Dibatalkan, Gagal |

Standar badge:

- Tinggi minimum 24 px.
- Padding horizontal 8 px.
- Font 12 px/500.
- Radius `--radius-full`.
- Nama status konsisten dengan `USER_FLOW.md` dan `DATABASE_DESIGN.md`.

---

## 11. Card, Empty State, Loading, dan Error

### Card

- Surface putih, border netral, radius 8–12 px.
- Shadow kecil hanya bila dibutuhkan untuk pemisahan; border lebih diutamakan.
- Card interaktif memiliki focus ring dan affordance yang jelas.

### Empty State

- Jelaskan kondisi, contoh: `Belum ada tagihan bulan ini.`
- Beri satu tindakan relevan bila pengguna memiliki izin.
- Jangan gunakan ilustrasi besar yang memperlambat halaman.

### Loading

- Gunakan skeleton untuk layout yang sudah diketahui.
- Gunakan spinner untuk aksi lokal seperti menyimpan atau upload.
- Jangan mengganti seluruh layar dengan spinner untuk refresh data kecil.

### Error dan Offline

- Pesan memakai bahasa pengguna, bukan error teknis.
- Sediakan tindakan **Coba lagi** bila aman.
- Tampilkan `request_id` pada error server bila pengguna memerlukan bantuan pengurus.
- Saat jaringan lemah/offline, tampilkan indikator koneksi serta jangan menghapus isian belum terkirim.

---

## 12. Dark Mode

Dark mode **opsional pada MVP**, tetapi token warna wajib mendukungnya sejak awal.

```css
[data-theme="dark"] {
  --color-bg: #0f172a;
  --color-bg-subtle: #111827;
  --color-surface: #1e293b;
  --color-surface-muted: #334155;
  --color-border: #334155;
  --color-border-strong: #475569;
  --color-text: #f8fafc;
  --color-text-secondary: #cbd5e1;
  --color-text-muted: #94a3b8;
  --color-text-disabled: #64748b;
}
```

Aturan:

- Ikuti `prefers-color-scheme` sebagai default bila tidak ada preferensi pengguna.
- Bila toggle disediakan, simpan hanya preferensi tema, bukan data pribadi.
- Gunakan border/surface untuk elevasi; kurangi shadow berat.
- Uji semua badge, form error, focus ring, dan kontras teks pada tema gelap.
- Mode gelap tidak boleh menurunkan kontras atau mengubah makna status.

---

## 13. Checklist Implementasi

- [ ] Semua interaksi memenuhi target sentuh 44 × 44 px.
- [ ] Semua input teks berukuran minimal 16 px.
- [ ] Semua komponen dapat digunakan dengan keyboard dan focus-visible.
- [ ] Semua status memiliki teks, bukan warna saja.
- [ ] Kontras WCAG AA diverifikasi untuk kombinasi teks/latar.
- [ ] Daftar data berubah menjadi kartu pada viewport kecil.
- [ ] Bottom navigation warga aman terhadap safe area perangkat.
- [ ] Form panjang mendukung validasi jelas dan perlindungan draft.
- [ ] Dialog dan bottom sheet memenuhi focus management.
- [ ] Light mode dan dark mode diuji bila dark mode diaktifkan.